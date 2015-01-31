package nes

import (
	"fmt"

	"github.com/nwidger/nintengo/rp2ago3"
)

type UNROMRegisters struct {
	BankSelect uint8
}

type UNROM struct {
	*ROMFile
	Registers UNROMRegisters
}

func (reg *UNROMRegisters) Reset() {
	reg.BankSelect = 0x00
}

func NewUNROM(romf *ROMFile) *UNROM {
	unrom := &UNROM{
		ROMFile: romf,
	}

	unrom.Registers.Reset()

	return unrom
}

func (unrom *UNROM) String() string {
	return unrom.ROMFile.String() +
		fmt.Sprintf("Mapper: 2 (UNROM)")
}

func (unrom *UNROM) Mappings(which rp2ago3.Mapping) (fetch, store []uint16) {
	fetch = []uint16{}
	store = []uint16{}

	switch which {
	case rp2ago3.PPU:
		if unrom.CHRBanks > 0 {
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
		if unrom.PRGBanks > 0 {
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

func (unrom *UNROM) Reset() {
	unrom.Registers.BankSelect = 0x00
}

func (unrom *UNROM) Fetch(address uint16) (value uint8) {
	switch {
	// PPU only
	case address >= 0x0000 && address <= 0x1fff:
		if unrom.CHRBanks > 0 {
			value = unrom.VROMBanks[0][address]
		}
	// CPU only
	case address >= 0x8000 && address <= 0xffff:
		index := address & 0x3fff

		switch {
		// PRG bank 1
		case address >= 0x8000 && address <= 0xbfff:
			if unrom.PRGBanks > 0 {
				value = unrom.ROMBanks[unrom.Registers.BankSelect][index]
			}
		// PRG bank 2
		case address >= 0xc000 && address <= 0xffff:
			if unrom.PRGBanks > 0 {
				value = unrom.ROMBanks[unrom.PRGBanks-1][index]
			}
		}
	}

	return
}

func (unrom *UNROM) Store(address uint16, value uint8) (oldValue uint8) {
	// PPU only
	switch {
	// CHR banks 1 & 2
	case address >= 0x0000 && address <= 0x1fff:
		if unrom.CHRBanks > 0 {
			unrom.VROMBanks[0][address] = value
		}
	// CPU only
	// PRG banks 1 & 2
	case address >= 0x8000 && address <= 0xffff:
		unrom.Registers.BankSelect = value & 0x07
	}

	return
}
