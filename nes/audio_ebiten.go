package nes

import (
	"sync"

	"github.com/hajimehoshi/ebiten/audio"
)

type EbitenAudio struct {
	input        chan int16
	frequency    int
	sampleSize   int
	audioContext *audio.Context
	s            *stream
	player       *audio.Player
}

// frequency = 44100, sampleSize = 2048
func NewAudio(frequency int, sampleSize int) (a *EbitenAudio, err error) {
	a = &EbitenAudio{
		input:      make(chan int16, sampleSize),
		frequency:  frequency,
		sampleSize: sampleSize,
		s:          &stream{},
	}
	a.audioContext, err = audio.NewContext(frequency)
	if err != nil {
		return nil, err
	}
	a.player, err = audio.NewPlayer(a.audioContext, a.s)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (a *EbitenAudio) Input() chan int16 {
	return a.input
}

func (a *EbitenAudio) Run() {
	a.player.Play()
	defer a.player.Close()
	buf := make([]byte, a.sampleSize)
	i := 0
	for {
		s := <-a.input
		buf[i] = byte(s)
		buf[i+1] = byte(s >> 8)
		buf[i+2] = byte(s)
		buf[i+3] = byte(s >> 8)
		i += 4
		if i == a.sampleSize {
			a.s.append(buf)
			i = 0
		}
	}
}

func (a *EbitenAudio) TogglePaused() {
	if a.player.IsPlaying() {
		a.player.Pause()
	} else {
		a.player.Play()
	}
}

func (a *EbitenAudio) SetSpeed(speed float32) {

}

type stream struct {
	sync.Mutex
	remaining []byte
	samples   [][]byte
}

func (s *stream) Read(buf []byte) (int, error) {
	if len(s.remaining) == 0 {
		s.Lock()
		if len(s.samples) > 0 {
			s.remaining = s.samples[0]
			s.samples = s.samples[1:]
		}
		s.Unlock()
	}
	if len(s.remaining) > 0 {
		n := copy(buf, s.remaining)
		s.remaining = s.remaining[n:]
		return n, nil
	}
	return len(buf), nil
}

func (s *stream) append(p []byte) {
	s.Lock()
	defer s.Unlock()
	s.samples = append(s.samples, p)
}

func (s *stream) Close() error {
	return nil
}
