package rp2cgo2

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"

	"github.com/nwidger/nintengo/m65go2"
	"github.com/nwidger/nintengo/rp2ago3"
)

type Mirroring uint8

const (
	Horizontal Mirroring = iota
	Vertical
	FourScreen
)

func (m Mirroring) String() string {
	switch m {
	case Horizontal:
		return "Horizontal"
	case Vertical:
		return "Vertical"
	case FourScreen:
		return "FourScreen"
	}

	return "Unknown"
}

type ControllerFlag uint8

const (
	BaseNametableAddress ControllerFlag = 1 << iota
	_
	VRAMAddressIncrement
	SpritePatternAddress
	BackgroundPatternAddress
	SpriteSize
	_
	NMIOnVBlank
)

type MaskFlag uint8

const (
	Grayscale MaskFlag = 1 << iota
	ShowBackgroundLeft
	ShowSpritesLeft
	ShowBackground
	ShowSprites
	IntensifyReds
	IntensifyGreens
	IntensifyBlues
)

type StatusFlag uint8

const (
	_ StatusFlag = 1 << iota
	_
	_
	_
	_
	SpriteOverflow
	Sprite0Hit
	VBlankStarted
)

type AddressFlag uint16

const (
	CoarseXScroll AddressFlag = 1 << iota
	_
	_
	_
	_
	CoarseYScroll
	_
	_
	_
	_
	NametableSelect
	_
	FineYScroll
	_
	_
	_
)

type SpriteFlag uint32

const (
	// byte 0
	YPosition SpriteFlag = 1 << iota
	_
	_
	_
	_
	_
	_
	_
	// byte 1
	TileBank
	TileNumber
	_
	_
	_
	_
	_
	_
	// byte 2
	SpritePalette
	_
	_
	_
	_
	Priority
	FlipHorizontally
	FlipVertically
	// byte 3
	XPosition
	_
	_
	_
	_
	_
	_
	_
)

type Registers struct {
	Controller uint8
	Mask       uint8
	Status     uint8
	OAMAddress uint8
	Scroll     uint16
	Address    uint16
	Data       uint8
}

func (reg *Registers) Reset() {
	reg.Controller = 0x00
	reg.Mask = 0x00
	reg.Status = 0x00
	reg.OAMAddress = 0x00
	reg.Scroll = 0x00
	reg.Address = 0x00
	reg.Data = 0x00
}

const (
	CYCLES_PER_SCANLINE uint16 = 341
	NUM_SCANLINES              = 262
	POWERUP_SCANLINE           = 241
)

type Sprite struct {
	TileLow   uint8
	TileHigh  uint8
	Sprite    uint32
	XPosition uint8
}

type RP2C02 struct {
	decode bool

	frame    uint16
	scanline uint16
	cycle    uint16

	Output    chan []uint8
	colors    []uint8
	Registers Registers
	Memory    *rp2ago3.MappedMemory
	Interrupt func(state bool)
	oam       *OAM

	latch        bool
	latchAddress uint16
	latchValue   uint8

	addressLine    uint16
	patternAddress uint16

	attributeNext  uint8
	attributeLatch uint8
	attributes     uint16

	tilesLatch uint16
	tilesLow   uint16
	tilesHigh  uint16

	sprites        [8]Sprite
	ShowBackground bool
	ShowSprites    bool
}

func NewRP2C02(interrupt func(bool)) *RP2C02 {
	mem := rp2ago3.NewMappedMemory(m65go2.NewBasicMemory(m65go2.DEFAULT_MEMORY_SIZE))
	mirrors := make(map[uint32]uint32)

	// Mirrored nametables
	for i := uint32(0x3000); i <= 0x3eff; i++ {
		mirrors[i] = i - 0x1000
	}

	// Mirrored palette
	for _, i := range []uint32{0x3f10, 0x3f14, 0x3f18, 0x3f1c} {
		mirrors[i] = i - 0x0010
	}

	for i := uint32(0x3f20); i <= 0x3fff; i++ {
		mirrors[i] = i - 0x0020
	}

	mem.AddMirrors(mirrors)

	return &RP2C02{
		Output:         make(chan []uint8),
		colors:         make([]uint8, 0xf000),
		Memory:         mem,
		Interrupt:      interrupt,
		oam:            NewOAM(),
		ShowBackground: true,
		ShowSprites:    true,
	}
}

