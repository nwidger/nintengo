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
	return &JSVideo{
		input:  make(chan []uint8),
		events: events,
	}, nil
}

func (video *JSVideo) Input() chan []uint8 {
	return video.input
}

func (video *JSVideo) Events() chan Event {
	return video.events
}

func (video *JSVideo) SetCaption(caption string) {

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
precision mediump float;

varying vec2 texCoord;
uniform sampler2D texture;

void main() {
vec4 c = texture2D(texture, texCoord);
gl_FragColor = vec4(c.r, c.g, c.b, c.a);
}
`

var JSPalette []uint32 = []uint32{
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

// file:///Users/niels/go/src/github.com/nwidger/nintengo/index.html
func (video *JSVideo) Run() {
	document := js.Global.Get("document")
	canvas := document.Call("createElement", "canvas")
	canvas.Call("setAttribute", "width", "256")
	canvas.Call("setAttribute", "height", "240")
	document.Get("body").Call("appendChild", canvas)

	img := document.Call("createElement", "img")
	img.Call("setAttribute", "width", "256")
	img.Call("setAttribute", "height", "240")
	canvas.Call("appendChild", img)

	attrs := webgl.DefaultAttributes()
	attrs.Alpha = false

	gl, err := webgl.NewContext(canvas, attrs)
	if err != nil {
		js.Global.Call("alert", "Error: "+err.Error())
	}

	gl.ClearColor(0.0, 0.0, 0.0, 1.0)

	gl.Enable(gl.CULL_FACE)
	gl.Enable(gl.DEPTH_TEST)

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

	posAttrib := gl.GetAttribLocation(prog, "vPosition")
	gl.EnableVertexAttribArray(posAttrib)

	texCoordAttr := gl.GetAttribLocation(prog, "vPosition")
	gl.EnableVertexAttribArray(texCoordAttr)

	// textureUni := gl.GetAttribLocation(prog, "vPosition")

	vertVBO := gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, vertVBO)
	verts := []float32{-1.0, 1.0, -1.0, -1.0, 1.0, -1.0, 1.0, -1.0, 1.0, 1.0, -1.0, 1.0}
	gl.BufferData(gl.ARRAY_BUFFER, verts, gl.STATIC_DRAW)

	textCoorBuf := gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, textCoorBuf)
	texVerts := []float32{0.0, 1.0, 0.0, 0.0, 1.0, 0.0, 1.0, 0.0, 1.0, 1.0, 0.0, 1.0}
	gl.BufferData(gl.ARRAY_BUFFER, texVerts, gl.STATIC_DRAW)

	texture := gl.CreateTexture()
	gl.ActiveTexture(gl.TEXTURE0)

	loaded := make(chan int, 1)

	handleTextureLoaded := func() {
		fmt.Println("in handleTextureLoaded")
		gl.BindTexture(gl.TEXTURE_2D, texture)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.RGBA, gl.UNSIGNED_BYTE, img)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
		loaded <- 1
	}

	img.Set("onload", handleTextureLoaded)

	<-loaded

	gl.VertexAttribPointer(posAttrib, 2, gl.FLOAT, false, 0, 0)
	gl.VertexAttribPointer(texCoordAttr, 2, gl.FLOAT, false, 0, 0)

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

		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		gl.UseProgram(prog)

		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, texture)

		buf.Reset()
		png.Encode(buf, frame)
		img.Call("setAttribute", "src", "data:image/png;base64,"+base64.StdEncoding.EncodeToString(buf.Bytes()))

		// gl.UNSIGNED_INT_8_8_8_8
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.RGBA, gl.UNSIGNED_BYTE, img)

		gl.DrawArrays(gl.TRIANGLES, 0, 6)

		// if video.screen != nil {
		// 	sdl.GL_SwapBuffers()
		// }
	}
}
