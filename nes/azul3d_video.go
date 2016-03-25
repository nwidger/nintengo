// +build !sdl,!js

package nes

import (
	"image"
	"image/color"
	"math"

	"azul3d.org/engine/gfx"
	"azul3d.org/engine/gfx/camera"
	"azul3d.org/engine/gfx/window"
	"azul3d.org/engine/keyboard"
	"azul3d.org/engine/lmath"
)

var Azul3DPalette []color.RGBA = []color.RGBA{
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
}

type Azul3DVideo struct {
	input         chan []uint8
	width, height int
	palette       []color.RGBA
	events        chan Event
	overscan      bool
	caption       string
	fps           float64
}

func NewVideo(caption string, events chan Event, fps float64) (video *Azul3DVideo, err error) {
	video = &Azul3DVideo{
		input:    make(chan []uint8, 128),
		events:   events,
		palette:  Azul3DPalette,
		overscan: true,
		caption:  caption,
		fps:      fps,
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

var glslVert = []byte(`
#version 120

attribute vec3 Vertex;
attribute vec2 TexCoord0;

uniform mat4 MVP;

uniform vec3 shift;
uniform vec3 scale;

varying vec2 tc;

void main()
{
	tc = scale.xy * TexCoord0;
	tc += shift.xy;

	gl_Position = MVP * vec4(Vertex, 1.0);
}
`)

var glslFrag = []byte(`
#version 120

varying vec2 tc;

uniform sampler2D Texture0; // palette indices
uniform sampler2D Texture1; // palette

void main() {
	vec4 t = texture2D(Texture0, tc);
	gl_FragColor = texture2D(Texture1, vec2((t.r*256.0)/64.0, 0));
}
`)

func (video *Azul3DVideo) handleInput(ev keyboard.ButtonEvent, w *window.Window) (running bool) {
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
			event = &FPSEvent{2.}
		case keyboard.F9:
			event = &FPSEvent{1.}
		case keyboard.F10:
			event = &FPSEvent{.75}
		case keyboard.F11:
			event = &FPSEvent{.5}
		case keyboard.F12:
			event = &FPSEvent{.25}
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
		case keyboard.NumFive:
			event = &MuteDMCEvent{}

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
			Button: button,
			Down:   ev.State == keyboard.Down,
		}
	}

	if event != nil {
		video.events <- event
	}

	return
}

func (video *Azul3DVideo) gfxLoop(w window.Window, d gfx.Device) {
	running := true

	// Create a simple shader.
	shader := gfx.NewShader("SimpleShader")
	shader.GLSL = &gfx.GLSLSources{
		Vertex:   glslVert,
		Fragment: glslFrag,
	}

	// Setup a camera using an orthographic projection.
	cam := camera.NewOrtho(d.Bounds())

	// Move the camera back two units away from the card.
	cam.SetPos(lmath.Vec3{0, -2, 0})

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

	palette := gfx.NewTexture()
	palette.MinFilter = gfx.Nearest
	palette.MagFilter = gfx.Nearest
	paletteSrc := image.NewRGBA(image.Rect(0, 0, 64, 2))
	for x, c := range Azul3DPalette {
		paletteSrc.SetRGBA(x, 0, c)
	}
	palette.Source = paletteSrc

	// Create a card object.
	card := gfx.NewObject()
	card.State = gfx.NewState()
	card.Shader = shader
	card.Textures = []*gfx.Texture{nil, palette}
	card.Meshes = []*gfx.Mesh{cardMesh}

	type textureUpdate struct {
		t            *gfx.Texture
		scale, shift gfx.Vec3
	}
	texUpdate := make(chan textureUpdate)

	img := image.NewRGBA(image.Rect(0, 0, 256, 256))

	updateTex := func(colors []uint8) {
		for i, c := range colors {
			img.Pix[i<<2] = c
		}

		var cropPx float32 = 0.0

		if video.overscan {
			cropPx = 8.0
		}

		var btmPx float32 = (16.0 - cropPx) / 256.0

		nx := 1.0 / float32(img.Bounds().Dx())
		ny := 1.0 / float32(img.Bounds().Dy())

		scale := gfx.Vec3{
			X: 1.0 - (nx * cropPx * 2),
			Y: 1.0 - (ny * cropPx * 2) - btmPx,
		}
		shift := gfx.Vec3{
			X: nx * cropPx,
			Y: ny * cropPx,
		}

		// Create new texture and ask the renderer to load it. We don't use DXT
		// compression because those textures cannot be downloaded.
		tex := gfx.NewTexture()
		tex.Source = img
		tex.MinFilter = gfx.Nearest
		tex.MagFilter = gfx.Nearest

		onLoad := make(chan *gfx.Texture, 1)
		d.LoadTexture(tex, onLoad)
		<-onLoad

		// Swap the texture with the old one on the card.
		texUpdate <- textureUpdate{
			t:     tex,
			scale: scale,
			shift: shift,
		}
	}

	// Create an event mask for the events we are interested in.
	evMask := window.KeyboardButtonEvents

	// Create a channel of events.
	events := make(chan window.Event, 256)

	// Have the window notify our channel whenever events occur.
	w.Notify(events, evMask)

	go func() {
		for running {
			select {
			case colors := <-video.input:
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
				updateTex(colors)
			}
		}
	}()

	for running {
		window.Poll(events, func(e window.Event) {
			switch ev := e.(type) {
			case keyboard.ButtonEvent:
				running = video.handleInput(ev, &w)
			}
		})

		// Center the card in the window.
		b := d.Bounds()
		cam.Update(b)

		select {
		case u := <-texUpdate:
			card.Textures[0] = u.t
			shader.Inputs["scale"] = u.scale
			shader.Inputs["shift"] = u.shift
		}

		card.SetPos(lmath.Vec3{float64(b.Dx()) / 2.0, 0, float64(b.Dy()) / 2.0})

		// Scale the card to fit the window, we divide by two because the
		// card is two units wide.
		var s float64
		if b.Dy() > b.Dx() {
			s = float64(b.Dx()) / 2.0
		} else {
			s = float64(b.Dy()) / 2.0
		}
		card.SetScale(lmath.Vec3{s, s, s})

		// clear the entire area (empty rectangle means "the whole area").
		d.Clear(d.Bounds(), gfx.Color{0, 0, 0, 1})
		d.ClearDepth(d.Bounds(), 1.0)

		// Draw the card to the screen.
		d.Draw(d.Bounds(), card, cam)

		// Render the whole frame.
		d.Render()
	}

	w.Close()
}

func (video *Azul3DVideo) Run() {
	props := window.NewProps()

	props.SetSize(512, 480)
	props.SetTitle("nintengo - " + video.caption + " - {FPS}")

	window.Run(video.gfxLoop, props)
}

func (video *Azul3DVideo) SetCaption(caption string) {
	video.caption = caption
}
