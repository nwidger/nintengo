// +build !sdl

package nes

import (
	"image"
	"image/color"
	"math"

	"azul3d.org/gfx.v1"
	"azul3d.org/gfx/window.v2"
	"azul3d.org/keyboard.v1"
	"azul3d.org/lmath.v1"
)

type Azul3DVideo struct {
	input         chan []uint8
	width, height int
	palette       []color.Color
	events        chan Event
	overscan      bool
	caption       string
}

func NewVideo(caption string, events chan Event) (video *Azul3DVideo, err error) {
	video = &Azul3DVideo{
		input:    make(chan []uint8, 128),
		events:   events,
		palette:  RGBAPalette,
		overscan: true,
		caption:  caption,
	}

	return
}

func convertColor(c color.Color) gfx.Color {
	r, g, b, a := c.RGBA()

	return gfx.Color{
		float32(r) / float32(math.MaxUint16),
		float32(g) / float32(math.MaxUint16),
		float32(b) / float32(math.MaxUint16),
		float32(a) / float32(math.MaxUint16),
	}
}

func (video *Azul3DVideo) Events() chan Event {
	return video.events
}

func (video *Azul3DVideo) Input() chan []uint8 {
	return video.input
}

func (video *Azul3DVideo) frameWidth() int {
	width := 256

	if video.overscan {
		width -= 16
	}

	return width
}

func (video *Azul3DVideo) frameHeight() int {
	height := 240

	if video.overscan {
		height -= 16
	}

	return height
}

var glslVert = []byte(`
#version 120

attribute vec3 Vertex;
attribute vec2 TexCoord0;

uniform mat4 MVP;

varying vec2 tc0;

void main()
{
	tc0 = TexCoord0;
	gl_Position = MVP * vec4(Vertex, 1.0);
}
`)

var glslFrag = []byte(`
#version 120

varying vec2 tc0;

uniform vec3 scale;
uniform vec3 shift;
uniform sampler2D Texture0;

void main()
{
	vec2 tc = scale.xy * tc0;
	tc += shift.xy;
	gl_FragColor = texture2D(Texture0, tc);
}
`)

func (video *Azul3DVideo) handleInput(ev keyboard.StateEvent, w *window.Window) (running bool) {
	var event Event

	setSize := func(width, height int) {
		props := (*w).Props()
		props.SetSize(width, height)
		(*w).Request(props)
	}

	running = true

	if ev.State == keyboard.Down {
		switch ev.Key {
		case keyboard.Tilde:
			video.overscan = !video.overscan
		case keyboard.One:
			setSize(256, 240)
		case keyboard.Two:
			setSize(512, 480)
		case keyboard.Three:
			setSize(768, 720)
		case keyboard.Four:
			setSize(1024, 960)
		case keyboard.Five:
			setSize(2560, 1440)
		case keyboard.P:
			event = &PauseEvent{}
		case keyboard.N:
			event = &FrameStepEvent{}
		case keyboard.Q:
			running = false
			event = &QuitEvent{}
		case keyboard.L:
			event = &SavePatternTablesEvent{}
		case keyboard.R:
			event = &ResetEvent{}
		case keyboard.S:
			event = &RecordEvent{}
		case keyboard.D:
			event = &StopEvent{}
		case keyboard.NumAdd:
			event = &AudioRecordEvent{}
		case keyboard.NumSubtract:
			event = &AudioStopEvent{}
		case keyboard.O:
			event = &CPUDecodeEvent{}
		case keyboard.I:
			event = &PPUDecodeEvent{}
		case keyboard.Nine:
			event = &ShowBackgroundEvent{}
		case keyboard.Zero:
			event = &ShowSpritesEvent{}
		case keyboard.F1:
			event = &SaveStateEvent{}
		case keyboard.F5:
			event = &LoadStateEvent{}
		case keyboard.F8:
			event = &FastForwardEvent{}
		case keyboard.F9:
			event = &FPS100Event{}
		case keyboard.F10:
			event = &FPS75Event{}
		case keyboard.F11:
			event = &FPS50Event{}
		case keyboard.F12:
			event = &FPS25Event{}
		case keyboard.NumZero:
			event = &MuteEvent{}
		case keyboard.NumOne:
			event = &MutePulse1Event{}
		case keyboard.NumTwo:
			event = &MutePulse2Event{}
		case keyboard.NumThree:
			event = &MuteTriangleEvent{}
		case keyboard.NumFour:
			event = &MuteNoiseEvent{}
		}
	}

	if event == nil {
		button := One

		switch ev.Key {
		case keyboard.Z:
			button = A
		case keyboard.X:
			button = B
		case keyboard.Enter:
			button = Start
		case keyboard.RightShift:
			button = Select
		case keyboard.ArrowUp:
			button = Up
		case keyboard.ArrowDown:
			button = Down
		case keyboard.ArrowLeft:
			button = Left
		case keyboard.ArrowRight:
			button = Right
		}

		event = &ControllerEvent{
			button: button,
			down:   ev.State == keyboard.Down,
		}
	}

	if event != nil {
		go func() { video.events <- event }()
	}

	return
}

