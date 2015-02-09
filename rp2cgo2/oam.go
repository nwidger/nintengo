package rp2cgo2

import "github.com/nwidger/nintengo/m65go2"

type CycleFunc func(oam *OAM, scanline uint16, cycle uint16, size uint16) (spriteOverflow bool)

const (
	ClearBuffer int = iota
	CopyYPosition
	CopyIndex
	CopyAttributes
	CopyXPosition
	EvaluateYPosition
	EvaluateIndex
	EvaluateAttributes
	EvaluateXPosition
	FailCopyYPosition
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
	cycleFuncs := []CycleFunc{
		ClearBuffer:        clearBuffer,
		CopyYPosition:      copyYPosition,
		CopyIndex:          copyIndex,
		CopyAttributes:     copyAttributes,
		CopyXPosition:      copyXPosition,
		EvaluateYPosition:  evaluateYPosition,
		EvaluateIndex:      evaluateIndex,
		EvaluateAttributes: evaluateAttributes,
		EvaluateXPosition:  evaluateXPosition,
		FailCopyYPosition:  failCopyYPosition,
	}

	return &OAM{
		BasicMemory:        m65go2.NewBasicMemory(256),
		Buffer:             m65go2.NewBasicMemory(32),
		SpriteZeroInBuffer: false,
		cycleFuncs:         cycleFuncs,
		WriteCycle:         FailCopyYPosition,
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
			oam.WriteCycle = ClearBuffer
		case 65:
			oam.Address = 0
			oam.Latch = 0xff
			oam.Index = 0

			oam.DisableReads = false
			oam.WriteCycle = CopyYPosition
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
		oam.WriteCycle = CopyIndex
		oam.incrementAddress(0x00ff)
	} else {
		oam.Address += 4

		if oam.Address == 0x0100 {
			oam.WriteCycle = FailCopyYPosition
		}
	}

	return
}

func copyIndex(oam *OAM, scanline uint16, cycle uint16, size uint16) (spriteOverflow bool) {
	oam.Buffer.Store(oam.Index+1, oam.Latch)
	oam.WriteCycle = CopyAttributes
	oam.incrementAddress(0x00ff)
	return
}

func copyAttributes(oam *OAM, scanline uint16, cycle uint16, size uint16) (spriteOverflow bool) {
	oam.Buffer.Store(oam.Index+2, oam.Latch)
	oam.WriteCycle = CopyXPosition
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
		oam.WriteCycle = FailCopyYPosition
	case oam.Index < 32:
		oam.WriteCycle = CopyYPosition
	default:
		oam.Buffer.DisableWrites = true
		oam.Address &= 0x00fc
		oam.WriteCycle = EvaluateYPosition
	}

	return
}

func evaluateYPosition(oam *OAM, scanline uint16, cycle uint16, size uint16) (spriteOverflow bool) {
	if scanline-uint16(uint32(oam.Latch)) < size {
		spriteOverflow = true
		oam.Address = (oam.Address + 1) & 0x00ff
		oam.WriteCycle = EvaluateIndex
	} else {
		oam.Address = ((oam.Address + 4) & 0x00fc) + ((oam.Address + 1) & 0x0003)

		if oam.Address <= 0x0005 {
			oam.Address &= 0x00fc
			oam.WriteCycle = FailCopyYPosition
		}
	}

	return
}

func evaluateIndex(oam *OAM, scanline uint16, cycle uint16, size uint16) (spriteOverflow bool) {
	oam.Address = (oam.Address + 1) & 0x00ff
	oam.WriteCycle = EvaluateAttributes
	return
}

func evaluateAttributes(oam *OAM, scanline uint16, cycle uint16, size uint16) (spriteOverflow bool) {
	oam.Address = (oam.Address + 1) & 0x00ff
	oam.WriteCycle = EvaluateXPosition
	return
}

func evaluateXPosition(oam *OAM, scanline uint16, cycle uint16, size uint16) (spriteOverflow bool) {
	oam.Address = (oam.Address + 1) & 0x00ff

	if (oam.Address & 0x0003) == 0x0003 {
		oam.incrementAddress(0x00ff)
	}

	oam.Address &= 0x00fc
	oam.WriteCycle = FailCopyYPosition

	return
}

func failCopyYPosition(oam *OAM, scanline uint16, cycle uint16, size uint16) (spriteOverflow bool) {
	oam.Address = (oam.Address + 4) & 0x00ff

	return
}
