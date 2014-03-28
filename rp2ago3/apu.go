package rp2ago3

type PulseChannel [4]uint8
type TriangleChannel [3]uint8
type NoiseChannel [3]uint8
type DMCChannel [4]uint8
type Control uint8
type Status uint8
type FrameCounter uint8

type Registers struct {
	Pulse1       PulseChannel
	Pulse2       PulseChannel
	Triangle     TriangleChannel
	Noise        NoiseChannel
	Dmc          DMCChannel
	Control      Control
	Status       Status
	FrameCounter FrameCounter
}

type APU struct {
	Registers Registers
}

func NewAPU() *APU {
	return &APU{}
}

func (apu *APU) Reset() {
	for i := range apu.Registers.Pulse1 {
		apu.Registers.Pulse1[i] = 0
	}

	for i := range apu.Registers.Pulse2 {
		apu.Registers.Pulse2[i] = 0
	}

	for i := range apu.Registers.Triangle {
		apu.Registers.Triangle[i] = 0
	}

	for i := range apu.Registers.Noise {
		apu.Registers.Noise[i] = 0
	}

	for i := range apu.Registers.Dmc {
		apu.Registers.Dmc[i] = 0
	}

	apu.Registers.Control = 0
	apu.Registers.Status = 0
	apu.Registers.FrameCounter = 0
}

func (apu *APU) Mappings(which Mapping) (fetch, store []uint16) {
	switch which {
	case CPU:
		fetch = []uint16{0x4015}
		store = []uint16{
			0x4000, 0x4001, 0x4002, 0x4003, 0x4004,
			0x4005, 0x4006, 0x4007, 0x4008, 0x400a,
			0x400b, 0x400c, 0x400e, 0x400f, 0x4010,
			0x4011, 0x4012, 0x4013, 0x4015, 0x4017,
		}
	}

	return
}

func (apu *APU) Fetch(address uint16) (value uint8) {
	value = 0

	switch address {
	// Status
	case 0x4015:
		value = uint8(apu.Registers.Status)
	}

	return
}

func (apu *APU) Store(address uint16, value uint8) (oldValue uint8) {
	oldValue = 0

	switch {
	// Pulse 1 channel
	case address >= 0x4000 && address <= 0x4003:
		index := address - 0x4000
		oldValue = apu.Registers.Pulse1[index]
		apu.Registers.Pulse1[index] = value
	// Pulse 2 channel
	case address >= 0x4004 && address <= 0x4007:
		index := address - 0x4004
		oldValue = apu.Registers.Pulse2[index]
		apu.Registers.Pulse2[index] = value
	// Triangle channel
	case address >= 0x4008 && address <= 0x400b:
		index := address - 0x4008

		switch address {
		case 0x4009: // 0x4009 is not mapped
			break
		case 0x400b:
			fallthrough
		case 0x400a:
			index--
			fallthrough
		case 0x4008:
			oldValue = apu.Registers.Triangle[index]
			apu.Registers.Triangle[index] = value
		}
	// Noise channel
	case address >= 0x400c && address <= 0x400f:
		index := address - 0x400c

		switch address {
		case 0x400d: // 0x400d is not mapped
			break
		case 0x400f:
			fallthrough
		case 0x400e:
			index--
			fallthrough
		case 0x400c:
			oldValue = apu.Registers.Noise[index]
			apu.Registers.Noise[index] = value
		}
	// DMC channel
	case address >= 0x4010 && address <= 0x4013:
		index := address - 0x4010
		oldValue = apu.Registers.Dmc[index]
		apu.Registers.Dmc[index] = value
	// Control
	case address == 0x4015:
		oldValue = uint8(apu.Registers.Control)
		apu.Registers.Control = Control(value)
	// Frame counter
	case address == 0x4017:
		oldValue = uint8(apu.Registers.FrameCounter)
		apu.Registers.FrameCounter = FrameCounter(value)
	}

	return
}
