// adapted from github.com/scottferg/Go-SDL/gfx/framerate.go

package nes

import "time"

const DEFAULT_FPS float64 = 60.0988

type FPS struct {
	frames float64
	rate   float64
	ticks  uint64
}

func NewFPS(rate float64) *FPS {
	fps := &FPS{}

	fps.SetRate(rate)

	return fps
}

func (fps *FPS) SetRate(rate float64) {
	fps.frames = 0
	fps.rate = 1000.0 / rate
	fps.ticks = uint64(time.Now().UnixNano()) / 1e6
}

func (fps *FPS) Delay() {
	// next frame
	fps.frames++

	// get/calc ticks
	current := uint64(time.Now().UnixNano()) / 1e6
	target := fps.ticks + uint64(fps.frames*fps.rate)

	if current <= target {
		time.Sleep(time.Duration((target - current) * 1e6))
	} else {
		fps.frames = 0.0
		fps.ticks = uint64(time.Now().UnixNano()) / 1e6
	}
}