func (ppu *RP2C02) ToggleDecode() bool {
	ppu.decode = !ppu.decode
	return ppu.decode
}

func (ppu *RP2C02) Reset() {
	ppu.latch = false
	ppu.Registers.Reset()
	ppu.Memory.Reset()

	ppu.frame = 0
	ppu.cycle = 0
	ppu.scanline = POWERUP_SCANLINE
}

func (ppu *RP2C02) controller(flag ControllerFlag) (value uint16) {
	byte := ppu.Registers.Controller
	bit := byte & uint8(flag)

	switch flag {
	case BaseNametableAddress:
		// 0x2000 | 0x2400 | 0x2800 | 0x2c00
		value = 0x2000 | (uint16(byte&0x03) << 10)
	case VRAMAddressIncrement:
		switch bit {
		case 0:
			value = 1
		default:
			value = 32
		}
	case SpritePatternAddress:
		// 0x0000 | 0x1000
		switch bit {
		case 0:
			value = 0x0000
		default:
			value = 0x1000
		}
	case BackgroundPatternAddress:
		// 0x0000 | 0x1000
		switch bit {
		case 0:
			value = 0x0000
		default:
			value = 0x1000
		}

	case SpriteSize:
		// 8x8 | 8x16
		switch bit {
		case 0:
			value = 8
		default:
			value = 16
		}
	case NMIOnVBlank:
		switch bit {
		case 0:
			value = 0
		default:
			value = 1
		}
	}

	return
}

func (ppu *RP2C02) mask(flag MaskFlag) (value bool) {
	if ppu.Registers.Mask&uint8(flag) != 0 {
		value = true
	}

	return
}

func (ppu *RP2C02) status(flag StatusFlag) (value bool) {
	if ppu.Registers.Status&uint8(flag) != 0 {
		value = true
	}

	return
}

func (ppu *RP2C02) address(flag AddressFlag) (value uint16) {
	word := ppu.Registers.Address

	switch flag {
	case CoarseXScroll:
		value = word & 0x001f
	case CoarseYScroll:
		value = (word & 0x03e0) >> 5
	case NametableSelect:
		value = (word & 0x0c00) >> 10
	case FineYScroll:
		value = (word & 0x7000) >> 12
	}

	return
}

func (ppu *RP2C02) sprite(sprite uint32, flag SpriteFlag) (value uint8) {
	switch flag {
	case YPosition:
		value = uint8(sprite >> 24)
	case TileBank:
		value = uint8((sprite & 0x00010000) >> 16)
	case TileNumber:
		value = uint8((sprite & 0x00ff0000) >> 16)

		if ppu.controller(SpriteSize) == 16 {
			value >>= 1
		}
	case SpritePalette:
		value = uint8((sprite & 0x00000300) >> 8)
	case Priority:
		value = uint8((sprite & 0x00002000) >> 13)
	case FlipHorizontally:
		value = uint8((sprite & 0x00004000) >> 14)
	case FlipVertically:
		value = uint8((sprite & 0x00008000) >> 15)
	case XPosition:
		value = uint8(sprite)
	}

	return
}

func (ppu *RP2C02) Mappings(which rp2ago3.Mapping) (fetch, store []uint16) {
	fetch = []uint16{}
	store = []uint16{}

	switch which {
	case rp2ago3.CPU:
		for i := uint16(0x2000); i <= 0x2007; i++ {
			switch i {
			case 0x2000:
				store = append(store, i)
			case 0x2001:
				fetch = append(fetch, i)
				store = append(store, i)
			case 0x2002:
				fetch = append(fetch, i)
			case 0x2003:
				store = append(store, i)
			case 0x2004:
				fetch = append(fetch, i)
				store = append(store, i)
			case 0x2005:
				store = append(store, i)
			case 0x2006:
				store = append(store, i)
			case 0x2007:
				fetch = append(fetch, i)
				store = append(store, i)
			}
		}
	}

	return
}

func (ppu *RP2C02) Fetch(address uint16) (value uint8) {
	switch address {
	// Mask
	case 0x2001:
		value = ppu.latchValue
	// Status
	case 0x2002:
		value = (ppu.Registers.Status & 0xe0) | (ppu.latchValue & 0x1f)
		ppu.Registers.Status &^= uint8(VBlankStarted)
		ppu.latch = false
	// OAMData
	case 0x2004:
		value = ppu.oam.Fetch(uint16(ppu.Registers.OAMAddress))
	// Data
	case 0x2007:
		value = ppu.Registers.Data

		vramAddress := ppu.Registers.Address & 0x3fff
		ppu.Registers.Data = ppu.Memory.Fetch(vramAddress)

		if vramAddress&0x3f00 == 0x3f00 {
			value = ppu.Registers.Data
		}

		ppu.incrementAddress()
	}

	return
}

