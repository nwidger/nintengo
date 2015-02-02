package nes

import (
	"fmt"

	"github.com/nwidger/nintengo/rp2ago3"
	"github.com/nwidger/nintengo/rp2cgo2"
)

type MMC3BankSelectFlag uint8

const (
	BankRegister MMC3BankSelectFlag = 1 << iota
	PRGROMBankMode
	CHRA12Inversion
)

type MMC3Registers struct {
	BankSelect    uint8
	BankData      uint8
	Mirroring     uint8
	PRGRAMProtect uint8

	IRQLatch   uint8
	IRQReload  bool
	IRQEnable  bool
	IRQCounter uint8

	CHRBank1 uint8
	CHRBank2 uint8
	CHRBank3 uint8
	CHRBank4 uint8
	CHRBank5 uint8
	CHRBank6 uint8

	PRGBankLow  uint8
	PRGBankHigh uint8
}

type MMC3 struct {
	*ROMFile
	Registers MMC3Registers
}

func (reg *MMC3Registers) Reset() {
	reg.BankSelect = 0x00
	reg.BankData = 0x00
	reg.Mirroring = 0x00
	reg.PRGRAMProtect = 0x00

	reg.IRQLatch = 0x00
	reg.IRQReload = false
	reg.IRQEnable = true
	reg.IRQCounter = 0x00

	reg.CHRBank1 = 0x00
	reg.CHRBank2 = 0x02
	reg.CHRBank3 = 0x04
	reg.CHRBank4 = 0x05
	reg.CHRBank5 = 0x06
	reg.CHRBank6 = 0x07
	reg.PRGBankLow = 0x00
	reg.PRGBankHigh = 0x01
}

func NewMMC3(romf *ROMFile) *MMC3 {
	mmc3 := &MMC3{
		ROMFile: romf,
	}

	// divide 8KB CHR banks into 1KB banks
	if romf.CHRBanks > 0 {
		offset := 0x0400
		vromBanks := make([][]uint8, uint16(romf.CHRBanks)*8)

		for n := 0; n < int(romf.CHRBanks); n++ {
			for i := 0; i < 8; i++ {
				vromBanks[(8*n)+i] = romf.VROMBanks[n][(offset * i):((offset * i) + offset)]
			}
		}

		romf.VROMBanks = vromBanks
		romf.CHRBanks *= 8
	}

	// divide 16KB PRG banks into 8KB banks since we may be
	// swapping 8KB banks
	if romf.PRGBanks > 0 {
		romBanks := make([][]uint8, uint16(romf.PRGBanks)*2)

		for n := 0; n < int(romf.PRGBanks); n++ {
			romBanks[2*n] = romf.ROMBanks[n][0x0000:0x2000]
			romBanks[(2*n)+1] = romf.ROMBanks[n][0x2000:0x4000]
		}

		romf.ROMBanks = romBanks
		romf.PRGBanks *= 2
	}

	mmc3.Registers.Reset()
	mmc3.setTables(mmc3.Tables())

	return mmc3
}

func (mmc3 *MMC3) String() string {
	return mmc3.ROMFile.String() +
		fmt.Sprintf("Mapper: 4 (MMC3)")
}

