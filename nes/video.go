package nes

import (
	"bufio"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"os"
	"unsafe"

	"github.com/go-gl/gl"
	"github.com/scottferg/Go-SDL/gfx"
	"github.com/scottferg/Go-SDL/sdl"
)

type Video interface {
	Input() chan []uint8
	Run()
}

type SDLVideo struct {
	input         chan []uint8
	screen        *sdl.Surface
	fps           *gfx.FPSmanager
	prog          gl.Program
	texture       gl.Texture
	width, height int
	textureUni    gl.AttribLocation
	palette       [64]uint32
	controllers   chan ControllerEvent
}

func NewSDLVideo(controllers chan ControllerEvent) (video *SDLVideo, err error) {
	video = &SDLVideo{
		input:       make(chan []uint8),
		controllers: controllers,
		palette: [64]uint32{
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
		},
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

	sdl.WM_SetCaption("nintengo", "")

	video.initGL()
	video.Reshape(int(video.screen.W), int(video.screen.H))

	video.fps = gfx.NewFramerate()
	video.fps.SetFramerate(60)

	return
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
		panic(fmt.Errorf("gl error: %v", err))
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

func (video *SDLVideo) Run() {
	running := true
	frame := make([]uint32, 0xf000)

	for running {
		select {
		case ev := <-sdl.Events:
			switch e := ev.(type) {
			case sdl.ResizeEvent:
				video.ResizeEvent(int(e.W), int(e.H))
			case sdl.QuitEvent:
				running = false
			case sdl.KeyboardEvent:
				switch e.Keysym.Sym {
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
				}

				switch e.Type {
				case sdl.KEYDOWN:
					video.controllers <- ControllerEvent{
						controller: 0,
						down:       true,
						button:     button(e),
					}
				case sdl.KEYUP:
					video.controllers <- ControllerEvent{
						controller: 0,
						down:       false,
						button:     button(e),
					}
				}
			}
		case colors := <-video.input:
			for i, c := range colors {
				frame[i] = video.palette[c] << 8
			}

			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

			video.prog.Use()

			gl.ActiveTexture(gl.TEXTURE0)
			video.texture.Bind(gl.TEXTURE_2D)

			gl.TexImage2D(gl.TEXTURE_2D, 0, 3, 256, 240, 0, gl.RGBA,
				gl.UNSIGNED_INT_8_8_8_8, frame)

			gl.DrawArrays(gl.TRIANGLES, 0, 6)

			if video.screen != nil {
				sdl.GL_SwapBuffers()
				video.fps.FramerateDelay()
			}

			video.input <- []uint8{}
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

type JPEGVideo struct {
	palette [64]color.RGBA
	input   chan []uint8
}

func NewJPEGVideo() (video *JPEGVideo, err error) {
	video = &JPEGVideo{
		input: make(chan []uint8),
		palette: [64]color.RGBA{
			color.RGBA{0x66, 0x66, 0x66, 0xff},
			color.RGBA{0x00, 0x2A, 0x88, 0xff},
			color.RGBA{0x14, 0x12, 0xA7, 0xff},
			color.RGBA{0x3B, 0x00, 0xA4, 0xff},
			color.RGBA{0x5C, 0x00, 0x7E, 0xff},
			color.RGBA{0x6E, 0x00, 0x40, 0xff},
			color.RGBA{0x6C, 0x06, 0x00, 0xff},
			color.RGBA{0x56, 0x1D, 0x00, 0xff},
			color.RGBA{0x33, 0x35, 0x00, 0xff},
			color.RGBA{0x0B, 0x48, 0x00, 0xff},
			color.RGBA{0x00, 0x52, 0x00, 0xff},
			color.RGBA{0x00, 0x4F, 0x08, 0xff},
			color.RGBA{0x00, 0x40, 0x4D, 0xff},
			color.RGBA{0x00, 0x00, 0x00, 0xff},
			color.RGBA{0x00, 0x00, 0x00, 0xff},
			color.RGBA{0x00, 0x00, 0x00, 0xff},
			color.RGBA{0xAD, 0xAD, 0xAD, 0xff},
			color.RGBA{0x15, 0x5F, 0xD9, 0xff},
			color.RGBA{0x42, 0x40, 0xFF, 0xff},
			color.RGBA{0x75, 0x27, 0xFE, 0xff},
			color.RGBA{0xA0, 0x1A, 0xCC, 0xff},
			color.RGBA{0xB7, 0x1E, 0x7B, 0xff},
			color.RGBA{0xB5, 0x31, 0x20, 0xff},
			color.RGBA{0x99, 0x4E, 0x00, 0xff},
			color.RGBA{0x6B, 0x6D, 0x00, 0xff},
			color.RGBA{0x38, 0x87, 0x00, 0xff},
			color.RGBA{0x0C, 0x93, 0x00, 0xff},
			color.RGBA{0x00, 0x8F, 0x32, 0xff},
			color.RGBA{0x00, 0x7C, 0x8D, 0xff},
			color.RGBA{0x00, 0x00, 0x00, 0xff},
			color.RGBA{0x00, 0x00, 0x00, 0xff},
			color.RGBA{0x00, 0x00, 0x00, 0xff},
			color.RGBA{0xFF, 0xFE, 0xFF, 0xff},
			color.RGBA{0x64, 0xB0, 0xFF, 0xff},
			color.RGBA{0x92, 0x90, 0xFF, 0xff},
			color.RGBA{0xC6, 0x76, 0xFF, 0xff},
			color.RGBA{0xF3, 0x6A, 0xFF, 0xff},
			color.RGBA{0xFE, 0x6E, 0xCC, 0xff},
			color.RGBA{0xFE, 0x81, 0x70, 0xff},
			color.RGBA{0xEA, 0x9E, 0x22, 0xff},
			color.RGBA{0xBC, 0xBE, 0x00, 0xff},
			color.RGBA{0x88, 0xD8, 0x00, 0xff},
			color.RGBA{0x5C, 0xE4, 0x30, 0xff},
			color.RGBA{0x45, 0xE0, 0x82, 0xff},
			color.RGBA{0x48, 0xCD, 0xDE, 0xff},
			color.RGBA{0x4F, 0x4F, 0x4F, 0xff},
			color.RGBA{0x00, 0x00, 0x00, 0xff},
			color.RGBA{0x00, 0x00, 0x00, 0xff},
			color.RGBA{0xFF, 0xFE, 0xFF, 0xff},
			color.RGBA{0xC0, 0xDF, 0xFF, 0xff},
			color.RGBA{0xD3, 0xD2, 0xFF, 0xff},
			color.RGBA{0xE8, 0xC8, 0xFF, 0xff},
			color.RGBA{0xFB, 0xC2, 0xFF, 0xff},
			color.RGBA{0xFE, 0xC4, 0xEA, 0xff},
			color.RGBA{0xFE, 0xCC, 0xC5, 0xff},
			color.RGBA{0xF7, 0xD8, 0xA5, 0xff},
			color.RGBA{0xE4, 0xE5, 0x94, 0xff},
			color.RGBA{0xCF, 0xEF, 0x96, 0xff},
			color.RGBA{0xBD, 0xF4, 0xAB, 0xff},
			color.RGBA{0xB3, 0xF3, 0xCC, 0xff},
			color.RGBA{0xB5, 0xEB, 0xF2, 0xff},
			color.RGBA{0xB8, 0xB8, 0xB8, 0xff},
			color.RGBA{0x00, 0x00, 0x00, 0xff},
			color.RGBA{0x00, 0x00, 0x00, 0xff},
		},
	}

	return
}

func (video *JPEGVideo) Input() chan []uint8 {
	return video.input
}

func (video *JPEGVideo) Run() {
	frame := image.NewRGBA(image.Rect(0, 0, 256, 240))

	fo, _ := os.Create(fmt.Sprintf("frame.jpg"))
	w := bufio.NewWriter(fo)

	for {
		select {
		case colors := <-video.input:
			x, y := 0, 0

			for _, c := range colors {
				frame.Set(x, y, video.palette[c])

				switch x {
				case 255:
					x = 0
					y++
				default:
					x++
				}
			}

			jpeg.Encode(w, frame, &jpeg.Options{Quality: 100})

			video.input <- []uint8{}
		}
	}
}
