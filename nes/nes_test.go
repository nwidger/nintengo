package nes

import (
	"testing"

	"github.com/nwidger/nintengo/rp2cgo2"
)

func TestVerticalMirroring(t *testing.T) {
	ppu := rp2cgo2.NewRP2C02(nil, "NTSC")
	ppu.Nametable.SetTables(0, 1, 0, 1)

	// Mirror nametable #2 to #0
	for i := uint16(0x2800); i <= 0x2bff; i++ {
		ppu.Memory.Store(i-0x0800, 0xff)

		if ppu.Memory.Fetch(i) != 0xff {
			t.Error("Memory is not 0xff")
		}

		ppu.Memory.Store(i-0x0800, 0x00)

		if ppu.Memory.Fetch(i) != 0x00 {
			t.Error("Memory is not 0x00")
		}

		ppu.Memory.Store(i, 0xff)

		if ppu.Memory.Fetch(i-0x0800) != 0xff {
			t.Error("Memory is not 0xff")
		}

		ppu.Memory.Store(i, 0x00)

		if ppu.Memory.Fetch(i) != 0x00 {
			t.Error("Memory is not 0x00")
		}
	}

	// Mirror nametable #3 to #1
	for i := uint16(0x2c00); i <= 0x2fff; i++ {
		ppu.Memory.Store(i-0x0800, 0xff)

		if ppu.Memory.Fetch(i) != 0xff {
			t.Error("Memory is not 0xff")
		}

		ppu.Memory.Store(i-0x0800, 0x00)

		if ppu.Memory.Fetch(i) != 0x00 {
			t.Error("Memory is not 0x00")
		}

		ppu.Memory.Store(i, 0xff)

		if ppu.Memory.Fetch(i-0x0800) != 0xff {
			t.Error("Memory is not 0xff")
		}

		ppu.Memory.Store(i, 0x00)

		if ppu.Memory.Fetch(i) != 0x00 {
			t.Error("Memory is not 0x00")
		}
	}

	// Mirror nametable #2 to #0
	for i := uint16(0x3000); i <= 0x33ff; i++ {
		ppu.Memory.Store(i-0x1000, 0xff)

		if ppu.Memory.Fetch(i) != 0xff {
			t.Error("Memory is not 0xff")
		}

		ppu.Memory.Store(i-0x1000, 0x00)

		if ppu.Memory.Fetch(i) != 0x00 {
			t.Error("Memory is not 0x00")
		}

		ppu.Memory.Store(i, 0xff)

		if ppu.Memory.Fetch(i-0x1000) != 0xff {
			t.Error("Memory is not 0xff")
		}

		ppu.Memory.Store(i, 0x00)

		if ppu.Memory.Fetch(i) != 0x00 {
			t.Error("Memory is not 0x00")
		}
	}

	// Mirror nametable #3 to #1
	for i := uint16(0x3400); i <= 0x37ff; i++ {
		ppu.Memory.Store(i-0x1000, 0xff)

		if ppu.Memory.Fetch(i) != 0xff {
			t.Error("Memory is not 0xff")
		}

		ppu.Memory.Store(i-0x1000, 0x00)

		if ppu.Memory.Fetch(i) != 0x00 {
			t.Error("Memory is not 0x00")
		}

		ppu.Memory.Store(i, 0xff)

		if ppu.Memory.Fetch(i-0x1000) != 0xff {
			t.Error("Memory is not 0xff")
		}

		ppu.Memory.Store(i, 0x00)

		if ppu.Memory.Fetch(i) != 0x00 {
			t.Error("Memory is not 0x00")
		}
	}
}

func TestHorizontalMirroring(t *testing.T) {
	ppu := rp2cgo2.NewRP2C02(nil, "NTSC")
	ppu.Nametable.SetTables(0, 0, 1, 1)

	// Mirror nametable #1 to #0
	for i := uint16(0x2400); i <= 0x27ff; i++ {
		ppu.Memory.Store(i-0x0400, 0xff)

		if ppu.Memory.Fetch(i) != 0xff {
			t.Error("Memory is not 0xff")
		}

		ppu.Memory.Store(i-0x0400, 0x00)
		ppu.Memory.Store(i, 0xff)

		if ppu.Memory.Fetch(i-0x0400) != 0xff {
			t.Error("Memory is not 0xff")
		}

		ppu.Memory.Store(i, 0x00)
	}

	// Mirror nametable #3 to #2
	for i := uint16(0x2c00); i <= 0x2fff; i++ {
		ppu.Memory.Store(i-0x0400, 0xff)

		if ppu.Memory.Fetch(i) != 0xff {
			t.Error("Memory is not 0xff")
		}

		ppu.Memory.Store(i-0x0400, 0x00)
		ppu.Memory.Store(i, 0xff)

		if ppu.Memory.Fetch(i-0x0400) != 0xff {
			t.Error("Memory is not 0xff")
		}

		ppu.Memory.Store(i, 0x00)
	}

	// Mirror nametable #1 to #0
	for i := uint16(0x3400); i <= 0x37ff; i++ {
		ppu.Memory.Store(i-0x1400, 0xff)

		if ppu.Memory.Fetch(i) != 0xff {
			t.Error("Memory is not 0xff")
		}

		ppu.Memory.Store(i-0x1400, 0x00)
		ppu.Memory.Store(i, 0xff)

		if ppu.Memory.Fetch(i-0x1400) != 0xff {
			t.Error("Memory is not 0xff")
		}

		ppu.Memory.Store(i, 0x00)
	}

	// Mirror nametable #3 to #2
	for i := uint16(0x3c00); i <= 0x3eff; i++ {
		ppu.Memory.Store(i-0x1400, 0xff)

		if ppu.Memory.Fetch(i) != 0xff {
			t.Error("Memory is not 0xff")
		}

		ppu.Memory.Store(i-0x1400, 0x00)
		ppu.Memory.Store(i, 0xff)

		if ppu.Memory.Fetch(i-0x1400) != 0xff {
			t.Error("Memory is not 0xff")
		}

		ppu.Memory.Store(i, 0x00)
	}
}
