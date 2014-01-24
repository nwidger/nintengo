package nintengo

import (
	"errors"
	"fmt"
	"github.com/nwidger/m65go2"
	"github.com/nwidger/rp2ago3"
	"github.com/nwidger/rp2cgo2"
	"time"
)

const NTSC_PPU_CLOCK_DIVISOR uint64 = 4
const PAL_PPU_CLOCK_DIVISOR uint64 = 5

type NES struct {
	cpu    *rp2ago3.RP2A03
	ppu    *rp2cgo2.RP2C02
	memory *rp2ago3.MappedMemory
	clock  m65go2.Clocker
	rom    ROM
}

func NewNES(filename string) (nes *NES, err error) {
	var rate time.Duration
	var cpuDivisor, ppuDivisor uint64

	rom, err := NewROM(filename)

	if err != nil {
		err = errors.New(fmt.Sprintf("Error loading ROM: %v: %v", filename, err))
		return
	}

	switch rom.Region() {
	case NTSC:
		rate = rp2ago3.NTSC_CLOCK_RATE
		cpuDivisor = rp2ago3.NTSC_CPU_CLOCK_DIVISOR
		ppuDivisor = NTSC_PPU_CLOCK_DIVISOR
	case PAL:
		rate = rp2ago3.PAL_CLOCK_RATE
		cpuDivisor = rp2ago3.PAL_CPU_CLOCK_DIVISOR
		ppuDivisor = PAL_PPU_CLOCK_DIVISOR
	}

	clock := m65go2.NewClock(rate)
	mem := rp2ago3.NewMappedMemory(m65go2.NewBasicMemory())
	cpu := rp2ago3.NewRP2A03(mem, clock, cpuDivisor)
	ppu := rp2cgo2.NewRP2C02(clock, ppuDivisor)

	mem.AddMappings(ppu)
	mem.AddMappings(rom)

	nes = &NES{cpu: cpu, ppu: ppu, memory: mem, clock: clock, rom: rom}
	return
}

func (nes *NES) Reset() {
	nes.cpu.Reset()
	nes.ppu.Reset()
	nes.memory.Reset()
}

func (nes *NES) Run() (err error) {
	nes.Reset()
	nes.clock.Start()

	for {
		err = nes.cpu.Run()

		if err != nil {
			fmt.Printf("%v\n", err)
			break
		}
	}

	return
}
