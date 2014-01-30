package nintengo

import (
	"github.com/nwidger/rp2ago3"
)

type Button uint8

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
