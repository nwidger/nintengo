package rp2ago3

import "testing"

var apu *APU

func Setup() {
	apu = NewAPU(nil)
	apu.Reset()
}

func Teardown() {

}

type PulseTest struct {
	flag     PulseFlag
	index    int
	value    uint8
	expected uint8
}

func TestPulse1Channel(t *testing.T) {
	Setup()

	address := uint16(0x4000)

	for i := range apu.Pulse1.Registers {
		apu.Pulse1.Registers[i] = 0x00
		apu.Store(address+uint16(i), 0xff)

		if apu.Pulse1.Registers[i] != 0xff {
			t.Error("Register is not 0xff")
		}

		apu.Pulse1.Registers[i] = 0xff
		value := apu.Fetch(address + uint16(i))

		if value != 0x00 {
			t.Error("Value is not 0x00")
		}
	}

	pts := []PulseTest{}

	pts = append(pts, PulseTest{Duty, 0, 0xc0, 0x03})
	pts = append(pts, PulseTest{Duty, 0, 0x3f, 0x00})

	pts = append(pts, PulseTest{PulseEnvelopeLoopLengthCounterHalt, 0, 0x20, 0x01})
	pts = append(pts, PulseTest{PulseEnvelopeLoopLengthCounterHalt, 0, 0xdf, 0x00})

	pts = append(pts, PulseTest{PulseConstantVolume, 0, 0x10, 0x01})
	pts = append(pts, PulseTest{PulseConstantVolume, 0, 0xef, 0x00})

	pts = append(pts, PulseTest{PulseVolumeEnvelope, 0, 0x0f, 0x0f})
	pts = append(pts, PulseTest{PulseVolumeEnvelope, 0, 0xf0, 0x00})

	pts = append(pts, PulseTest{SweepEnabled, 1, 0x80, 0x01})
	pts = append(pts, PulseTest{SweepEnabled, 1, 0x7f, 0x00})

	pts = append(pts, PulseTest{SweepPeriod, 1, 0x70, 0x07})
	pts = append(pts, PulseTest{SweepPeriod, 1, 0x8f, 0x00})

	pts = append(pts, PulseTest{SweepNegate, 1, 0x08, 0x01})
	pts = append(pts, PulseTest{SweepNegate, 1, 0xf7, 0x00})

	pts = append(pts, PulseTest{SweepShift, 1, 0x07, 0x07})
	pts = append(pts, PulseTest{SweepShift, 1, 0xf8, 0x00})

	pts = append(pts, PulseTest{PulseTimerLow, 2, 0xff, 0xff})
	pts = append(pts, PulseTest{PulseTimerLow, 2, 0x00, 0x00})

	pts = append(pts, PulseTest{PulseLengthCounterLoad, 3, 0xf8, 0x1f})
	pts = append(pts, PulseTest{PulseLengthCounterLoad, 3, 0x07, 0x00})

	pts = append(pts, PulseTest{PulseTimerHigh, 3, 0x07, 0x07})
	pts = append(pts, PulseTest{PulseTimerHigh, 3, 0xf8, 0x00})

	for _, pt := range pts {
		apu.Pulse1.Registers[pt.index] = pt.value

		actual := apu.Pulse1.registers(pt.flag)
		expected := pt.expected

		if actual != expected {
			t.Errorf("Value is %02x not %02x\n", actual, expected)
		}
	}

	Teardown()
}

func TestPulse2Channel(t *testing.T) {
	Setup()

	address := uint16(0x4004)

	for i := range apu.Pulse2.Registers {
		apu.Pulse2.Registers[i] = 0x00
		apu.Store(address+uint16(i), 0xff)

		if apu.Pulse2.Registers[i] != 0xff {
			t.Error("Register is not 0xff")
		}

		apu.Pulse2.Registers[i] = 0xff
		value := apu.Fetch(address + uint16(i))

		if value != 0x00 {
			t.Error("Value is not 0x00")
		}
	}

	Teardown()
}

type TriangleTest struct {
	flag     TriangleFlag
	index    int
	value    uint8
	expected uint8
}

func TestTriangleChannel(t *testing.T) {
	Setup()

	var address uint16

	for i := range apu.Triangle.Registers {
		switch i {
		case 0:
			address = 0x4008
		case 1:
			address = 0x400a
		case 2:
			address = 0x400b
		}

		apu.Triangle.Registers[i] = 0x00
		apu.Store(address, 0xff)

		if apu.Triangle.Registers[i] != 0xff {
			t.Error("Register is not 0xff")
		}

		apu.Triangle.Registers[i] = 0xff
		value := apu.Fetch(address)

		if value != 0x00 {
			t.Error("Value is not 0x00")
		}
	}

	pts := []TriangleTest{}

	pts = append(pts, TriangleTest{LengthCounterHaltLinearCounterControl, 0, 0x80, 0x01})
	pts = append(pts, TriangleTest{LengthCounterHaltLinearCounterControl, 0, 0x7f, 0x00})

	pts = append(pts, TriangleTest{LinearCounterLoad, 0, 0x7f, 0x7f})
	pts = append(pts, TriangleTest{LinearCounterLoad, 0, 0x80, 0x00})

	pts = append(pts, TriangleTest{TriangleTimerLow, 1, 0xff, 0xff})
	pts = append(pts, TriangleTest{TriangleTimerLow, 1, 0x00, 0x00})

	pts = append(pts, TriangleTest{TriangleLengthCounterLoad, 2, 0xf8, 0x1f})
	pts = append(pts, TriangleTest{TriangleLengthCounterLoad, 2, 0x07, 0x00})

	pts = append(pts, TriangleTest{TriangleTimerHigh, 2, 0x07, 0x07})
	pts = append(pts, TriangleTest{TriangleTimerHigh, 2, 0xf8, 0x00})

	for _, pt := range pts {
		apu.Triangle.Registers[pt.index] = pt.value

		actual := apu.Triangle.registers(pt.flag)
		expected := pt.expected

		if actual != expected {
			t.Errorf("Value is %02x not %02x\n", actual, expected)
		}
	}

	Teardown()
}

