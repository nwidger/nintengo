package rp2ago3

type PulseChannel [4]uint8
type TriangleChannel [3]uint8
type NoiseChannel [3]uint8
type DMCChannel [4]uint8
type Control uint8
type Status uint8
type FrameCounter uint8

type PulseFlag uint32

const (
	Duty PulseFlag = 1 << iota
	_
	PulseEnvelopeLoopLengthCounterHalt
	PulseConstantVolume
	PulseVolumeEnvelope
	_
	_
	_
	SweepEnabled
	SweepPeriod
	_
	_
	SweepNegate
	SweepShift
	_
	_
	PulseTimerLow
	_
	_
	_
	_
	_
	_
	_
	PulseLengthCounterLoad
	_
	_
	_
	_
	PulseTimerHigh
	_
	_
)

type TriangleFlag uint32

const (
	LengthCounterHaltLinearCounterControl = 1 << iota
	LinearCounterLoad
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	TriangleTimerLow
	_
	_
	_
	_
	_
	_
	_
	TriangleLengthCounterLoad
	_
	_
	_
	_
	TriangleTimerHigh
	_
	_
	_
)

type NoiseFlag uint32

const (
	_ NoiseFlag = 1 << iota
	_
	NoiseEnvelopeLoopLengthCounterHalt
	NoiseConstantVolume
	NoiseVolumeEnvelope
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	LoopNoise
	_
	_
	_
	NoisePeriod
	_
	_
	_
	NoiseLengthCounterLoad
	_
	_
	_
	_
	_
	_
	_
)

type DMCFlag uint32

const (
	IRQEnable DMCFlag = 1 << iota
	Loop
	_
	_
	Frequency
	_
	_
	_
	_
	LoadCounter
	_
	_
	_
	_
	_
	_
	SampleAddress
	_
	_
	_
	_
	_
	_
	_
	SampleLength
	_
	_
	_
	_
	_
	_
	_
)

type Registers struct {
	Pulse1       PulseChannel
	Pulse2       PulseChannel
	Triangle     TriangleChannel
	Noise        NoiseChannel
	DMC          DMCChannel
	Control      Control
	Status       Status
	FrameCounter FrameCounter
}

type APU struct {
	Registers Registers
	Samples   chan int16
}

func NewAPU() *APU {
	return &APU{
		Samples: make(chan int16),
	}
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

	for i := range apu.Registers.DMC {
		apu.Registers.DMC[i] = 0
	}

	apu.Registers.Control = 0
	apu.Registers.Status = 0
	apu.Registers.FrameCounter = 0
}

func (apu *APU) pulse1(flag PulseFlag) (value uint8) {
	return apu.pulse(apu.Registers.Pulse1, flag)
}

func (apu *APU) pulse2(flag PulseFlag) (value uint8) {
	return apu.pulse(apu.Registers.Pulse2, flag)
}

func (apu *APU) pulse(pulse PulseChannel, flag PulseFlag) (value uint8) {
	switch flag {
	case Duty:
		value = pulse[0] >> 6
	case PulseEnvelopeLoopLengthCounterHalt:
		value = (pulse[0] >> 5) & 0x01
	case PulseConstantVolume:
		value = (pulse[0] >> 4) & 0x01
	case PulseVolumeEnvelope:
		value = pulse[0] & 0x0f
	case SweepEnabled:
		value = pulse[1] >> 7
	case SweepPeriod:
		value = (pulse[1] >> 4) & 0x07
	case SweepNegate:
		value = (pulse[1] >> 3) & 0x01
	case SweepShift:
		value = (pulse[1] & 0x07)
	case PulseTimerLow:
		value = pulse[2]
	case PulseLengthCounterLoad:
		value = pulse[3] >> 3
	case PulseTimerHigh:
		value = pulse[3] & 0x07
	}

	return
}

func (apu *APU) triangle(flag TriangleFlag) (value uint8) {
	switch flag {
	case LengthCounterHaltLinearCounterControl:
		value = apu.Registers.Triangle[0] >> 7
	case LinearCounterLoad:
		value = apu.Registers.Triangle[0] & 0x7f
	case TriangleTimerLow:
		value = apu.Registers.Triangle[1]
	case TriangleLengthCounterLoad:
		value = apu.Registers.Triangle[2] >> 3
	case TriangleTimerHigh:
		value = apu.Registers.Triangle[2] & 0x07
	}

	return
}

func (apu *APU) noise(flag NoiseFlag) (value uint8) {
	switch flag {
	case NoiseEnvelopeLoopLengthCounterHalt:
		value = (apu.Registers.Noise[0] >> 5) & 0x01
	case NoiseConstantVolume:
		value = (apu.Registers.Noise[0] >> 4) & 0x01
	case NoiseVolumeEnvelope:
		value = apu.Registers.Noise[0] & 0x0f
	case LoopNoise:
		value = apu.Registers.Noise[1] >> 7
	case NoisePeriod:
		value = apu.Registers.Noise[1] & 0x0f
	case NoiseLengthCounterLoad:
		value = apu.Registers.Noise[2] >> 3
	}

	return
}

func (apu *APU) dmc(flag DMCFlag) (value uint8) {
	switch flag {
	case IRQEnable:
		value = apu.Registers.DMC[0] >> 7
	case Loop:
		value = (apu.Registers.DMC[0] >> 6) & 0x01
	case Frequency:
		value = apu.Registers.DMC[0] & 0x0f
	case LoadCounter:
		value = apu.Registers.DMC[1] & 0x7f
	case SampleAddress:
		value = apu.Registers.DMC[2]
	case SampleLength:
		value = apu.Registers.DMC[3]
	}

	return
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
		oldValue = apu.Registers.DMC[index]
		apu.Registers.DMC[index] = value
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

func (apu *APU) Execute() {

}
