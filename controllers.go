package nintengo

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
)

type Controllers struct {
	strobe      uint8
	controller1 [8]uint8
	controller2 [8]uint8
}

func (ctrls *Controllers) Reset() {

}

func (ctrls *Controllers) Mappings() (fetch, store []uint16) {
	fetch = []uint16{0x4016, 0x4017}
	store = []uint16{0x4016}
	return
}

func (ctrls *Controllers) Fetch(address uint16) (value uint8) {
	switch address {
	case 0x4016:
		value = ctrls.controller1[ctrls.strobe]
	case 0x4017:
		value = ctrls.controller2[ctrls.strobe]
	}

	return
}

func (ctrls *Controllers) Store(address uint16, value uint8) (oldValue uint8) {
	switch address {
	case 0x4016:
		oldValue = ctrls.strobe
		ctrls.strobe = value & 0x01
	}

	return
}
