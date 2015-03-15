package m65go2

const (
	DefaultMemorySize uint32 = 65536
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
	M             []uint8
	DisableReads  bool
	DisableWrites bool
}

// Returns a pointer to a new BasicMemory with all memory initialized
// to zero.
func NewBasicMemory(size uint32) *BasicMemory {
	return &BasicMemory{
		M: make([]uint8, size),
	}
}

// Resets all memory locations to zero
func (mem *BasicMemory) Reset() {
	for i := range mem.M {
		mem.M[i] = 0xff
	}
}

// Returns the value stored at the given memory address
func (mem *BasicMemory) Fetch(address uint16) (value uint8) {
	if mem.DisableReads {
		value = 0xff
	} else {
		value = mem.M[address]
	}

	return
}

// Stores the value at the given memory address
func (mem *BasicMemory) Store(address uint16, value uint8) (oldValue uint8) {
	if !mem.DisableWrites {
		oldValue = mem.M[address]
		mem.M[address] = value
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
