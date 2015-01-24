package rp2cgo2

import "github.com/kaicheng/nintengo/m65go2"

type CycleFunc func(oam *OAM, scanline uint16, cycle uint16, size uint16) (spriteOverflow bool)

const (
	CLEAR_BUFFER int = iota
	COPY_Y_POSITION
	COPY_INDEX
	COPY_ATTRIBUTES
	COPY_X_POSITION
	EVALUATE_Y_POSITION
	EVALUATE_INDEX
	EVALUATE_ATTRIBUTES
	EVALUATE_X_POSITION
	FAIL_COPY_Y_POSITION
)

type OAM struct {
	*m65go2.BasicMemory
	Address            uint16
	Latch              uint8
	Buffer             *m65go2.BasicMemory
	SpriteZeroInBuffer bool
	Index              uint16
	cycleFuncs         []CycleFunc
	WriteCycle         int
}

func NewOAM() *OAM {
	cycleFuncs := make([]CycleFunc, 11)

	cycleFuncs[CLEAR_BUFFER] = clearBuffer
	cycleFuncs[COPY_Y_POSITION] = copyYPosition
	cycleFuncs[COPY_INDEX] = copyIndex
	cycleFuncs[COPY_ATTRIBUTES] = copyAttributes
	cycleFuncs[COPY_X_POSITION] = copyXPosition
	cycleFuncs[EVALUATE_Y_POSITION] = evaluateYPosition
	cycleFuncs[EVALUATE_INDEX] = evaluateIndex
	cycleFuncs[EVALUATE_ATTRIBUTES] = evaluateAttributes
	cycleFuncs[EVALUATE_X_POSITION] = evaluateXPosition
	cycleFuncs[FAIL_COPY_Y_POSITION] = failCopyYPosition

	return &OAM{
		BasicMemory:        m65go2.NewBasicMemory(256),
		Buffer:             m65go2.NewBasicMemory(32),
		SpriteZeroInBuffer: false,
		cycleFuncs:         cycleFuncs,
		WriteCycle:         FAIL_COPY_Y_POSITION,
	}
}

func (oam *OAM) Sprite(index uint8) uint32 {
	address := uint16(index) << 2

	return (uint32(oam.Buffer.Fetch(address))<<24 |
		uint32(oam.Buffer.Fetch(address+1))<<16 |
		uint32(oam.Buffer.Fetch(address+2))<<8 |
		uint32(oam.Buffer.Fetch(address+3)))
}

func (oam *OAM) SpriteEvaluation(scanline uint16, cycle uint16, size uint16) (spriteOverflow bool) {
	if scanline != 261 {
		switch cycle {
		case 1:
			oam.Address = 0
			oam.Latch = 0xff
			oam.Index = 0
			oam.SpriteZeroInBuffer = false

			oam.Buffer.DisableWrites = false
			oam.DisableReads = true
			oam.WriteCycle = CLEAR_BUFFER
		case 65:
			oam.Address = 0
			oam.Latch = 0xff
			oam.Index = 0

			oam.DisableReads = false
			oam.WriteCycle = COPY_Y_POSITION
		}

		switch cycle & 0x1 {
		case 1: // odd cycle
			oam.fetchAddress(scanline, cycle, size)
		case 0: // even cycle
			spriteOverflow = oam.cycleFuncs[oam.WriteCycle](oam, scanline, cycle, size)
		}
	}

	return
}

func (oam *OAM) incrementAddress(mask uint16) uint16 {
	oam.Address = (oam.Address + 1) & mask
	return oam.Address
}

func (oam *OAM) fetchAddress(scanline uint16, cycle uint16, size uint16) {
	if oam.Address < 0x0100 {
		oam.Latch = oam.Fetch(oam.Address)
	}
}

func clearBuffer(oam *OAM, scanline uint16, cycle uint16, size uint16) (spriteOverflow bool) {
	oam.Buffer.Store(oam.Address, oam.Latch)
	oam.incrementAddress(0x001f)
	return
}

func copyYPosition(oam *OAM, scanline uint16, cycle uint16, size uint16) (spriteOverflow bool) {
	if scanline-uint16(oam.Latch) < size {
		oam.Buffer.Store(oam.Index+0, oam.Latch)
		oam.WriteCycle = COPY_INDEX
		oam.incrementAddress(0x00ff)
	} else {
		oam.Address += 4

		if oam.Address == 0x0100 {
			oam.WriteCycle = FAIL_COPY_Y_POSITION
		}
	}

	return
}

func copyIndex(oam *OAM, scanline uint16, cycle uint16, size uint16) (spriteOverflow bool) {
	oam.Buffer.Store(oam.Index+1, oam.Latch)
	oam.WriteCycle = COPY_ATTRIBUTES
	oam.incrementAddress(0x00ff)
	return
}

func copyAttributes(oam *OAM, scanline uint16, cycle uint16, size uint16) (spriteOverflow bool) {
	oam.Buffer.Store(oam.Index+2, oam.Latch)
	oam.WriteCycle = COPY_X_POSITION
	oam.incrementAddress(0x00ff)
	return
}

func copyXPosition(oam *OAM, scanline uint16, cycle uint16, size uint16) (spriteOverflow bool) {
	oam.Buffer.Store(oam.Index+3, oam.Latch)

	if oam.Index == 0 {
		oam.SpriteZeroInBuffer = true
	}

	oam.Index += 4
	oam.incrementAddress(0x00ff)

	switch {
	case oam.Address == 0x0100:
		oam.WriteCycle = FAIL_COPY_Y_POSITION
	case oam.Index < 32:
		oam.WriteCycle = COPY_Y_POSITION
	default:
		oam.Buffer.DisableWrites = true
		oam.Address &= 0x00fc
		oam.WriteCycle = EVALUATE_Y_POSITION
	}

	return
}

func evaluateYPosition(oam *OAM, scanline uint16, cycle uint16, size uint16) (spriteOverflow bool) {
	if scanline-uint16(uint32(oam.Latch)) < size {
		spriteOverflow = true
		oam.Address = (oam.Address + 1) & 0x00ff
		oam.WriteCycle = EVALUATE_INDEX
	} else {
		oam.Address = ((oam.Address + 4) & 0x00fc) + ((oam.Address + 1) & 0x0003)

		if oam.Address <= 0x0005 {
			oam.Address &= 0x00fc
			oam.WriteCycle = FAIL_COPY_Y_POSITION
		}
	}

	return
}

func evaluateIndex(oam *OAM, scanline uint16, cycle uint16, size uint16) (spriteOverflow bool) {
	oam.Address = (oam.Address + 1) & 0x00ff
	oam.WriteCycle = EVALUATE_ATTRIBUTES
	return
}

func evaluateAttributes(oam *OAM, scanline uint16, cycle uint16, size uint16) (spriteOverflow bool) {
	oam.Address = (oam.Address + 1) & 0x00ff
	oam.WriteCycle = EVALUATE_X_POSITION
	return
}

func evaluateXPosition(oam *OAM, scanline uint16, cycle uint16, size uint16) (spriteOverflow bool) {
	oam.Address = (oam.Address + 1) & 0x00ff

	if (oam.Address & 0x0003) == 0x0003 {
		oam.incrementAddress(0x00ff)
	}

	oam.Address &= 0x00fc
	oam.WriteCycle = FAIL_COPY_Y_POSITION

	return
}

func failCopyYPosition(oam *OAM, scanline uint16, cycle uint16, size uint16) (spriteOverflow bool) {
	oam.Address = (oam.Address + 4) & 0x00ff

	return
}
