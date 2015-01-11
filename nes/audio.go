// adapted from github.com/scottferg/Fergulator/audio.go

package nes

import (
	"fmt"
	"os"

	"github.com/cryptix/wav"
)

type Audio interface {
	Input() chan int16
	Run()
	TogglePaused()
	SetSpeed(speed float32)
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
	wavWriter *wav.Writer
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

	meta := wav.File{
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

	if err = wr.wavWriter.Close(); err != nil {
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