func (ppu *RP2C02) Store(address uint16, value uint8) (oldValue uint8) {
	ppu.latchValue = value

	switch address {
	// Controller
	case 0x2000:
		// t: ...BA.. ........ = d: ......BA
		oldValue = ppu.Registers.Controller
		ppu.Registers.Controller = value
		ppu.latchAddress = (ppu.latchAddress & 0x73ff) | uint16(value&0x03)<<10
	// Mask
	case 0x2001:
		oldValue = ppu.Registers.Mask
		old := ppu.rendering()
		ppu.Registers.Mask = value
		new := ppu.rendering()

		if new != old && ppu.decode {
			fmt.Printf("*** $2001 (mask) rendering switch from %v to %v: %08b -> %08b\n", old, new, oldValue, value)
		}
	// OAMAddress
	case 0x2003:
		oldValue = ppu.Registers.OAMAddress
		ppu.Registers.OAMAddress = value
	// OAMData
	case 0x2004:
		oldValue = ppu.oam.Fetch(uint16(ppu.Registers.OAMAddress))
		ppu.oam.Store(uint16(ppu.Registers.OAMAddress), value)
		ppu.Registers.OAMAddress++
	// Scroll
	case 0x2005:
		if !ppu.latch {
			// t: ....... ...HGFED = d: HGFED...
			// x:              CBA = d: .....CBA
			ppu.latchAddress = (ppu.latchAddress & 0x7fe0) | uint16(value>>3)
			ppu.Registers.Scroll = uint16(value & 0x07)
		} else {
			// t: CBA..HG FED..... = d: HGFEDCBA
			ppu.latchAddress = (ppu.latchAddress & 0x0c1f) | ((uint16(value)<<2 | uint16(value)<<12) & 0x73e0)
		}

		ppu.latch = !ppu.latch
	// Address
	case 0x2006:
		if !ppu.latch {
			// t: .FEDCBA ........ = d: ..FEDCBA
			// t: X...... ........ = 0
			ppu.latchAddress = (ppu.latchAddress & 0x00ff) | (uint16(value&0x3f) << 8)
		} else {
			// t: ....... HGFEDCBA = d: HGFEDCBA
			// v                   = t
			ppu.latchAddress = (ppu.latchAddress & 0x7f00) | uint16(value)
			ppu.Registers.Address = ppu.latchAddress
		}

		ppu.latch = !ppu.latch
	// Data
	case 0x2007:
		oldValue = ppu.Registers.Data
		ppu.Memory.Store(ppu.Registers.Address&0x3fff, value)
		ppu.incrementAddress()
	}

	return
}

func (ppu *RP2C02) transferX() {
	// v: ....F.. ...EDCBA = t: ....F.. ...EDCBA
	ppu.Registers.Address = (ppu.Registers.Address & 0x7be0) | (ppu.latchAddress & 0x041f)
}

func (ppu *RP2C02) transferY() {
	// v: IHGF.ED CBA..... = t: IHGF.ED CBA.....
	ppu.Registers.Address = (ppu.Registers.Address & 0x041f) | (ppu.latchAddress & 0x7be0)
}

func (ppu *RP2C02) incrementX() {
	// v: .yyy NN YYYYY XXXXX
	//     ||| || ||||| +++++-- coarse X scroll
	//     ||| || +++++-------- coarse Y scroll
	//     ||| ++-------------- nametable select
	//     +++----------------- fine Y scroll
	v := ppu.Registers.Address

	switch v & 0x001f {
	case 0x001f: // coarse X == 31
		v ^= 0x041f // coarse X = 0, switch horizontal nametable
	default:
		v++ // increment coarse X
	}

	ppu.Registers.Address = v
}

