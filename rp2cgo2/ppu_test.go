package rp2cgo2

import "testing"

func TestController(t *testing.T) {
	ppu := NewRP2C02(nil)

	ppu.Registers.Controller = 0x00
	ppu.Store(0x2000, 0xff)

	if ppu.Registers.Controller != 0xff {
		t.Errorf("Register is %02X not 0xff\n", ppu.Registers.Controller)
	}

	// BaseNametableAddress
	ppu.Registers.Controller = 0xff - 0x03

	if ppu.controller(BaseNametableAddress) != 0x2000 {
		t.Errorf("BaseNametableAddress is %04X not 0x2000", ppu.controller(BaseNametableAddress))
	}

	ppu.Registers.Controller = 0xff - 0x02

	if ppu.controller(BaseNametableAddress) != 0x2400 {
		t.Errorf("BaseNametableAddress is %04X not 0x2400", ppu.controller(BaseNametableAddress))
	}

	ppu.Registers.Controller = 0xff - 0x01

	if ppu.controller(BaseNametableAddress) != 0x2800 {
		t.Errorf("BaseNametableAddress is %04X not 0x2800", ppu.controller(BaseNametableAddress))
	}

	ppu.Registers.Controller = 0xff

	if ppu.controller(BaseNametableAddress) != 0x2c00 {
		t.Errorf("BaseNametableAddress is %04X not 0x2c00", ppu.controller(BaseNametableAddress))
	}

	// VRAMAddressIncrement
	ppu.Registers.Controller = ^uint8(VRAMAddressIncrement)

	if ppu.controller(VRAMAddressIncrement) != 1 {
		t.Error("VRAMAddressIncrement is not 1")
	}

	ppu.Registers.Controller = uint8(VRAMAddressIncrement)

	if ppu.controller(VRAMAddressIncrement) != 32 {
		t.Error("VRAMAddressIncrement is not 32")
	}

	// SpritePatternAddress
	ppu.Registers.Controller = ^uint8(SpritePatternAddress)

	if ppu.controller(SpritePatternAddress) != 0x0000 {
		t.Error("SpritePatternAddress is not 0x0000")
	}

	ppu.Registers.Controller = uint8(SpritePatternAddress)

	if ppu.controller(SpritePatternAddress) != 0x1000 {
		t.Error("SpritePatternAddress is not 0x1000")
	}

	// BackgroundPatternAddress
	ppu.Registers.Controller = ^uint8(BackgroundPatternAddress)

	if ppu.controller(BackgroundPatternAddress) != 0x0000 {
		t.Error("BackgroundPatternAddress is not 0x0000")
	}

	ppu.Registers.Controller = uint8(BackgroundPatternAddress)

	if ppu.controller(BackgroundPatternAddress) != 0x1000 {
		t.Error("BackgroundPatternAddress is not 0x1000")
	}

	// SpriteSize
	ppu.Registers.Controller = ^uint8(SpriteSize)

	if ppu.controller(SpriteSize) != 8 {
		t.Error("SpriteSize is not 8")
	}

	ppu.Registers.Controller = uint8(SpriteSize)

	if ppu.controller(SpriteSize) != 16 {
		t.Error("SpriteSize is not 16")
	}

	// NMIOnVBlank
	ppu.Registers.Controller = ^uint8(NMIOnVBlank)

	if ppu.controller(NMIOnVBlank) != 0 {
		t.Error("NMIOnVBlank is not 0")
	}

	ppu.Registers.Controller = uint8(NMIOnVBlank)

	if ppu.controller(NMIOnVBlank) != 1 {
		t.Error("NMIOnVBlank is not 1")
	}
}

func TestMask(t *testing.T) {
	ppu := NewRP2C02(nil)

	ppu.Registers.Mask = 0x00
	value := uint8(0xff)
	ppu.Store(0x2001, value)

	if ppu.Registers.Mask != value {
		t.Errorf("Register is %02X not %02X\n", ppu.Registers.Mask, value)
	}

	for _, m := range []MaskFlag{
		Grayscale, ShowBackgroundLeft, ShowSpritesLeft, ShowBackground,
		ShowSprites, IntensifyReds, IntensifyGreens, IntensifyBlues,
	} {
		ppu.Registers.Mask = uint8(m)

		if !ppu.mask(m) {
			t.Errorf("Mask %v is not true", m)
		}
	}
}

