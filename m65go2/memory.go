package m65go2

import (
	"io"
	"os"
)

const (
	DEFAULT_MEMORY_SIZE uint32 = 65536
)

// Represents the RAM memory available to the 6502 CPU.  Stores 8-bit
// values using a 16-bit address for a total of 65,536 possible 8-bit
// values.
type Memory interface {
	Reset()                                             // Sets all memory locations to zero
	Fetch(address uint16) (value uint8)                 // Returns the value stored at the given memory address
	Store(address uint16, value uint8) (oldValue uint8) // Stores the value at the given memory address
}

// Represents the 6502 CPU's memory using a static array of uint8's.
type BasicMemory struct {
	m             []uint8
	disableReads  bool
	disableWrites bool
}

// Returns a pointer to a new BasicMemory with all memory initialized
// to zero.
func NewBasicMemory(size uint32) *BasicMemory {
	return &BasicMemory{
		m: make([]uint8, size),
	}
}

func (mem *BasicMemory) DisableReads() {
	mem.disableReads = true
}

func (mem *BasicMemory) EnableReads() {
	mem.disableReads = false
}

func (mem *BasicMemory) DisableWrites() {
	mem.disableWrites = true
}

func (mem *BasicMemory) EnableWrites() {
	mem.disableWrites = false
}

// Resets all memory locations to zero
func (mem *BasicMemory) Reset() {
	for i := range mem.m {
		mem.m[i] = 0x00
	}
}

// Returns the value stored at the given memory address
func (mem *BasicMemory) Fetch(address uint16) (value uint8) {
	if mem.disableReads {
		value = 0xff
	} else {
		value = mem.m[address]
	}

	return
}

// Stores the value at the given memory address
func (mem *BasicMemory) Store(address uint16, value uint8) (oldValue uint8) {
	if !mem.disableWrites {
		oldValue = mem.m[address]
		mem.m[address] = value
	}

	return
}

func (mem *BasicMemory) load(path string) {
	fi, err := os.Open(path)

	if err != nil {
		panic(err)
	}

	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()

	total := 0
	buf := make([]byte, 65536)

	for {
		n, err := fi.Read(buf)

		if err != nil && err != io.EOF {
			panic(err)
		}

		if n == 0 {
			break
		}

		total++
	}

	j := 0xc000

	for i, b := range buf {
		if i <= 15 {
			continue
		}

		mem.m[j] = b
		j++

		if j == 0xffff {
			break
		}
	}

	return
}

// Returns true iff the two addresses are located in the same page in
// memory.  Two addresses are on the same page if their high bytes are
// both the same, i.e. 0x0101 and 0x0103 are on the same page but
// 0x0101 and 0203 are not.
func SamePage(addr1 uint16, addr2 uint16) bool {
	return (addr1^addr2)>>8 == 0
}