func (ppu *RP2C02) incrementY() {
	// v: .yyy NN YYYYY XXXXX
	//     ||| || ||||| +++++-- coarse X scroll
	//     ||| || +++++-------- coarse Y scroll
	//     ||| ++-------------- nametable select
	//     +++----------------- fine Y scroll
	v := ppu.Registers.Address

	if (v & 0x7000) != 0x7000 { // if fine Y < 7
		v += 0x1000 // increment fine Y
	} else {
		v &= 0x0fff

		switch v & 0x3e0 {
		case 0x03a0: // coarse Y = 29
			v ^= 0x0ba0 // switch vertical nametable
		case 0x03e0: // coarse Y = 31
			v ^= 0x03e0 // coarse Y = 0, nametable not switched
		default:
			v += 0x0020 // increment coarse Y
		}
	}

	ppu.Registers.Address = v
}

func (ppu *RP2C02) incrementAddress() {
	if (ppu.scanline > 239 && ppu.scanline != 261) || !ppu.rendering() {
		ppu.Registers.Address =
			(ppu.Registers.Address + ppu.controller(VRAMAddressIncrement)) & 0x7fff
	} else { // (ppu.scanline <= 239 || ppu.scanline == 261) && ppu.rendering()
		if ppu.controller(VRAMAddressIncrement) == 32 {
			ppu.incrementY()
		} else {
			ppu.Registers.Address++
		}
	}
}

func (ppu *RP2C02) reloadBackgroundTiles() {
	switch ppu.cycle {
	case 9, 17, 25, 33, 41, 49, 57, 65, 73, 81, 89, 97, 105, 113, 121, 129, 137, 145, 153,
		161, 169, 177, 185, 193, 201, 209, 217, 225, 233, 241, 249, 257, 329, 337:
		ppu.tilesLow = (ppu.tilesLow & 0xff00) | (ppu.tilesLatch & 0x00ff)
		ppu.tilesHigh = (ppu.tilesHigh & 0xff00) | ((ppu.tilesLatch >> 8) & 0x00ff)
		ppu.attributeLatch = ppu.attributeNext
	}
}

func (ppu *RP2C02) shiftBackgroundTiles() {
	if (ppu.cycle >= 2 && ppu.cycle <= 257) || (ppu.cycle >= 322 && ppu.cycle <= 337) {
		ppu.tilesLow <<= 1
		ppu.tilesHigh <<= 1
		ppu.attributes = (ppu.attributes << 2) | uint16(ppu.attributeLatch)
	}
}

func (ppu *RP2C02) fetchSprites() {
	var s *Sprite

	switch ppu.cycle {
	case 263, 271, 279, 287, 295, 303, 311, 319:
		index := uint8((ppu.cycle >> 3) & 0x07)
		sprite := ppu.oam.Sprite(index)

		s = &ppu.sprites[index]

		s.Sprite = sprite
		s.XPosition = ppu.sprite(sprite, XPosition)

		s.TileLow = 0x00
		s.TileHigh = 0x00

		address := ppu.spriteAddress(sprite)
		s.TileLow = ppu.Memory.Fetch(address)
		s.TileHigh = ppu.Memory.Fetch(address | 0x0008)

		if ppu.sprite(sprite, FlipHorizontally) != 0 {
			reverse := func(x uint8) uint8 {
				x = (x&0x55)<<1 | (x&0xaa)>>1
				x = (x&0x33)<<2 | (x&0xcc)>>2
				x = (x&0x0f)<<4 | (x&0xf0)>>4
				return x
			}

			s.TileLow = reverse(s.TileLow)
			s.TileHigh = reverse(s.TileHigh)
		}
	}
}

func (ppu *RP2C02) decrementSprites() {
	var i int
	var s *Sprite

	for i = range ppu.sprites {
		s = &ppu.sprites[i]

		if s.XPosition > 0 {
			s.XPosition--
		}
	}
}

func (ppu *RP2C02) shiftSprites() {
	var i int
	var s *Sprite

	for i = range ppu.sprites {
		s = &ppu.sprites[i]

		if s.XPosition == 0 {
			s.TileLow <<= 1
			s.TileHigh <<= 1
		}
	}
}

func (ppu *RP2C02) rendering() bool {
	return ppu.mask(ShowBackground) || ppu.mask(ShowSprites)
}

func (ppu *RP2C02) openName(address uint16) uint16 {
	//               NNii iiii iiii
	// 0x2000 = 0010 0000 0000 0000
	// 0x2400 = 0010 0100 0000 0000
	// 0x2800 = 0010 1000 0000 0000
	// 0x2c00 = 0010 1100 0000 0000
	return 0x2000 | address&0x0fff
}

