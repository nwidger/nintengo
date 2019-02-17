// +build js

package nes

import "syscall/js"

type JSAudio struct {
	input      chan int16
	sampleSize int
}

func NewAudio(frequency int, sampleSize int) (audio *JSAudio, err error) {
	audio = &JSAudio{
		input:      make(chan int16),
		sampleSize: sampleSize,
	}
	return
}

func (audio *JSAudio) Input() chan int16 {
	return audio.input
}

func (audio *JSAudio) Run() {
	ctx := js.Global().Get("AudioContext").New()

	endedChan := make(chan bool, 1)
	playing := false

	buffer := ctx.Call("createBuffer", 1, audio.sampleSize, 44100)
	data := buffer.Call("getChannelData", 0)

	slice := make([]float32, data.Length())
	buf := js.TypedArrayOf(slice)
	defer buf.Release()

	onendedCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			endedChan <- true
		}()
		return nil
	})
	defer onendedCallback.Release()

	for {
		for i := 0; i < audio.sampleSize; i++ {
			slice[i] = float32(<-audio.input) / float32(0x7fff)
		}
		buffer.Call("copyToChannel", buf, 0)

		source := ctx.Call("createBufferSource")
		source.Set("buffer", buffer)
		source.Call("connect", ctx.Get("destination"))

		source.Set("onended", onendedCallback)

		if playing {
			<-endedChan
		}

		source.Call("start", 0)
		playing = true
	}
}

func (audio *JSAudio) TogglePaused() {
}

func (audio *JSAudio) SetSpeed(speed float32) {
}

func (audio *JSAudio) Close() {
}
