package rp2ago3

type Control uint8
type Status uint8

type PulseFlag uint32

const (
	Duty PulseFlag = 1 << iota
	PulseEnvelopeLoopLengthCounterHalt
	PulseConstantVolume
	PulseVolumeEnvelope
	SweepEnabled
	SweepPeriod
	SweepNegate
	SweepShift
	PulseTimerLow
	PulseLengthCounterLoad
	PulseTimerHigh
)

type TriangleFlag uint32

const (
	LengthCounterHaltLinearCounterControl = 1 << iota
	LinearCounterLoad
	TriangleTimerLow
	TriangleLengthCounterLoad
	TriangleTimerHigh
)

type NoiseFlag uint32

const (
	NoiseEnvelopeLoopLengthCounterHalt NoiseFlag = 1 << iota
	NoiseConstantVolume
	NoiseVolumeEnvelope
	LoopNoise
	NoisePeriod
	NoiseLengthCounterLoad
)

type DMCFlag uint32

const (
	IRQEnable DMCFlag = 1 << iota
	Loop
	Frequency
	LoadCounter
	SampleAddress
	SampleLength
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
	IRQInhibit FrameCounterFlag = 1 << iota
	Mode
)

var LengthCounterLUT [32]uint8 = [32]uint8{
	0x0a, 0xfe, 0x14, 0x02,
	0x28, 0x04, 0x50, 0x06,
	0xa0, 0x08, 0x3c, 0x0a,
	0x0e, 0x0c, 0x1a, 0x0e,
	0x0c, 0x10, 0x18, 0x12,
	0x30, 0x14, 0x60, 0x16,
	0xc0, 0x18, 0x48, 0x1a,
	0x10, 0x1c, 0x20, 0x1e,
}

var SequencerLUT [8][]uint8 = [8][]uint8{
	[]uint8{0, 1, 0, 0, 0, 0, 0, 0},
	[]uint8{0, 1, 1, 0, 0, 0, 0, 0},
	[]uint8{0, 1, 1, 1, 1, 0, 0, 0},
	[]uint8{1, 0, 0, 1, 1, 1, 1, 1},
}

type Registers struct {
	Control Control
	Status  Status
}

type APU struct {
	Muted     bool      `json:"-"`
	Registers Registers `json:"APURegisters"`

	Pulse1       Pulse
	Pulse2       Pulse
	Triangle     Triangle
	Noise        Noise
	DMC          DMC
	FrameCounter FrameCounter

	Cycles       uint64
	TargetCycles uint64

	HipassStrong int64
	HipassWeak   int64
	pulseLUT     [31]float64
	tndLUT       [203]float64

	Interrupt func(state bool) `json:"-"`
}

func NewAPU(targetCycles uint64, interrupt func(bool)) *APU {
	apu := &APU{
		TargetCycles: targetCycles,
		Interrupt:    interrupt,
		Pulse1: Pulse{
			MinusOne: true,
			Divider: Divider{
				PlusOne:  true,
				TimesTwo: true,
			},
			SequencerLUT:     SequencerLUT,
			LengthCounterLUT: LengthCounterLUT,
		},
		Pulse2: Pulse{
			Divider: Divider{
				TimesTwo: true,
			},
			SequencerLUT:     SequencerLUT,
			LengthCounterLUT: LengthCounterLUT,
		},
		Noise: Noise{
			Envelope: Envelope{
				Divider: Divider{
					PlusOne: true,
				},
			},
			PeriodLUT: [16]int16{
				// NTSC
				4, 8, 16, 32, 64, 96, 128, 160, 202,
				254, 380, 508, 762, 1016, 2034, 4068,
				// PAL
				// 4, 8, 14, 30, 60, 88, 118, 148, 188,
				// 236, 354, 472, 708, 944, 1890, 3778,
			},
			LengthCounterLUT: LengthCounterLUT,
		},
		Triangle: Triangle{
			Divider: Divider{
				PlusOne:  true,
				TimesTwo: true,
			},
			Sequencer: Sequencer{
				Values: []uint8{
					15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0,
					0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
				},
			},
			LengthCounterLUT: LengthCounterLUT,
		},
	}

	for i := 0; i < len(apu.pulseLUT); i++ {
		apu.pulseLUT[i] = 95.52 / (8128.0/float64(i) + 100.0)
	}

	for i := 0; i < len(apu.tndLUT); i++ {
		apu.tndLUT[i] = 163.67 / (24329.0/float64(i) + 100.0)
	}

	return apu
}

