package nes

import (
	"fmt"

	"github.com/nwidger/nintengo/rp2ago3"
	"github.com/nwidger/nintengo/rp2cgo2"
)

type MMC2Registers struct {
	PRGBank   uint8
	CHRBank0  uint8
	CHRBank1  uint8
	CHRBank2  uint8
	CHRBank3  uint8
	Mirroring uint8
	Latch0    uint8
	Latch1    uint8
}

type MMC2 struct {
	*ROMFile
	Registers     MMC2Registers
	refreshTables bool
}

func (reg *MMC2Registers) Reset() {
	reg.PRGBank = 0x00
	reg.CHRBank0 = 0x00
	reg.CHRBank1 = 0x00
	reg.CHRBank2 = 0x00
	reg.CHRBank3 = 0x00
	reg.Mirroring = 0x00
	reg.Latch0 = 0x00
	reg.Latch1 = 0x00
}

func NewMMC2(romf *ROMFile) *MMC2 {
	mmc2 := &MMC2{
		ROMFile: romf,
	}

	// divide 8KB CHR banks into 4KB banks since we may be
	// swapping 4KB banks
	vromBanks := make([][]uint8, romf.chrBanks*2)

	for n := 0; n < int(romf.chrBanks); n++ {
		vromBanks[2*n] = romf.vromBanks[n][0x0000:0x1000]
		vromBanks[(2*n)+1] = romf.vromBanks[n][0x1000:0x2000]
	}

	romf.vromBanks = vromBanks
	romf.chrBanks *= 2

	// divide 16KB PRG banks into 8KB banks since we may be
	// swapping 8KB banks
	romBanks := make([][]uint8, romf.prgBanks*2)

	for n := 0; n < int(romf.prgBanks); n++ {
		romBanks[2*n] = romf.romBanks[n][0x0000:0x2000]
		romBanks[(2*n)+1] = romf.romBanks[n][0x2000:0x4000]
	}

	romf.romBanks = romBanks
	romf.prgBanks *= 2

	mmc2.Registers.Reset()

	return mmc2
}

func (mmc2 *MMC2) String() string {
	return mmc2.ROMFile.String() +
		fmt.Sprintf("Mapper: 9 (MMC2)")
}

