// adapted from github.com/scottferg/Fergulator/audio.go

package nes

import (
	"errors"
	"fmt"
	"os"

	"github.com/cryptix/wav"
	"github.com/scottferg/Go-SDL/sdl"
	sdl_audio "github.com/scottferg/Go-SDL/sdl/audio"
)

type Audio interface {
	Input() chan int16
	Run()
	TogglePaused()
}

type SDLAudio struct {
	paused  bool
	spec    sdl_audio.AudioSpec
	samples []int16
	input   chan int16
}

func NewSDLAudio(frequency int, sampleSize int) (audio *SDLAudio, err error) {
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

type AudioRecorder interface {
	Input() chan int16
	Record()
	Stop()
	Quit()
	Run()
}

type WAVRecorder struct {
	file      *os.File
	wavWriter *wav.WavWriter
	input     chan int16
	stop      chan uint8
}

func NewWAVRecorder() (wr *WAVRecorder, err error) {
	wr = &WAVRecorder{
		wavWriter: nil,
		input:     make(chan int16),
		stop:      make(chan uint8),
	}

	return
}

func (wr *WAVRecorder) Input() chan int16 {
	return wr.input
}

func (wr *WAVRecorder) Record() {
	var err error

	fmt.Println("*** Audio recording started")

	if wr.file, err = os.Create(fmt.Sprintf("audio.wav")); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	meta := wav.WavFile{
		Channels:        1,
		SampleRate:      44100 * 2,
		SignificantBits: 16,
	}

	if wr.wavWriter, err = meta.NewWriter(wr.file); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
}

func (wr *WAVRecorder) Stop() {
	var err error

	if err = wr.wavWriter.CloseFile(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	if err = wr.file.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	wr.wavWriter = nil

	fmt.Println("*** Audio recording stopped")
}

func (wr *WAVRecorder) Quit() {
	wr.stop <- 1
	<-wr.stop
}

func (wr *WAVRecorder) Run() {
	var err error

	for {
		select {
		case s := <-wr.input:
			if wr.wavWriter != nil {
				if err = wr.wavWriter.WriteInt32(int32(s)); err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
					break
				}
			}
		case <-wr.stop:
			wr.stop <- 1
			break
		}
	}
}