func (apu *APU) Reset() {
	apu.Cycles = 0

	apu.Registers.Control = 0x00
	apu.Registers.Status = 0x00

	apu.Pulse1.Reset()
	apu.Pulse2.Reset()
	apu.Noise.Reset()
	apu.Triangle.Reset()

	apu.FrameCounter.Reset()

	apu.HipassStrong = 0.0
	apu.HipassWeak = 0.0
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
	switch address {
	// Status
	case 0x4015:
		value = apu.FetchUpdatedStatus()
	}

	return
}

func (apu *APU) Store(address uint16, value uint8) (oldValue uint8) {
	switch {
	// Pulse 1 channel
	case address >= 0x4000 && address <= 0x4003:
		oldValue = apu.Pulse1.Store(address-0x4000, value)
	// Pulse 2 channel
	case address >= 0x4004 && address <= 0x4007:
		oldValue = apu.Pulse2.Store(address-0x4004, value)
	// Triangle channel
	case address >= 0x4008 && address <= 0x400b:
		index := address - 0x4008

		switch address {
		case 0x4009: // 0x4009 is not mapped
			break
		case 0x400a, 0x400b:
			index--
			fallthrough
		case 0x4008:
			oldValue = apu.Triangle.Store(index, value)
		}
	// Noise channel
	case address >= 0x400c && address <= 0x400f:
		index := address - 0x400c

		switch address {
		case 0x400d: // 0x400d is not mapped
			break
		case 0x400e, 0x400f:
			index--
			fallthrough
		case 0x400c:
			oldValue = apu.Noise.Store(index, value)
		}
	// DMC channel
	case address >= 0x4010 && address <= 0x4013:
		oldValue = apu.DMC.Store(address-0x4010, value)
	// Control
	case address == 0x4015:
		oldValue = uint8(apu.Registers.Control)
		apu.Registers.Control = Control(value)
		apu.Registers.Status &= Status(^DMCInterrupt)

		apu.Pulse1.SetEnabled(apu.control(EnablePulseChannel1))
		apu.Pulse2.SetEnabled(apu.control(EnablePulseChannel2))
		apu.Noise.SetEnabled(apu.control(EnableNoise))
		apu.Triangle.SetEnabled(apu.control(EnableTriangle))
	// Frame counter
	case address == 0x4017:
		var executeFrameCounter bool

		if oldValue, executeFrameCounter = apu.FrameCounter.Store(value); executeFrameCounter {
			apu.ExecuteFrameCounter()
		}

		apu.status(FrameInterrupt, apu.FrameCounter.register(IRQInhibit) != 1)
	}

	return
}

func (apu *APU) FetchUpdatedStatus() (value uint8) {
	apu.status(Pulse1LengthCounterNotZero, apu.Pulse1.LengthCounter > 0)
	apu.status(Pulse2LengthCounterNotZero, apu.Pulse2.LengthCounter > 0)
	apu.status(NoiseLengthCounterNotZero, apu.Noise.LengthCounter > 0)
	apu.status(TriangleLengthCounterNotZero, apu.Triangle.LengthCounter > 0)

	value = uint8(apu.Registers.Status)

	apu.status(FrameInterrupt, false)

	return
}

func (apu *APU) hipassStrong(s int16) int16 {
	HiPassStrong := int64(225574)

	apu.HipassStrong += (((int64(s) << 16) - (apu.HipassStrong >> 16)) * HiPassStrong) >> 16
	return int16(int64(s) - (apu.HipassStrong >> 32))
}

func (apu *APU) hipassWeak(s int16) int16 {
	HiPassWeak := int64(57593)

	apu.HipassWeak += (((int64(s) << 16) - (apu.HipassWeak >> 16)) * HiPassWeak) >> 16
	return int16(int64(s) - (apu.HipassWeak >> 32))
}