func (ppu *RP2C02) fetchName(address uint16) uint16 {
	// 000p NNNN NNNN vvvv
	return ppu.controller(BackgroundPatternAddress) |
		uint16(ppu.Memory.Fetch(address))<<4 | ppu.address(FineYScroll)
}

func (ppu *RP2C02) openAttribute(address uint16) uint16 {
	// 0x23c0 = 0010 0011 1100 0000
	//               NN = 0x0c00
	//                      ii i = 0x0038
	//                          jjj = 0x0007
	return 0x23c0 | (address & 0x0c00) | (address >> 4 & 0x0038) | (address >> 2 & 0x0007)
}

func (ppu *RP2C02) fetchAttribute(address uint16) uint8 {
	// combine 2nd X- and Y-bit of loopy_v to
	// determine which 2-bits of AT byte to use:
	//
	// value = (topleft << 0) | (topright << 2) | (bottomleft << 4) | (bottomright << 6)
	//
	// v: .yyy NNYY YYYX XXXX|
	//    .... .... .... ..X.|
	// v >> 4: .... .>>> >Y..|....
	//         .X. = 000 = 0
	//         Y..   010 = 2
	//               100 = 4
	//               110 = 6
	return (ppu.Memory.Fetch(address) >>
		((ppu.Registers.Address & 0x2) | (ppu.Registers.Address >> 4 & 0x4))) & 0x03
}

func (ppu *RP2C02) spriteAddress(sprite uint32) (address uint16) {
	comparitor := (ppu.scanline - uint16(ppu.sprite(sprite, YPosition)))

	if ppu.sprite(sprite, FlipVertically) != 0 {
		comparitor ^= 0x000f
	}

	switch ppu.controller(SpriteSize) {
	case 8:
		address = ppu.controller(SpritePatternAddress) |
			(uint16(ppu.sprite(sprite, TileNumber)) << 4)
	case 16:
		address = (uint16(ppu.sprite(sprite, TileBank)) << 12) |
			(uint16(ppu.sprite(sprite, TileNumber)) << 5) |
			((comparitor & 0x08) << 1)
	}

	address |= comparitor & 0x07

	return
}

func (ppu *RP2C02) priorityMultiplexer(bgAddress, spriteAddress uint16, spritePriority uint8) (address uint16) {
	bgIndex := bgAddress & 0x0003
	spriteIndex := spriteAddress & 0x0003

	if !ppu.ShowBackground {
		bgIndex = 0
	}

	if !ppu.ShowSprites {
		spriteIndex = 0
	}

	switch bgIndex {
	case 0:
		switch spriteIndex {
		case 0:
			address = 0x3f00
		default:
			address = spriteAddress
		}
	default:
		switch spriteIndex {
		case 0:
			address = bgAddress
		default:
			switch spritePriority {
			case 0:
				address = spriteAddress
			case 1:
				address = bgAddress
			}
		}
	}

	return
}

func (ppu *RP2C02) renderBackground() (bgAddress, bgIndex uint16) {
	if ppu.mask(ShowBackground) && (ppu.mask(ShowBackgroundLeft) || ppu.cycle > 8) {
		scroll := 15 - ppu.Registers.Scroll
		bgIndex = (((ppu.tilesHigh >> scroll) & 0x0001) << 1) | ((ppu.tilesLow >> scroll) & 0x0001)
		bgAttribute := uint16((ppu.attributes>>(14-(2*ppu.Registers.Scroll)))&0x0003) << 2
		bgAddress = uint16(0x3f00 | bgAttribute | bgIndex)
	}

	return
}

func (ppu *RP2C02) renderSprites() (spriteAddress, spriteIndex uint16, spritePriority uint8, spriteUnit int) {
	var i int
	var s *Sprite

	if ppu.mask(ShowSprites) && (ppu.mask(ShowSpritesLeft) || ppu.cycle > 8) {
		for i = range ppu.sprites {
			s = &ppu.sprites[i]

			if s.XPosition > 0 {
				continue
			}

			high := s.TileHigh & 0x80
			low := s.TileLow & 0x80

			if high == 0x00 && low == 0x00 {
				continue
			}

			index := uint16((high >> 6) | (low >> 7))

			if index != 0 {
				spriteIndex = index
				spriteAttribute := uint16(ppu.sprite(s.Sprite, SpritePalette)) << 2
				spriteAddress = uint16(0x3f10 | spriteAttribute | spriteIndex)
				spritePriority = ppu.sprite(s.Sprite, Priority)
				spriteUnit = i
				break
			}
		}
	}

	return
}

