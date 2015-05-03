// +build js

package nes

import (
	"strconv"

	"github.com/gopherjs/gopherjs/js"
)

var JSPalette []uint32 = []uint32{
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
	input    chan []uint8
	events   chan Event
	overscan bool
}

func NewVideo(caption string, events chan Event, fps float64) (video *JSVideo, err error) {
	video = &JSVideo{
		input:    make(chan []uint8),
		events:   events,
		overscan: true,
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

func (video *JSVideo) Run() {
	width, height := 256, 240
	document := js.Global.Get("document")

	handleKey := func(e *js.Object, down bool) {
		var event Event

		code := e.Get("keyCode").Int()

		if down {
			switch code {
			case 192: // backtick `
				video.overscan = !video.overscan
			case 82: // r
				event = &ResetEvent{}
			case 80: // p
				event = &PauseEvent{}
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

	document.Set("onkeydown", func(e *js.Object) {
		handleKey(e, true)
	})

	document.Set("onkeyup", func(e *js.Object) {
		handleKey(e, false)
	})

	canvas := document.Call("createElement", "canvas")
	canvas.Call("setAttribute", "width", strconv.Itoa(width))
	canvas.Call("setAttribute", "height", strconv.Itoa(height))
	document.Get("body").Call("appendChild", canvas)

	ctx := canvas.Call("getContext", "2d")
	img := ctx.Call("getImageData", 0, 0, width, height)

	ctx.Set("fillStyle", "black")
	ctx.Call("fillRect", 0, 0, width, height)
	ctx.Set("lineWidth", 16)
	ctx.Set("strokeStyle", "white")

	for i := 3; i < img.Get("data").Length()-3; i += 4 {
		img.Get("data").SetIndex(i, 0xff)
	}

	prev := make([]uint8, width*height)
	data := img.Get("data")

	buf := js.Global.Get("ArrayBuffer").New(data.Length())
	buf8 := js.Global.Get("Uint8ClampedArray").New(buf)
	buf32 := js.Global.Get("Uint32Array").New(buf)

	for {
		colors := <-video.input

		for i, c := range colors {
			if c != prev[i] {
				buf32.SetIndex(i, JSPalette[c])
				prev[i] = c
			}
		}

		data.Call("set", buf8)
		ctx.Call("putImageData", img, 0, 0)

		if video.overscan {
			ctx.Call("strokeRect", 0, 0, width, height)
		}
	}
}