func TestStatus(t *testing.T) {
	ppu := NewRP2C02(nil)

	ppu.Registers.Status = 0xff
	ppu.latch = true
	ppu.latchValue = 0x00

	value := ppu.Fetch(0x2002)

	if value != 0xe0 {
		t.Errorf("Memory is %02X not 0xe0\n", value)
	}

	if ppu.Registers.Status != 0x7f {
		t.Error("VBlankStarted flag is set")
	}

	if ppu.latch {
		t.Error("Latch is true")
	}

	for _, s := range []StatusFlag{
		SpriteOverflow,
		Sprite0Hit,
		VBlankStarted,
	} {
		ppu.Registers.Status = uint8(s)

		if !ppu.status(s) {
			t.Errorf("Status %v is not true", s)
		}
	}
}

func TestAddress(t *testing.T) {
	ppu := NewRP2C02(nil)

	ppu.Registers.Address = 0x00
	ppu.Store(0x2006, 0xff)
	ppu.Store(0x2006, 0xff)

	if ppu.Registers.Address != 0x3fff {
		t.Errorf("Register is %04X not 0x3fff\n", ppu.Registers.Address)
	}

	// CoarseXScroll
	ppu.Registers.Address = 0x0000

	if ppu.address(CoarseXScroll) != 0x0000 {
		t.Errorf("CoarseXScroll is %04X not 0x0000", ppu.address(CoarseXScroll))
	}

	ppu.Registers.Address = 0xffff

	if ppu.address(CoarseXScroll) != 0x001f {
		t.Errorf("CoarseXScroll is %04X not 0x001f", ppu.address(CoarseXScroll))
	}

	// CoarseYScroll
	ppu.Registers.Address = 0x0000

	if ppu.address(CoarseYScroll) != 0x0000 {
		t.Errorf("CoarseYScroll is %04X not 0x0000", ppu.address(CoarseYScroll))
	}

	ppu.Registers.Address = 0xffff

	if ppu.address(CoarseYScroll) != 0x001f {
		t.Errorf("CoarseYScroll is %04X not 0x001f", ppu.address(CoarseYScroll))
	}

	// NametableSelect
	ppu.Registers.Address = 0x0000

	if ppu.address(NametableSelect) != 0x0000 {
		t.Errorf("NametableSelect is %04X not 0x0000", ppu.address(NametableSelect))
	}

	ppu.Registers.Address = 0xffff

	if ppu.address(NametableSelect) != 0x0003 {
		t.Errorf("NametableSelect is %04X not 0x0003", ppu.address(NametableSelect))
	}

	// FineYScroll
	ppu.Registers.Address = 0x0000

	if ppu.address(FineYScroll) != 0x0000 {
		t.Errorf("FineYScroll is %04X not 0x0000", ppu.address(FineYScroll))
	}

	ppu.Registers.Address = 0xffff

	if ppu.address(FineYScroll) != 0x0007 {
		t.Errorf("FineYScroll is %04X not 0x0007", ppu.address(FineYScroll))
	}
}

