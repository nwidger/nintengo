package nes

import (
	"fmt"

	"github.com/nwidger/nintengo/rp2ago3"
)

type CNROMRegisters struct {
	BankSelect uint8
}

type CNROM struct {
	*ROMFile  `json:"-"`
	Registers CNROMRegisters
}

func (reg *CNROMRegisters) Reset() {
	reg.BankSelect = 0x00
}

func NewCNROM(romf *ROMFile) *CNROM {
	cnrom := &CNROM{
		ROMFile: romf,
	}

	cnrom.Registers.Reset()

	return cnrom
}

func (cnrom *CNROM) String() string {
	return cnrom.ROMFile.String() +
		fmt.Sprintf("Mapper: 3 (CNROM)")
}

func (cnrom *CNROM) Mappings(which rp2ago3.Mapping) (fetch, store []uint16) {
	fetch = []uint16{}
	store = []uint16{}

	switch which {
	case rp2ago3.PPU:
		if cnrom.ROMFile.chrBanks > 0 {
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
		if cnrom.ROMFile.prgBanks > 0 {
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

func (cnrom *CNROM) Reset() {
	cnrom.Registers.BankSelect = 0x00
}

func (cnrom *CNROM) Fetch(address uint16) (value uint8) {
	switch {
	// PPU only
	case address >= 0x0000 && address <= 0x1fff:
		if cnrom.ROMFile.chrBanks > 0 {
			value = cnrom.ROMFile.vromBanks[cnrom.Registers.BankSelect][address]
		}
	// CPU only
	case address >= 0x8000 && address <= 0xffff:
		index := address & 0x3fff

		switch {
		// PRG bank 1
		case address >= 0x8000 && address <= 0xbfff:
			if cnrom.ROMFile.prgBanks > 0 {
				value = cnrom.ROMFile.romBanks[0][index]
			}
		// PRG bank 2
		case address >= 0xc000 && address <= 0xffff:
			if cnrom.ROMFile.prgBanks > 0 {
				value = cnrom.ROMFile.romBanks[cnrom.ROMFile.prgBanks-1][index]
			}
		}
	}

	return
}

func (cnrom *CNROM) Store(address uint16, value uint8) (oldValue uint8) {
	// PPU only
	switch {
	// CHR banks 1 & 2
	case address >= 0x0000 && address <= 0x1fff:
		if cnrom.ROMFile.chrBanks > 0 {
			cnrom.ROMFile.vromBanks[cnrom.Registers.BankSelect][address] = value
		}
	// CPU only
	// PRG banks 1 & 2
	case address >= 0x8000 && address <= 0xffff:
		cnrom.Registers.BankSelect = value & 0x03
	}

	return
}