func (video *Azul3DVideo) Run() {
	colors := []uint8{}
	running := true

	gfxLoop := func(w window.Window, r gfx.Renderer) {
		r.Clock().SetMaxFrameRate(DEFAULT_FPS)

		// Create a simple shader.
		shader := gfx.NewShader("SimpleShader")

		shader.GLSLVert = glslVert
		shader.GLSLFrag = glslFrag

		// Setup a camera using an orthographic projection.
		camera := gfx.NewCamera()
		camNear := 0.01
		camFar := 1000.0
		camera.SetOrtho(r.Bounds(), camNear, camFar)

		// Move the camera back two units away from the card.
		camera.SetPos(lmath.Vec3{0, -2, 0})

		// Create a card mesh.
		cardMesh := gfx.NewMesh()

		cardMesh.Vertices = []gfx.Vec3{
			// Left triangle.
			{-1, 0, 1},  // Left-Top
			{-1, 0, -1}, // Left-Bottom
			{1, 0, -1},  // Right-Bottom

			// Right triangle.
			{-1, 0, 1}, // Left-Top
			{1, 0, -1}, // Right-Bottom
			{1, 0, 1},  // Right-Top
		}

		cardMesh.TexCoords = []gfx.TexCoordSet{
			{
				Slice: []gfx.TexCoord{
					// Left triangle.
					{0, 0},
					{0, 1},
					{1, 1},

					// Right triangle.
					{0, 0},
					{1, 1},
					{1, 0},
				},
			},
		}

		// Create a card object.
		card := gfx.NewObject()

		card.Shader = shader
		card.Textures = []*gfx.Texture{nil}
		card.Meshes = []*gfx.Mesh{cardMesh}

		img := image.NewPaletted(image.Rect(0, 0, 256, 240), video.palette)

		updateTex := func() {
			x, y := 0, 0

			for _, c := range colors {
				img.Set(x, y, video.palette[c])

				switch x {
				case 255:
					x = 0
					y++
				default:
					x++
				}
			}

			scale := gfx.Vec3{1.0, 1.0, 0.0}
			shift := gfx.Vec3{0, 0, 0}
			if video.overscan {
				var cropPx float32 = 8.0

				nx := 1.0 / float32(img.Bounds().Dx())
				ny := 1.0 / float32(img.Bounds().Dy())

				scale = gfx.Vec3{
					X: 1.0 - (nx * cropPx * 2),
					Y: 1.0 - (ny * cropPx * 2),
				}
				shift = gfx.Vec3{
					X: nx * cropPx,
					Y: ny * cropPx,
				}
			}
			shader.Inputs["scale"] = scale
			shader.Inputs["shift"] = shift

			// Create new texture and ask the renderer to load it. We don't use DXT
			// compression because those textures cannot be downloaded.
			tex := gfx.NewTexture()

			tex.Source = img
			tex.MinFilter = gfx.Nearest
			tex.MagFilter = gfx.Nearest

			onLoad := make(chan *gfx.Texture, 1)
			r.LoadTexture(tex, onLoad)
			<-onLoad

			// Swap the texture with the old one on the card.
			card.Lock()
			card.Textures[0] = tex
			card.Unlock()
		}

		updateTex()

		go func() {
			// Create an event mask for the events we are interested in.
			evMask := window.KeyboardStateEvents

			// Create a channel of events.
			events := make(chan window.Event, 256)

			// Have the window notify our channel whenever events occur.
			w.Notify(events, evMask)

			for running {
				select {
				case colors = <-video.input:
					// We drop any pending frames and grab the most recent one. This is
					// because frame display is tied to the runProcessors loop and can
					// cause audio stuttering.
				frameDrop:
					for {
						select {
						case colors = <-video.input:
						default:
							break frameDrop
						}
					}

					// Update the texture using the most recent frame.
					updateTex()

				case e := <-events:
					switch ev := e.(type) {
					case keyboard.StateEvent:
						running = video.handleInput(ev, &w)
					}
				}
			}
		}()

		for running {
			// Center the card in the window.
			b := r.Bounds()
			camera.SetOrtho(b, camNear, camFar)
			card.SetPos(lmath.Vec3{float64(b.Dx()) / 2.0, 0, float64(b.Dy()) / 2.0})

			// Scale the card to fit the window.
			s := float64(b.Dy()) / 2.0 // Card is two units wide, so divide by two.
			card.SetScale(lmath.Vec3{s, s, s})

			// clear the entire area (empty rectangle means "the whole area").
			r.Clear(image.Rect(0, 0, 0, 0), gfx.Color{0, 0, 0, 1})
			r.ClearDepth(image.Rect(0, 0, 0, 0), 1.0)

			// Draw the card to the screen.
			r.Draw(image.Rect(0, 0, 0, 0), card, camera)

			// Render the whole frame.
			r.Render()
		}

		w.Close()
	}

	props := window.NewProps()

	props.SetSize(512, 480)
	props.SetTitle("nintengo - " + video.caption + " - {FPS}")

	window.Run(gfxLoop, props)
}
