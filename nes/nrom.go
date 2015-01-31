package nes

import (
	"fmt"

	"github.com/nwidger/nintengo/rp2ago3"
)

type NROM struct {
	*ROMFile
}

func NewNROM(romf *ROMFile) *NROM {
	return &NROM{ROMFile: romf}
}

func (nrom *NROM) String() string {
	return nrom.ROMFile.String() +
		fmt.Sprintf("Mapper: 0 (NROM)")
}

func (nrom *NROM) Mappings(which rp2ago3.Mapping) (fetch, store []uint16) {
	fetch = []uint16{}
	store = []uint16{}

	switch which {
	case rp2ago3.PPU:
		if nrom.CHRBanks > 0 {
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
		if nrom.PRGBanks > 0 {
			// PRG bank 1
			for i := uint32(0x8000); i <= 0xbfff; i++ {
				fetch = append(fetch, uint16(i))
			}

			// PRG bank 2
			for i := uint32(0xc000); i <= 0xffff; i++ {
				fetch = append(fetch, uint16(i))
			}
		}
	}

	return
}

func (nrom *NROM) Reset() {

}

func (nrom *NROM) Fetch(address uint16) (value uint8) {
	switch {
	// PPU only
	case address >= 0x0000 && address <= 0x1fff:
		if nrom.CHRBanks > 0 {
			value = nrom.VROMBanks[0][address]
		}
	// CPU only
	case address >= 0x8000 && address <= 0xffff:
		index := address & 0x3fff

		switch {
		// PRG bank 1
		case address >= 0x8000 && address <= 0xbfff:
			if nrom.PRGBanks > 0 {
				value = nrom.ROMBanks[0][index]
			}
		// PRG bank 2
		case address >= 0xc000 && address <= 0xffff:
			if nrom.PRGBanks > 0 {
				value = nrom.ROMBanks[nrom.PRGBanks-1][index]
			}
		}
	}

	return
}

func (nrom *NROM) Store(address uint16, value uint8) (oldValue uint8) {
	// PPU only
	switch {
	// CHR banks 1 & 2
	case address >= 0x0000 && address <= 0x1fff:
		if nrom.CHRBanks > 0 {
			nrom.VROMBanks[0][address] = value
		}
	}

	return
}
