// +build sdl

package nes

import (
	"errors"
	"fmt"
	"math"
	"unsafe"

	"github.com/go-gl/gl"
	"github.com/go-gl/glu"
	"github.com/scottferg/Go-SDL/sdl"
)

var SDLPalette []uint32 = []uint32{
	0x666666, 0x002A88, 0x1412A7, 0x3B00A4, 0x5C007E,
	0x6E0040, 0x6C0600, 0x561D00, 0x333500, 0x0B4800,
	0x005200, 0x004F08, 0x00404D, 0x000000, 0x000000,
	0x000000, 0xADADAD, 0x155FD9, 0x4240FF, 0x7527FE,
	0xA01ACC, 0xB71E7B, 0xB53120, 0x994E00, 0x6B6D00,
	0x388700, 0x0C9300, 0x008F32, 0x007C8D, 0x000000,
	0x000000, 0x000000, 0xFFFEFF, 0x64B0FF, 0x9290FF,
	0xC676FF, 0xF36AFF, 0xFE6ECC, 0xFE8170, 0xEA9E22,
	0xBCBE00, 0x88D800, 0x5CE430, 0x45E082, 0x48CDDE,
	0x4F4F4F, 0x000000, 0x000000, 0xFFFEFF, 0xC0DFFF,
	0xD3D2FF, 0xE8C8FF, 0xFBC2FF, 0xFEC4EA, 0xFECCC5,
	0xF7D8A5, 0xE4E594, 0xCFEF96, 0xBDF4AB, 0xB3F3CC,
	0xB5EBF2, 0xB8B8B8, 0x000000, 0x000000,
}

type SDLVideo struct {
	input         chan []uint8
	screen        *sdl.Surface
	prog          gl.Program
	texture       gl.Texture
	width, height int
	textureUni    gl.AttribLocation
	palette       []uint32
	events        chan Event
	overscan      bool
	caption       string
	fps           float64
}

func NewVideo(caption string, events chan Event, fps float64) (video *SDLVideo, err error) {
	video = &SDLVideo{
		input:    make(chan []uint8),
		events:   events,
		palette:  SDLPalette,
		overscan: true,
		caption:  caption,
		fps:      fps,
	}

	for i, _ := range video.palette {
		video.palette[i] <<= 8
	}

	if sdl.Init(sdl.INIT_VIDEO|sdl.INIT_JOYSTICK|sdl.INIT_AUDIO) != 0 {
		err = errors.New(sdl.GetError())
		return
	}

	video.screen = sdl.SetVideoMode(512, 480, 32,
		sdl.OPENGL|sdl.RESIZABLE|sdl.GL_DOUBLEBUFFER)

	if video.screen == nil {
		err = errors.New("Error setting video mode")
		return
	}

	sdl.WM_SetCaption("nintengo - "+video.caption, "")

	video.initGL()
	video.Reshape(int(video.screen.W), int(video.screen.H))

	return
}

func (video *SDLVideo) SetCaption(caption string) {
	sdl.WM_SetCaption("nintengo - "+video.caption, "")
}

func (video *SDLVideo) Events() chan Event {
	return video.events
}

const vertShaderSrcDef = `
attribute vec4 vPosition;
attribute vec2 vTexCoord;
varying vec2 texCoord;

void main() {
texCoord = vec2(vTexCoord.x, -vTexCoord.y);
gl_Position = vec4((vPosition.xy * 2.0) - 1.0, vPosition.zw);
}
`

const fragShaderSrcDef = `
varying vec2 texCoord;
uniform sampler2D texture;

void main() {
vec4 c = texture2D(texture, texCoord);
gl_FragColor = vec4(c.r, c.g, c.b, c.a);
}
`

func createProgram(vertShaderSrc string, fragShaderSrc string) gl.Program {
	vertShader := loadShader(gl.VERTEX_SHADER, vertShaderSrc)
	fragShader := loadShader(gl.FRAGMENT_SHADER, fragShaderSrc)

	prog := gl.CreateProgram()

	prog.AttachShader(vertShader)
	prog.AttachShader(fragShader)
	prog.Link()

	if prog.Get(gl.LINK_STATUS) != gl.TRUE {
		log := prog.GetInfoLog()
		panic(fmt.Errorf("Failed to link program: %v", log))
	}

	return prog
}

func loadShader(shaderType gl.GLenum, source string) gl.Shader {
	shader := gl.CreateShader(shaderType)
	if err := gl.GetError(); err != gl.NO_ERROR {
		str, _ := glu.ErrorString(err)
		panic(fmt.Errorf("gl error: %v", str))
	}

	shader.Source(source)
	shader.Compile()

	if shader.Get(gl.COMPILE_STATUS) != gl.TRUE {
		log := shader.GetInfoLog()
		panic(fmt.Errorf("Failed to compile shader: %v, shader: %v", log, source))
	}

	return shader
}