func (apu *APU) Sample() (sample int16) {
	if !apu.Muted {
		pulse := apu.pulseLUT[apu.Pulse1.Sample()+apu.Pulse2.Sample()]
		tnd := apu.tndLUT[(3*apu.Triangle.Sample())+(2*apu.Noise.Sample())+apu.DMC.Sample()]

		sample = int16((pulse + tnd) * 40000)
		sample = apu.hipassStrong(sample)
		sample = apu.hipassWeak(sample)
	}

	return
}

func (apu *APU) control(flag ControlFlag, state ...bool) (value bool) {
	if len(state) == 0 {
		if (apu.Registers.Control & Control(flag)) != 0 {
			value = true
		}
	} else {
		value = state[0]

		if !value {
			apu.Registers.Control &= Control(^flag)
		} else {
			apu.Registers.Control |= Control(flag)
		}
	}

	return
}

func (apu *APU) status(flag StatusFlag, state ...bool) (value bool) {
	if len(state) == 0 {
		if (apu.Registers.Status & Status(flag)) != 0 {
			value = true
		}
	} else {
		value = state[0]

		if !value {
			apu.Registers.Status &= Status(^flag)
		} else {
			apu.Registers.Status |= Status(flag)
		}
	}

	return
}

func (apu *APU) ClockEnvelopes() {
	apu.Pulse1.ClockEnvelope()
	apu.Pulse2.ClockEnvelope()
	apu.Noise.ClockEnvelope()
}

func (apu *APU) ClockLengthCounters() {
	apu.Pulse1.ClockLengthCounter()
	apu.Pulse2.ClockLengthCounter()
	apu.Triangle.ClockLengthCounter()
	apu.Noise.ClockLengthCounter()
}

func (apu *APU) ClockSweepUnits() {
	apu.Pulse1.ClockSweepUnit()
	apu.Pulse2.ClockSweepUnit()
}

func (apu *APU) ExecuteFrameCounter() {
	if changed, step := apu.FrameCounter.Clock(); changed {
		// mode 0:    mode 1:       function
		// ---------  -----------  -----------------------------
		//  - - - f    - - - - -    IRQ (if bit 6 is clear)
		//  - l - l    l - l - -    Length counter and sweep
		//  e e e e    e e e e -    Envelope and linear counter

		if step != 5 {
			// clock env & tri's linear counter
			apu.ClockEnvelopes()
			apu.Triangle.ClockLinearCounter()
		}

		if (apu.FrameCounter.register(Mode) == 4 && (step == 2 || step == 4)) ||
			(apu.FrameCounter.register(Mode) == 5 && (step == 1 || step == 3)) {
			// clock length counters & sweep units
			apu.ClockLengthCounters()
			apu.ClockSweepUnits()
		}

		if step == apu.FrameCounter.register(Mode) {
			// set frame interrupt flag if interrupt inhibit is clear
			if step == 4 && apu.FrameCounter.register(IRQInhibit) == 0 {
				apu.status(FrameInterrupt, true)
			}

			apu.FrameCounter.Reset()
		}
	}
}

func (apu *APU) Execute() (sample int16, haveSample bool) {
	if apu.control(EnablePulseChannel1) {
		apu.Pulse1.ClockDivider()
	}

	if apu.control(EnablePulseChannel2) {
		apu.Pulse2.ClockDivider()
	}

	if apu.control(EnableTriangle) {
		apu.Triangle.ClockDivider()
	}

	if apu.control(EnableNoise) {
		apu.Noise.ClockDivider()
	}

	if apu.control(EnableDMC) {

	}

	apu.ExecuteFrameCounter()

	if apu.Cycles++; apu.Cycles == apu.TargetCycles {
		sample = apu.Sample()
		haveSample = true

		apu.Cycles = 0.0
		apu.TargetCycles ^= 0x1
	}

	return
}

type Pulse struct {
	Muted     bool `json:"-"`
	Enabled   bool
	MinusOne  bool
	Registers [4]uint8

	Envelope         Envelope
	SweepUnit        SweepUnit
	Divider          Divider
	Sequencer        Sequencer
	SequencerLUT     [8][]uint8 `json:"-"`
	LengthCounter    uint8
	LengthCounterLUT [32]uint8 `json:"-"`
}

