package rp2ago3

import (
	"testing"

	"github.com/nwidger/m65go2"
)

var apu *APU
var master *m65go2.Clock

func Setup() {
	apu = NewAPU()
	apu.Reset()
}

func Teardown() {

}

func TestPulse1Channel(t *testing.T) {
	Setup()

	address := uint16(0x4000)

	for i := range apu.Registers.Pulse1 {
		apu.Registers.Pulse1[i] = 0x00
		apu.Store(address+uint16(i), 0xff)

		if apu.Registers.Pulse1[i] != 0xff {
			t.Error("Register is not 0xff")
		}

		apu.Registers.Pulse1[i] = 0xff
		value := apu.Fetch(address + uint16(i))

		if value != 0x00 {
			t.Error("Value is not 0x00")
		}
	}

	Teardown()
}

func TestPulse2Channel(t *testing.T) {
	Setup()

	address := uint16(0x4004)

	for i := range apu.Registers.Pulse2 {
		apu.Registers.Pulse2[i] = 0x00
		apu.Store(address+uint16(i), 0xff)

		if apu.Registers.Pulse2[i] != 0xff {
			t.Error("Register is not 0xff")
		}

		apu.Registers.Pulse2[i] = 0xff
		value := apu.Fetch(address + uint16(i))

		if value != 0x00 {
			t.Error("Value is not 0x00")
		}
	}

	Teardown()
}

func TestTriangleChannel(t *testing.T) {
	Setup()

	var address uint16

	for i := range apu.Registers.Triangle {
		switch i {
		case 0:
			address = 0x4008
		case 1:
			address = 0x400a
		case 2:
			address = 0x400b
		}

		apu.Registers.Triangle[i] = 0x00
		apu.Store(address, 0xff)

		if apu.Registers.Triangle[i] != 0xff {
			t.Error("Register is not 0xff")
		}

		apu.Registers.Triangle[i] = 0xff
		value := apu.Fetch(address)

		if value != 0x00 {
			t.Error("Value is not 0x00")
		}
	}

	Teardown()
}

func TestNoiseChannel(t *testing.T) {
	Setup()

	var address uint16

	for i := range apu.Registers.Noise {
		switch i {
		case 0:
			address = 0x400c
		case 1:
			address = 0x400e
		case 2:
			address = 0x400f
		}

		apu.Registers.Noise[i] = 0x00
		apu.Store(address, 0xff)

		if apu.Registers.Noise[i] != 0xff {
			t.Error("Register is not 0xff")
		}

		apu.Registers.Noise[i] = 0xff
		value := apu.Fetch(address)

		if value != 0x00 {
			t.Error("Value is not 0x00")
		}
	}

	Teardown()
}

func TestDmcChannel(t *testing.T) {
	Setup()

	address := uint16(0x4010)

	for i := range apu.Registers.Dmc {
		apu.Registers.Dmc[i] = 0x00
		apu.Store(address+uint16(i), 0xff)

		if apu.Registers.Dmc[i] != 0xff {
			t.Error("Register is not 0xff")
		}

		apu.Registers.Dmc[i] = 0xff
		value := apu.Fetch(address + uint16(i))

		if value != 0x00 {
			t.Error("Value is not 0x00")
		}
	}

	Teardown()
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

	Teardown()
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

	if apu.Registers.Status != 0xff {
		t.Error("Register is not 0xff")
	}

	Teardown()
}
