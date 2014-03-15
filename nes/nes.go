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
	cpu         *rp2ago3.RP2A03
	ppu         *rp2cgo2.RP2C02
	controllers *Controllers
	rom         ROM
	video       Video
}

type Options struct {
	Video     string
	CPUDecode bool
}

func NewNES(filename string, options *Options) (nes *NES, err error) {
	var video Video
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

	cycles := make(chan uint16)
	cpu := rp2ago3.NewRP2A03(cpuDivisor, cycles)

	if options.CPUDecode {
		cpu.EnableDecode()
	}

	ctrls := NewControllers()

	switch options.Video {
	case "sdl":
		video, err = NewSDLVideo(ctrls.Input)
	case "jpeg":
		video, err = NewJPEGVideo()
	default:
		err = errors.New(fmt.Sprintf("Error creating video: unknown video output %v", options.Video))
		return
	}

	if err != nil {
		err = errors.New(fmt.Sprintf("Error creating video: %v", err))
		return
	}

	ppu := rp2cgo2.NewRP2C02(cpu.InterruptLine(m65go2.Nmi), rom.Mirroring(), video.Input(), cycles)

	cpu.Memory.AddMappings(ppu, rp2ago3.CPU)
	cpu.Memory.AddMappings(rom, rp2ago3.CPU)
	cpu.Memory.AddMappings(ctrls, rp2ago3.CPU)

	ppu.Memory.AddMappings(rom, rp2ago3.PPU)

	nes = &NES{
		cpu:         cpu,
		ppu:         ppu,
		rom:         rom,
		video:       video,
		controllers: ctrls,
	}

	return
}

func (nes *NES) Reset() {
	nes.cpu.Reset()
	nes.ppu.Reset()
}

func (nes *NES) Run() (err error) {
	fmt.Println(nes.rom)

	runtime.GOMAXPROCS(runtime.NumCPU())
	nes.Reset()

	go nes.controllers.Run()
	go nes.cpu.Run()
	go nes.ppu.Run()

	runtime.LockOSThread()
	nes.video.Run()

	return
}