func (pulse *Pulse) Reset() {
	pulse.Enabled = false

	for i := range pulse.Registers {
		pulse.Registers[i] = 0x00
	}

	pulse.Envelope.Reset()
	pulse.SweepUnit.Reset()
	pulse.Divider.Reset()
	pulse.Sequencer.Reset()

	pulse.LengthCounter = 0x00
}

func (pulse *Pulse) SetEnabled(enabled bool) {
	if pulse.Enabled = enabled; !enabled {
		pulse.LengthCounter = 0
	}
}

func (pulse *Pulse) Store(index uint16, value uint8) (oldValue uint8) {
	oldValue = pulse.Registers[index]
	pulse.Registers[index] = value

	switch index {
	// $4000 / $4004
	case 0:
		pulse.Envelope.Counter = pulse.registers(PulseVolumeEnvelope)
		pulse.Envelope.Loop = pulse.registers(PulseEnvelopeLoopLengthCounterHalt) == 1
		pulse.Sequencer.Values = pulse.SequencerLUT[pulse.registers(Duty)]
	// $4001 / $4005
	case 1:
		pulse.SweepUnit.Enabled = pulse.registers(SweepEnabled) == 1
		pulse.SweepUnit.Divider.Period = int16(pulse.registers(SweepPeriod))
		pulse.SweepUnit.Reload = true
	// $4002 / $4006
	case 2:
		pulse.Divider.Period = (pulse.Divider.Period & 0x0700) | int16(pulse.registers(PulseTimerLow))
	// $4003 / $4007
	case 3:
		if pulse.Enabled {
			pulse.LengthCounter = pulse.LengthCounterLUT[pulse.registers(PulseLengthCounterLoad)]
		}

		pulse.Divider.Period = (pulse.Divider.Period & 0x00ff) | (int16(pulse.registers(PulseTimerHigh)) << 8)
		pulse.Sequencer.Reset()
		pulse.Envelope.Start = true
	}

	return
}

func (pulse *Pulse) registers(flag PulseFlag, state ...uint8) (value uint8) {
	if len(state) == 0 {
		switch flag {
		case Duty:
			value = pulse.Registers[0] >> 6
		case PulseEnvelopeLoopLengthCounterHalt:
			value = (pulse.Registers[0] >> 5) & 0x01
		case PulseConstantVolume:
			value = (pulse.Registers[0] >> 4) & 0x01
		case PulseVolumeEnvelope:
			value = pulse.Registers[0] & 0x0f
		case SweepEnabled:
			value = pulse.Registers[1] >> 7
		case SweepPeriod:
			value = (pulse.Registers[1] >> 4) & 0x07
		case SweepNegate:
			value = (pulse.Registers[1] >> 3) & 0x01
		case SweepShift:
			value = (pulse.Registers[1] & 0x07)
		case PulseTimerLow:
			value = pulse.Registers[2]
		case PulseLengthCounterLoad:
			value = pulse.Registers[3] >> 3
		case PulseTimerHigh:
			value = pulse.Registers[3] & 0x07
		}
	} else {
		value = state[0]

		switch flag {
		case Duty:
			value = (pulse.Registers[0] & 0x3f) | ((value & 0x03) << 6)
		case PulseEnvelopeLoopLengthCounterHalt:
			value = (pulse.Registers[0] & 0xdf) | ((value & 0x01) << 5)
		case PulseConstantVolume:
			value = (pulse.Registers[0] & 0xef) | ((value & 0x01) << 4)
		case PulseVolumeEnvelope:
			value = (pulse.Registers[0] & 0xf0) | (value & 0x0f)
		case SweepEnabled:
			value = (pulse.Registers[1] & 0x7f) | ((value & 0x01) << 7)
		case SweepPeriod:
			value = (pulse.Registers[1] & 0x8f) | ((value & 0x07) << 4)
		case SweepNegate:
			value = (pulse.Registers[1] & 0xf7) | ((value & 0x01) << 3)
		case SweepShift:
			value = (pulse.Registers[1] & 0xf8) | (value & 0x07)
		case PulseTimerLow:
			pulse.Registers[2] = value
		case PulseLengthCounterLoad:
			value = (pulse.Registers[3] & 0x07) | ((value & 0x1f) << 3)
		case PulseTimerHigh:
			value = (pulse.Registers[3] & 0xf8) | (value & 0x07)
		}
	}

	return
}

