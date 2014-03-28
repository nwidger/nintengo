package nes

import "github.com/nwidger/nintengo/rp2ago3"

type Button uint8

func (btn Button) String() string {
	switch btn {
	case A:
		return "A"
	case B:
		return "B"
	case Select:
		return "Select"
	case Start:
		return "Start"
	case Up:
		return "Up"
	case Down:
		return "Down"
	case Left:
		return "Left"
	case Right:
		return "Right"
	default:
		return "Unknown"
	}
}

func (btn Button) Valid() bool {
	switch btn {
	case A:
		fallthrough
	case B:
		fallthrough
	case Select:
		fallthrough
	case Start:
		fallthrough
	case Up:
		fallthrough
	case Down:
		fallthrough
	case Left:
		fallthrough
	case Right:
		return true
	}

	return false
}

const (
	A = iota
	B
	Select
	Start
	Up
	Down
	Left
	Right
	One
)

type Controller struct {
	strobe  Button
	buttons uint8
}

type Controllers struct {
	last        uint8
	controllers [2]Controller
	input       chan PressButton
}

type PressButton struct {
	controller int
	down       bool
	button     Button
}

func NewControllers() *Controllers {
	return &Controllers{
		input: make(chan PressButton),
	}
}

func (ctrls *Controllers) Reset() {
	for i := range ctrls.controllers {
		ctrls.controllers[i].strobe = A
		ctrls.controllers[i].buttons = 0
	}
}

func (ctrls *Controllers) Input() chan PressButton {
	return ctrls.input
}

func (ctrls *Controllers) Mappings(which rp2ago3.Mapping) (fetch, store []uint16) {
	switch which {
	case rp2ago3.CPU:
		fetch = []uint16{0x4016, 0x4017}
		store = []uint16{0x4016}
	}

	return
}

func (ctrls *Controllers) Fetch(address uint16) (value uint8) {
	switch address {
	case 0x4016:
		fallthrough
	case 0x4017:
		index := address - 0x4016
		ctrl := &ctrls.controllers[index]

		if ctrl.strobe == One {
			value = 1
		} else {
			value = (ctrl.buttons >> ctrl.strobe) & 0x01
			ctrl.strobe++
		}

		value |= 0x40
	}

	return
}

func (ctrls *Controllers) Store(address uint16, value uint8) (oldValue uint8) {
	switch address {
	case 0x4016:
		oldValue = ctrls.last
		ctrls.last = value & 0x01

		if oldValue == 1 && value == 0 {
			for i := range ctrls.controllers {
				ctrls.controllers[i].strobe = A
			}
		}
	}

	return
}

func (ctrls *Controllers) KeyIsDown(controller int, btn Button) bool {
	return ctrls.controllers[controller].buttons&(1<<btn) != 0
}

func (ctrls *Controllers) ValidKeyDown(controller int, btn Button) (valid bool) {
	valid = btn.Valid()

	switch {
	case btn == Up && ctrls.KeyIsDown(controller, Down):
		fallthrough
	case btn == Down && ctrls.KeyIsDown(controller, Up):
		fallthrough
	case btn == Left && ctrls.KeyIsDown(controller, Right):
		fallthrough
	case btn == Right && ctrls.KeyIsDown(controller, Left):
		valid = false
	}

	return
}

func (ctrls *Controllers) KeyDown(controller int, btn Button) {
	if ctrls.ValidKeyDown(controller, btn) {
		ctrls.controllers[controller].buttons |= (1 << uint8(btn))
	}
}

func (ctrls *Controllers) KeyUp(controller int, btn Button) {
	if btn.Valid() {
		ctrls.controllers[controller].buttons &^= (1 << uint8(btn))
	}
}

func (ctrls *Controllers) Run() {
	for {
		select {
		case e := <-ctrls.input:
			if e.down {
				ctrls.KeyDown(e.controller, e.button)
			} else {
				ctrls.KeyUp(e.controller, e.button)
			}
		}
	}
}
