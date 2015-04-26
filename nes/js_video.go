// +build js

package nes

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/webgl"
)

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

func render(img *js.Object, canvas *js.Object, gl *webgl.Context) {
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

	canvas.Call("appendChild", img)

	loaded := make(chan int, 1)

	handleTextureLoaded := func() {
		loaded <- 1
	}

	// img.Set("onload", handleTextureLoaded)
	// <-loaded

	attrs := webgl.DefaultAttributes()
	attrs.Alpha = false

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

	handleTextureLoaded = func() {
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.RGBA, gl.UNSIGNED_BYTE, img)
		gl.DrawArrays(gl.TRIANGLES, 0, 6)
	}

	img.Set("onload", handleTextureLoaded)

	frame := image.NewPaletted(image.Rect(0, 0, 256, 240), RGBAPalette)
	buf := new(bytes.Buffer)

	for {
		colors := <-video.input
		x, y := 0, 0

		for _, c := range colors {
			frame.Set(x, y, RGBAPalette[c])

			switch x {
			case 255:
				x = 0
				y++
			default:
				x++
			}
		}

		buf.Reset()
		png.Encode(buf, frame)
		img.Call("setAttribute", "src", "data:image/png;base64,"+base64.StdEncoding.EncodeToString(buf.Bytes()))

		fmt.Println("frame!")
	}
}