func (mmc3 *MMC3) Mappings(which rp2ago3.Mapping) (fetch, store []uint16) {
	fetch = []uint16{}
	store = []uint16{}

	switch which {
	case rp2ago3.PPU:
		if mmc3.CHRBanks > 0 {
			// CHR bank 1
			for i := uint32(0x0000); i <= 0x03ff; i++ {
				fetch = append(fetch, uint16(i))
				store = append(store, uint16(i))
			}

			// CHR bank 2
			for i := uint32(0x0400); i <= 0x07ff; i++ {
				fetch = append(fetch, uint16(i))
				store = append(store, uint16(i))
			}

			// CHR bank 3
			for i := uint32(0x0800); i <= 0x0bff; i++ {
				fetch = append(fetch, uint16(i))
				store = append(store, uint16(i))
			}

			// CHR bank 4
			for i := uint32(0x0c00); i <= 0x0fff; i++ {
				fetch = append(fetch, uint16(i))
				store = append(store, uint16(i))
			}

			// CHR bank 5
			for i := uint32(0x1000); i <= 0x13ff; i++ {
				fetch = append(fetch, uint16(i))
				store = append(store, uint16(i))
			}

			// CHR bank 6
			for i := uint32(0x1400); i <= 0x17ff; i++ {
				fetch = append(fetch, uint16(i))
				store = append(store, uint16(i))
			}

			// CHR bank 7
			for i := uint32(0x1800); i <= 0x1bff; i++ {
				fetch = append(fetch, uint16(i))
				store = append(store, uint16(i))
			}

			// CHR bank 8
			for i := uint32(0x1c00); i <= 0x1fff; i++ {
				fetch = append(fetch, uint16(i))
				store = append(store, uint16(i))
			}
		}
	case rp2ago3.CPU:
		if mmc3.RAMBanks > 0 {
			// PRG RAM bank
			for i := uint32(0x6000); i <= 0x7fff; i++ {
				store = append(store, uint16(i))
				fetch = append(fetch, uint16(i))
			}
		}

		if mmc3.PRGBanks > 0 {
			// PRG bank 1
			for i := uint32(0x8000); i <= 0x9fff; i++ {
				store = append(store, uint16(i))
				fetch = append(fetch, uint16(i))
			}

			// PRG bank 2
			for i := uint32(0xa000); i <= 0xbfff; i++ {
				store = append(store, uint16(i))
				fetch = append(fetch, uint16(i))
			}

			// PRG bank 3
			for i := uint32(0xc000); i <= 0xdfff; i++ {
				store = append(store, uint16(i))
				fetch = append(fetch, uint16(i))
			}

			// PRG bank 4
			for i := uint32(0xe000); i <= 0xffff; i++ {
				store = append(store, uint16(i))
				fetch = append(fetch, uint16(i))
			}
		}
	}

	return
}

func (mmc3 *MMC3) Reset() {
	mmc3.Registers.Reset()
}

func (mmc3 *MMC3) Fetch(address uint16) (value uint8) {
	switch {
	// PPU only
	// CHR banks 1-8
	case address >= 0x0000 && address <= 0x1fff:
		index := address & 0x03ff
		bank1, bank2, bank3, bank4, bank5, bank6, bank7, bank8 := mmc3.chrBanks()

		switch {
		// CHR bank 1
		case address >= 0x0000 && address <= 0x03ff:
			value = mmc3.VROMBanks[bank1][index]
		// CHR bank 2
		case address >= 0x0400 && address <= 0x07ff:
			value = mmc3.VROMBanks[bank2][index]
		// CHR bank 3
		case address >= 0x0800 && address <= 0x0bff:
			value = mmc3.VROMBanks[bank3][index]
		// CHR bank 4
		case address >= 0x0c00 && address <= 0x0fff:
			value = mmc3.VROMBanks[bank4][index]
		// CHR bank 5
		case address >= 0x1000 && address <= 0x13ff:
			value = mmc3.VROMBanks[bank5][index]
		// CHR bank 6
		case address >= 0x1400 && address <= 0x17ff:
			value = mmc3.VROMBanks[bank6][index]
		// CHR bank 7
		case address >= 0x1800 && address <= 0x1bff:
			value = mmc3.VROMBanks[bank7][index]
		// CHR bank 8
		case address >= 0x1c00 && address <= 0x1fff:
			value = mmc3.VROMBanks[bank8][index]
		}
	// CPU only
	case address >= 0x6000 && address <= 0x7fff:
		if chipEnable, _ := mmc3.prgRAMProtect(); chipEnable {
			index := address & 0x1fff
			value = mmc3.WRAMBanks[0][index]
		}
	case address >= 0x8000 && address <= 0xffff:
		index := address & 0x1fff
		bank1, bank2, bank3, bank4 := mmc3.prgBanks()

		switch {
		// PRG bank 1
		case address >= 0x8000 && address <= 0x9fff:
			value = mmc3.ROMBanks[bank1][index]
		// PRG bank 2
		case address >= 0xa000 && address <= 0xbfff:
			value = mmc3.ROMBanks[bank2][index]
		// PRG bank 3
		case address >= 0xc000 && address <= 0xdfff:
			value = mmc3.ROMBanks[bank3][index]
		// PRG bank 4
		case address >= 0xe000 && address <= 0xffff:
			value = mmc3.ROMBanks[bank4][index]
		}
	}

	return
}

