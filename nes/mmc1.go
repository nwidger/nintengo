package nes

import (
	"fmt"

	"github.com/nwidger/nintengo/rp2ago3"
)

type MMC1Mirroring uint8

const (
	OneScreenLowerBank MMC1Mirroring = iota
	OneScreenUpperBank
	Vertical
	Horizontal
)

type ControlFlag uint8

const (
	Mirroring ControlFlag = 1 << iota
	_
	PRGRomBankMode
	_
	CHRRomBankMode
)

type PRGBankFlag uint8

const (
	PRGBankSelect PRGBankFlag = 1 << iota
	_
	_
	_
	PRGRAMChipEnable
)

type MMC1Registers struct {
	Load       uint8
	Control    uint8
	CHRBank0   uint8
	CHRBank1   uint8
	PRGBank    uint8
	Shift      uint8
	ShiftCount uint8
}

type MMC1 struct {
	*ROMFile
	Registers     MMC1Registers
	refreshTables bool
}

func (reg *MMC1Registers) Reset() {
	reg.Load = 0x00
	reg.Control = 0x0f
	reg.CHRBank0 = 0x00
	reg.CHRBank1 = 0x00
	reg.PRGBank = 0x00
	reg.Shift = 0x00
	reg.ShiftCount = 0x00
}

func NewMMC1(romf *ROMFile) *MMC1 {
	mmc1 := &MMC1{
		ROMFile: romf,
	}

	// divide 8KB CHR banks into 4KB banks since we may be
	// swapping 4KB banks
	vromBanks := make([][]uint8, romf.chrBanks*2)

	for n := 0; n < int(romf.chrBanks); n++ {
		vromBanks[2*n] = romf.vromBanks[n][0x000:0x1000]
		vromBanks[(2*n)+1] = romf.vromBanks[n][0x1000:0x2000]
	}

	romf.vromBanks = vromBanks
	romf.chrBanks *= 2

	mmc1.Registers.Reset()

	return mmc1
}

func (mmc1 *MMC1) String() string {
	return mmc1.ROMFile.String() +
		fmt.Sprintf("Mapper: 1 (MMC1)")
}