func (pulse *Pulse) TargetPeriod() (target int16) {
	current := pulse.Divider.Period
	delta := (current >> pulse.registers(SweepShift))

	if pulse.registers(SweepNegate) == 1 {
		delta *= -1

		if pulse.MinusOne {
			delta -= 1
		}
	}

	target = current + delta
	return
}

func (pulse *Pulse) Sample() (sample int16) {
	if !pulse.Muted && pulse.Sequencer.Output != 0 &&
		(pulse.Divider.Period >= 0x0008 && pulse.TargetPeriod() <= 0x07ff) &&
		pulse.LengthCounter != 0 && pulse.Divider.Counter >= 8 {
		if pulse.registers(PulseConstantVolume) == 0 {
			sample = int16(pulse.Envelope.Counter)
		} else {
			sample = int16(pulse.registers(PulseVolumeEnvelope))
		}
	}

	return
}

func (pulse *Pulse) ClockEnvelope() {
	pulse.Envelope.Clock()
}

func (pulse *Pulse) ClockDivider() {
	if pulse.Divider.Clock() {
		pulse.ClockSequencer()
	}

	return
}

func (pulse *Pulse) ClockSequencer() {
	pulse.Sequencer.Clock()
}

func (pulse *Pulse) ClockLengthCounter() {
	if pulse.Enabled && pulse.registers(PulseEnvelopeLoopLengthCounterHalt) != 0 &&
		pulse.LengthCounter > 0 {
		pulse.LengthCounter--
	}

	return
}

func (pulse *Pulse) ClockSweepUnit() {
	current := pulse.Divider.Period
	target := pulse.TargetPeriod()

	if adjustPeriod := pulse.SweepUnit.Clock(); (current >= 0x0008 && target <= 0x07ff) &&
		pulse.Enabled && pulse.registers(SweepShift) > 0 && adjustPeriod {
		pulse.Divider.Period = target
	}
}

type Triangle struct {
	Muted     bool `json:"-"`
	Enabled   bool
	Registers [3]uint8

	Divider          Divider
	LinearCounter    LinearCounter
	Sequencer        Sequencer
	LengthCounter    uint8
	LengthCounterLUT [32]uint8 `json:"-"`
}

func (triangle *Triangle) Reset() {
	triangle.Enabled = false

	for i := range triangle.Registers {
		triangle.Registers[i] = 0x00
	}

	triangle.Divider.Reset()
	triangle.LinearCounter.Reset()
	triangle.Sequencer.Reset()
	triangle.LengthCounter = 0x00
}

func (triangle *Triangle) SetEnabled(enabled bool) {
	if triangle.Enabled = enabled; !enabled {
		triangle.LengthCounter = 0
	}
}

func (triangle *Triangle) Store(index uint16, value uint8) (oldValue uint8) {
	oldValue = triangle.Registers[index]
	triangle.Registers[index] = value

	switch index {
	// $4008
	case 0:
		// C---.----: control flag (and length counter halt flag)
		// -RRR.RRRR: counter reload value
		triangle.LinearCounter.Control = triangle.registers(LengthCounterHaltLinearCounterControl) != 0
		triangle.LinearCounter.ReloadValue = triangle.registers(LinearCounterLoad)
	// $400a
	case 1:
		// LLLL.LLLL: timer low
		triangle.Divider.Period = (triangle.Divider.Period & 0x0700) | int16(triangle.registers(TriangleTimerLow))
	// $400b
	case 2:
		// llll.lHHH: length counter load, timer high
		if triangle.Enabled {
			triangle.LengthCounter = triangle.LengthCounterLUT[triangle.registers(TriangleLengthCounterLoad)]
		}

		triangle.Divider.Period = (triangle.Divider.Period & 0x00ff) | (int16(triangle.registers(TriangleTimerHigh)) << 8)
		triangle.LinearCounter.Halt = true
	}

	return
}

