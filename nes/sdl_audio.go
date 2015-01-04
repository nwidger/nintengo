// +build sdl

// adapted from github.com/scottferg/Fergulator/audio.go

package nes

import (
	"errors"

	"github.com/scottferg/Go-SDL/sdl"
	sdl_audio "github.com/scottferg/Go-SDL/sdl/audio"
)

type SDLAudio struct {
	paused  bool
	spec    sdl_audio.AudioSpec
	samples []int16
	input   chan int16
}

func NewAudio(frequency int, sampleSize int) (audio *SDLAudio, err error) {
	spec := sdl_audio.AudioSpec{
		Freq:        frequency,
		Format:      sdl_audio.AUDIO_S16SYS,
		Channels:    1,
		Out_Silence: 0,
		Samples:     uint16(sampleSize),
		Out_Size:    0,
	}

	if sdl_audio.OpenAudio(&spec, nil) < 0 {
		err = errors.New(sdl.GetError())
		return
	}

	sdl_audio.PauseAudio(false)

	audio = &SDLAudio{
		samples: make([]int16, sampleSize),
		input:   make(chan int16),
	}

	return
}

func (audio *SDLAudio) Input() chan int16 {
	return audio.input
}

func (audio *SDLAudio) Run() {
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
}

func (audio *SDLAudio) TogglePaused() {
	audio.paused = !audio.paused
	sdl_audio.PauseAudio(audio.paused)
}

func (audio *SDLAudio) Close() {
	sdl_audio.PauseAudio(true)
	sdl_audio.CloseAudio()
}