func (video *SDLVideo) initGL() {
	if gl.Init() != 0 {
		panic(sdl.GetError())
	}

	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	gl.Enable(gl.CULL_FACE)
	gl.Enable(gl.DEPTH_TEST)

	video.prog = createProgram(vertShaderSrcDef, fragShaderSrcDef)
	posAttrib := video.prog.GetAttribLocation("vPosition")
	texCoordAttr := video.prog.GetAttribLocation("vTexCoord")
	video.textureUni = video.prog.GetAttribLocation("texture")

	video.texture = gl.GenTexture()
	gl.ActiveTexture(gl.TEXTURE0)
	video.texture.Bind(gl.TEXTURE_2D)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	video.prog.Use()
	posAttrib.EnableArray()
	texCoordAttr.EnableArray()

	vertVBO := gl.GenBuffer()
	vertVBO.Bind(gl.ARRAY_BUFFER)
	verts := []float32{-1.0, 1.0, -1.0, -1.0, 1.0, -1.0, 1.0, -1.0, 1.0, 1.0, -1.0, 1.0}
	gl.BufferData(gl.ARRAY_BUFFER, len(verts)*int(unsafe.Sizeof(verts[0])), &verts[0], gl.STATIC_DRAW)

	textCoorBuf := gl.GenBuffer()
	textCoorBuf.Bind(gl.ARRAY_BUFFER)
	texVerts := []float32{0.0, 1.0, 0.0, 0.0, 1.0, 0.0, 1.0, 0.0, 1.0, 1.0, 0.0, 1.0}
	gl.BufferData(gl.ARRAY_BUFFER, len(texVerts)*int(unsafe.Sizeof(texVerts[0])), &texVerts[0], gl.STATIC_DRAW)

	posAttrib.AttribPointer(2, gl.FLOAT, false, 0, uintptr(0))
	texCoordAttr.AttribPointer(2, gl.FLOAT, false, 0, uintptr(0))
}

func (video *SDLVideo) ResizeEvent(width, height int) {
	video.screen = sdl.SetVideoMode(width, height, 32, sdl.OPENGL|sdl.RESIZABLE)
	video.Reshape(width, height)
}

func (video *SDLVideo) Reshape(width int, height int) {
	x_offset := 0
	y_offset := 0

	r := ((float64)(height)) / ((float64)(width))

	if r > 0.9375 { // Height taller than ratio
		h := (int)(math.Floor((float64)(0.9375 * (float64)(width))))
		y_offset = (height - h) / 2
		height = h
	} else if r < 0.9375 { // Width wider
		w := (int)(math.Floor((float64)((256.0 / 240.0) * (float64)(height))))
		x_offset = (width - w) / 2
		width = w
	}

	video.width = width
	video.height = height

	gl.Viewport(x_offset, y_offset, width, height)
}

func (video *SDLVideo) Input() chan []uint8 {
	return video.input
}

func (video *SDLVideo) frameWidth() int {
	width := 256

	if video.overscan {
		width -= 16
	}

	return width
}

func (video *SDLVideo) frameHeight() int {
	height := 240

	if video.overscan {
		height -= 16
	}

	return height
}

