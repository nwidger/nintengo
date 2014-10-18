package rp2ago3

import "github.com/nwidger/nintengo/m65go2"

const NO_PENDING uint32 = 0xffffffff

type DMA struct {
	Memory  m65go2.Memory `json:"-"`
	Pending uint32
}

func NewDMA(memory m65go2.Memory) *DMA {
	return &DMA{
		Memory:  memory,
		Pending: NO_PENDING,
	}
}

func (dma *DMA) PerformDMA() (cycles uint16) {
	if dma.Pending == NO_PENDING {
		cycles = 0
	} else {
		start := dma.Pending
		end := uint32(start + 0x0100)

		for address := start; address < end; address++ {
			dma.Memory.Store(0x2004, dma.Memory.Fetch(uint16(address)))
		}

		cycles = 512
		dma.Pending = NO_PENDING
	}

	return
}

func (dma *DMA) Reset() {
	dma.Pending = NO_PENDING
}

func (dma *DMA) Mappings(which Mapping) (fetch, store []uint16) {
	switch which {
	case CPU:
		store = []uint16{0x4014}
	}

	return
}

func (dma *DMA) Fetch(address uint16) (value uint8) {
	// nothing to do
	return
}

func (dma *DMA) Store(address uint16, value uint8) (oldValue uint8) {
	switch address {
	case 0x4014:
		dma.Pending = uint32(value) << 8
	}

	return
}