type NoiseTest struct {
	flag     NoiseFlag
	index    int
	value    uint8
	expected uint8
}

func TestNoiseChannel(t *testing.T) {
	Setup()

	var address uint16

	for i := range apu.Noise.Registers {
		switch i {
		case 0:
			address = 0x400c
		case 1:
			address = 0x400e
		case 2:
			address = 0x400f
		}

		apu.Noise.Registers[i] = 0x00
		apu.Store(address, 0xff)

		if apu.Noise.Registers[i] != 0xff {
			t.Error("Register is not 0xff")
		}

		apu.Noise.Registers[i] = 0xff
		value := apu.Fetch(address)

		if value != 0x00 {
			t.Error("Value is not 0x00")
		}
	}

	pts := []NoiseTest{}

	pts = append(pts, NoiseTest{NoiseEnvelopeLoopLengthCounterHalt, 0, 0x20, 0x01})
	pts = append(pts, NoiseTest{NoiseEnvelopeLoopLengthCounterHalt, 0, 0xdf, 0x00})

	pts = append(pts, NoiseTest{NoiseConstantVolume, 0, 0x10, 0x01})
	pts = append(pts, NoiseTest{NoiseConstantVolume, 0, 0xef, 0x00})

	pts = append(pts, NoiseTest{NoiseVolumeEnvelope, 0, 0x0f, 0x0f})
	pts = append(pts, NoiseTest{NoiseVolumeEnvelope, 0, 0xf0, 0x00})

	pts = append(pts, NoiseTest{LoopNoise, 1, 0x80, 0x01})
	pts = append(pts, NoiseTest{LoopNoise, 1, 0x7f, 0x00})

	pts = append(pts, NoiseTest{NoisePeriod, 1, 0x0f, 0x0f})
	pts = append(pts, NoiseTest{NoisePeriod, 1, 0xf0, 0x00})

	pts = append(pts, NoiseTest{NoiseLengthCounterLoad, 2, 0xf8, 0x1f})
	pts = append(pts, NoiseTest{NoiseLengthCounterLoad, 2, 0x07, 0x00})

	for _, pt := range pts {
		apu.Noise.Registers[pt.index] = pt.value

		actual := apu.Noise.registers(pt.flag)
		expected := pt.expected

		if actual != expected {
			t.Errorf("Value is %02x not %02x\n", actual, expected)
		}
	}

	Teardown()
}

type DMCTest struct {
	flag     DMCFlag
	index    int
	value    uint8
	expected uint8
}

func TestDMCChannel(t *testing.T) {
	Setup()

	address := uint16(0x4010)

	for i := range apu.DMC.Registers {
		apu.DMC.Registers[i] = 0x00
		apu.Store(address+uint16(i), 0xff)

		if apu.DMC.Registers[i] != 0xff {
			t.Error("Register is not 0xff")
		}

		apu.DMC.Registers[i] = 0xff
		value := apu.Fetch(address + uint16(i))

		if value != 0x00 {
			t.Error("Value is not 0x00")
		}
	}

	pts := []DMCTest{}

	pts = append(pts, DMCTest{IRQEnable, 0, 0x80, 0x01})
	pts = append(pts, DMCTest{IRQEnable, 0, 0x7f, 0x00})

	pts = append(pts, DMCTest{Loop, 0, 0x40, 0x01})
	pts = append(pts, DMCTest{Loop, 0, 0xbf, 0x00})

	pts = append(pts, DMCTest{Frequency, 0, 0x0f, 0x0f})
	pts = append(pts, DMCTest{Frequency, 0, 0xf0, 0x00})

	pts = append(pts, DMCTest{LoadCounter, 1, 0x7f, 0x7f})
	pts = append(pts, DMCTest{LoadCounter, 1, 0x80, 0x00})

	pts = append(pts, DMCTest{SampleAddress, 2, 0xff, 0xff})
	pts = append(pts, DMCTest{SampleAddress, 2, 0x00, 0x00})

	pts = append(pts, DMCTest{SampleLength, 3, 0xff, 0xff})
	pts = append(pts, DMCTest{SampleLength, 3, 0x00, 0x00})

	for _, pt := range pts {
		apu.DMC.Registers[pt.index] = pt.value

		actual := apu.DMC.registers(pt.flag)
		expected := pt.expected

		if actual != expected {
			t.Errorf("Value is %02x not %02x\n", actual, expected)
		}
	}

	Teardown()
}