func (mmc3 *MMC3) Store(address uint16, value uint8) (oldValue uint8) {
	switch {
	// PPU only
	// CHR banks 1-8
	case address >= 0x0000 && address <= 0x1fff:
		index := address & 0x03ff
		bank1, bank2, bank3, bank4, bank5, bank6, bank7, bank8 := mmc3.chrBanks()

		switch {
		// CHR bank 1
		case address >= 0x0000 && address <= 0x03ff:
			mmc3.VROMBanks[bank1][index] = value
		// CHR bank 2
		case address >= 0x0400 && address <= 0x07ff:
			mmc3.VROMBanks[bank2][index] = value
		// CHR bank 3
		case address >= 0x0800 && address <= 0x0bff:
			mmc3.VROMBanks[bank3][index] = value
		// CHR bank 4
		case address >= 0x0c00 && address <= 0x0fff:
			mmc3.VROMBanks[bank4][index] = value
		// CHR bank 5
		case address >= 0x1000 && address <= 0x13ff:
			mmc3.VROMBanks[bank5][index] = value
		// CHR bank 6
		case address >= 0x1400 && address <= 0x17ff:
			mmc3.VROMBanks[bank6][index] = value
		// CHR bank 7
		case address >= 0x1800 && address <= 0x1bff:
			mmc3.VROMBanks[bank7][index] = value
		// CHR bank 8
		case address >= 0x1c00 && address <= 0x1fff:
			mmc3.VROMBanks[bank8][index] = value
		}
	// CPU only
	// PRG RAM bank
	case address >= 0x6000 && address <= 0x7fff:
		if _, allowWrites := mmc3.prgRAMProtect(); allowWrites {
			index := address & 0x1fff
			mmc3.WRAMBanks[0][index] = value
		}
	// Bank select (even) / Bank data (odd)
	case address >= 0x8000 && address <= 0x9fff:
		if (address & 0x0001) == 0x0000 { // even
			mmc3.Registers.BankSelect = value
		} else { // odd
			switch mmc3.bankSelect(BankRegister) {
			case 0:
				mmc3.Registers.CHRBank1 = value
			case 1:
				mmc3.Registers.CHRBank2 = value
			case 2:
				mmc3.Registers.CHRBank3 = value
			case 3:
				mmc3.Registers.CHRBank4 = value
			case 4:
				mmc3.Registers.CHRBank5 = value
			case 5:
				mmc3.Registers.CHRBank6 = value
			case 6:
				mmc3.Registers.PRGBankLow = value
			case 7:
				mmc3.Registers.PRGBankHigh = value
			}

			mmc3.Registers.BankData = value
		}
	// PRG RAM protect (odd) / Mirroring (even)
	case address >= 0xa000 && address <= 0xbfff:
		if (address & 0x0001) == 0x0001 { // odd
			mmc3.Registers.PRGRAMProtect = value
		} else { // even
			oldMirroring := mmc3.mirroring()
			mmc3.Registers.Mirroring = value

			if mmc3.mirroring() != oldMirroring {
				mmc3.setTables(mmc3.Tables())
			}
		}
	// IRQ latch (even) / IRQ reload (odd)
	case address >= 0xc000 && address <= 0xdfff:
		if (address & 0x0001) == 0x0000 { // even
			mmc3.Registers.IRQLatch = value
		} else { // odd
			mmc3.Registers.IRQReload = true
		}
	// IRQ enable (odd) / IRQ disable (even)
	case address >= 0xe000 && address <= 0xffff:
		if (address & 0x0001) == 0x0001 { // odd
			mmc3.Registers.IRQEnable = true
		} else { // even
			mmc3.Registers.IRQEnable = false
			mmc3.irq(false)
		}
	}

	return
}

