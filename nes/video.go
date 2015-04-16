package nes

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"os"
)

type Video interface {
	Input() chan []uint8
	Events() chan Event
	Run()
	SetCaption(caption string)
}

var RGBAPalette []color.Color = []color.Color{
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

type Recorder interface {
	Input() chan []uint8
	Record()
	Stop()
	Quit()
	Run()
}

type JPEGRecorder struct {
	frame     *image.Paletted
	palette   []color.Color
	input     chan []uint8
	recording bool
	stop      chan uint8
}

func NewJPEGRecorder() (video *JPEGRecorder, err error) {
	video = &JPEGRecorder{
		frame:   nil,
		input:   make(chan []uint8),
		palette: RGBAPalette,
		stop:    make(chan uint8),
	}

	return
}

func (video *JPEGRecorder) Input() chan []uint8 {
	return video.input
}

func (video *JPEGRecorder) Record() {
	if video.frame != nil {
		fo, _ := os.Create(fmt.Sprintf("frame.jpg"))
		w := bufio.NewWriter(fo)
		jpeg.Encode(w, video.frame, &jpeg.Options{Quality: 100})
		fmt.Println("*** Screenshot saved")
	}

	video.frame = image.NewPaletted(image.Rect(0, 0, 256, 240), video.palette)
}

func (video *JPEGRecorder) Stop() {
	video.frame = image.NewPaletted(image.Rect(0, 0, 256, 240), video.palette)
}

func (video *JPEGRecorder) Quit() {
	video.stop <- 1
	<-video.stop
}

func (video *JPEGRecorder) Run() {
	video.frame = image.NewPaletted(image.Rect(0, 0, 256, 240), video.palette)

	for {
		select {
		case colors := <-video.input:
			if video.frame == nil {
				continue
			}

			x, y := 0, 0

			for _, c := range colors {
				video.frame.Set(x, y, video.palette[c])

				switch x {
				case 255:
					x = 0
					y++
				default:
					x++
				}
			}
		case <-video.stop:
			video.stop <- 1
			break
		}
	}
}

type GIFRecorder struct {
	gif     *gif.GIF
	palette []color.Color
	input   chan []uint8
	stop    chan uint8
}

func NewGIFRecorder() (video *GIFRecorder, err error) {
	video = &GIFRecorder{
		gif:     nil,
		input:   make(chan []uint8),
		palette: RGBAPalette,
		stop:    make(chan uint8),
	}

	return
}

func (video *GIFRecorder) Input() chan []uint8 {
	return video.input
}

func (video *GIFRecorder) Record() {
	fmt.Println("*** Recording started")

	video.gif = &gif.GIF{
		Image:     []*image.Paletted{},
		Delay:     []int{},
		LoopCount: 0xfffffff,
	}
}

func (video *GIFRecorder) Stop() {
	if video.gif != nil {
		fmt.Println("*** Recording stopped")

		fo, _ := os.Create(fmt.Sprintf("frame.gif"))
		w := bufio.NewWriter(fo)
		gif.EncodeAll(w, video.gif)

		video.gif = nil
	}
}

func (video *GIFRecorder) Quit() {
	video.stop <- 1
	<-video.stop
}

func (video *GIFRecorder) Run() {
	for {
		select {
		case colors := <-video.input:
			if video.gif == nil {
				continue
			}

			frame := image.NewPaletted(image.Rect(0, 0, 256, 240), video.palette)

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

			video.gif.Image = append(video.gif.Image, frame)
			video.gif.Delay = append(video.gif.Delay, 3)
		case <-video.stop:
			video.stop <- 1
			break
		}
	}
}

func pixelInFrame(x, y int, overscan bool) bool {
	return !overscan || (y >= 8 && y <= 231 && x >= 8 && x <= 247)
}