func TestSprite(t *testing.T) {
	ppu := NewRP2C02(nil)

	sprite := uint32(0)

	// YPosition
	sprite = 0x00000000

	if ppu.sprite(sprite, YPosition) != 0x00 {
		t.Error("YPosition is not 0x00")
	}

	sprite = 0xffffffff

	if ppu.sprite(sprite, YPosition) != 0xff {
		t.Error("YPosition is not 0xff")
	}

	// TileBank
	sprite = 0x00000000

	if ppu.sprite(sprite, TileBank) != 0x00 {
		t.Error("TileBank is not 0x00")
	}

	sprite = 0xffffffff

	if ppu.sprite(sprite, TileBank) != 0x01 {
		t.Error("TileBank is not 0x01")
	}

	// TileNumber
	sprite = 0x00000000

	if ppu.sprite(sprite, TileNumber) != 0x00 {
		t.Error("TileNumber is not 0x00")
	}

	sprite = 0xffffffff

	if ppu.sprite(sprite, TileNumber) != 0xff {
		t.Error("TileNumber is not 0xff")
	}

	// SpritePalette
	sprite = 0x00000000

	if ppu.sprite(sprite, SpritePalette) != 0x00 {
		t.Error("SpritePalette is not 0x00")
	}

	sprite = 0xffffffff

	if ppu.sprite(sprite, SpritePalette) != 0x03 {
		t.Error("SpritePalette is not 0x03")
	}

	// Priority
	sprite = 0x00000000

	if ppu.sprite(sprite, Priority) != 0x00 {
		t.Error("Priority is not 0x00")
	}

	sprite = 0xffffffff

	if ppu.sprite(sprite, Priority) != 0x01 {
		t.Error("Priority is not 0x01")
	}

	// FlipHorizontally
	sprite = 0x00000000

	if ppu.sprite(sprite, FlipHorizontally) != 0x00 {
		t.Error("FlipHorizontally is not 0x00")
	}

	sprite = 0xffffffff

	if ppu.sprite(sprite, FlipHorizontally) != 0x01 {
		t.Error("FlipHorizontally is not 0x01")
	}

	// FlipVertically
	sprite = 0x00000000

	if ppu.sprite(sprite, FlipVertically) != 0x00 {
		t.Error("FlipVertically is not 0x00")
	}

	sprite = 0xffffffff

	if ppu.sprite(sprite, FlipVertically) != 0x01 {
		t.Error("FlipVertically is not 0x01")
	}

	// XPosition
	sprite = 0x00000000

	if ppu.sprite(sprite, XPosition) != 0x00 {
		t.Error("XPosition is not 0x00")
	}

	sprite = 0xffffffff

	if ppu.sprite(sprite, XPosition) != 0xff {
		t.Error("XPosition is not 0xff")
	}
}

func TestOAMAddress(t *testing.T) {
	ppu := NewRP2C02(nil)

	ppu.Registers.OAMAddress = 0x00

	ppu.Store(0x2003, 0xff)

	if ppu.Registers.OAMAddress != 0xff {
		t.Errorf("Register is %02X not 0xff\n", ppu.Registers.OAMAddress)
	}
}

func TestOAMData(t *testing.T) {
	ppu := NewRP2C02(nil)

	ppu.Registers.OAMAddress = 0x00

	for i := uint16(0x0000); i <= 0x00ff; i++ {
		ppu.Store(0x2004, uint8(i))
	}

	for i := uint16(0x0000); i <= 0x00ff; i++ {
		if ppu.oam.Fetch(uint16(i)) != uint8(i) {
			t.Errorf("Memory is %02X not %02X\n", ppu.oam.Fetch(uint16(i)), uint8(i))
		}
	}
}

func TestPaletteMirroring(t *testing.T) {
	ppu := NewRP2C02(nil)

	// Mirrored palette
	for _, i := range []uint16{0x3f10, 0x3f14, 0x3f18, 0x3f1c} {
		ppu.Memory.Store(i-0x0010, 0xff)

		if ppu.Memory.Fetch(i) != 0xff {
			t.Error("Memory is not 0xff")
		}

		ppu.Memory.Store(i-0x0010, 0x00)
		ppu.Memory.Store(i, 0xff)

		if ppu.Memory.Fetch(i-0x0010) != 0xff {
			t.Error("Memory is not 0xff")
		}
	}

	for i := uint16(0x3f20); i <= 0x3fff; i++ {
		ppu.Memory.Store(i-0x0020, 0xff)

		if ppu.Memory.Fetch(i) != 0xff {
			t.Error("Memory is not 0xff")
		}

		ppu.Memory.Store(i-0x0020, 0x00)
		ppu.Memory.Store(i, 0xff)

		if ppu.Memory.Fetch(i-0x0020) != 0xff {
			t.Error("Memory is not 0xff")
		}
	}
}

