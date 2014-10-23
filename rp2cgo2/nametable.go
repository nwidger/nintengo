package rp2cgo2

import "github.com/nwidger/nintengo/rp2ago3"

type Nametable struct {
	Tables [4]int
	Memory [2][0x0400]uint8
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
			nametable.Tables[i] = 0
		case 1:
			nametable.Tables[i] = 1
		}
	}
}

func (nametable *Nametable) Reset() {
	for i := range nametable.Memory {
		for j := range nametable.Memory[i] {
			nametable.Memory[i][j] = 0xff
		}
	}

	nametable.SetTables(0, 0, 1, 1)
}

func (nametable *Nametable) Fetch(address uint16) (value uint8) {
	switch {
	// PPU only
	case address >= 0x2000 && address <= 0x2fff:
		table := (address >> 10) & 0x0003
		index := address & 0x03ff
		i := nametable.Tables[table]
		value = nametable.Memory[i][index]
	}

	return
}

func (nametable *Nametable) Store(address uint16, value uint8) (oldValue uint8) {
	// PPU only
	switch {
	case address >= 0x2000 && address <= 0x2fff:
		table := (address >> 10) & 0x0003
		index := address & 0x03ff

		oldValue = nametable.Memory[nametable.Tables[table]][index]
		i := nametable.Tables[table]
		nametable.Memory[i][index] = value
	}

	return
}