func (triangle *Triangle) registers(flag TriangleFlag, state ...uint8) (value uint8) {
	if len(state) == 0 {
		switch flag {
		case LengthCounterHaltLinearCounterControl:
			value = triangle.Registers[0] >> 7
		case LinearCounterLoad:
			value = triangle.Registers[0] & 0x7f
		case TriangleTimerLow:
			value = triangle.Registers[1]
		case TriangleLengthCounterLoad:
			value = triangle.Registers[2] >> 3
		case TriangleTimerHigh:
			value = triangle.Registers[2] & 0x07
		}
	} else {
		value = state[0]

		switch flag {
		case LengthCounterHaltLinearCounterControl:
			value = (triangle.Registers[0] & 0x7f) | ((value & 0x01) << 7)
		case LinearCounterLoad:
			value = (triangle.Registers[0] & 0x80) | (value & 0x7f)
		case TriangleTimerLow:
			triangle.Registers[1] = value
		case TriangleLengthCounterLoad:
			value = (triangle.Registers[2] & 0x07) | ((value & 0x1f) << 3)
		case TriangleTimerHigh:
			value = (triangle.Registers[2] & 0xf8) | (value & 0x07)
		}
	}

	return
}

func (triangle *Triangle) Sample() (sample int16) {
	if !triangle.Muted && triangle.Enabled &&
		triangle.LinearCounter.Counter > 0 && triangle.LengthCounter > 0 {
		sample = int16(triangle.Sequencer.Output)
	}

	return
}

func (triangle *Triangle) ClockDivider() {
	if triangle.Divider.Clock() &&
		triangle.LinearCounter.Counter > 0 && triangle.LengthCounter > 0 {
		triangle.ClockSequencer()
	}
}

func (triangle *Triangle) ClockLinearCounter() {
	if triangle.Enabled && triangle.registers(LengthCounterHaltLinearCounterControl) != 0 {
		triangle.LinearCounter.Clock()
	}

	return
}

func (triangle *Triangle) ClockLengthCounter() {
	if triangle.Enabled && triangle.registers(LengthCounterHaltLinearCounterControl) != 0 &&
		triangle.LengthCounter > 0 {
		triangle.LengthCounter--
	}

	return
}

func (triangle *Triangle) ClockSequencer() {
	triangle.Sequencer.Clock()
	return
}

type Noise struct {
	Muted     bool `json:"-"`
	Enabled   bool
	Registers [3]uint8

	Envelope         Envelope
	Divider          Divider
	Shift            uint16
	LengthCounter    uint8
	LengthCounterLUT [32]uint8 `json:"-"`
	PeriodLUT        [16]int16 `json:"-"`
}

func (noise *Noise) Reset() {
	noise.Enabled = false

	for i := range noise.Registers {
		noise.Registers[i] = 0x00
	}

	noise.Envelope.Reset()
	noise.Divider.Reset()

	noise.Shift = 0x0001
	noise.LengthCounter = 0x00
}

func (noise *Noise) SetEnabled(enabled bool) {
	if noise.Enabled = enabled; !enabled {
		noise.LengthCounter = 0
	}
}

func (noise *Noise) Store(index uint16, value uint8) (oldValue uint8) {
	oldValue = noise.Registers[index]
	noise.Registers[index] = value

	switch index {
	// $400c
	case 0:
		noise.Envelope.Loop = noise.registers(NoiseEnvelopeLoopLengthCounterHalt) != 0
		noise.Envelope.Divider.Period = int16(noise.registers(NoiseVolumeEnvelope))
	// $400e
	case 1:
		noise.Divider.Period = noise.PeriodLUT[noise.registers(NoisePeriod)]
		noise.Divider.Reload()
	// $400f
	case 2:
		noise.Envelope.Start = true

		if noise.Enabled {
			noise.LengthCounter = noise.LengthCounterLUT[noise.registers(NoiseLengthCounterLoad)]
		}
	}

	return
}