func (ppu *RP2C02) renderVisibleScanline() {
	ppu.reloadBackgroundTiles()

	switch ppu.cycle {
	// skipped on BG+odd
	case 0:

	// open NT byte
	case 1, 9, 17, 25, 33, 41, 49, 57, 65, 73, 81, 89, 97, 105, 113, 121, 129, 137,
		145, 153, 161, 169, 177, 185, 193, 201, 209, 217, 225, 233, 241, 249,
		321, 329, 337, 339:
		ppu.addressLine = ppu.openName(ppu.Registers.Address)

	// fetch NT byte
	case 2, 10, 18, 26, 34, 42, 50, 58, 66, 74, 82, 90, 98, 106, 114, 122, 130, 138,
		146, 154, 162, 170, 178, 186, 194, 202, 210, 218, 226, 234, 242, 250,
		322, 330, 338, 340:
		ppu.patternAddress = ppu.fetchName(ppu.addressLine)

	// open AT byte
	case 3, 11, 19, 27, 35, 43, 51, 59, 67, 75, 83, 91, 99, 107, 115, 123, 131, 139,
		147, 155, 163, 171, 179, 187, 195, 203, 211, 219, 227, 235, 243, 251,
		323, 331:
		ppu.addressLine = ppu.openAttribute(ppu.Registers.Address)

	// fetch AT byte
	case 4, 12, 20, 28, 36, 44, 52, 60, 68, 76, 84, 92, 100, 108, 116, 124, 132, 140,
		148, 156, 164, 172, 180, 188, 196, 204, 212, 220, 228, 236, 244, 252,
		324, 332:
		ppu.attributeNext = ppu.fetchAttribute(ppu.addressLine)

	// open low BG tile byte (color bit 0)
	case 5, 13, 21, 29, 37, 45, 53, 61, 69, 77, 85, 93, 101, 109, 117, 125, 133, 141,
		149, 157, 165, 173, 181, 189, 197, 205, 213, 221, 229, 237, 245, 253,
		325, 333:
		// Fetch color bit 0 for next 8 dots
		ppu.addressLine = ppu.patternAddress

	// fetch BG tile byte (color bit 0)
	case 6, 14, 22, 30, 38, 46, 54, 62, 70, 78, 86, 94, 102, 110, 118, 126, 134, 142,
		150, 158, 166, 174, 182, 190, 198, 206, 214, 222, 230, 238, 246, 254,
		326, 334:
		// Fetch color bit 0 for next 8 dots
		ppu.tilesLatch = (ppu.tilesLatch & 0xff00) | uint16(ppu.Memory.Fetch(ppu.addressLine))

	// open high BG tile byte (color bit 1)
	case 7, 15, 23, 31, 39, 47, 55, 63, 71, 79, 87, 95, 103, 111, 119, 127, 135, 143,
		151, 159, 167, 175, 183, 191, 199, 207, 215, 223, 231, 239, 247, 255,
		327, 335:
		// Fetch color bit 1 for next 8 dots
		ppu.addressLine = ppu.patternAddress | 0x0008

	// fetch high BG tile byte (color bit 1)
	case 8, 16, 24, 32, 40, 48, 56, 64, 72, 80, 88, 96, 104, 112, 120, 128, 136, 144,
		152, 160, 168, 176, 184, 192, 200, 208, 216, 224, 232, 240, 248, 256,
		328, 336:
		// Fetch color bit 1 for next 8 dots
		ppu.tilesLatch = (ppu.tilesLatch & 0x00ff) | uint16(ppu.Memory.Fetch(ppu.addressLine))<<8

		// inc hori(v)
		ppu.incrementX()

		// inc vert(v)
		if ppu.cycle == 256 {
			ppu.incrementY()
		}

	// hori(v) = hori(t)
	case 257:
		ppu.transferX()

	// vert(v) = vert(t)
	case 280, 281, 282, 283, 284, 285, 286, 287, 288, 289, 290, 291, 292,
		293, 294, 295, 296, 297, 298, 299, 300, 301, 302, 303, 304:
		if ppu.scanline == 261 {
			ppu.transferY()
		}
	}

	if ppu.cycle >= 1 && ppu.cycle <= 256 {
		address := uint16(0)

		bgAddress, bgIndex := ppu.renderBackground()

		ppu.decrementSprites()
		spriteAddress, spriteIndex, spritePriority, spriteUnit := ppu.renderSprites()
		ppu.shiftSprites()

		address = ppu.priorityMultiplexer(bgAddress, spriteAddress, spritePriority)

		if spriteUnit == 0 && ppu.oam.SpriteZeroInBuffer && bgIndex != 0 && spriteIndex != 0 &&
			(ppu.cycle > 8 || (ppu.mask(ShowBackgroundLeft) && ppu.mask(ShowSpritesLeft))) &&
			ppu.cycle < 255 && (ppu.mask(ShowBackground) && ppu.mask(ShowSprites)) {
			ppu.Registers.Status |= uint8(Sprite0Hit)
		}

		color := ppu.Memory.Fetch(address) & 0x3f

		if ppu.scanline >= 0 && ppu.scanline <= 239 {
			ppu.colors[(ppu.scanline*256)+(ppu.cycle-1)] = color
		}

		if ppu.oam.SpriteEvaluation(ppu.scanline, ppu.cycle, ppu.controller(SpriteSize)) {
			ppu.Registers.Status |= uint8(SpriteOverflow)
		}
	}

	ppu.shiftBackgroundTiles()
	ppu.fetchSprites()

	return
}

