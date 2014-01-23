package nintengo

import (
	"errors"
	"fmt"
	"github.com/nwidger/m65go2"
	"github.com/nwidger/rp2ago3"
)

type NES struct {
	cpu    *rp2ago3.RP2A03
	memory *rp2ago3.MappedMemory
	clock  m65go2.Clocker
	rom    ROM
}

func NewNES(filename string) (nes *NES, err error) {
	clock := m65go2.NewClock(rp2ago3.NTSC_CLOCK_RATE)
	mem := rp2ago3.NewMappedMemory(m65go2.NewBasicMemory())
	cpu := rp2ago3.NewRP2A03(mem, clock, rp2ago3.NTSC_CLOCK_DIVISOR)

	rom, err := NewROM(filename)

	if err != nil {
		err = errors.New(fmt.Sprintf("Error loading ROM: %v: %v", filename, err))
		return
	}

	mem.AddMappings(rom)

	nes = &NES{cpu: cpu, memory: mem, clock: clock, rom: rom}
	return
}

func (nes *NES) Reset() {
	nes.cpu.Reset()
	nes.memory.Reset()
}

func (nes *NES) Run() (err error) {
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
