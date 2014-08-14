package rp2ago3

import "testing"

func TestStore(t *testing.T) {
	cpu := NewRP2A03()
	cpu.Reset()

	cpu.APU.Pulse1.Registers[0] = 0xde
	cpu.Memory.Store(0x4000, 0xff)

	if cpu.APU.Pulse1.Registers[0] != 0xff {
		t.Error("Register is not 0xff")
	}

	cpu.Memory.Store(0x0800, 0xff)

	if cpu.Memory.Fetch(0x0000) != 0xff {
		t.Error("Memory is not 0xff")
	}

	cpu.Memory.Store(0x0800, 0x00)

	if cpu.Memory.Fetch(0x0000) != 0x00 {
		t.Error("Memory is not 0x00")
	}
}

type FakePPU struct {
	i      uint16
	memory [65536]uint8
}

func (ppu *FakePPU) Reset() {

}

func (ppu *FakePPU) Mappings(which Mapping) (fetch, store []uint16) {
	store = []uint16{0x2004}
	return
}

func (ppu *FakePPU) Fetch(address uint16) (value uint8) {
	return
}

func (ppu *FakePPU) Store(address uint16, value uint8) (oldValue uint8) {
	switch address {
	case 0x2004:
		ppu.memory[ppu.i] = value
		ppu.i++
	}

	return
}

func TestDMA(t *testing.T) {
	cpu := NewRP2A03()
	ppu := &FakePPU{}

	cpu.Memory.AddMappings(ppu, CPU)

	for address := uint32(0xff00); address <= 0xffff; address++ {
		cpu.Memory.Store(uint16(address), uint8(address&0x00ff))
	}

	cpu.Memory.Store(0x4014, 0xff)
	cpu.dma.PerformDMA()

	for address := uint32(0x0000); address <= 0x00ff; address++ {
		if ppu.memory[uint16(address)] != uint8(address) {
			t.Errorf("Memory is not %02X", uint8(address))
		}
	}
}
