package nes

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/nwidger/m65go2"
	"github.com/nwidger/rp2ago3"
	"github.com/nwidger/rp2cgo2"
)

type NES struct {
	running     bool
	cpu         *rp2ago3.RP2A03
	cpuDivisor  uint16
	ppu         *rp2cgo2.RP2C02
	controllers *Controllers
	rom         ROM
	video       Video
	recorder    Recorder
}

type Options struct {
	Recorder  string
	CPUDecode bool
}

func NewNES(filename string, options *Options) (nes *NES, err error) {
	var video Video
	var recorder Recorder
	var cpuDivisor uint16

	rom, err := NewROM(filename)

	if err != nil {
		err = errors.New(fmt.Sprintf("Error loading ROM: %v", err))
		return
	}

	switch rom.Region() {
	case NTSC:
		cpuDivisor = rp2ago3.NTSC_CPU_CLOCK_DIVISOR
	case PAL:
		cpuDivisor = rp2ago3.PAL_CPU_CLOCK_DIVISOR
	}

	cpu := rp2ago3.NewRP2A03()

	if options.CPUDecode {
		cpu.EnableDecode()
	}

	ctrls := NewControllers()

	video, err = NewSDLVideo()

	if err != nil {
		err = errors.New(fmt.Sprintf("Error creating video: %v", err))
		return
	}

	switch options.Recorder {
	case "none":
		// none
	case "jpeg":
		recorder, err = NewJPEGRecorder()
	case "gif":
		recorder, err = NewGIFRecorder()
	}

	if err != nil {
		err = errors.New(fmt.Sprintf("Error creating recorder: %v", err))
		return
	}

	ppu := rp2cgo2.NewRP2C02(cpu.InterruptLine(m65go2.Nmi))

	cpu.Memory.AddMappings(ppu, rp2ago3.CPU)
	cpu.Memory.AddMappings(rom, rp2ago3.CPU)
	cpu.Memory.AddMappings(ctrls, rp2ago3.CPU)

	ppu.Memory.AddMirrors(rom.Mirrors())
	ppu.Memory.AddMappings(rom, rp2ago3.PPU)

	nes = &NES{
		cpu:         cpu,
		cpuDivisor:  cpuDivisor,
		ppu:         ppu,
		rom:         rom,
		video:       video,
		recorder:    recorder,
		controllers: ctrls,
	}

	return
}

func (nes *NES) Reset() {
	nes.cpu.Reset()
	nes.ppu.Reset()
}

type PressPause uint8
type PressQuit uint8

func (nes *NES) pause() {
	for done := false; !done; {
		switch (<-nes.video.ButtonPresses()).(type) {
		case PressPause:
			done = true
		}
	}
}

func (nes *NES) route() {
	for nes.running {
		select {
		case e := <-nes.video.ButtonPresses():
			switch i := e.(type) {
			case PressButton:
				go func() {
					nes.controllers.Input() <- i
				}()
			case PressPause:
				nes.pause()
			case PressQuit:
				nes.running = false
			}
		case e := <-nes.cpu.Cycles:
			go func() {
				nes.ppu.Cycles <- (e * nes.cpuDivisor)
				ok := <-nes.ppu.Cycles
				nes.cpu.Cycles <- ok
			}()
		case e := <-nes.ppu.Output:
			if nes.recorder != nil {
				nes.recorder.Input() <- e
			}

			go func() {
				nes.video.Input() <- e
				ok := <-nes.video.Input()
				nes.ppu.Output <- ok
			}()
		}
	}
}

func (nes *NES) Run() (err error) {
	fmt.Println(nes.rom)

	nes.Reset()

	nes.running = true

	go nes.controllers.Run()
	go nes.cpu.Run()
	go nes.ppu.Run()
	go nes.route()

	if nes.recorder != nil {
		go nes.recorder.Run()
	}

	runtime.LockOSThread()
	nes.video.Run()

	if nes.recorder != nil {
		nes.recorder.Stop()
	}

	return
}