func (ppu *RP2C02) Execute() {
	switch {
	// visible scanlines (0-239), pre-render scanline (261)
	case (ppu.scanline >= 0 && ppu.scanline <= 239) || ppu.scanline == 261:
		if ppu.cycle == 1 && ppu.scanline == 261 {
			ppu.Registers.Status &^= uint8(VBlankStarted | Sprite0Hit | SpriteOverflow)
		}

		if ppu.rendering() {
			ppu.renderVisibleScanline()
		}
	// post-render scanline (240), vertical blanking scanlines (241-260)
	default:
		if ppu.scanline == 241 && ppu.cycle == 1 {
			ppu.Registers.Status |= uint8(VBlankStarted)

			if ppu.status(VBlankStarted) &&
				ppu.controller(NMIOnVBlank) != 0 &&
				ppu.Interrupt != nil {
				ppu.Interrupt(true)
			}
		}
	}

	if ppu.cycle++; ppu.cycle == CYCLES_PER_SCANLINE {
		ppu.cycle = 0

		if ppu.scanline++; ppu.scanline == NUM_SCANLINES {
			if ppu.rendering() {
				ppu.Output <- ppu.colors
				<-ppu.Output
			}

			ppu.scanline = 0
			ppu.frame++
		}
	}
}

func (ppu *RP2C02) SavePatternTables() (left, right *image.RGBA) {
	left = image.NewRGBA(image.Rect(0, 0, 128, 128))
	right = image.NewRGBA(image.Rect(0, 0, 128, 128))

	colors := [4]color.RGBA{
		color.RGBA{0, 0, 0, 255},
		color.RGBA{203, 79, 15, 255},
		color.RGBA{255, 155, 59, 255},
		color.RGBA{255, 231, 163, 255},
	}

	x_base := 0
	y_base := 0

	ptimg := left

	for address := uint16(0x0000); address <= 0x1fff; address += 0x0010 {
		if address < 0x1000 {
			ptimg = left
		} else {
			ptimg = right
		}

		for row := uint16(0); row <= 7; row++ {
			low := ppu.Memory.Fetch(address + row)
			high := ppu.Memory.Fetch(address + row + 8)

			for i := int16(7); i >= 0; i-- {
				b := ((low >> uint16(i)) & 0x0001) | (((high >> uint16(i)) & 0x0001) << 1)
				ptimg.Set(x_base+(8-int(i+1)), y_base+int(row), colors[b])
			}
		}

		x_base += 8

		if x_base == 128 {
			x_base = 0
			y_base = (y_base + 8) % 128
		}
	}

	fo, _ := os.Create(fmt.Sprintf("left.jpg"))
	w := bufio.NewWriter(fo)
	jpeg.Encode(w, left, &jpeg.Options{Quality: 100})

	fo, _ = os.Create(fmt.Sprintf("right.jpg"))
	w = bufio.NewWriter(fo)
	jpeg.Encode(w, right, &jpeg.Options{Quality: 100})

	return
}

func (ppu *RP2C02) Run() {
	for {
		ppu.Execute()
	}
}
