package rp2ago3

import (
	"errors"

	"github.com/nwidger/nintengo/m65go2"
)

type Mapping uint8

const (
	CPU Mapping = iota
	PPU
)

type MappableMemory interface {
	m65go2.Memory
	Mappings(which Mapping) (fetch, store []uint16)
}

type MappedMemory struct {
	mirrors map[uint16]uint16
	fetch   map[uint16]m65go2.Memory
	store   map[uint16]m65go2.Memory
	m65go2.Memory
}

func NewMappedMemory(base m65go2.Memory) *MappedMemory {
	return &MappedMemory{
		mirrors: make(map[uint16]uint16),
		fetch:   make(map[uint16]m65go2.Memory),
		store:   make(map[uint16]m65go2.Memory),
		Memory:  base,
	}
}

func (mem *MappedMemory) AddMirrors(mirrors map[uint16]uint16) (err error) {
	for from, to := range mirrors {
		if from == to {
			err = errors.New("Address cannot be mirrored to itself")
			break
		}

		if _, ok := mem.mirrors[from]; ok {
			err = errors.New("Address is already mirrored")
			break
		}

		mem.mirrors[from] = to
	}

	return
}

func (mem *MappedMemory) AddMappings(mappable MappableMemory, which Mapping) (err error) {
	fetch, store := mappable.Mappings(which)

	for _, address := range fetch {
		if _, ok := mem.fetch[address]; ok {
			err = errors.New("Address is already mapped for fetch")
			return
		}

		mem.fetch[address] = mappable
	}

	for _, address := range store {
		if _, ok := mem.store[address]; ok {
			err = errors.New("Address is already mapped for store")
			return
		}

		mem.store[address] = mappable
	}

	return
}

func (mem *MappedMemory) Reset() {
	// don't clear mappings
	mem.Memory.Reset()
}

func (mem *MappedMemory) mirror(address uint16) (newAddress uint16) {
	newAddress = address

	for {
		if mapAddress, ok := mem.mirrors[newAddress]; !ok {
			break
		} else {
			newAddress = mapAddress
		}
	}

	return
}

func (mem *MappedMemory) Fetch(address uint16) (value uint8) {
	address = mem.mirror(address)

	if mmap, ok := mem.fetch[address]; ok {
		return mmap.Fetch(address)
	}

	return mem.Memory.Fetch(address)
}

func (mem *MappedMemory) Store(address uint16, value uint8) (oldValue uint8) {
	address = mem.mirror(address)

	if mmap, ok := mem.store[address]; ok {
		return mmap.Store(address, value)
	}

	return mem.Memory.Store(address, value)
}
