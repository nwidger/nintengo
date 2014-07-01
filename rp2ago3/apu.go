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

type ControlFlag uint8

const (
	EnablePulseChannel1 ControlFlag = 1 << iota
	EnablePulseChannel2
	EnableTriangle
	EnableNoise
	EnableDMC
	_
	_
	_
)

type StatusFlag uint8

const (
	Pulse1LengthCounterNotZero StatusFlag = 1 << iota
	Pulse2LengthCounterNotZero
	TriangleLengthCounterNotZero
	NoiseLengthCounterNotZero
	DMCActive
	_
	FrameInterrupt
	DMCInterrupt
)

type FrameCounterFlag uint8

const (
	_ FrameCounterFlag = 1 << iota
	_
	_
	_
	_
	_
	IRQInhibit
	Mode
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

type Divider struct {
	Counter int16
	Period  int16
}

type FrameCounterSequencer struct {
	NumSteps uint8
	Step     uint8
	Cycles   uint16
}

type Envelope struct {
	Start bool
	Loop  bool
	Divider
	Counter uint8
}

type Noise struct {
	Enabled bool
	Envelope
	Divider
	Shift         uint16
	LengthCounter uint8
}

type APU struct {
	Registers        Registers
	Samples          chan int16
	NoisePeriodLUT   [16]int16
	LengthCounterLUT [32]uint8
	Interrupt        func(state bool)
	FrameCounter     FrameCounterSequencer
}

func NewAPU(interrupt func(bool)) *APU {
	return &APU{
		Samples:   make(chan int16),
		Interrupt: interrupt,
		NoisePeriodLUT: [16]int16{
			// NTSC
			4, 8, 16, 32, 64, 96, 128, 160, 202,
			254, 380, 508, 762, 1016, 2034, 4068,
			// PAL
			// 4, 8, 14, 30, 60, 88, 118, 148, 188,
			// 236, 354, 472, 708, 944, 1890, 3778,
		},
		LengthCounterLUT: [32]uint8{
			0x0a, 0xfe, 0x14, 0x02,
			0x28, 0x04, 0x50, 0x06,
			0xa0, 0x08, 0x3c, 0x0a,
			0x0e, 0x0c, 0x1a, 0x0e,
			0x0c, 0x10, 0x18, 0x12,
			0x30, 0x14, 0x60, 0x16,
			0xc0, 0x18, 0x48, 0x1a,
			0x10, 0x1c, 0x20, 0x1e,
		},
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

func (apu *APU) control(flag ControlFlag) (value bool) {
	if (apu.Registers.Control & Control(flag)) != 0 {
		value = true
	}

	return
}

func (apu *APU) status(flag StatusFlag) (value bool) {
	if (apu.Registers.Status & Status(flag)) != 0 {
		value = true
	}

	return
}

func (apu *APU) frameCounter(flag FrameCounterFlag) (value uint8) {
	switch flag {
	case Mode:
		switch uint8(apu.Registers.FrameCounter >> 7) {
		case 0:
			value = 4
		case 1:
			value = 5
		}
	case IRQInhibit:
		value = uint8(apu.Registers.FrameCounter>>6) & 0x01
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
		apu.Registers.Status &= Status(^FrameInterrupt)
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
		apu.Registers.Status &= Status(^DMCInterrupt)
	// Frame counter
	case address == 0x4017:
		oldValue = uint8(apu.Registers.FrameCounter)
		apu.Registers.FrameCounter = FrameCounter(value)
	}

	return
}

func (envelope *Envelope) Clock() {
	if envelope.Start {
		envelope.Start = false
		envelope.Counter = 0x0f
		envelope.Divider.Reload()
	} else if envelope.Divider.Clock() {
		if envelope.Counter > 0 {
			envelope.Counter--
		} else if envelope.Loop {
			envelope.Counter = 0x0f
		}
	}
}

func (divider *Divider) Clock() (output bool) {
	divider.Counter--

	if divider.Counter == 0 {
		divider.Reload()
		output = true
	}

	return
}

func (divider *Divider) Reload() {
	divider.Counter = divider.Period
}

func (frameCounter *FrameCounterSequencer) Clock() (changed bool, newStep uint8) {
	// 2 CPU cycles = 1 APU cycle
	frameCounter.Cycles++

	oldStep := frameCounter.Step

	switch frameCounter.Cycles {
	case 3729 * 2, 7457 * 2, 11186 * 2,
		14915 * 2, 18641 * 2:
		frameCounter.Step++
	}

	newStep = frameCounter.Step

	if oldStep != newStep {
		changed = true
	}

	return
}

func (frameCounter *FrameCounterSequencer) Reset() {
	frameCounter.Step = 0
}

func (apu *APU) Execute() {
	if changed, step := apu.FrameCounter.Clock(); changed {
		switch step {
		case 1:
			// clock env & tri's linear counter
		case 2:
			// clock env & tri's linear counter
			// clock length counters & sweep units
		case 3:
			// clock env & tri's linear counter
		case 4:
			if apu.FrameCounter.NumSteps == 4 {
				// clock env & tri's linear counter
				// clock length counters & sweep units
				// set frame interrupt flag if interrupt inhibit is clear
			}
		case 5:
			if apu.FrameCounter.NumSteps == 5 {
				// clock env & tri's linear counter
				// clock length counters & sweep units
			}
		}

		if step == apu.FrameCounter.NumSteps {
			apu.FrameCounter.Reset()
		}
	}

	if apu.control(EnableNoise) {
		switch apu.noise(NoiseConstantVolume) {
		case 1:
			volume := apu.noise(NoiseVolumeEnvelope)
			_ = volume
		case 0:
			period := apu.NoisePeriodLUT[apu.noise(NoisePeriod)]
			_ = period
		}
	}
}