func (video *SDLVideo) Run() {
	running := true
	frame := make([]uint32, 0xf000)

	for running {
		select {
		case ev := <-sdl.Events:
			var event Event

			switch e := ev.(type) {
			case sdl.QuitEvent:
				running = false
				event = &QuitEvent{}
			case sdl.KeyboardEvent:
				switch e.Keysym.Sym {
				case sdl.K_BACKQUOTE:
					if e.Type == sdl.KEYDOWN {
						video.overscan = !video.overscan
					}
				case sdl.K_1:
					if e.Type == sdl.KEYDOWN {
						video.ResizeEvent(256, 240)
					}
				case sdl.K_2:
					if e.Type == sdl.KEYDOWN {
						video.ResizeEvent(512, 480)
					}
				case sdl.K_3:
					if e.Type == sdl.KEYDOWN {
						video.ResizeEvent(768, 720)
					}
				case sdl.K_4:
					if e.Type == sdl.KEYDOWN {
						video.ResizeEvent(1024, 960)
					}
				case sdl.K_5:
					if e.Type == sdl.KEYDOWN {
						video.ResizeEvent(2560, 1440)
					}
				case sdl.K_p:
					if e.Type == sdl.KEYDOWN {
						event = &PauseEvent{}
					}
				case sdl.K_n:
					if e.Type == sdl.KEYDOWN {
						event = &FrameStepEvent{}
					}
				case sdl.K_q:
					if e.Type == sdl.KEYDOWN {
						running = false
						event = &QuitEvent{}
					}
				case sdl.K_l:
					if e.Type == sdl.KEYDOWN {
						event = &SavePatternTablesEvent{}
					}
				case sdl.K_r:
					if e.Type == sdl.KEYDOWN {
						event = &ResetEvent{}
					}
				case sdl.K_s:
					if e.Type == sdl.KEYDOWN {
						event = &RecordEvent{}
					}
				case sdl.K_d:
					if e.Type == sdl.KEYDOWN {
						event = &StopEvent{}
					}
				case sdl.K_KP_PLUS:
					if e.Type == sdl.KEYDOWN {
						event = &AudioRecordEvent{}
					}
				case sdl.K_KP_MINUS:
					if e.Type == sdl.KEYDOWN {
						event = &AudioStopEvent{}
					}

				case sdl.K_o:
					if e.Type == sdl.KEYDOWN {
						event = &CPUDecodeEvent{}
					}
				case sdl.K_i:
					if e.Type == sdl.KEYDOWN {
						event = &PPUDecodeEvent{}
					}
				case sdl.K_9:
					if e.Type == sdl.KEYDOWN {
						event = &ShowBackgroundEvent{}
					}
				case sdl.K_0:
					if e.Type == sdl.KEYDOWN {
						event = &ShowSpritesEvent{}
					}
				case sdl.K_F1:
					if e.Type == sdl.KEYDOWN {
						event = &SaveStateEvent{}
					}
				case sdl.K_F5:
					if e.Type == sdl.KEYDOWN {
						event = &LoadStateEvent{}
					}
				case sdl.K_F8:
					if e.Type == sdl.KEYDOWN {
						event = &FPSEvent{2.}
					}
				case sdl.K_F9:
					if e.Type == sdl.KEYDOWN {
						event = &FPSEvent{1.}
					}
				case sdl.K_F10:
					if e.Type == sdl.KEYDOWN {
						event = &FPSEvent{.75}
					}
				case sdl.K_F11:
					if e.Type == sdl.KEYDOWN {
						event = &FPSEvent{.5}
					}
				case sdl.K_F12:
					if e.Type == sdl.KEYDOWN {
						event = &FPSEvent{.25}
					}
				case sdl.K_KP0:
					if e.Type == sdl.KEYDOWN {
						event = &MuteEvent{}
					}
				case sdl.K_KP1:
					if e.Type == sdl.KEYDOWN {
						event = &MutePulse1Event{}
					}
				case sdl.K_KP2:
					if e.Type == sdl.KEYDOWN {
						event = &MutePulse2Event{}
					}
				case sdl.K_KP3:
					if e.Type == sdl.KEYDOWN {
						event = &MuteTriangleEvent{}
					}
				case sdl.K_KP4:
					if e.Type == sdl.KEYDOWN {
						event = &MuteNoiseEvent{}
					}
				case sdl.K_KP5:
					if e.Type == sdl.KEYDOWN {
						event = &MuteDMCEvent{}
					}
				}

				if event == nil && running {
					event = &ControllerEvent{
						Button: button(e),
						Down:   e.Type == sdl.KEYDOWN,
					}
				}
			}

			if event != nil {
				go func() { video.events <- event }()
			}
		case colors := <-video.input:
			index := 0
			x, y := 0, 0

			for _, c := range colors {
				if pixelInFrame(x, y, video.overscan) {
					frame[index] = video.palette[c]
					index++
				}

				switch x {
				case 255:
					x = 0
					y++
				default:
					x++
				}
			}

			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

			video.prog.Use()

			gl.ActiveTexture(gl.TEXTURE0)
			video.texture.Bind(gl.TEXTURE_2D)

			gl.TexImage2D(gl.TEXTURE_2D, 0, 3, video.frameWidth(), video.frameHeight(), 0, gl.RGBA,
				gl.UNSIGNED_INT_8_8_8_8, frame)

			gl.DrawArrays(gl.TRIANGLES, 0, 6)

			if video.screen != nil {
				sdl.GL_SwapBuffers()
			}
		}
	}
}

func button(ev interface{}) Button {
	if k, ok := ev.(sdl.KeyboardEvent); ok {
		switch k.Keysym.Sym {
		case sdl.K_z: // A
			return A
		case sdl.K_x: // B
			return B
		case sdl.K_RSHIFT: // Select
			return Select
		case sdl.K_RETURN: // Start
			return Start
		case sdl.K_UP: // Up
			return Up
		case sdl.K_DOWN: // Down
			return Down
		case sdl.K_LEFT: // Left
			return Left
		case sdl.K_RIGHT: // Right
			return Right
		}
	}

	return One
}