func TestAddressFetchStore(t *testing.T) {
	ppu := NewRP2C02(nil)

	ppu.Registers.Address = 0x0000
	ppu.Fetch(0x2002)

	ppu.Store(0x2006, 0x3f)
	ppu.Store(0x2006, 0xff)

	if ppu.Registers.Address != 0x3fff {
		t.Error("Register is not 0x3fff")
	}

	ppu.Fetch(0x2002)

	ppu.Store(0x2006, 0x01)
	ppu.Store(0x2006, 0x01)

	if ppu.Registers.Address != 0x0101 {
		t.Error("Register is not 0x0101")
	}
}

func TestDataIncrement1(t *testing.T) {
	ppu := NewRP2C02(nil)

	ppu.Registers.Address = 0x0000
	ppu.Fetch(0x2002)

	ppu.Store(0x2006, 0x01)
	ppu.Store(0x2006, 0x00)

	if ppu.Registers.Address != 0x0100 {
		t.Error("Register is not 0x0100")
	}

	ppu.Store(0x2007, 0xff)
	ppu.Store(0x2007, 0xff)
	ppu.Store(0x2007, 0xff)

	if ppu.Memory.Fetch(0x0100) != 0xff {
		t.Error("Memory is not 0xff")
	}

	if ppu.Memory.Fetch(0x0101) != 0xff {
		t.Error("Memory is not 0xff")
	}

	if ppu.Memory.Fetch(0x0102) != 0xff {
		t.Error("Memory is not 0xff")
	}

	if ppu.Registers.Address != 0x0103 {
		t.Error("Register is not 0x0103")
	}

	ppu.Memory.Store(0x0103, 0xff)
	ppu.Fetch(0x2007)

	if ppu.Fetch(0x2007) != 0xff {
		t.Error("Memory is not 0xff")
	}

	if ppu.Registers.Address != 0x0105 {
		t.Error("Register is not 0x0105")
	}

	ppu.Registers.Address = 0x0000
	ppu.Fetch(0x2002)

	ppu.Store(0x2006, 0x3f)
	ppu.Store(0x2006, 0x00)

	if ppu.Registers.Address != 0x3f00 {
		t.Error("Register is not 0x3f00")
	}

	ppu.Memory.Store(0x3f00, 0xff)
	ppu.Memory.Store(0x3f01, 0xff)
	ppu.Memory.Store(0x3f02, 0xff)

	if ppu.Fetch(0x2007) != 0xff {
		t.Error("Memory is not 0xff")
	}

	if ppu.Registers.Address != 0x3f01 {
		t.Error("Register is not 0x3f01")
	}
}

func TestDataIncrement32(t *testing.T) {
	ppu := NewRP2C02(nil)

	ppu.Store(0x2000, 0x04)

	ppu.Registers.Address = 0x0000
	ppu.Fetch(0x2002)

	ppu.Store(0x2006, 0x01)
	ppu.Store(0x2006, 0x00)

	if ppu.Registers.Address != 0x0100 {
		t.Error("Register is not 0x0100")
	}

	ppu.Store(0x2007, 0xff)
	ppu.Store(0x2007, 0xff)
	ppu.Store(0x2007, 0xff)

	if ppu.Memory.Fetch(0x0100) != 0xff {
		t.Error("Memory is not 0xff")
	}

	if ppu.Memory.Fetch(0x0120) != 0xff {
		t.Error("Memory is not 0xff")
	}

	if ppu.Memory.Fetch(0x0140) != 0xff {
		t.Error("Memory is not 0xff")
	}
}

