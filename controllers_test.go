package nintengo

import (
	"testing"
)

func TestControllers(t *testing.T) {
	ctrls := NewControllers()

	ctrls.controllers[0].buttons = 0x01

	if ctrls.Fetch(0x4016) != 0x41 {
		t.Error("Memory is not 0x41")
	}

	ctrls.controllers[0].buttons = 0x00

	if ctrls.Fetch(0x4016) != 0x40 {
		t.Error("Memory is not 0x40")
	}

	ctrls.Store(0x4016, 1)
	ctrls.Store(0x4016, 0)

	ctrls.controllers[0].buttons = 0xff
	ctrls.controllers[1].buttons = 0xff

	for i := 0; i < 20; i++ {
		if ctrls.Fetch(0x4016) != 0x41 {
			t.Error("Memory is not 0x41")
		}
	}

	for i := 0; i < 20; i++ {
		if ctrls.Fetch(0x4017) != 0x41 {
			t.Error("Memory is not 0x41")
		}
	}

	ctrls.Store(0x4016, 1)
	ctrls.Store(0x4016, 0)

	ctrls.controllers[0].buttons = 0x00
	ctrls.controllers[1].buttons = 0x00

	for i := 0; i < 20; i++ {
		if i < 8 {
			if ctrls.Fetch(0x4016) != 0x40 {
				t.Error("Memory is not 0x40")
			}
		} else {
			if ctrls.Fetch(0x4016) != 0x41 {
				t.Error("Memory is not 0x41")
			}
		}
	}

	for i := 0; i < 20; i++ {
		if i < 8 {
			if ctrls.Fetch(0x4017) != 0x40 {
				t.Error("Memory is not 0x40")
			}
		} else {
			if ctrls.Fetch(0x4017) != 0x41 {
				t.Error("Memory is not 0x41")
			}
		}
	}
}