type ControlTest struct {
	flag     ControlFlag
	value    uint8
	expected bool
}

func TestControl(t *testing.T) {
	Setup()

	address := uint16(0x4015)

	apu.Registers.Control = 0x00
	apu.Store(address, 0xff)

	if apu.Registers.Control != 0xff {
		t.Error("Register is not 0xff")
	}

	apu.Registers.Control = 0x00
	apu.Registers.Status = 0xff

	value := apu.Fetch(address)

	if value != 0xff {
		t.Error("Value is not 0xff")
	}

	pts := []ControlTest{}

	pts = append(pts, ControlTest{EnableDMC, 0x10, true})
	pts = append(pts, ControlTest{EnableDMC, 0xef, false})

	pts = append(pts, ControlTest{EnableNoise, 0x08, true})
	pts = append(pts, ControlTest{EnableNoise, 0xf7, false})

	pts = append(pts, ControlTest{EnableTriangle, 0x04, true})
	pts = append(pts, ControlTest{EnableTriangle, 0xfb, false})

	pts = append(pts, ControlTest{EnablePulseChannel2, 0x02, true})
	pts = append(pts, ControlTest{EnablePulseChannel2, 0xfd, false})

	pts = append(pts, ControlTest{EnablePulseChannel1, 0x01, true})
	pts = append(pts, ControlTest{EnablePulseChannel1, 0xfe, false})

	for _, pt := range pts {
		apu.Registers.Control = Control(pt.value)

		actual := apu.control(pt.flag)
		expected := pt.expected

		if actual != expected {
			t.Errorf("Value is %v not %v\n", actual, expected)
		}
	}

	Teardown()
}

type StatusTest struct {
	flag     StatusFlag
	value    uint8
	expected bool
}

func TestStatus(t *testing.T) {
	Setup()

	address := uint16(0x4015)

	apu.Registers.Status = 0xff
	value := apu.Fetch(address)

	if value != 0xff {
		t.Error("Value is not 0xff")
	}

	apu.Registers.Status = 0xff
	apu.Store(address, 0x00)

	if apu.Registers.Status != 0x7f {
		t.Error("Register is not 0x7f")
	}

	pts := []StatusTest{}

	pts = append(pts, StatusTest{DMCInterrupt, 0x80, true})
	pts = append(pts, StatusTest{DMCInterrupt, 0x7f, false})

	pts = append(pts, StatusTest{FrameInterrupt, 0x40, true})
	pts = append(pts, StatusTest{FrameInterrupt, 0xbf, false})

	pts = append(pts, StatusTest{DMCActive, 0x10, true})
	pts = append(pts, StatusTest{DMCActive, 0xef, false})

	pts = append(pts, StatusTest{NoiseLengthCounterNotZero, 0x08, true})
	pts = append(pts, StatusTest{NoiseLengthCounterNotZero, 0xf7, false})

	pts = append(pts, StatusTest{TriangleLengthCounterNotZero, 0x04, true})
	pts = append(pts, StatusTest{TriangleLengthCounterNotZero, 0xfb, false})

	pts = append(pts, StatusTest{Pulse2LengthCounterNotZero, 0x02, true})
	pts = append(pts, StatusTest{Pulse2LengthCounterNotZero, 0xfd, false})

	pts = append(pts, StatusTest{Pulse1LengthCounterNotZero, 0x01, true})
	pts = append(pts, StatusTest{Pulse1LengthCounterNotZero, 0xfe, false})

	for _, pt := range pts {
		apu.Registers.Status = Status(pt.value)

		actual := apu.status(pt.flag)
		expected := pt.expected

		if actual != expected {
			t.Errorf("Value is %v not %v\n", actual, expected)
		}
	}

	Teardown()
}

type FrameCounterTest struct {
	flag     FrameCounterFlag
	value    uint8
	expected uint8
}

func TestFrameCounter(t *testing.T) {
	Setup()

	address := uint16(0x4017)

	apu.FrameCounter.Register = 0xff
	apu.Store(address, 0x00)

	if apu.FrameCounter.Register != 0x00 {
		t.Error("Register is not 0x00")
	}

	pts := []FrameCounterTest{}

	pts = append(pts, FrameCounterTest{Mode, 0x80, 5})
	pts = append(pts, FrameCounterTest{Mode, 0x7f, 4})

	pts = append(pts, FrameCounterTest{IRQInhibit, 0x40, 0x01})
	pts = append(pts, FrameCounterTest{IRQInhibit, 0xbf, 0x00})

	for _, pt := range pts {
		apu.FrameCounter.Register = pt.value

		actual := apu.FrameCounter.register(pt.flag)
		expected := pt.expected

		if actual != expected {
			t.Errorf("Value is %02x not %02x\n", actual, expected)
		}
	}

	Teardown()
}