func TestIncrementX(t *testing.T) {
	ppu := NewRP2C02(nil)

	ppu.Registers.Address = 0x0000
	ppu.incrementX()

	if ppu.Registers.Address != 0x0001 {
		t.Error("Register is not 0x0001")
	}

	ppu.Registers.Address = 0x001e
	ppu.incrementX()

	if ppu.Registers.Address != 0x001f {
		t.Error("Register is not 0x001f")
	}

	ppu.Registers.Address = 0x001f
	ppu.incrementX()

	if ppu.Registers.Address != 0x0400 {
		t.Error("Register is not 0x0400")
	}

	ppu.Registers.Address = 0x041f
	ppu.incrementX()

	if ppu.Registers.Address != 0x0000 {
		t.Error("Register is not 0x0000")
	}
}

func TestTransferX(t *testing.T) {
	ppu := NewRP2C02(nil)

	ppu.Registers.Address = 0x7be0
	ppu.latchAddress = 0x041f

	ppu.transferX()

	if ppu.Registers.Address != 0x7fff {
		t.Error("Registers is not 0x7fff")
	}

	ppu.Registers.Address = 0x7be0
	ppu.latchAddress = 0xffff

	ppu.transferX()

	if ppu.Registers.Address != 0x7fff {
		t.Error("Registers is not 0x7fff")
	}

	ppu.Registers.Address = 0xffff
	ppu.latchAddress = 0x0000

	ppu.transferX()

	if ppu.Registers.Address != 0x7be0 {
		t.Error("Registers is not 0x7be0")
	}

	ppu.Registers.Address = 0x0000
	ppu.latchAddress = 0xffff

	ppu.transferX()

	if ppu.Registers.Address != 0x041f {
		t.Error("Registers is not 0x041f")
	}
}

func TestIncrementY(t *testing.T) {
	ppu := NewRP2C02(nil)

	ppu.Registers.Address = 0x0000
	ppu.incrementY()

	if ppu.Registers.Address != 0x1000 {
		t.Error("Register is not 0x1000")
	}

	ppu.Registers.Address = 0x1000
	ppu.incrementY()

	if ppu.Registers.Address != 0x2000 {
		t.Error("Register is not 0x2000")
	}

	ppu.Registers.Address = 0x6000
	ppu.incrementY()

	if ppu.Registers.Address != 0x7000 {
		t.Error("Register is not 0x7000")
	}

	ppu.Registers.Address = 0x7000
	ppu.incrementY()

	if ppu.Registers.Address != 0x0020 {
		t.Error("Register is not 0x0020")
	}

	ppu.Registers.Address = 0x73d0
	ppu.incrementY()

	if ppu.Registers.Address != 0x03f0 {
		t.Error("Register is not 0x03f0")
	}

	ppu.Registers.Address = 0x7fa0
	ppu.incrementY()

	if ppu.Registers.Address != 0x0400 {
		t.Errorf("Register is %04x not 0x0400\n", ppu.Registers.Address)
	}

	ppu.Registers.Address = 0x73a0
	ppu.incrementY()

	if ppu.Registers.Address != 0x0800 {
		t.Errorf("Register is %04x not 0x0800\n", ppu.Registers.Address)
	}

	ppu.Registers.Address = 0x73e1
	ppu.incrementY()

	if ppu.Registers.Address != 0x0001 {
		t.Error("Register is not 0x0001")
	}
}

func TestTransferY(t *testing.T) {
	ppu := NewRP2C02(nil)

	ppu.Registers.Address = 0x041f
	ppu.latchAddress = 0x7be0

	ppu.transferY()

	if ppu.Registers.Address != 0x7fff {
		t.Error("Registers is not 0x7fff")
	}

	ppu.Registers.Address = 0x041f
	ppu.latchAddress = 0xffff

	ppu.transferY()

	if ppu.Registers.Address != 0x7fff {
		t.Error("Registers is not 0x7fff")
	}

	ppu.Registers.Address = 0xffff
	ppu.latchAddress = 0x0000

	ppu.transferY()

	if ppu.Registers.Address != 0x041f {
		t.Error("Registers is not 0x041f")
	}

	ppu.Registers.Address = 0x0000
	ppu.latchAddress = 0xffff

	ppu.transferY()

	if ppu.Registers.Address != 0x7be0 {
		t.Error("Registers is not 0x7be0")
	}
}
