// +build js

package nes

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/jpeg"

	"github.com/gopherjs/jquery"
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

func (video *JSVideo) Run() {
	prev := ""
	frame := image.NewPaletted(image.Rect(0, 0, 256, 240), RGBAPalette)
	buf := &bytes.Buffer{}

	for {
		select {
		case colors := <-video.input:
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
			jpeg.Encode(buf, frame, nil)
			src := "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())

			if src != prev {
				jQuery := jquery.NewJQuery

				jQuery("document").Ready(func() {
					jQuery("#frame").SetAttr("src", src)
					prev = src
				})
			}
		}
	}
}
