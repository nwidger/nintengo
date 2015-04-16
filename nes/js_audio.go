// +build js

package nes

type JSAudio struct {
	input chan int16
}

func NewAudio(frequency int, sampleSize int) (audio *JSAudio, err error) {
	audio = &JSAudio{
		input: make(chan int16),
	}
	return
}

func (audio *JSAudio) Input() chan int16 {
	return audio.input
}

func (audio *JSAudio) Run() {
	for {
		<-audio.input
	}
}

func (audio *JSAudio) TogglePaused() {
}

func (audio *JSAudio) SetSpeed(speed float32) {
}

func (audio *JSAudio) Close() {
}
