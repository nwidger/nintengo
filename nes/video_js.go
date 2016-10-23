// +build js

package nes

import (
	"strconv"
	"sync"

	"github.com/gopherjs/gopherjs/js"
)

var JSPalette []uint = []uint{
	0xff666666, 0xff882a00, 0xffa71214, 0xffa4003b, 0xff7e005c,
	0xff40006e, 0xff00066c, 0xff001d56, 0xff003533, 0xff00480b,
	0xff005200, 0xff084f00, 0xff4d4000, 0xff000000, 0xff000000,
	0xff000000, 0xffadadad, 0xffd95f15, 0xffff4042, 0xfffe2775,
	0xffcc1aa0, 0xff7b1eb7, 0xff2031b5, 0xff004e99, 0xff006d6b,
	0xff008738, 0xff00930c, 0xff328f00, 0xff8d7c00, 0xff000000,
	0xff000000, 0xff000000, 0xfffffeff, 0xffffb064, 0xffff9092,
	0xffff76c6, 0xffff6af3, 0xffcc6efe, 0xff7081fe, 0xff229eea,
	0xff00bebc, 0xff00d888, 0xff30e45c, 0xff82e045, 0xffdecd48,
	0xff4f4f4f, 0xff000000, 0xff000000, 0xfffffeff, 0xffffdfc0,
	0xffffd2d3, 0xffffc8e8, 0xffffc2fb, 0xffeac4fe, 0xffc5ccfe,
	0xffa5d8f7, 0xff94e5e4, 0xff96efcf, 0xffabf4bd, 0xffccf3b3,
	0xfff2ebb5, 0xffb8b8b8, 0xff000000, 0xff000000,
}

type JSVideo struct {
	input         chan []uint8
	events        chan Event
	framePool     *sync.Pool
	canvas        *js.Object
	width, height int
	overscan      bool
}

func NewVideo(caption string, events chan Event, framePool *sync.Pool, fps float64) (video *JSVideo, err error) {
	video = &JSVideo{
		input:     make(chan []uint8),
		events:    events,
		framePool: framePool,
		overscan:  true,
		width:     256,
		height:    240,
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

func (video *JSVideo) handleKey(code int, down bool) {
	var event Event

	setSize := func(width, height int) {
		video.canvas.Get("style").Set("width", strconv.Itoa(width)+"px")
		video.canvas.Get("style").Set("height", strconv.Itoa(height)+"px")
	}

	if down {
		switch code {
		case 192: // backtick `
			video.overscan = !video.overscan
		case 82: // r
			event = &ResetEvent{}
		case 80: // p
			event = &PauseEvent{}
		case 49: // 1
			setSize(256, 240)
		case 50: // 2
			setSize(512, 480)
		case 51: // 3
			setSize(768, 720)
		case 52: // 4
			setSize(1024, 960)
		case 53: // 5
			setSize(2560, 1440)
		case 57: // 9
			event = &ShowBackgroundEvent{}
		case 48: // 0
			event = &ShowSpritesEvent{}
		case 96: // NP-0
			event = &MuteEvent{}
		case 97: // NP-1
			event = &MutePulse1Event{}
		case 98: // NP-2
			event = &MutePulse2Event{}
		case 99: // NP-3
			event = &MuteTriangleEvent{}
		case 100: // NP-4
			event = &MuteNoiseEvent{}
		case 101: // NP-5
			event = &MuteDMCEvent{}
		}
	}

	if event == nil {
		button := button(code)
		if button != One {
			event = &ControllerEvent{
				Button: button,
				Down:   down,
			}
		}
	}

	if event != nil {
		go func() { video.events <- event }()
	}
}

func (video *JSVideo) Run() {
	imgWidth, imgHeight := 256, 240

	document := js.Global.Get("document")

	document.Set("onkeydown", func(e *js.Object) {
		video.handleKey(e.Get("keyCode").Int(), true)
	})

	document.Set("onkeyup", func(e *js.Object) {
		video.handleKey(e.Get("keyCode").Int(), false)
	})

	canvas := document.Call("createElement", "canvas")
	canvas.Call("setAttribute", "width", strconv.Itoa(imgWidth))
	canvas.Call("setAttribute", "height", strconv.Itoa(imgHeight))
	canvas.Get("style").Set("width", strconv.Itoa(video.width*2)+"px")
	canvas.Get("style").Set("height", strconv.Itoa(video.height*2)+"px")
	document.Get("body").Call("appendChild", canvas)

	video.canvas = canvas

	ctx := canvas.Call("getContext", "2d")
	img := ctx.Call("getImageData", 0, 0, imgWidth, imgHeight)

	ctx.Set("fillStyle", "black")
	ctx.Call("fillRect", 0, 0, imgWidth, imgHeight)
	ctx.Set("lineWidth", 16)
	ctx.Set("strokeStyle", "white")

	for i := 3; i < img.Get("data").Length()-3; i += 4 {
		img.Get("data").SetIndex(i, 0xff)
	}

	data := img.Get("data")

	arrBuf := js.Global.Get("ArrayBuffer").New(data.Length())
	buf8 := js.Global.Get("Uint8ClampedArray").New(arrBuf)
	buf32 := js.Global.Get("Uint32Array").New(arrBuf)

	buf := buf32.Interface().([]uint)

	for {
		colors := <-video.input

		for i, c := range colors {
			buf[i] = JSPalette[c]
		}
		video.framePool.Put(colors)

		data.Call("set", buf8)
		ctx.Call("putImageData", img, 0, 0)

		if video.overscan {
			ctx.Call("strokeRect", 0, 0, imgWidth, imgHeight)
		}
	}
}
