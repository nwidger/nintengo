// +build sdl,apudebug

package nes

import (
	"fmt"
	"time"
)

type APUDebugAudio struct {
	frequency  int
	sampleSize int
	input      chan int16
}

func NewAudio(frequency int, sampleSize int) (audio *APUDebugAudio, err error) {
	audio = &APUDebugAudio{
		frequency:  frequency,
		sampleSize: sampleSize,
		input:      make(chan int16),
	}
	return
}

func (audio *APUDebugAudio) Input() chan int16 {
	return audio.input
}

func (audio *APUDebugAudio) Run() {
	print := time.NewTicker(1 * time.Second)
	start := time.Now()
	samples := 0
	for {
		select {
		case <-audio.input:
			samples++
		case <-print.C:
			t := time.Since(start)
			fmt.Println("\nAPU DEBUG:")
			fmt.Println("Total samples:", samples)
			fmt.Println("Total time:", t)
			fmt.Printf("%f samples/sec\n", float64(samples)/t.Seconds())
		}
	}
	/*
		i := 0

		for {
			select {
			case s := <-audio.input:
				audio.samples[i] = s

				if i++; i == len(audio.samples) {
					sdl_audio.SendAudio_int16(audio.samples)
					i = 0
				}
			}
		}
	*/
}

func (audio *APUDebugAudio) TogglePaused() {
}

func (audio *APUDebugAudio) Close() {
}
