package rp2cgo2

import "github.com/nwidger/nintengo/rp2ago3"

type Nametable struct {
	Tables [4]*[0x0400]uint8
	Table0 [0x0400]uint8
	Table1 [0x0400]uint8
}

func NewNametable() *Nametable {
	nametable := &Nametable{}
	nametable.Reset()
	return nametable
}

func (nametable *Nametable) Mappings(which rp2ago3.Mapping) (fetch, store []uint16) {
	fetch = []uint16{}
	store = []uint16{}

	switch which {
	case rp2ago3.PPU:
		for i := uint16(0x2000); i <= 0x2fff; i++ {
			fetch = append(fetch, i)
			store = append(store, i)
		}
	}

	return
}

func (nametable *Nametable) SetTables(t0, t1, t2, t3 int) {
	for i, t := range []int{t0, t1, t2, t3} {
		switch t & 0x01 {
		case 0:
			nametable.Tables[i] = &nametable.Table0
		case 1:
			nametable.Tables[i] = &nametable.Table1
		}
	}
}

func (nametable *Nametable) Reset() {
	for i := range nametable.Table0 {
		nametable.Table0[i] = 0xff
	}

	for i := range nametable.Table1 {
		nametable.Table1[i] = 0xff
	}

	nametable.SetTables(0, 0, 1, 1)
}

func (nametable *Nametable) Fetch(address uint16) (value uint8) {
	switch {
	// PPU only
	case address >= 0x2000 && address <= 0x2fff:
		table := (address >> 10) & 0x0003
		index := address & 0x03ff
		value = nametable.Tables[table][index]
	}

	return
}

func (nametable *Nametable) Store(address uint16, value uint8) (oldValue uint8) {
	// PPU only
	switch {
	case address >= 0x2000 && address <= 0x2fff:
		table := (address >> 10) & 0x0003
		index := address & 0x03ff

		oldValue = nametable.Tables[table][index]
		nametable.Tables[table][index] = value
	}

	return
}
