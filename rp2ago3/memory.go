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

const (
	UNMIRRORED uint32 = 0x10000
)

type MappableMemory interface {
	m65go2.Memory
	Mappings(which Mapping) (fetch, store []uint16)
}

type MappedMemory struct {
	mirrors [65536]uint32
	fetch   [65536]m65go2.Memory
	store   [65536]m65go2.Memory
	m65go2.Memory
}

func NewMappedMemory(base m65go2.Memory) *MappedMemory {
	mem := &MappedMemory{
		Memory: base,
	}

	for i := range mem.mirrors {
		mem.mirrors[i] = UNMIRRORED
	}

	return mem
}

func (mem *MappedMemory) AddMirrors(mirrors map[uint16]uint16) (err error) {
	for from, to := range mirrors {
		if from == to {
			err = errors.New("Address cannot be mirrored to itself")
			break
		}

		if mem.mirrors[from] != UNMIRRORED {
			err = errors.New("Address is already mirrored")
			break
		}

		mem.mirrors[from] = uint32(to)
	}

	return
}

func (mem *MappedMemory) AddMappings(mappable MappableMemory, which Mapping) (err error) {
	fetch, store := mappable.Mappings(which)

	for _, address := range fetch {
		if mmap := mem.fetch[address]; mmap != nil {
			err = errors.New("Address is already mapped for fetch")
			return
		}

		mem.fetch[address] = mappable
	}

	for _, address := range store {
		if mmap := mem.store[address]; mmap != nil {
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
		if mapAddress := mem.mirrors[newAddress]; mapAddress == UNMIRRORED {
			break
		} else {
			newAddress = uint16(mapAddress)
		}
	}

	return
}

func (mem *MappedMemory) Fetch(address uint16) (value uint8) {
	address = mem.mirror(address)

	if mmap := mem.fetch[address]; mmap != nil {
		return mmap.Fetch(address)
	}

	return mem.Memory.Fetch(address)
}

func (mem *MappedMemory) Store(address uint16, value uint8) (oldValue uint8) {
	address = mem.mirror(address)

	if mmap := mem.store[address]; mmap != nil {
		return mmap.Store(address, value)
	}

	return mem.Memory.Store(address, value)
}