func (mmc2 *MMC2) Mappings(which rp2ago3.Mapping) (fetch, store []uint16) {
	fetch = []uint16{}
	store = []uint16{}

	switch which {
	case rp2ago3.PPU:
		if mmc2.ROMFile.chrBanks > 0 {
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
		if mmc2.ROMFile.prgBanks > 0 {
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

func (mmc2 *MMC2) Reset() {
	mmc2.Registers.Reset()
}

func (mmc2 *MMC2) Fetch(address uint16) (value uint8) {
	switch {
	// PPU only
	// CHR banks 1 & 2
	case address >= 0x0000 && address <= 0x1fff:
		index := address & 0x0fff
		lower, upper := mmc2.chrBanks()

		switch {
		// CHR bank 1
		case address >= 0x0000 && address <= 0x0fff:
			if mmc2.ROMFile.chrBanks > 0 {
				value = mmc2.ROMFile.vromBanks[lower][index]
			}
		// CHR bank 2
		case address >= 0x1000 && address <= 0x1fff:
			if mmc2.ROMFile.chrBanks > 0 {
				value = mmc2.ROMFile.vromBanks[upper][index]
			}
		}

		switch {
		case address == 0x0fd8:
			mmc2.Registers.Latch0 = 0xfd
		case address == 0x0fe8:
			mmc2.Registers.Latch0 = 0xfe
		case address >= 0x1fd8 && address <= 0x1fdf:
			mmc2.Registers.Latch1 = 0xfd
		case address >= 0x1fe8 && address <= 0x1fef:
			mmc2.Registers.Latch1 = 0xfe
		}
	// CPU only
	case address >= 0x8000 && address <= 0xffff:
		index := address & 0x1fff

		switch {
		// PRG bank 1
		case address >= 0x8000 && address <= 0x9fff:
			if mmc2.ROMFile.prgBanks > 0 {
				bank := mmc2.prgBank()
				value = mmc2.ROMFile.romBanks[bank][index]
			}
		// PRG bank 2
		case address >= 0xa000 && address <= 0xbfff:
			if mmc2.ROMFile.prgBanks > 0 {
				value = mmc2.ROMFile.romBanks[mmc2.ROMFile.prgBanks-3][index]
			}
		// PRG bank 3
		case address >= 0xc000 && address <= 0xdfff:
			if mmc2.ROMFile.prgBanks > 0 {
				value = mmc2.ROMFile.romBanks[mmc2.ROMFile.prgBanks-2][index]
			}
		// PRG bank 4
		case address >= 0xe000 && address <= 0xffff:
			if mmc2.ROMFile.prgBanks > 0 {
				value = mmc2.ROMFile.romBanks[mmc2.ROMFile.prgBanks-1][index]
			}
		}
	}

	return
}

func (mmc2 *MMC2) Store(address uint16, value uint8) (oldValue uint8) {
	switch {
	// PPU only
	// CHR banks 1 & 2
	case address >= 0x0000 && address <= 0x1fff:
		index := address & 0x0fff
		lower, upper := mmc2.chrBanks()

		switch {
		// CHR bank 1
		case address >= 0x0000 && address <= 0x0fff:
			if mmc2.ROMFile.chrBanks > 0 {
				mmc2.ROMFile.vromBanks[lower][index] = value
			}
		// CHR bank 2
		case address >= 0x1000 && address <= 0x1fff:
			if mmc2.ROMFile.chrBanks > 0 {
				mmc2.ROMFile.vromBanks[upper][index] = value
			}
		}
	// CPU only
	// PRG banks 1 & 2
	case address >= 0x8000 && address <= 0x9fff:
		bank := mmc2.prgBank()
		index := address & 0x1fff

		if mmc2.ROMFile.prgBanks > 0 {
			mmc2.ROMFile.romBanks[bank][index] = value
		}
	// PRG bank select
	case address >= 0xa000 && address <= 0xafff:
		mmc2.Registers.PRGBank = value
	// CHR $fd/0000 bank select
	case address >= 0xb000 && address <= 0xbfff:
		mmc2.Registers.CHRBank0 = value
	// CHR $fe/0000 bank select
	case address >= 0xc000 && address <= 0xcfff:
		mmc2.Registers.CHRBank1 = value
	// CHR $fd/1000 bank select
	case address >= 0xd000 && address <= 0xdfff:
		mmc2.Registers.CHRBank2 = value
	// CHR $fe/1000 bank select
	case address >= 0xe000 && address <= 0xefff:
		mmc2.Registers.CHRBank3 = value
	// Mirroring
	case address >= 0xf000 && address <= 0xffff:
		oldMirroring := mmc2.mirroring()

		mmc2.Registers.Mirroring = value

		if mmc2.mirroring() != oldMirroring {
			mmc2.refreshTables = true
		}
	}

	return
}

func (mmc2 *MMC2) prgBank() uint8 {
	return mmc2.Registers.PRGBank & 0x0f
}

func (mmc2 *MMC2) chrBank0() (bank uint8) {
	return mmc2.Registers.CHRBank0 & 0x1f
}

func (mmc2 *MMC2) chrBank1() (bank uint8) {
	return mmc2.Registers.CHRBank1 & 0x1f
}

func (mmc2 *MMC2) chrBank2() (bank uint8) {
	return mmc2.Registers.CHRBank2 & 0x1f
}

func (mmc2 *MMC2) chrBank3() (bank uint8) {
	return mmc2.Registers.CHRBank3 & 0x1f
}

func (mmc2 *MMC2) mirroring() rp2cgo2.Mirroring {
	switch mmc2.Registers.Mirroring & 0x01 {
	case 0:
		return rp2cgo2.Vertical
	default:
		return rp2cgo2.Horizontal
	}
}

func (mmc2 *MMC2) chrBanks() (lower, upper uint8) {
	switch mmc2.Registers.Latch0 {
	case 0xfd:
		lower = mmc2.chrBank0()
	case 0xfe:
		lower = mmc2.chrBank1()
	}

	switch mmc2.Registers.Latch1 {
	case 0xfd:
		upper = mmc2.chrBank2()
	case 0xfe:
		upper = mmc2.chrBank3()
	}

	return
}

func (mmc2 *MMC2) Tables() (t0, t1, t2, t3 int) {
	switch mmc2.mirroring() {
	case rp2cgo2.Vertical:
		t0, t1, t2, t3 = 0, 1, 0, 1
	case rp2cgo2.Horizontal:
		t0, t1, t2, t3 = 0, 0, 1, 1
	}

	return
}

func (mmc2 *MMC2) RefreshTables() (refresh bool) {
	refresh = mmc2.refreshTables

	if mmc2.refreshTables {
		mmc2.refreshTables = false
	}

	return refresh
}