func (noise *Noise) registers(flag NoiseFlag, state ...uint8) (value uint8) {
	if len(state) == 0 {
		switch flag {
		case NoiseEnvelopeLoopLengthCounterHalt:
			value = (noise.Registers[0] >> 5) & 0x01
		case NoiseConstantVolume:
			value = (noise.Registers[0] >> 4) & 0x01
		case NoiseVolumeEnvelope:
			value = noise.Registers[0] & 0x0f
		case LoopNoise:
			value = noise.Registers[1] >> 7
		case NoisePeriod:
			value = noise.Registers[1] & 0x0f
		case NoiseLengthCounterLoad:
			value = noise.Registers[2] >> 3
		}
	} else {
		value = state[0]

		switch flag {
		case NoiseEnvelopeLoopLengthCounterHalt:
			value = (noise.Registers[0] & 0xdf) | ((value & 0x01) << 5)
		case NoiseConstantVolume:
			value = (noise.Registers[0] & 0xef) | ((value & 0x01) << 4)
		case NoiseVolumeEnvelope:
			value = (noise.Registers[0] & 0xf0) | (value & 0x0f)
		case LoopNoise:
			value = (noise.Registers[1] & 0x7f) | ((value & 0x01) << 7)
		case NoisePeriod:
			value = (noise.Registers[1] & 0xf0) | (value & 0x0f)
		case NoiseLengthCounterLoad:
			value = (noise.Registers[2] & 0x07) | ((value & 0x1f) << 3)
		}
	}

	return
}

func (noise *Noise) ClockLengthCounter() {
	if noise.Enabled && noise.registers(NoiseEnvelopeLoopLengthCounterHalt) == 0 &&
		noise.LengthCounter > 0 {
		noise.LengthCounter--
	}

	return
}

func (noise *Noise) ClockEnvelope() {
	noise.Envelope.Clock()
}

func (noise *Noise) ClockDivider() {
	var tmp uint16

	if noise.Divider.Clock() {
		if noise.registers(LoopNoise) == 1 {
			tmp = 6
		} else {
			tmp = 1
		}

		bit := (noise.Shift >> tmp) & 0x0001
		feedback := (noise.Shift & 0x0001) ^ bit

		noise.Shift = (noise.Shift >> 1) | (feedback << 14)
	}
}

func (noise *Noise) Sample() (sample int16) {
	if !noise.Muted && noise.Enabled && (noise.Shift&0x0001) == 0 && noise.LengthCounter != 0 {
		if noise.registers(NoiseConstantVolume) == 0 {
			sample = int16(noise.Envelope.Counter)
		} else {
			sample = int16(noise.registers(NoiseVolumeEnvelope))
		}
	}

	return
}

type DMC struct {
	Registers [4]uint8
}

func (dmc *DMC) Store(index uint16, value uint8) (oldValue uint8) {
	oldValue = dmc.Registers[index]
	dmc.Registers[index] = value

	return
}

func (dmc *DMC) registers(flag DMCFlag, state ...uint8) (value uint8) {
	if len(state) == 0 {
		switch flag {
		case IRQEnable:
			value = dmc.Registers[0] >> 7
		case Loop:
			value = (dmc.Registers[0] >> 6) & 0x01
		case Frequency:
			value = dmc.Registers[0] & 0x0f
		case LoadCounter:
			value = dmc.Registers[1] & 0x7f
		case SampleAddress:
			value = dmc.Registers[2]
		case SampleLength:
			value = dmc.Registers[3]
		}
	} else {
		value = state[0]

		switch flag {
		case IRQEnable:
			value = (dmc.Registers[0] & 0x7f) | ((value & 0x01) << 7)
		case Loop:
			value = (dmc.Registers[0] & 0xbf) | ((value & 0x01) << 6)
		case Frequency:
			value = (dmc.Registers[0] & 0xf0) | (value & 0x0f)
		case LoadCounter:
			value = (dmc.Registers[1] & 0x80) | (value & 0x7f)
		case SampleAddress:
			dmc.Registers[2] = value
		case SampleLength:
			dmc.Registers[3] = value
		}
	}

	return
}

func (dmc *DMC) Sample() (sample int16) {
	return
}

type FrameCounter struct {
	Register uint8
	Step     uint8
	Cycles   float64
}

func (frameCounter *FrameCounter) Reset() {
	frameCounter.Cycles = 0
	frameCounter.Step = 0
}

