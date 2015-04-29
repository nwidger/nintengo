// +build js

package nes

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/gopherjs/gopherjs/js"
)

type JSAudio struct {
	input      chan int16
	sampleSize int
}

func NewAudio(frequency int, sampleSize int) (audio *JSAudio, err error) {
	audio = &JSAudio{
		input:      make(chan int16),
		sampleSize: sampleSize * 4,
	}
	return
}

func (audio *JSAudio) Input() chan int16 {
	return audio.input
}

func wavHeader(sampleSize int) (*bytes.Buffer, error) {
	n := sampleSize * 4
	buf := &bytes.Buffer{}

	// write 'RIFF' chunkSize 'WAVE'
	header := struct {
		Ftype       [4]byte
		ChunkSize   uint32
		ChunkFormat [4]byte
	}{
		Ftype:       [4]byte{'R', 'I', 'F', 'F'},
		ChunkSize:   uint32(n + 36),
		ChunkFormat: [4]byte{'W', 'A', 'V', 'E'},
	}

	err := binary.Write(buf, binary.LittleEndian, header)
	if err != nil {
		return nil, err
	}

	// write 'fmt '
	_, err = buf.Write([]byte{'f', 'm', 't', ' '})
	if err != nil {
		return nil, err
	}

	// write RIFF chunk format
	chunkFmt := struct {
		LengthOfHeader uint32
		AudioFormat    uint16 // 1 = PCM not compressed
		NumChannels    uint16
		SampleRate     uint32
		BytesPerSec    uint32
		BytesPerBloc   uint16
		BitsPerSample  uint16
	}{
		LengthOfHeader: 16,
		AudioFormat:    1,
		NumChannels:    1,
		SampleRate:     44100 * 2,
		BytesPerSec:    44100 * 4,
		BytesPerBloc:   2,
		BitsPerSample:  16,
	}

	err = binary.Write(buf, binary.LittleEndian, chunkFmt)
	if err != nil {
		return nil, err
	}

	// write 'data'
	_, err = buf.Write([]byte{'d', 'a', 't', 'a'})
	if err != nil {
		return nil, err
	}

	// write dataSize
	err = binary.Write(buf, binary.LittleEndian, int32(n))
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func (audio *JSAudio) Run() {
	hdr, err := wavHeader(audio.sampleSize)
	if err != nil {
		fmt.Println(err)
		return
	}

	header := hdr.Bytes()

	context := js.Global.Get("AudioContext").New()

	bufChan := make(chan *js.Object, 1)
	endedChan := make(chan bool, 1)
	playing := false

	for {
		buf := bytes.NewBuffer(header)

		for i := 0; i < audio.sampleSize; i++ {
			err := binary.Write(buf, binary.LittleEndian, int32(<-audio.input))
			if err != nil {
				fmt.Println(err)
				break
			}
		}

		data := js.NewArrayBuffer(buf.Bytes())

		if data == js.Undefined {
			fmt.Println("data is undefined")
			break
		}

		context.Call("decodeAudioData", data, func(buffer *js.Object) {
			bufChan <- buffer
		}, func() {
			fmt.Println("error decoding audio")
			bufChan <- js.Undefined
		})

		buffer := <-bufChan

		if buffer == js.Undefined {
			fmt.Println("buffer is undefined")
			break
		}

		source := context.Call("createBufferSource")
		source.Set("buffer", buffer)
		source.Call("connect", context.Get("destination"))

		source.Set("onended", func(event *js.Object) {
			endedChan <- true
		})

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
