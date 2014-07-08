package nes

import (
	"fmt"

	"github.com/nwidger/nintengo/rp2ago3"
)

type ANROMRegisters struct {
	BankSelect uint8
}

type ANROM struct {
	*ROMFile
	Registers      ANROMRegisters
	NTMirrors      []map[uint32]uint32
	refreshMirrors bool
}

func (reg *ANROMRegisters) Reset() {
	reg.BankSelect = 0x00
}

func NewANROM(romf *ROMFile) *ANROM {
	anrom := &ANROM{
		ROMFile:   romf,
		NTMirrors: makeNTMirrors(),
	}

	anrom.Registers.Reset()

	return anrom
}

func (anrom *ANROM) String() string {
	return anrom.ROMFile.String() +
		fmt.Sprintf("Mapper: 7 (ANROM)")
}

func (anrom *ANROM) Mappings(which rp2ago3.Mapping) (fetch, store []uint16) {
	fetch = []uint16{}
	store = []uint16{}

	switch which {
	case rp2ago3.PPU:
		if anrom.ROMFile.chrBanks > 0 {
			// CHR bank 1
			for i := uint32(0x0000); i <= 0x0fff; i++ {
				fetch = append(fetch, uint16(i))
				store = append(store, uint16(i))
			}

			// CHR bank 2
			for i := uint32(0x1000); i <= 0x1fff; i++ {
				fetch = append(fetch, uint16(i))
				store = append(store, uint16(i))
			}
		}
	case rp2ago3.CPU:
		if anrom.ROMFile.prgBanks > 0 {
			// PRG bank 1
			for i := uint32(0x8000); i <= 0xbfff; i++ {
				fetch = append(fetch, uint16(i))
				store = append(store, uint16(i))
			}

			// PRG bank 2
			for i := uint32(0xc000); i <= 0xffff; i++ {
				fetch = append(fetch, uint16(i))
				store = append(store, uint16(i))
			}
		}
	}

	return
}

func (anrom *ANROM) Reset() {
	anrom.Registers.BankSelect = 0x00
}

func (anrom *ANROM) Fetch(address uint16) (value uint8) {
	switch {
	// PPU only
	case address >= 0x0000 && address <= 0x1fff:
		if anrom.ROMFile.chrBanks > 0 {
			value = anrom.ROMFile.vromBanks[0][address]
		}
	// CPU only
	case address >= 0x8000 && address <= 0xffff:
		index := address & 0x3fff
		lower, upper := anrom.prgBanks()

		switch {
		// PRG bank 1
		case address >= 0x8000 && address <= 0xbfff:
			if anrom.ROMFile.prgBanks > 0 {
				value = anrom.ROMFile.romBanks[lower][index]
			}
		// PRG bank 2
		case address >= 0xc000 && address <= 0xffff:
			if anrom.ROMFile.prgBanks > 0 {
				value = anrom.ROMFile.romBanks[upper][index]
			}
		}
	}

	return
}

func (anrom *ANROM) Store(address uint16, value uint8) (oldValue uint8) {
	// PPU only
	switch {
	// CHR banks 1 & 2
	case address >= 0x0000 && address <= 0x1fff:
		if anrom.ROMFile.chrBanks > 0 {
			anrom.ROMFile.vromBanks[0][address] = value
		}
	// CPU only
	// PRG banks 1 & 2
	case address >= 0x8000 && address <= 0xffff:
		oldMirrors := anrom.mirroring()
		anrom.Registers.BankSelect = value

		if anrom.mirroring() != oldMirrors {
			anrom.refreshMirrors = true
		}
	}

	return
}

func (anrom *ANROM) mirroring() uint8 {
	return (anrom.Registers.BankSelect >> 4) & 0x01
}

func (anrom *ANROM) prgBanks() (lower, upper uint8) {
	bank := (anrom.Registers.BankSelect & 0x07) << 1

	lower = bank
	upper = lower + 1

	return
}

func (anrom *ANROM) Mirrors() (mirrors map[uint32]uint32) {
	return anrom.NTMirrors[MMC1Mirroring(anrom.mirroring())]
}

func (anrom *ANROM) RefreshMirrors() (refresh bool) {
	refresh = anrom.refreshMirrors

	if anrom.refreshMirrors {
		anrom.refreshMirrors = false
	}

	return refresh
}