func (frameCounter *FrameCounter) Store(value uint8) (oldValue uint8, executeFrameCounter bool) {
	oldValue = frameCounter.Register
	frameCounter.Register = value

	oldValue = uint8(frameCounter.Register)
	frameCounter.Register = value

	frameCounter.Reset()

	if frameCounter.register(Mode) == 5 {
		executeFrameCounter = true
	}

	return
}

func (frameCounter *FrameCounter) register(flag FrameCounterFlag, state ...uint8) (value uint8) {
	if len(state) == 0 {
		switch flag {
		case Mode:
			switch uint8(frameCounter.Register >> 7) {
			case 0:
				value = 4
			case 1:
				value = 5
			}
		case IRQInhibit:
			value = uint8(frameCounter.Register>>6) & 0x01
		}
	} else {
		value = state[0]

		switch flag {
		case Mode:
			switch value {
			case 4:
				frameCounter.Register &= 0x7f
			case 5:
				frameCounter.Register |= 0x80
			}
		case IRQInhibit:
			switch value {
			case 0:
				frameCounter.Register &= 0xbf
			case 1:
				frameCounter.Register |= 0x40
			}
		}
	}

	return
}

func (frameCounter *FrameCounter) Clock() (changed bool, newStep uint8) {
	frameCounter.Cycles += 1.0

	mod := 7457.0

	switch frameCounter.Cycles {
	//   1        2        3        4        5
	case mod * 1, mod * 2, mod * 3, mod * 4, mod * 5:
		frameCounter.Step++
		changed = true
	}

	newStep = frameCounter.Step

	return
}

type Sequencer struct {
	Values []uint8
	Index  int
	Output uint8
}

func (sequencer *Sequencer) Clock() (output uint8) {
	if sequencer.Values != nil {
		sequencer.Output = sequencer.Values[sequencer.Index]
		output = sequencer.Output

		sequencer.Index++

		if sequencer.Index == len(sequencer.Values) {
			sequencer.Index = 0
		}
	}

	return
}

func (sequencer *Sequencer) Reset() {
	sequencer.Index = 0
	sequencer.Output = 0
}

type LinearCounter struct {
	Control     bool
	Halt        bool
	ReloadValue uint8
	Counter     uint8
}

func (linearCounter *LinearCounter) Clock() (counter uint8) {
	if linearCounter.Halt {
		linearCounter.Counter = linearCounter.ReloadValue
	} else if linearCounter.Counter > 0 {
		linearCounter.Counter--
	}

	if !linearCounter.Control {
		linearCounter.Halt = false
	}

	counter = linearCounter.Counter

	return
}

func (linearCounter *LinearCounter) Reset() {
	linearCounter.Control = false
	linearCounter.Halt = false
	linearCounter.Counter = 0
	linearCounter.ReloadValue = 0
}

type SweepUnit struct {
	Enabled bool
	Reload  bool
	Divider Divider
}

func (sweepUnit *SweepUnit) Reset() {
	sweepUnit.Enabled = false
	sweepUnit.Reload = false
	sweepUnit.Divider.Reset()
}

func (sweepUnit *SweepUnit) Clock() (adjustPeriod bool) {
	if sweepUnit.Reload {
		if sweepUnit.Divider.Counter == 0 && sweepUnit.Enabled {
			// adjust pulse's period if in range
			adjustPeriod = true
		}

		sweepUnit.Divider.Reload()
		sweepUnit.Reload = false
	} else if sweepUnit.Divider.Clock() && sweepUnit.Enabled {
		sweepUnit.Divider.Reload()
		// adjust pulse's period if in range
		adjustPeriod = true
	}

	return
}

type Envelope struct {
	Start   bool
	Loop    bool
	Divider Divider
	Counter uint8
}

func (envelope *Envelope) Reset() {
	envelope.Start = false
	envelope.Loop = false
	envelope.Divider.Reset()
	envelope.Counter = 0x00
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

type Divider struct {
	Counter  int16
	Period   int16
	PlusOne  bool
	TimesTwo bool
}

func (divider *Divider) Reset() {
	divider.Counter = 0
	divider.Period = 0
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

	if divider.PlusOne {
		divider.Counter++
	}

	if divider.TimesTwo {
		divider.Counter *= 2
	}
}
