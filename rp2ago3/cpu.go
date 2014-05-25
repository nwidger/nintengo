package rp2ago3

import "github.com/nwidger/nintengo/m65go2"

const NTSC_CPU_CLOCK_DIVISOR float32 = 3
const PAL_CPU_CLOCK_DIVISOR float32 = 3.2

type RP2A03 struct {
	*m65go2.M6502
	*APU
	dma    *DMA
	Memory *MappedMemory
}

func NewRP2A03() *RP2A03 {
	mem := NewMappedMemory(m65go2.NewBasicMemory(m65go2.DEFAULT_MEMORY_SIZE))
	mirrors := make(map[uint16]uint16)

	// Mirrored 2KB internal RAM
	for i := uint16(0x0800); i <= 0x1fff; i++ {
		mirrors[i] = i % 0x0800
	}

	// Mirrored PPU registers
	for i := uint16(0x2008); i <= 0x3fff; i++ {
		mirrors[i] = 0x2000 + (i & 0x0007)
	}

	mem.AddMirrors(mirrors)

	cpu := m65go2.NewM6502(mem)
	cpu.DisableDecimalMode()
	apu := NewAPU()

	// APU memory maps
	mem.AddMappings(apu, CPU)

	dma := NewDMA(mem)

	// DMA memory maps
	mem.AddMappings(dma, CPU)

	return &RP2A03{
		Memory: mem,
		M6502:  cpu,
		APU:    apu,
		dma:    dma,
	}
}

func (cpu *RP2A03) Reset() {
	cpu.M6502.Reset()
	cpu.APU.Reset()
	cpu.Memory.Reset()
}

func (cpu *RP2A03) Execute() (cycles uint16, err error) {
	if cycles, err = cpu.M6502.Execute(); err != nil {
		return
	}

	cycles += cpu.dma.PerformDMA()

	cpu.APU.Execute()

	return
}

func (cpu *RP2A03) Run() (err error) {
	for {
		if _, err = cpu.Execute(); err != nil {
			break
		}
	}

	return
}
