// +build !sdl

package nes

import (
	"errors"
	"fmt"
	"os"
	"unsafe"

	"azul3d.org/native/al.v1-dev"
)

type Azul3DAudio struct {
	paused     bool
	frequency  int
	device     *al.Device
	source     uint32
	buffers    []uint32
	sampleSize int
	input      chan int16
}

func NewAudio(frequency int, sampleSize int) (audio *Azul3DAudio, err error) {
	var device *al.Device

	device, err = al.OpenDevice("", nil)

	if err != nil {
		return
	}

	audio = &Azul3DAudio{
		frequency:  frequency,
		device:     device,
		sampleSize: sampleSize,
		buffers:    make([]uint32, 2),
		input:      make(chan int16),
	}

	al.SetErrorHandler(func(e error) {
		err = e
	})

	device.GenSources(1, &audio.source)

	if !device.IsSource(audio.source) {
		err = errors.New("IsSource returned false")
		return
	}

	device.Sourcei(audio.source, al.LOOPING, al.FALSE)

	device.GenBuffers(int32(len(audio.buffers)), &audio.buffers[0])

	for i := range audio.buffers {
		if !device.IsBuffer(audio.buffers[i]) {
			err = errors.New(fmt.Sprintf("IsBuffer[%v] returned false", i))
			return
		}
	}

	return
}

func (audio *Azul3DAudio) Input() chan int16 {
	return audio.input
}

func (audio *Azul3DAudio) stream(flush chan bool, schan chan []int16) {
	samples := []int16{}
	doFlush := false

	for {
		select {
		case s := <-audio.input:
			samples = append(samples, s)
		case <-flush:
			doFlush = true
		}

		if doFlush || len(samples) == audio.sampleSize {
			if len(samples) == 0 {
				samples = append(samples, <-audio.input)
			}

			schan <- samples
			samples = []int16{}
			doFlush = false
		}
	}
}

func (audio *Azul3DAudio) bufferData(buffer uint32, samples []int16) (err error) {
	al.SetErrorHandler(func(e error) {
		err = e
	})

	audio.device.BufferData(buffer, al.FORMAT_MONO16, unsafe.Pointer(&samples[0]),
		int32(int(unsafe.Sizeof(samples[0]))*len(samples)), int32(audio.frequency))

	return
}

func (audio *Azul3DAudio) Run() {
	running := true

	handler := func(e error) {
		if e != nil {
			fmt.Fprintf(os.Stderr, "%v\n", e)
			running = false
		}
	}

	al.SetErrorHandler(handler)

	flush := make(chan bool)
	schan := make(chan []int16, 256)
	empty := []int16{0, 0}

	go audio.stream(flush, schan)

	for i := range audio.buffers {
		handler(audio.bufferData(audio.buffers[i], empty))
	}

	audio.device.SourceQueueBuffers(audio.source, audio.buffers)
	audio.device.SourcePlay(audio.source)

	state := al.PLAYING
	processed := int32(0)
	samples := []int16{0}

	for running {
		if audio.device.GetSourcei(audio.source, al.BUFFERS_PROCESSED, &processed); processed > 0 {
			pbuffers := make([]uint32, processed)

			audio.device.SourceUnqueueBuffers(audio.source, pbuffers)

			for i := range pbuffers {
				if samples == nil {
					select {
					case samples = <-schan:
					default:
						flush <- true
						samples = <-schan
					}
				}

				handler(audio.bufferData(pbuffers[i], samples))
				samples = nil
			}

			audio.device.SourceQueueBuffers(audio.source, pbuffers)

			if audio.device.GetSourcei(audio.source, al.SOURCE_STATE, &state); state != al.PLAYING {
				audio.device.SourcePlay(audio.source)
			}
		}

		if samples == nil {
			samples = <-schan
		}
	}
}

func (audio *Azul3DAudio) TogglePaused() {
	audio.device.SourcePause(audio.source)
	audio.paused = !audio.paused
}

func (audio *Azul3DAudio) Close() {
	audio.device.DeleteSources(1, &audio.source)
	audio.device.DeleteBuffers(int32(len(audio.buffers)), &audio.buffers[0])
	audio.device.Close()
}