func (mmc3 *MMC3) scanlineCounter() {
	if mmc3.Registers.IRQReload {
		mmc3.Registers.IRQReload = false
		mmc3.Registers.IRQCounter = mmc3.Registers.IRQLatch
	} else if mmc3.Registers.IRQCounter == 0x00 {
		mmc3.Registers.IRQCounter = mmc3.Registers.IRQLatch
	} else {
		mmc3.Registers.IRQCounter--
	}

	if mmc3.Registers.IRQCounter == 0x00 && mmc3.Registers.IRQEnable {
		mmc3.irq(true)
	}
}

func (mmc3 *MMC3) bankSelect(flag MMC3BankSelectFlag) (value uint8) {
	reg := mmc3.Registers.BankSelect

	switch flag {
	case BankRegister:
		value = reg & 0x07
	case PRGROMBankMode:
		value = (reg >> 6) & 0x01
	case CHRA12Inversion:
		value = (reg >> 7) & 0x01
	}

	return
}

func (mmc3 *MMC3) mirroring() rp2cgo2.Mirroring {
	switch mmc3.Registers.Mirroring & 0x01 {
	case 0:
		return rp2cgo2.Vertical
	default:
		return rp2cgo2.Horizontal
	}
}

func (mmc3 *MMC3) prgRAMProtect() (chipEnable, allowWrites bool) {
	if (mmc3.Registers.PRGRAMProtect & 0x80) != 0 {
		chipEnable = true
	}

	if (mmc3.Registers.PRGRAMProtect & 0x40) == 0 {
		allowWrites = true
	}

	return
}

func (mmc3 *MMC3) prgBanks() (bank1, bank2, bank3, bank4 uint16) {
	switch mmc3.bankSelect(PRGROMBankMode) {
	// $8000-$9fff swappable,
	// $c000-$dfff fixed to second-last bank
	case 0:
		bank1 = uint16(mmc3.Registers.PRGBankLow) & 0x001f
		bank2 = uint16(mmc3.Registers.PRGBankHigh) & 0x001f
		bank3 = mmc3.PRGBanks - 2
		bank4 = mmc3.PRGBanks - 1
	// $c000-$dfff swappable,
	// $8000-$9fff fixed to second-last bank
	case 1:
		bank1 = mmc3.PRGBanks - 2
		bank2 = uint16(mmc3.Registers.PRGBankHigh) & 0x001f
		bank3 = uint16(mmc3.Registers.PRGBankLow) & 0x001f
		bank4 = mmc3.PRGBanks - 1
	}

	return
}

func (mmc3 *MMC3) chrBanks() (bank1, bank2, bank3, bank4, bank5, bank6, bank7, bank8 uint8) {
	switch mmc3.bankSelect(CHRA12Inversion) {
	// two 2 KB banks at $0000-$0FFF,
	// four 1 KB banks at $1000-$1FFF
	case 0:
		bank1 = mmc3.Registers.CHRBank1 & 0xfe
		bank2 = mmc3.Registers.CHRBank1 | 0x01
		bank3 = mmc3.Registers.CHRBank2 & 0xfe
		bank4 = mmc3.Registers.CHRBank2 | 0x01
		bank5 = mmc3.Registers.CHRBank3
		bank6 = mmc3.Registers.CHRBank4
		bank7 = mmc3.Registers.CHRBank5
		bank8 = mmc3.Registers.CHRBank6
	// two 2 KB banks at $1000-$1FFF,
	// four 1 KB banks at $0000-$0FFF
	case 1:
		bank1 = mmc3.Registers.CHRBank3
		bank2 = mmc3.Registers.CHRBank4
		bank3 = mmc3.Registers.CHRBank5
		bank4 = mmc3.Registers.CHRBank6
		bank5 = mmc3.Registers.CHRBank1 & 0xfe
		bank6 = mmc3.Registers.CHRBank1 | 0x01
		bank7 = mmc3.Registers.CHRBank2 & 0xfe
		bank8 = mmc3.Registers.CHRBank2 | 0x01
	}

	return
}

func (mmc3 *MMC3) Tables() (t0, t1, t2, t3 int) {
	switch mmc3.mirroring() {
	case rp2cgo2.Vertical:
		t0, t1, t2, t3 = 0, 1, 0, 1
	case rp2cgo2.Horizontal:
		t0, t1, t2, t3 = 0, 0, 1, 1
	}

	return
}