func (mmc1 *MMC1) Mappings(which rp2ago3.Mapping) (fetch, store []uint16) {
	fetch = []uint16{}
	store = []uint16{}

	switch which {
	case rp2ago3.PPU:
		if mmc1.ROMFile.chrBanks > 0 {
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
		if mmc1.ROMFile.ramBanks > 0 {
			// PRG RAM bank
			for i := uint32(0x6000); i <= 0x7fff; i++ {
				store = append(store, uint16(i))
				fetch = append(fetch, uint16(i))
			}
		}

		if mmc1.ROMFile.prgBanks > 0 {
			// PRG bank 1
			for i := uint32(0x8000); i <= 0xbfff; i++ {
				store = append(store, uint16(i))
				fetch = append(fetch, uint16(i))
			}

			// PRG bank 2
			for i := uint32(0xc000); i <= 0xffff; i++ {
				store = append(store, uint16(i))
				fetch = append(fetch, uint16(i))
			}
		}

	}

	return
}

func (mmc1 *MMC1) Reset() {
	mmc1.Registers.Reset()
}

func (mmc1 *MMC1) Fetch(address uint16) (value uint8) {
	switch {
	// PPU only
	// CHR banks 1 & 2
	case address >= 0x0000 && address <= 0x1fff:
		index := address & 0x0fff
		lower, upper := mmc1.chrBanks()

		switch {
		// CHR bank 1
		case address >= 0x0000 && address <= 0x0fff:
			if mmc1.ROMFile.chrBanks > 0 {
				value = mmc1.ROMFile.vromBanks[lower][index]
			}
		// CHR bank 2
		case address >= 0x1000 && address <= 0x1fff:
			if mmc1.ROMFile.chrBanks > 0 {
				value = mmc1.ROMFile.vromBanks[upper][index]
			}
		}
	// CPU only
	case address >= 0x6000 && address <= 0xffff:
		lower, upper := mmc1.prgBanks()

		switch {
		// PRG RAM bank
		case address >= 0x6000 && address <= 0x7fff:
			index := address & 0x1fff
			value = mmc1.ROMFile.wramBanks[0][index]
		// PRG bank 1
		case address >= 0x8000 && address <= 0xbfff:
			index := address & 0x3fff

			if mmc1.ROMFile.prgBanks > 0 {
				value = mmc1.ROMFile.romBanks[lower][index]
			}
		// PRG bank 2
		case address >= 0xc000 && address <= 0xffff:
			index := address & 0x3fff

			if mmc1.ROMFile.prgBanks > 0 {
				value = mmc1.ROMFile.romBanks[upper][index]
			}
		}
	}

	return
}

func (mmc1 *MMC1) Store(address uint16, value uint8) (oldValue uint8) {
	switch {
	// PPU only
	// CHR banks 1 & 2
	case address >= 0x0000 && address <= 0x1fff:
		index := address & 0x0fff
		lower, upper := mmc1.chrBanks()

		switch {
		// CHR bank 1
		case address >= 0x0000 && address <= 0x0fff:
			if mmc1.ROMFile.chrBanks > 0 {
				mmc1.ROMFile.vromBanks[lower][index] = value
			}
		// CHR bank 2
		case address >= 0x1000 && address <= 0x1fff:
			if mmc1.ROMFile.chrBanks > 0 {
				mmc1.ROMFile.vromBanks[upper][index] = value
			}
		}
	// CPU only
	// PRG RAM bank
	case address >= 0x6000 && address <= 0x7fff:
		index := address & 0x1fff
		mmc1.ROMFile.wramBanks[0][index] = value
	// PRG banks 1 & 2
	case address >= 0x8000 && address <= 0xffff:
		oldMirrors := mmc1.control(Mirroring)

		oldValue = mmc1.Registers.Load

		if (value & 0x80) != 0 {
			mmc1.Registers.Load = 0x00

			mmc1.Registers.Shift = 0x00
			mmc1.Registers.ShiftCount = 0x00

			mmc1.Registers.Control |= 0x0c
		} else {
			mmc1.Registers.Load = value

			mmc1.Registers.Shift |= (value & 0x01) << mmc1.Registers.ShiftCount
			mmc1.Registers.ShiftCount++

			if mmc1.Registers.ShiftCount == 0x05 {
				switch (address >> 13) & 0x0003 {
				case 0x0000:
					mmc1.Registers.Control = mmc1.Registers.Shift
				case 0x0001:
					mmc1.Registers.CHRBank0 = mmc1.Registers.Shift
				case 0x0002:
					mmc1.Registers.CHRBank1 = mmc1.Registers.Shift
				case 0x0003:
					mmc1.Registers.PRGBank = mmc1.Registers.Shift
				}

				mmc1.Registers.Shift = 0x00
				mmc1.Registers.ShiftCount = 0x00
			}
		}

		if mmc1.control(Mirroring) != oldMirrors {
			mmc1.refreshTables = true
		}
	}

	return
}

// 4bit0
// -----
// CPPMM
// |||||
// |||++- Mirroring (0: one-screen, lower bank; 1: one-screen, upper bank;
// |||               2: vertical; 3: horizontal)
// |++--- PRG ROM bank mode (0, 1: switch 32 KB at $8000, ignoring low bit of bank number;
// |                         2: fix first bank at $8000 and switch 16 KB bank at $C000;
// |                         3: fix last bank at $C000 and switch 16 KB bank at $8000)
// +----- CHR ROM bank mode (0: switch 8 KB at a time; 1: switch two separate 4 KB banks)
func (mmc1 *MMC1) control(flag ControlFlag) (value uint8) {
	reg := mmc1.Registers.Control

	switch flag {
	case Mirroring:
		value = reg & 0x03
	case PRGRomBankMode:
		value = (reg >> 2) & 0x03
	case CHRRomBankMode:
		value = (reg >> 4) & 0x01
	}

	return
}

// 4bit0
// -----
// CCCCC
// |||||
// +++++- Select 4 KB or 8 KB CHR bank at PPU $0000 (low bit ignored in 8 KB mode)
func (mmc1 *MMC1) chrBank0() (bank uint8) {
	reg := mmc1.Registers.CHRBank0

	switch mmc1.control(CHRRomBankMode) {
	// 8 KB at a time
	case 0:
		bank = reg & 0x1e
	// switch two separate 4 KB banks
	case 1:
		bank = reg & 0x1f
	}

	return
}

// 4bit0
// -----
// CCCCC
// |||||
// +++++- Select 4 KB CHR bank at PPU $1000 (ignored in 8 KB mode)
func (mmc1 *MMC1) chrBank1() (bank uint8) {
	bank = mmc1.Registers.CHRBank1 & 0x1f
	return
}

// 4bit0
// -----
// RPPPP
// |||||
// |++++- Select 16 KB PRG ROM bank (low bit ignored in 32 KB mode)
// +----- PRG RAM chip enable (0: enabled; 1: disabled; ignored on MMC1A)
func (mmc1 *MMC1) prgBank(flag PRGBankFlag) (value uint8) {
	reg := mmc1.Registers.PRGBank

	switch flag {
	case PRGBankSelect:
		// (low bit ignored in 32 KB mode)
		if mmc1.control(PRGRomBankMode) < 2 {
			value = reg & 0x0e
		} else {
			value = reg & 0x0f
		}
	case PRGRAMChipEnable:
		value = (reg >> 4) & 0x01
	}

	return
}

func (mmc1 *MMC1) chrBanks() (lower, upper uint8) {
	switch mmc1.control(CHRRomBankMode) {
	// 8 KB
	case 0:
		lower = mmc1.chrBank0()
		upper = lower | 0x01
	// 4 KB
	case 1:
		lower = mmc1.chrBank0()
		upper = mmc1.chrBank1()
	}

	return
}

// |++--- PRG ROM bank mode (0, 1: switch 32 KB at $8000, ignoring low bit of bank number;
// |                         2: fix first bank at $8000 and switch 16 KB bank at $C000;
// |                         3: fix last bank at $C000 and switch 16 KB bank at $8000)
func (mmc1 *MMC1) prgBanks() (lower, upper uint8) {
	switch mmc1.control(PRGRomBankMode) {
	// 32 KB
	case 0, 1:
		lower = mmc1.prgBank(PRGBankSelect)
		upper = lower | 0x01
	// 16 KB
	case 2:
		lower = 0
		upper = mmc1.prgBank(PRGBankSelect)
	case 3:
		lower = mmc1.prgBank(PRGBankSelect)
		upper = mmc1.ROMFile.prgBanks - 1
	}

	return
}

func (mmc1 *MMC1) Tables() (t0, t1, t2, t3 int) {
	switch MMC1Mirroring(mmc1.control(Mirroring)) {
	case OneScreenLowerBank:
		t0, t1, t2, t3 = 1, 1, 1, 1
	case OneScreenUpperBank:
		t0, t1, t2, t3 = 0, 0, 0, 0
	case Vertical:
		t0, t1, t2, t3 = 0, 1, 0, 1
	case Horizontal:
		t0, t1, t2, t3 = 0, 0, 1, 1
	}

	return
}

func (mmc1 *MMC1) RefreshTables() (refresh bool) {
	refresh = mmc1.refreshTables

	if mmc1.refreshTables {
		mmc1.refreshTables = false
	}

	return refresh
}
