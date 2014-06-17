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

type Registers struct {
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
	Registers      Registers
	NTMirrors      []map[uint32]uint32
	refreshMirrors bool
}

func (reg *Registers) Reset() {
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
		ROMFile:   romf,
		NTMirrors: makeNTMirrors(),
	}

	mmc1.Registers.Reset()

	return mmc1
}

func makeNTMirrors() []map[uint32]uint32 {
	m := make([]map[uint32]uint32, 4)

	for mirroring := range [4]MMC1Mirroring{OneScreenLowerBank, OneScreenUpperBank, Vertical, Horizontal} {
		mirrors := make(map[uint32]uint32, 0x1000)

		for i := uint32(0x2000); i <= 0x2fff; i++ {
			mirrors[i] = rp2ago3.UNMIRRORED
		}

		switch mirroring {
		case int(OneScreenLowerBank):
			// Mirror nametable #0 to #1
			for i := uint32(0x2000); i <= 0x23ff; i++ {
				mirrors[i] = i + 0x0400
			}

			// Mirror nametable #2 to #1
			for i := uint32(0x2800); i <= 0x2bff; i++ {
				mirrors[i] = i - 0x0400
			}

			// Mirror nametable #3 to #1
			for i := uint32(0x2c00); i <= 0x2fff; i++ {
				mirrors[i] = i - 0x0800
			}
		case int(OneScreenUpperBank):
			// Mirror nametable #1 to #0
			for i := uint32(0x2400); i <= 0x27ff; i++ {
				mirrors[i] = i - 0x0400
			}

			// Mirror nametable #2 to #0
			for i := uint32(0x2800); i <= 0x2bff; i++ {
				mirrors[i] = i - 0x0800
			}

			// Mirror nametable #3 to #0
			for i := uint32(0x2c00); i <= 0x2fff; i++ {
				mirrors[i] = i - 0x0c00
			}
		case int(Horizontal):
			// Mirror nametable #1 to #0
			for i := uint32(0x2400); i <= 0x27ff; i++ {
				mirrors[i] = i - 0x0400
			}

			// Mirror nametable #3 to #2
			for i := uint32(0x2c00); i <= 0x2fff; i++ {
				mirrors[i] = i - 0x0400
			}
		case int(Vertical):
			// Mirror nametable #2 to #0
			for i := uint32(0x2800); i <= 0x2bff; i++ {
				mirrors[i] = i - 0x0800
			}

			// Mirror nametable #3 to #1
			for i := uint32(0x2c00); i <= 0x2fff; i++ {
				mirrors[i] = i - 0x0800
			}
		}

		m[mirroring] = mirrors
	}

	return m
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
	case rp2ago3.CPU:
		// PRG RAM bank
		for i := uint32(0x6000); i <= 0x7fff; i++ {
			store = append(store, uint16(i))
			fetch = append(fetch, uint16(i))
		}

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
		index := address & 0x3fff
		lower, upper := mmc1.prgBanks()

		switch {
		// PRG RAM bank
		case address >= 0x6000 && address <= 0x7fff:
			// value = mmc1.ROMFile.ramBank[address]
		// PRG bank 1
		case address >= 0x8000 && address <= 0xbfff:
			if mmc1.ROMFile.prgBanks > 0 {
				value = mmc1.ROMFile.romBanks[lower][index]
			}
		// PRG bank 2
		case address >= 0xc000 && address <= 0xffff:
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
		// mmc1.ROMFile.ramBank[address] = value
	// PRG banks 1 & 2
	case address >= 0x8000 && address <= 0xffff:
		oldMirrors := mmc1.control(Mirroring)

		if (value & 0x80) != 0 {
			oldValue = mmc1.Registers.Load
			mmc1.Registers.Load = 0x00

			mmc1.Registers.Shift = 0x00
			mmc1.Registers.ShiftCount = 0x00

			mmc1.Registers.Control = mmc1.Registers.Control | 0x0c
		} else {
			oldValue = mmc1.Registers.Load
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
			mmc1.refreshMirrors = true
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
	switch mmc1.control(CHRRomBankMode) {
	// 8 KB at a time
	case 0:
		bank = mmc1.Registers.CHRBank0 & 0x1e
	// switch two separate 4 KB banks
	case 1:
		bank = mmc1.Registers.CHRBank0 & 0x1f
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
		upper = lower + 1
	// 4 KB
	case 1:
		lower = mmc1.chrBank0()
		upper = mmc1.chrBank1()
	}

	// fmt.Printf("mode = %v, chrBank0 = %v, chrBank1 = %v, lower = %v, upper = %v\n",
	// 	mmc1.control(CHRRomBankMode), mmc1.chrBank0(), mmc1.chrBank1(), lower, upper)

	return
}

func (mmc1 *MMC1) prgBanks() (lower, upper uint8) {
	switch mmc1.control(PRGRomBankMode) {
	// 32 KB
	case 0:
		fallthrough
	case 1:
		lower = mmc1.prgBank(PRGBankSelect)
		upper = lower + 1
	// 16 KB
	case 2:
		lower = 0
		upper = mmc1.prgBank(PRGBankSelect)
	case 3:
		lower = mmc1.prgBank(PRGBankSelect)
		upper = mmc1.ROMFile.prgBanks - 1
	}

	// fmt.Printf("mode = %v, prgBank = %v, lower = %v, upper = %v\n",
	// 	mmc1.control(PRGRomBankMode), mmc1.prgBank(PRGBankSelect), lower, upper)

	return
}

func (mmc1 *MMC1) Mirrors() (mirrors map[uint32]uint32) {
	return mmc1.NTMirrors[MMC1Mirroring(mmc1.control(Mirroring))]
}

func (mmc1 *MMC1) RefreshMirrors() (refresh bool) {
	refresh = mmc1.refreshMirrors

	if mmc1.refreshMirrors {
		mmc1.refreshMirrors = false
	}

	return refresh
}
