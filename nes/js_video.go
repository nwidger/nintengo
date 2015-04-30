// +build js

package nes

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/webgl"
)

var JSPalette []color.RGBA = []color.RGBA{
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

type JSVideo struct {
	input  chan []uint8
	events chan Event
}

func NewVideo(caption string, events chan Event, fps float64) (video *JSVideo, err error) {
	video = &JSVideo{
		input:  make(chan []uint8),
		events: events,
	}

	video.SetCaption(caption)

	return video, nil
}

func (video *JSVideo) Input() chan []uint8 {
	return video.input
}

func (video *JSVideo) Events() chan Event {
	return video.events
}

func (video *JSVideo) SetCaption(caption string) {
	if ts := js.Global.Get("document").Call("getElementsByTagName", "title"); ts.Length() > 0 {
		ts.Index(0).Set("innerHTML", "nintengo - "+caption)
	}
}

const vertShaderSrcDef = `
attribute vec2 a_position;
attribute vec2 a_texCoord;

uniform vec2 u_resolution;

varying vec2 v_texCoord;

void main() {
   // convert the rectangle from pixels to 0.0 to 1.0
   vec2 zeroToOne = a_position / u_resolution;

   // convert from 0->1 to 0->2
   vec2 zeroToTwo = zeroToOne * 2.0;

   // convert from 0->2 to -1->+1 (clipspace)
   vec2 clipSpace = zeroToTwo - 1.0;

   gl_Position = vec4(clipSpace * vec2(1, -1), 0, 1);

   // pass the texCoord to the fragment shader
   // The GPU will interpolate this value between points.
   v_texCoord = a_texCoord;
}
`

const fragShaderSrcDef = `
precision mediump float;

// our texture
uniform sampler2D u_image;

// the texCoords passed in from the vertex shader.
varying vec2 v_texCoord;

void main() {
   gl_FragColor = texture2D(u_image, v_texCoord);
}
`

func button(keyCode int) Button {
	switch keyCode {
	case 37:
		return Left
	case 38:
		return Up
	case 39:
		return Right
	case 40:
		return Down
	case 90:
		return A
	case 88:
		return B
	case 13:
		return Start
	case 16:
		return Select
	default:
		return One
	}
}

func setRectangle(gl *webgl.Context, x, y int, width, height float32) {
	x1 := float32(x)
	x2 := float32(x) + width
	y1 := float32(y)
	y2 := float32(y) + height

	verts := []float32{x1, y1, x2, y1, x1, y2, x1, y2, x2, y1, x2, y2}
	gl.BufferData(gl.ARRAY_BUFFER, verts, gl.STATIC_DRAW)
}

// file:///Users/niels/go/src/github.com/nwidger/nintengo/index.html
func (video *JSVideo) Run() {
	document := js.Global.Get("document")

	handleKey := func(e *js.Object, down bool) {
		var event Event

		if button := button(e.Get("keyCode").Int()); button != One {
			event = &ControllerEvent{
				Button: button,
				Down:   down,
			}
		}

		if event != nil {
			go func() { video.events <- event }()
		}
	}

	document.Set("onkeydown", func(e *js.Object) {
		handleKey(e, true)
	})

	document.Set("onkeyup", func(e *js.Object) {
		handleKey(e, false)
	})

	canvas := document.Call("createElement", "canvas")
	canvas.Call("setAttribute", "width", "256")
	canvas.Call("setAttribute", "height", "240")
	document.Get("body").Call("appendChild", canvas)

	img := document.Call("createElement", "img")
	img.Call("setAttribute", "width", "256")
	img.Call("setAttribute", "height", "240")

	// canvas.Call("appendChild", img)

	loaded := make(chan int, 1)

	handleTextureLoaded := func() {
		loaded <- 1
	}

	img.Set("onload", handleTextureLoaded)
	// http://garethrees.org/2007/11/14/pngcrush/
	img.Call("setAttribute", "src", "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAACklEQVR4nGMAAQAABQABDQottAAAAABJRU5ErkJggg==")

	<-loaded

	attrs := webgl.DefaultAttributes()
	attrs.Alpha = false
	attrs.Antialias = false

	gl, err := webgl.NewContext(canvas, attrs)
	if err != nil {
		js.Global.Call("alert", "Error: "+err.Error())
	}

	vertShader := gl.CreateShader(gl.VERTEX_SHADER)
	gl.ShaderSource(vertShader, vertShaderSrcDef)
	gl.CompileShader(vertShader)

	if !gl.GetShaderParameterb(vertShader, gl.COMPILE_STATUS) {
		fmt.Println("Vertex shader compilation failed:", gl.GetShaderInfoLog(vertShader))
		return
	}

	fragShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	gl.ShaderSource(fragShader, fragShaderSrcDef)
	gl.CompileShader(fragShader)

	if !gl.GetShaderParameterb(fragShader, gl.COMPILE_STATUS) {
		fmt.Println("Fragment shader compilation failed:", gl.GetShaderInfoLog(fragShader))
		return
	}

	prog := gl.CreateProgram()

	gl.AttachShader(prog, vertShader)
	gl.AttachShader(prog, fragShader)
	gl.LinkProgram(prog)

	if !gl.GetProgramParameterb(prog, gl.LINK_STATUS) {
		fmt.Println("Linking failed:", gl.GetProgramInfoLog(prog))
		return
	}

	gl.UseProgram(prog)

	posAttrib := gl.GetAttribLocation(prog, "a_position")
	texCoordAttr := gl.GetAttribLocation(prog, "a_texCoord")

	textCoorBuf := gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, textCoorBuf)
	texVerts := []float32{0.0, 0.0, 1.0, 0.0, 0.0, 1.0, 0.0, 1.0, 1.0, 0.0, 1.0, 1.0}
	gl.BufferData(gl.ARRAY_BUFFER, texVerts, gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(texCoordAttr)
	gl.VertexAttribPointer(texCoordAttr, 2, gl.FLOAT, false, 0, 0)

	texture := gl.CreateTexture()
	gl.BindTexture(gl.TEXTURE_2D, texture)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	resolutionLocation := gl.GetUniformLocation(prog, "u_resolution")

	gl.Uniform2f(resolutionLocation, float32(canvas.Get("width").Float()), float32(canvas.Get("height").Float()))

	vertVBO := gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, vertVBO)
	gl.EnableVertexAttribArray(posAttrib)
	gl.VertexAttribPointer(posAttrib, 2, gl.FLOAT, false, 0, 0)

	setRectangle(gl, 0, 0, float32(img.Get("width").Float()), float32(img.Get("height").Float()))

	gl.DrawArrays(gl.TRIANGLES, 0, 6)

	frame := image.NewRGBA(image.Rect(0, 0, 256, 240))
	buf := new(bytes.Buffer)

	for {
		colors := <-video.input

		for i, c := range colors {
			p := JSPalette[c]
			j := i << 2
			frame.Pix[j+0] = p.R
			frame.Pix[j+1] = p.G
			frame.Pix[j+2] = p.B
			frame.Pix[j+3] = p.A
		}

		buf.Reset()
		png.Encode(buf, frame)

		img = document.Call("createElement", "img")
		img.Call("setAttribute", "width", "256")
		img.Call("setAttribute", "height", "240")
		img.Set("onload", handleTextureLoaded)

		img.Call("setAttribute", "src", "data:image/png;base64,"+base64.StdEncoding.EncodeToString(buf.Bytes()))
		<-loaded

		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.RGBA, gl.UNSIGNED_BYTE, img)
		gl.DrawArrays(gl.TRIANGLES, 0, 6)
	}
}
