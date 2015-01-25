package nes

import "github.com/kaicheng/nintengo/rp2ago3"

//go:generate stringer -type=Button
type Button uint8

const (
	A Button = iota
	B
	Select
	Start
	Up
	Down
	Left
	Right
	One
)

func (btn Button) Valid() bool {
	switch btn {
	case A, B, Select, Start, Up, Down, Left, Right:
		return true
	}

	return false
}

type Controller struct {
	strobe  Button
	buttons uint8
}

type Controllers struct {
	last        uint8
	controllers [2]Controller
}

func NewControllers() *Controllers {
	return &Controllers{}
}

func (ctrls *Controllers) Reset() {
	for i := range ctrls.controllers {
		ctrls.controllers[i].strobe = A
		ctrls.controllers[i].buttons = 0
	}
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
	case 0x4016, 0x4017:
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
	case btn == Up && ctrls.KeyIsDown(controller, Down),
		btn == Down && ctrls.KeyIsDown(controller, Up),
		btn == Left && ctrls.KeyIsDown(controller, Right),
		btn == Right && ctrls.KeyIsDown(controller, Left):
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
