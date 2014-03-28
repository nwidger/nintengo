// Package m65go2 simulates the MOS 6502 CPU
package m65go2

import (
	"fmt"
	"strings"
)

// Flags used by P (Status) register
type Status uint8

const (
	C Status = 1 << iota // carry flag
	Z                    // zero flag
	I                    // interrupt disable
	D                    // decimal mode
	B                    // break command
	U                    // -UNUSED-
	V                    // overflow flag
	N                    // negative flag
)

// The 6502's registers, all registers are 8-bit values except for PC
// which is 16-bits.
type Registers struct {
	A  uint8  // accumulator
	X  uint8  // index register X
	Y  uint8  // index register Y
	P  Status // processor status
	SP uint8  // stack pointer
	PC uint16 // program counter
}

// Creates a new set of Registers.  All registers are initialized to
// 0.
func NewRegisters() (reg Registers) {
	reg = Registers{}
	reg.Reset()
	return
}

// Resets all registers.  Register P is initialized with only the I
// bit set, SP is initialized to 0xfd, PC is initialized to 0xfffc
// (the RESET vector) and all other registers are initialized to 0.
func (reg *Registers) Reset() {
	reg.A = 0
	reg.X = 0
	reg.Y = 0
	reg.P = I
	reg.SP = 0xfd
	reg.PC = 0xfffc
}

// Prints the values of each register to os.Stderr.
func (reg *Registers) String() string {
	return fmt.Sprintf("A:%02X X:%02X Y:%02X P:%02X SP:%02X", reg.A, reg.X, reg.Y, reg.P, reg.SP)
}

type Interrupt uint8

const (
	Irq Interrupt = iota
	Nmi
	Rst
)

type Index uint8

const (
	X Index = iota
	Y
)

type decode struct {
	enabled     bool
	pc          uint16
	opcode      OpCode
	args        string
	mneumonic   string
	decodedArgs string
	registers   string
	ticks       uint64
}

func (d *decode) String() string {
	return fmt.Sprintf("%04X  %02X %-5s %4s %-26s  %25s",
		d.pc, d.opcode, d.args, d.mneumonic, d.decodedArgs, d.registers)
}

// Represents the 6502 CPU.
type M6502 struct {
	decode       decode
	Nmi          bool
	Irq          bool
	Rst          bool
	Registers    Registers
	Memory       Memory
	Instructions InstructionTable
	decimalMode  bool
	breakError   bool
	Cycles       chan uint16
}

// Returns a pointer to a new CPU with the given Memory.
func NewM6502(mem Memory) *M6502 {
	instructions := NewInstructionTable()
	instructions.InitInstructions()

	return &M6502{
		decode:       decode{},
		Registers:    NewRegisters(),
		Memory:       mem,
		Instructions: instructions,
		decimalMode:  true,
		breakError:   false,
		Nmi:          false,
		Irq:          false,
		Rst:          false,
		Cycles:       make(chan uint16),
	}
}

// Resets the CPU by resetting both the registers and memory.
func (cpu *M6502) Reset() {
	cpu.Registers.Reset()
	cpu.Memory.Reset()
	cpu.PerformRst()
}

func (cpu *M6502) Interrupt(which Interrupt, state bool) {
	switch which {
	case Irq:
		cpu.Irq = state
	case Nmi:
		cpu.Nmi = state
	case Rst:
		cpu.Rst = state
	}
}

func (cpu *M6502) InterruptLine(which Interrupt) func(state bool) {
	return func(state bool) {
		if cpu != nil {
			cpu.Interrupt(which, state)
		}
	}
}

func (cpu *M6502) GetInterrupt(which Interrupt) (state bool) {
	switch which {
	case Irq:
		state = cpu.Irq
	case Nmi:
		state = cpu.Nmi
	case Rst:
		state = cpu.Rst
	}

	return
}

func (cpu *M6502) PerformInterrupts() {
	// check interrupts
	switch {
	case cpu.Irq && cpu.Registers.P&I == 0:
		cpu.PerformIrq()
		cpu.Irq = false
	case cpu.Nmi:
		cpu.PerformNmi()
		cpu.Nmi = false
	case cpu.Rst:
		cpu.PerformRst()
		cpu.Rst = false
	}
}

func (cpu *M6502) PerformIrq() {
	cpu.push16(cpu.Registers.PC)
	cpu.push(uint8(cpu.Registers.P))

	low := cpu.Memory.Fetch(0xfffe)
	high := cpu.Memory.Fetch(0xffff)

	cpu.Registers.PC = (uint16(high) << 8) | uint16(low)
}

func (cpu *M6502) PerformNmi() {
	cpu.push16(cpu.Registers.PC)
	cpu.push(uint8(cpu.Registers.P))

	low := cpu.Memory.Fetch(0xfffa)
	high := cpu.Memory.Fetch(0xfffb)

	cpu.Registers.PC = (uint16(high) << 8) | uint16(low)
}

func (cpu *M6502) PerformRst() {
	low := cpu.Memory.Fetch(0xfffc)
	high := cpu.Memory.Fetch(0xfffd)

	cpu.Registers.PC = (uint16(high) << 8) | uint16(low)
}

func (cpu *M6502) DisableDecimalMode() {
	cpu.decimalMode = false
}

func (cpu *M6502) EnableDecode() {
	cpu.decode.enabled = true
}

// Error type used to indicate that the CPU attempted to execute an
// invalid opcode
type BadOpCodeError OpCode

func (b BadOpCodeError) Error() string {
	return fmt.Sprintf("No such opcode %#02x", b)
}

// Error type used to indicate that the CPU executed a BRK instruction
type BrkOpCodeError OpCode

func (b BrkOpCodeError) Error() string {
	return fmt.Sprintf("Executed BRK opcode")
}

// Executes the instruction pointed to by the PC register in the
// number of cycles as returned by the instruction's Exec function.
// Returns the number of cycles executed and any error (such as
// BadOpCodeError).
func (cpu *M6502) Execute() (cycles uint16, error error) {
	// check interrupts
	cpu.PerformInterrupts()

	// fetch
	opcode := OpCode(cpu.Memory.Fetch(cpu.Registers.PC))
	inst, ok := cpu.Instructions[opcode]

	if !ok {
		return 0, BadOpCodeError(opcode)
	}

	// execute
	if cpu.decode.enabled {
		cpu.decode.pc = cpu.Registers.PC
		cpu.decode.opcode = opcode
		cpu.decode.args = ""
		cpu.decode.mneumonic = inst.Mneumonic
		cpu.decode.decodedArgs = ""
		cpu.decode.registers = cpu.Registers.String()
	}

	cpu.Registers.PC++
	cycles = inst.Exec(cpu)

	if cpu.decode.enabled {
		fmt.Println(cpu.decode.String())
	}

	if cpu.breakError && opcode == 0x00 {
		return cycles, BrkOpCodeError(opcode)
	}

	return cycles, nil
}

// Executes instruction until Execute() returns an error.
func (cpu *M6502) Run() (err error) {
	var cycles uint16

	for {
		if cycles, err = cpu.Execute(); err != nil {
			return
		}

		if cpu.Cycles != nil && cycles != 0 {
			cpu.Cycles <- cycles
			<-cpu.Cycles
		}
	}
}

func (cpu *M6502) setZFlag(value uint8) uint8 {
	if value == 0 {
		cpu.Registers.P |= Z
	} else {
		cpu.Registers.P &= ^Z
	}

	return value
}

func (cpu *M6502) setNFlag(value uint8) uint8 {
	cpu.Registers.P = (cpu.Registers.P & ^N) | Status(value&uint8(N))
	return value
}

func (cpu *M6502) setZNFlags(value uint8) uint8 {
	cpu.setZFlag(value)
	cpu.setNFlag(value)
	return value
}

func (cpu *M6502) setCFlagAddition(value uint16) uint16 {
	cpu.Registers.P = (cpu.Registers.P & ^C) | Status(value>>8&uint16(C))
	return value
}

func (cpu *M6502) setVFlagAddition(term1 uint16, term2 uint16, result uint16) uint16 {
	cpu.Registers.P = (cpu.Registers.P & ^V) | Status((^(term1^term2)&(term1^result)&uint16(N))>>1)
	return result
}

func (cpu *M6502) controlAddress(opcode OpCode, cycles *uint16) (address uint16) {
	// control opcodes end with 00

	if opcode&0x10 == 0 {
		switch (opcode >> 2) & 0x03 {
		case 0x00:
			*cycles = 2
			address = cpu.immediateAddress()
		case 0x01:
			*cycles = 3
			address = cpu.zeroPageAddress()
		case 0x02:
			*cycles = 4
			address = 0 // not used
		case 0x03:
			*cycles = 4
			address = cpu.absoluteAddress()
		}
	} else {
		switch (opcode >> 2) & 0x03 {
		case 0x00:
			*cycles = 2
			address = cpu.relativeAddress()
		case 0x01:
			*cycles = 4
			address = cpu.zeroPageIndexedAddress(X)
		case 0x02:
			*cycles = 2
			address = 0 // not used
		case 0x03:
			*cycles = 4
			address = cpu.absoluteIndexedAddress(X, cycles)
		}
	}

	return
}

func (cpu *M6502) aluAddress(opcode OpCode, cycles *uint16) (address uint16) {
	// alu opcodes end with 01

	if opcode&0x10 == 0 {
		switch (opcode >> 2) & 0x03 {
		case 0x00:
			*cycles = 6
			address = cpu.indexedIndirectAddress()
		case 0x01:
			*cycles = 3
			address = cpu.zeroPageAddress()
		case 0x02:
			*cycles = 2
			address = cpu.immediateAddress()
		case 0x03:
			*cycles = 4
			address = cpu.absoluteAddress()
		}
	} else {
		switch (opcode >> 2) & 0x03 {
		case 0x00:
			*cycles = 5
			address = cpu.indirectIndexedAddress(cycles)
		case 0x01:
			*cycles = 4
			address = cpu.zeroPageIndexedAddress(X)
		case 0x02:
			*cycles = 4
			address = cpu.absoluteIndexedAddress(Y, cycles)
		case 0x03:
			*cycles = 4
			address = cpu.absoluteIndexedAddress(X, cycles)
		}
	}

	return
}

func (cpu *M6502) rmwAddress(opcode OpCode, cycles *uint16) (address uint16) {
	// rmw opcodes end with 10
	var index Index

	if opcode&0x10 == 0 {
		switch (opcode >> 2) & 0x03 {
		case 0x00:
			*cycles = 2
			address = cpu.immediateAddress()
		case 0x01:
			*cycles = 3
			address = cpu.zeroPageAddress()
		case 0x02:
			*cycles = 2
			address = 0 // not used
		case 0x03:
			*cycles = 4
			address = cpu.absoluteAddress()
		}
	} else {
		switch (opcode >> 2) & 0x03 {
		case 0x00:
			*cycles = 2
			address = 0 // not used
		case 0x01:
			*cycles = 4

			switch opcode & 0xf0 {
			case 0x90:
				fallthrough
			case 0xb0:
				index = Y
			default:
				index = X
			}

			address = cpu.zeroPageIndexedAddress(index)
		case 0x02:
			*cycles = 2
			address = 0 // not used
		case 0x03:
			*cycles = 4

			switch opcode & 0xf0 {
			case 0x90:
				fallthrough
			case 0xb0:
				index = Y
			default:
				index = X
			}

			address = cpu.absoluteIndexedAddress(index, cycles)
		}
	}

	return
}

func (cpu *M6502) unofficialAddress(opcode OpCode, cycles *uint16) (address uint16) {
	// alu opcodes end with 11
	var index Index

	if opcode&0x10 == 0 {
		switch (opcode >> 2) & 0x03 {
		case 0x00:
			*cycles = 8
			address = cpu.indexedIndirectAddress()
		case 0x01:
			*cycles = 5
			address = cpu.zeroPageAddress()
		case 0x02:
			*cycles = 2
			address = cpu.immediateAddress()
		case 0x03:
			*cycles = 6
			address = cpu.absoluteAddress()
		}
	} else {
		switch (opcode >> 2) & 0x03 {
		case 0x00:
			*cycles = 8
			address = cpu.indirectIndexedAddress(cycles)
		case 0x01:
			*cycles = 6

			switch opcode & 0xf0 {
			case 0x90:
				fallthrough
			case 0xb0:
				index = Y
			default:
				index = X
			}

			address = cpu.zeroPageIndexedAddress(index)
		case 0x02:
			*cycles = 7
			address = cpu.absoluteIndexedAddress(Y, cycles)
		case 0x03:
			*cycles = 7

			switch opcode & 0xf0 {
			case 0x90:
				fallthrough
			case 0xb0:
				index = Y
			default:
				index = X
			}

			address = cpu.absoluteIndexedAddress(index, cycles)
		}
	}

	return
}

func (cpu *M6502) immediateAddress() (result uint16) {
	result = cpu.Registers.PC
	cpu.Registers.PC++

	if cpu.decode.enabled {
		value := cpu.Memory.Fetch(result)
		cpu.decode.args = fmt.Sprintf("%02X", value)
		cpu.decode.decodedArgs = fmt.Sprintf("#$")
	}

	return
}

func (cpu *M6502) zeroPageAddress() (result uint16) {
	result = uint16(cpu.Memory.Fetch(cpu.Registers.PC))
	cpu.Registers.PC++

	if cpu.decode.enabled {
		cpu.decode.args = fmt.Sprintf("%02X", result)
		cpu.decode.decodedArgs = fmt.Sprintf("$%02X", result)
	}

	return
}

func (cpu *M6502) IndexToRegister(which Index) uint8 {
	var index uint8

	switch which {
	case X:
		index = cpu.Registers.X
	case Y:
		index = cpu.Registers.Y
	}

	return index
}

func (which Index) String() string {
	switch which {
	case X:
		return "X"
	case Y:
		return "Y"
	default:
		return "?"
	}
}

func (cpu *M6502) zeroPageIndexedAddress(index Index) (result uint16) {
	value := cpu.Memory.Fetch(cpu.Registers.PC)
	result = uint16(value + cpu.IndexToRegister(index))
	cpu.Registers.PC++

	if cpu.decode.enabled {
		cpu.decode.args = fmt.Sprintf("%02X", value)
		cpu.decode.decodedArgs = fmt.Sprintf("$%02X,%s @ %02X",
			value, index.String(), result)
	}

	return
}

func (cpu *M6502) relativeAddress() (result uint16) {
	value := uint16(cpu.Memory.Fetch(cpu.Registers.PC))
	cpu.Registers.PC++

	var offset uint16

	if value > 0x7f {
		offset = -(0x0100 - value)
	} else {
		offset = value
	}

	result = cpu.Registers.PC + offset

	if cpu.decode.enabled {
		cpu.decode.args = fmt.Sprintf("%02X", value)
		cpu.decode.decodedArgs = fmt.Sprintf("$%04X", result)
	}

	return
}

func (cpu *M6502) absoluteAddress() (result uint16) {
	low := cpu.Memory.Fetch(cpu.Registers.PC)
	high := cpu.Memory.Fetch(cpu.Registers.PC + 1)
	cpu.Registers.PC += 2

	result = (uint16(high) << 8) | uint16(low)

	if cpu.decode.enabled {
		cpu.decode.args = fmt.Sprintf("%02X %02X", low, high)
		cpu.decode.decodedArgs = fmt.Sprintf("$%04X = ", result)
	}

	return
}

func (cpu *M6502) indirectAddress() (result uint16) {
	low := cpu.Memory.Fetch(cpu.Registers.PC)
	high := cpu.Memory.Fetch(cpu.Registers.PC + 1)
	cpu.Registers.PC += 2

	if cpu.decode.enabled {
		cpu.decode.args = fmt.Sprintf("%02X %02X", low, high)
	}

	// XXX: The 6502 had a bug in which it incremented only the
	// high byte instead of the whole 16-bit address when
	// computing the address.
	//
	// See http://www.obelisk.demon.co.uk/6502/reference.html#JMP
	// and http://www.6502.org/tutorials/6502opcodes.html#JMP for
	// details
	aHigh := (uint16(high) << 8) | uint16(low+1)
	aLow := (uint16(high) << 8) | uint16(low)

	low = cpu.Memory.Fetch(aLow)
	high = cpu.Memory.Fetch(aHigh)

	result = (uint16(high) << 8) | uint16(low)
	badResult := (uint16(cpu.Memory.Fetch(aLow+1)) << 8) | uint16(low)

	if cpu.decode.enabled {
		cpu.decode.decodedArgs = fmt.Sprintf("($%04X) = %04X", aLow, badResult)
	}

	return
}

func (cpu *M6502) absoluteIndexedAddress(index Index, cycles *uint16) (result uint16) {
	low := cpu.Memory.Fetch(cpu.Registers.PC)
	high := cpu.Memory.Fetch(cpu.Registers.PC + 1)
	cpu.Registers.PC += 2

	address := (uint16(high) << 8) | uint16(low)
	result = address + uint16(cpu.IndexToRegister(index))

	if cycles != nil && !SamePage(address, result) {
		*cycles++
	}

	if cpu.decode.enabled {
		cpu.decode.args = fmt.Sprintf("%02X %02X", low, high)
		cpu.decode.decodedArgs = fmt.Sprintf("$%04X,%s @ %04X = ", address, index.String(), result)
	}

	return
}

func (cpu *M6502) indexedIndirectAddress() (result uint16) {
	value := cpu.Memory.Fetch(cpu.Registers.PC)
	address := uint16(value + cpu.Registers.X)
	cpu.Registers.PC++

	low := cpu.Memory.Fetch(address)
	high := cpu.Memory.Fetch((address + 1) & 0x00ff)

	result = (uint16(high) << 8) | uint16(low)

	if cpu.decode.enabled {
		cpu.decode.args = fmt.Sprintf("%02X", value)
		cpu.decode.decodedArgs = fmt.Sprintf("($%02X,X) @ %02X = %04X = ", value, address, result)
	}

	return
}

func (cpu *M6502) indirectIndexedAddress(cycles *uint16) (result uint16) {
	value := cpu.Memory.Fetch(cpu.Registers.PC)
	address := uint16(value)
	cpu.Registers.PC++

	low := cpu.Memory.Fetch(address)
	high := cpu.Memory.Fetch((address + 1) & 0x00ff)

	address = (uint16(high) << 8) | uint16(low)

	result = address + uint16(cpu.Registers.Y)

	if cycles != nil && !SamePage(address, result) {
		*cycles++
	}

	if cpu.decode.enabled {
		cpu.decode.args = fmt.Sprintf("%02X", value)
		cpu.decode.decodedArgs = fmt.Sprintf("($%02X),Y = %04X @ %04X = ", value, address, result)
	}

	return
}

func (cpu *M6502) load(address uint16, register *uint8) {
	value := cpu.setZNFlags(cpu.Memory.Fetch(address))
	*register = value

	if cpu.decode.enabled {
		if !strings.HasPrefix(cpu.decode.decodedArgs, "#") &&
			!strings.HasSuffix(cpu.decode.decodedArgs, " = ") {
			cpu.decode.decodedArgs += fmt.Sprintf(" = ")
		}

		cpu.decode.decodedArgs += fmt.Sprintf("%02X", value)
	}
}

// Loads a byte of memory into the accumulator setting the zero and
// negative flags as appropriate.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if A = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of A is set
func (cpu *M6502) Lda(address uint16) {
	cpu.load(address, &cpu.Registers.A)
}

// Unofficial
//
// Loads a byte of memory into the accumulator and X setting the zero
// and negative flags as appropriate.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if A = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of A is set
func (cpu *M6502) Lax(address uint16) {
	cpu.Registers.X = cpu.Memory.Fetch(address)
	cpu.load(address, &cpu.Registers.A)
}

// Loads a byte of memory into the X register setting the zero and
// negative flags as appropriate.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if X = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of X is set
func (cpu *M6502) Ldx(address uint16) {
	cpu.load(address, &cpu.Registers.X)
}

// Loads a byte of memory into the Y register setting the zero and
// negative flags as appropriate.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if Y = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of Y is set
func (cpu *M6502) Ldy(address uint16) {
	cpu.load(address, &cpu.Registers.Y)
}

func (cpu *M6502) store(address uint16, value uint8) {
	oldValue := cpu.Memory.Store(address, value)

	if cpu.decode.enabled {
		if !strings.HasSuffix(cpu.decode.decodedArgs, " = ") {
			cpu.decode.decodedArgs += fmt.Sprintf(" = ")
		}

		cpu.decode.decodedArgs += fmt.Sprintf("%02X", oldValue)

	}
}

// Unofficial
func (cpu *M6502) Sax(address uint16) {
	cpu.store(address, cpu.Registers.A&cpu.Registers.X)
}

// Stores the contents of the accumulator into memory.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Sta(address uint16) {
	cpu.store(address, cpu.Registers.A)
}

// Stores the contents of the X register into memory.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Stx(address uint16) {
	cpu.store(address, cpu.Registers.X)
}

// Stores the contents of the Y register into memory.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Sty(address uint16) {
	cpu.store(address, cpu.Registers.Y)
}

func (cpu *M6502) transfer(from uint8, to *uint8) {
	*to = cpu.setZNFlags(from)
}

// Copies the current contents of the accumulator into the X register
// and sets the zero and negative flags as appropriate.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if X = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of X is set
func (cpu *M6502) Tax() {
	cpu.transfer(cpu.Registers.A, &cpu.Registers.X)
}

// Copies the current contents of the accumulator into the Y register
// and sets the zero and negative flags as appropriate.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if Y = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of Y is set
func (cpu *M6502) Tay() {
	cpu.transfer(cpu.Registers.A, &cpu.Registers.Y)
}

// Copies the current contents of the X register into the accumulator
// and sets the zero and negative flags as appropriate.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if A = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of A is set
func (cpu *M6502) Txa() {
	cpu.transfer(cpu.Registers.X, &cpu.Registers.A)
}

// Copies the current contents of the Y register into the accumulator
// and sets the zero and negative flags as appropriate.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if A = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of A is set
func (cpu *M6502) Tya() {
	cpu.transfer(cpu.Registers.Y, &cpu.Registers.A)
}

// Copies the current contents of the stack register into the X
// register and sets the zero and negative flags as appropriate.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if X = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of X is set
func (cpu *M6502) Tsx() {
	cpu.transfer(cpu.Registers.SP, &cpu.Registers.X)
}

// Copies the current contents of the X register into the stack
// register.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Txs() {
	cpu.Registers.SP = cpu.Registers.X
}

func (cpu *M6502) push(value uint8) {
	cpu.Memory.Store(0x0100|uint16(cpu.Registers.SP), value)
	cpu.Registers.SP--
}

func (cpu *M6502) push16(value uint16) {
	cpu.push(uint8(value >> 8))
	cpu.push(uint8(value))
}

func (cpu *M6502) pull() (value uint8) {
	cpu.Registers.SP++
	value = cpu.Memory.Fetch(0x0100 | uint16(cpu.Registers.SP))
	return
}

func (cpu *M6502) pull16() (value uint16) {
	low := cpu.pull()
	high := cpu.pull()

	value = (uint16(high) << 8) | uint16(low)
	return
}

// Pushes a copy of the accumulator on to the stack.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Pha() {
	cpu.push(cpu.Registers.A)
}

// Pushes a copy of the status flags on to the stack.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Php() {
	cpu.push(uint8(cpu.Registers.P | B | U))
}

// Pulls an 8 bit value from the stack and into the accumulator. The
// zero and negative flags are set as appropriate.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if A = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of A is set
func (cpu *M6502) Pla() {
	cpu.Registers.A = cpu.setZNFlags(cpu.pull())
}

// Pulls an 8 bit value from the stack and into the processor
// flags. The flags will take on new states as determined by the value
// pulled.
//
//         C 	Carry Flag 	  Set from stack
//         Z 	Zero Flag 	  Set from stack
//         I 	Interrupt Disable Set from stack
//         D 	Decimal Mode Flag Set from stack
//         B 	Break Command 	  Set from stack
//         V 	Overflow Flag 	  Set from stack
//         N 	Negative Flag 	  Set from stack
func (cpu *M6502) Plp() {
	cpu.Registers.P = Status(cpu.pull())
	cpu.Registers.P &^= B
	cpu.Registers.P |= U
}

// A logical AND is performed, bit by bit, on the accumulator contents
// using the contents of a byte of memory.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if A = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 set
func (cpu *M6502) And(address uint16) {
	value := cpu.Memory.Fetch(address)

	if cpu.decode.enabled {
		if !strings.HasPrefix(cpu.decode.decodedArgs, "#") &&
			!strings.HasSuffix(cpu.decode.decodedArgs, " = ") {
			cpu.decode.decodedArgs += fmt.Sprintf(" = ")
		}

		cpu.decode.decodedArgs += fmt.Sprintf("%02X", value)
	}

	cpu.Registers.A = cpu.setZNFlags(cpu.Registers.A & value)
}

// An exclusive OR is performed, bit by bit, on the accumulator
// contents using the contents of a byte of memory.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if A = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 set
func (cpu *M6502) Eor(address uint16) {
	value := cpu.Memory.Fetch(address)

	if cpu.decode.enabled {
		if !strings.HasPrefix(cpu.decode.decodedArgs, "#") &&
			!strings.HasSuffix(cpu.decode.decodedArgs, " = ") {
			cpu.decode.decodedArgs += fmt.Sprintf(" = ")
		}

		cpu.decode.decodedArgs += fmt.Sprintf("%02X", value)
	}

	cpu.Registers.A = cpu.setZNFlags(cpu.Registers.A ^ value)
}

// An inclusive OR is performed, bit by bit, on the accumulator
// contents using the contents of a byte of memory.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if A = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 set
func (cpu *M6502) Ora(address uint16) {
	value := cpu.Memory.Fetch(address)

	if cpu.decode.enabled {
		if !strings.HasPrefix(cpu.decode.decodedArgs, "#") &&
			!strings.HasSuffix(cpu.decode.decodedArgs, " = ") {
			cpu.decode.decodedArgs += fmt.Sprintf(" = ")
		}

		cpu.decode.decodedArgs += fmt.Sprintf("%02X", value)
	}

	cpu.Registers.A = cpu.setZNFlags(cpu.Registers.A | value)
}

// This instructions is used to test if one or more bits are set in a
// target memory location. The mask pattern in A is ANDed with the
// value in memory to set or clear the zero flag, but the result is
// not kept. Bits 7 and 6 of the value from memory are copied into the
// N and V flags.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if the result if the AND is zero
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Set to bit 6 of the memory value
//         N 	Negative Flag 	  Set to bit 7 of the memory value
func (cpu *M6502) Bit(address uint16) {
	value := cpu.Memory.Fetch(address)

	if cpu.decode.enabled {
		if !strings.HasPrefix(cpu.decode.decodedArgs, "#") &&
			!strings.HasSuffix(cpu.decode.decodedArgs, " = ") {
			cpu.decode.decodedArgs += fmt.Sprintf(" = ")
		}

		cpu.decode.decodedArgs += fmt.Sprintf("%02X", value)
	}

	cpu.setZFlag(value & cpu.Registers.A)
	cpu.Registers.P = (cpu.Registers.P & ^N & ^V) | Status(value&uint8(V|N))
}

func (cpu *M6502) addition(value uint16) {
	orig := uint16(cpu.Registers.A)

	if !cpu.decimalMode || cpu.Registers.P&D == 0 {
		result := cpu.setCFlagAddition(orig + value + uint16(cpu.Registers.P&C))
		cpu.Registers.A = cpu.setZNFlags(uint8(cpu.setVFlagAddition(orig, value, result)))
	} else {
		low := uint16(orig&0x000f) + uint16(value&0x000f) + uint16(cpu.Registers.P&C)
		high := uint16(orig&0x00f0) + uint16(value&0x00f0)

		if low >= 0x000a {
			low -= 0x000a
			high += 0x0010
		}

		if high >= 0x00a0 {
			high -= 0x00a0
		}

		result := cpu.setCFlagAddition(high | (low & 0x000f))
		cpu.Registers.A = cpu.setZNFlags(uint8(cpu.setVFlagAddition(orig, value, result)))
	}
}

// This instruction adds the contents of a memory location to the
// accumulator together with the carry bit. If overflow occurs the
// carry bit is set, this enables multiple byte addition to be
// performed.
//
//         C 	Carry Flag 	  Set if overflow in bit 7
//         Z 	Zero Flag 	  Set if A = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Set if sign bit is incorrect
//         N 	Negative Flag 	  Set if bit 7 set
func (cpu *M6502) Adc(address uint16) {
	value := uint16(cpu.Memory.Fetch(address))

	if cpu.decode.enabled {
		if !strings.HasPrefix(cpu.decode.decodedArgs, "#") &&
			!strings.HasSuffix(cpu.decode.decodedArgs, " = ") {
			cpu.decode.decodedArgs += fmt.Sprintf(" = ")
		}

		cpu.decode.decodedArgs += fmt.Sprintf("%02X", value)
	}

	cpu.addition(value)
}

// This instruction subtracts the contents of a memory location to the
// accumulator together with the not of the carry bit. If overflow
// occurs the carry bit is clear, this enables multiple byte
// subtraction to be performed.
//
//         C 	Carry Flag 	  Clear if overflow in bit 7
//         Z 	Zero Flag 	  Set if A = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Set if sign bit is incorrect
//         N 	Negative Flag 	  Set if bit 7 set
func (cpu *M6502) Sbc(address uint16) {
	value := uint16(cpu.Memory.Fetch(address))

	if cpu.decode.enabled {
		if !strings.HasPrefix(cpu.decode.decodedArgs, "#") &&
			!strings.HasSuffix(cpu.decode.decodedArgs, " = ") {
			cpu.decode.decodedArgs += fmt.Sprintf(" = ")
		}

		cpu.decode.decodedArgs += fmt.Sprintf("%02X", value)
	}

	if cpu.Registers.P&D == 0 {
		value ^= 0xff
	} else {
		value = 0x99 - value
	}

	cpu.addition(value)
}

func (cpu *M6502) compare(value uint16, register uint8) {
	if cpu.decode.enabled {
		if !strings.HasPrefix(cpu.decode.decodedArgs, "#") &&
			!strings.HasSuffix(cpu.decode.decodedArgs, " = ") {
			cpu.decode.decodedArgs += fmt.Sprintf(" = ")
		}

		cpu.decode.decodedArgs += fmt.Sprintf("%02X", value)
	}

	value = value ^ 0xff + 1
	cpu.setZNFlags(uint8(cpu.setCFlagAddition(uint16(register) + value)))
}

// Unofficial
func (cpu *M6502) Dcp(address uint16) {
	value := cpu.Memory.Fetch(address)

	if cpu.decode.enabled {
		if !strings.HasPrefix(cpu.decode.decodedArgs, "#") &&
			!strings.HasSuffix(cpu.decode.decodedArgs, " = ") {
			cpu.decode.decodedArgs += fmt.Sprintf(" = ")
		}

		cpu.decode.decodedArgs += fmt.Sprintf("%02X", value)
	}

	enabled := cpu.decode.enabled
	cpu.decode.enabled = false
	cpu.Dec(address)
	cpu.Cmp(address)
	cpu.decode.enabled = enabled
}

// Unofficial
func (cpu *M6502) Isb(address uint16) {
	value := cpu.Memory.Fetch(address)

	if cpu.decode.enabled {
		if !strings.HasPrefix(cpu.decode.decodedArgs, "#") &&
			!strings.HasSuffix(cpu.decode.decodedArgs, " = ") {
			cpu.decode.decodedArgs += fmt.Sprintf(" = ")
		}

		cpu.decode.decodedArgs += fmt.Sprintf("%02X", value)
	}

	enabled := cpu.decode.enabled
	cpu.decode.enabled = false
	cpu.Inc(address)
	cpu.Sbc(address)
	cpu.decode.enabled = enabled
}

// Unofficial
func (cpu *M6502) Slo(address uint16) {
	value := cpu.Memory.Fetch(address)

	if cpu.decode.enabled {
		if !strings.HasPrefix(cpu.decode.decodedArgs, "#") &&
			!strings.HasSuffix(cpu.decode.decodedArgs, " = ") {
			cpu.decode.decodedArgs += fmt.Sprintf(" = ")
		}

		cpu.decode.decodedArgs += fmt.Sprintf("%02X", value)
	}

	enabled := cpu.decode.enabled
	cpu.decode.enabled = false
	cpu.Asl(address)
	cpu.Ora(address)
	cpu.decode.enabled = enabled
}

// Unofficial
func (cpu *M6502) Rla(address uint16) {
	value := cpu.Memory.Fetch(address)

	if cpu.decode.enabled {
		if !strings.HasPrefix(cpu.decode.decodedArgs, "#") &&
			!strings.HasSuffix(cpu.decode.decodedArgs, " = ") {
			cpu.decode.decodedArgs += fmt.Sprintf(" = ")
		}

		cpu.decode.decodedArgs += fmt.Sprintf("%02X", value)
	}

	enabled := cpu.decode.enabled
	cpu.decode.enabled = false
	cpu.Rol(address)
	cpu.And(address)
	cpu.decode.enabled = enabled
}

// Unofficial
func (cpu *M6502) Sre(address uint16) {
	value := cpu.Memory.Fetch(address)

	if cpu.decode.enabled {
		if !strings.HasPrefix(cpu.decode.decodedArgs, "#") &&
			!strings.HasSuffix(cpu.decode.decodedArgs, " = ") {
			cpu.decode.decodedArgs += fmt.Sprintf(" = ")
		}

		cpu.decode.decodedArgs += fmt.Sprintf("%02X", value)
	}

	enabled := cpu.decode.enabled
	cpu.decode.enabled = false
	cpu.Lsr(address)
	cpu.Eor(address)
	cpu.decode.enabled = enabled
}

// Unofficial
func (cpu *M6502) Rra(address uint16) {
	value := cpu.Memory.Fetch(address)

	if cpu.decode.enabled {
		if !strings.HasPrefix(cpu.decode.decodedArgs, "#") &&
			!strings.HasSuffix(cpu.decode.decodedArgs, " = ") {
			cpu.decode.decodedArgs += fmt.Sprintf(" = ")
		}

		cpu.decode.decodedArgs += fmt.Sprintf("%02X", value)
	}

	enabled := cpu.decode.enabled
	cpu.decode.enabled = false
	cpu.Ror(address)
	cpu.Adc(address)
	cpu.decode.enabled = enabled
}

// This instruction compares the contents of the accumulator with
// another memory held value and sets the zero and carry flags as
// appropriate.
//
//         C 	Carry Flag 	  Set if A >= M
//         Z 	Zero Flag 	  Set if A = M
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of the result is set
func (cpu *M6502) Cmp(address uint16) {
	value := uint16(cpu.Memory.Fetch(address))
	cpu.compare(value, cpu.Registers.A)
}

// This instruction compares the contents of the X register with
// another memory held value and sets the zero and carry flags as
// appropriate.
//
//         C 	Carry Flag 	  Set if X >= M
//         Z 	Zero Flag 	  Set if X = M
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of the result is set
func (cpu *M6502) Cpx(address uint16) {
	value := uint16(cpu.Memory.Fetch(address))
	cpu.compare(value, cpu.Registers.X)
}

// This instruction compares the contents of the Y register with
// another memory held value and sets the zero and carry flags as
// appropriate.
//
//         C 	Carry Flag 	  Set if Y >= M
//         Z 	Zero Flag 	  Set if Y = M
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of the result is set
func (cpu *M6502) Cpy(address uint16) {
	value := uint16(cpu.Memory.Fetch(address))
	cpu.compare(value, cpu.Registers.Y)
}

// Adds one to the value held at a specified memory location setting
// the zero and negative flags as appropriate.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if result is zero
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of the result is set
func (cpu *M6502) Inc(address uint16) {
	value := cpu.Memory.Fetch(address)

	if cpu.decode.enabled {
		if !strings.HasPrefix(cpu.decode.decodedArgs, "#") &&
			!strings.HasSuffix(cpu.decode.decodedArgs, " = ") {
			cpu.decode.decodedArgs += fmt.Sprintf(" = ")
		}

		cpu.decode.decodedArgs += fmt.Sprintf("%02X", value)
	}

	cpu.Memory.Store(address, cpu.setZNFlags(value+1))
}

func (cpu *M6502) increment(register *uint8) {
	*register = cpu.setZNFlags(*register + 1)
}

// Adds one to the X register setting the zero and negative flags as
// appropriate.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if X is zero
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of X is set
func (cpu *M6502) Inx() {
	cpu.increment(&cpu.Registers.X)
}

// Adds one to the Y register setting the zero and negative flags as
// appropriate.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if Y is zero
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of Y is set
func (cpu *M6502) Iny() {
	cpu.increment(&cpu.Registers.Y)
}

// Subtracts one from the value held at a specified memory location
// setting the zero and negative flags as appropriate.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if result is zero
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of the result is set
func (cpu *M6502) Dec(address uint16) {
	value := cpu.Memory.Fetch(address)

	if cpu.decode.enabled {
		if !strings.HasPrefix(cpu.decode.decodedArgs, "#") &&
			!strings.HasSuffix(cpu.decode.decodedArgs, " = ") {
			cpu.decode.decodedArgs += fmt.Sprintf(" = ")
		}

		cpu.decode.decodedArgs += fmt.Sprintf("%02X", value)
	}

	cpu.Memory.Store(address, cpu.setZNFlags(value-1))
}

func (cpu *M6502) decrement(register *uint8) {
	*register = cpu.setZNFlags(*register - 1)
}

// Subtracts one from the X register setting the zero and negative
// flags as appropriate.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if X is zero
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of X is set
func (cpu *M6502) Dex() {
	cpu.decrement(&cpu.Registers.X)
}

// Subtracts one from the Y register setting the zero and negative
// flags as appropriate.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Set if Y is zero
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of Y is set
func (cpu *M6502) Dey() {
	cpu.decrement(&cpu.Registers.Y)
}

type direction int

const (
	left direction = iota
	right
)

func (cpu *M6502) shift(direction direction, value uint8, store func(uint8)) {
	if cpu.decode.enabled {
		if !strings.HasPrefix(cpu.decode.decodedArgs, "#") &&
			!strings.HasSuffix(cpu.decode.decodedArgs, " = ") {
			cpu.decode.decodedArgs += fmt.Sprintf(" = ")
		}

		cpu.decode.decodedArgs += fmt.Sprintf("%02X", value)
	}

	c := Status(0)

	switch direction {
	case left:
		c = Status((value & uint8(N)) >> 7)
		value <<= 1
	case right:
		c = Status(value & uint8(C))
		value >>= 1
	}

	cpu.Registers.P &= ^C
	cpu.Registers.P |= c

	store(cpu.setZNFlags(value))
}

// This operation shifts all the bits of the accumulator one bit
// left. Bit 0 is set to 0 and bit 7 is placed in the carry flag. The
// effect of this operation is to multiply the memory contents by 2
// (ignoring 2's complement considerations), setting the carry if the
// result will not fit in 8 bits.
//
//         C 	Carry Flag 	  Set to contents of old bit 7
//         Z 	Zero Flag 	  Set if A = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of the result is set
func (cpu *M6502) AslA() {
	cpu.shift(left, cpu.Registers.A, func(value uint8) { cpu.Registers.A = value })

	if cpu.decode.enabled {
		cpu.decode.decodedArgs = fmt.Sprintf("A")
	}
}

// This operation shifts all the bits of the memory contents one bit
// left. Bit 0 is set to 0 and bit 7 is placed in the carry flag. The
// effect of this operation is to multiply the memory contents by 2
// (ignoring 2's complement considerations), setting the carry if the
// result will not fit in 8 bits.
//
//         C 	Carry Flag 	  Set to contents of old bit 7
//         Z 	Zero Flag 	  Set if A = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of the result is set
func (cpu *M6502) Asl(address uint16) {
	cpu.shift(left, cpu.Memory.Fetch(address), func(value uint8) { cpu.Memory.Store(address, value) })
}

// Each of the bits in A is shift one place to the right. The bit that
// was in bit 0 is shifted into the carry flag. Bit 7 is set to zero.
//
//         C 	Carry Flag 	  Set to contents of old bit 0
//         Z 	Zero Flag 	  Set if result = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of the result is set
func (cpu *M6502) LsrA() {
	cpu.shift(right, cpu.Registers.A, func(value uint8) { cpu.Registers.A = value })

	if cpu.decode.enabled {
		cpu.decode.decodedArgs = fmt.Sprintf("A")
	}
}

// Each of the bits in M is shift one place to the right. The bit that
// was in bit 0 is shifted into the carry flag. Bit 7 is set to zero.
//
//         C 	Carry Flag 	  Set to contents of old bit 0
//         Z 	Zero Flag 	  Set if result = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of the result is set
func (cpu *M6502) Lsr(address uint16) {
	cpu.shift(right, cpu.Memory.Fetch(address), func(value uint8) { cpu.Memory.Store(address, value) })
}

func (cpu *M6502) rotate(direction direction, value uint8, store func(uint8)) {
	if cpu.decode.enabled {
		if !strings.HasPrefix(cpu.decode.decodedArgs, "#") &&
			!strings.HasSuffix(cpu.decode.decodedArgs, " = ") {
			cpu.decode.decodedArgs += fmt.Sprintf(" = ")
		}

		cpu.decode.decodedArgs += fmt.Sprintf("%02X", value)
	}

	c := Status(0)

	switch direction {
	case left:
		c = Status(value & uint8(N) >> 7)
		value = ((value << 1) & uint8(^C)) | uint8(cpu.Registers.P&C)
	case right:
		c = Status(value & uint8(C))
		value = ((value >> 1) & uint8(^N)) | uint8((cpu.Registers.P&C)<<7)
	}

	cpu.Registers.P &= ^C
	cpu.Registers.P |= c

	store(cpu.setZNFlags(value))
}

// Move each of the bits in A one place to the left. Bit 0 is filled
// with the current value of the carry flag whilst the old bit 7
// becomes the new carry flag value.
//
//         C 	Carry Flag 	  Set to contents of old bit 7
//         Z 	Zero Flag 	  Set if A = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of the result is set
func (cpu *M6502) RolA() {
	cpu.rotate(left, cpu.Registers.A, func(value uint8) { cpu.Registers.A = value })

	if cpu.decode.enabled {
		cpu.decode.decodedArgs = fmt.Sprintf("A")
	}
}

// Move each of the bits in A one place to the left. Bit 0 is filled
// with the current value of the carry flag whilst the old bit 7
// becomes the new carry flag value.
//
//         C 	Carry Flag 	  Set to contents of old bit 7
//         Z 	Zero Flag 	  Set if A = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of the result is set
func (cpu *M6502) Rol(address uint16) {
	cpu.rotate(left, cpu.Memory.Fetch(address), func(value uint8) { cpu.Memory.Store(address, value) })
}

// Move each of the bits in A one place to the right. Bit 7 is filled
// with the current value of the carry flag whilst the old bit 0
// becomes the new carry flag value.
//
//         C 	Carry Flag 	  Set to contents of old bit 0
//         Z 	Zero Flag 	  Set if A = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of the result is set
func (cpu *M6502) RorA() {
	cpu.rotate(right, cpu.Registers.A, func(value uint8) { cpu.Registers.A = value })

	if cpu.decode.enabled {
		cpu.decode.decodedArgs = fmt.Sprintf("A")
	}
}

// Move each of the bits in M one place to the right. Bit 7 is filled
// with the current value of the carry flag whilst the old bit 0
// becomes the new carry flag value.
//
//         C 	Carry Flag 	  Set to contents of old bit 0
//         Z 	Zero Flag 	  Set if A = 0
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Set if bit 7 of the result is set
func (cpu *M6502) Ror(address uint16) {
	cpu.rotate(right, cpu.Memory.Fetch(address), func(value uint8) { cpu.Memory.Store(address, value) })
}

// Sets the program counter to the address specified by the operand.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Jmp(address uint16) {
	if cpu.decode.enabled {
		if strings.HasPrefix(cpu.decode.decodedArgs, "$") {
			// delete ' = '
			cpu.decode.decodedArgs = cpu.decode.decodedArgs[:len(cpu.decode.decodedArgs)-3]
		}
	}

	cpu.Registers.PC = address
}

// The JSR instruction pushes the address (minus one) of the return
// point on to the stack and then sets the program counter to the
// target memory address.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Jsr(address uint16) {
	if cpu.decode.enabled {
		cpu.decode.decodedArgs = fmt.Sprintf("$%04X", address)
	}

	value := cpu.Registers.PC - 1

	cpu.push16(value)

	cpu.Registers.PC = address
}

// The RTS instruction is used at the end of a subroutine to return to
// the calling routine. It pulls the program counter (minus one) from
// the stack.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Rts() {
	cpu.Registers.PC = cpu.pull16() + 1
}

func (cpu *M6502) branch(address uint16, condition func() bool, cycles *uint16) {
	if condition() {
		*cycles++

		if !SamePage(cpu.Registers.PC, address) {
			*cycles++
		}

		cpu.Registers.PC = address
	}
}

// If the carry flag is clear then add the relative displacement to
// the program counter to cause a branch to a new location.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Bcc(address uint16, cycles *uint16) {
	cpu.branch(address, func() bool { return cpu.Registers.P&C == 0 }, cycles)
}

// If the carry flag is set then add the relative displacement to the
// program counter to cause a branch to a new location.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Bcs(address uint16, cycles *uint16) {
	cpu.branch(address, func() bool { return cpu.Registers.P&C != 0 }, cycles)
}

// If the zero flag is set then add the relative displacement to the
// program counter to cause a branch to a new location.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Beq(address uint16, cycles *uint16) {
	cpu.branch(address, func() bool { return cpu.Registers.P&Z != 0 }, cycles)
}

// If the negative flag is set then add the relative displacement to
// the program counter to cause a branch to a new location.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Bmi(address uint16, cycles *uint16) {
	cpu.branch(address, func() bool { return cpu.Registers.P&N != 0 }, cycles)
}

// If the zero flag is clear then add the relative displacement to the
// program counter to cause a branch to a new location.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Bne(address uint16, cycles *uint16) {
	cpu.branch(address, func() bool { return cpu.Registers.P&Z == 0 }, cycles)
}

// If the negative flag is clear then add the relative displacement to
// the program counter to cause a branch to a new location.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Bpl(address uint16, cycles *uint16) {
	cpu.branch(address, func() bool { return cpu.Registers.P&N == 0 }, cycles)
}

// If the overflow flag is clear then add the relative displacement to
// the program counter to cause a branch to a new location.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Bvc(address uint16, cycles *uint16) {
	cpu.branch(address, func() bool { return cpu.Registers.P&V == 0 }, cycles)
}

// If the overflow flag is set then add the relative displacement to
// the program counter to cause a branch to a new location.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Bvs(address uint16, cycles *uint16) {
	cpu.branch(address, func() bool { return cpu.Registers.P&V != 0 }, cycles)
}

// Set the carry flag to zero.
//
//         C 	Carry Flag 	  Set to 0
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Clc() {
	cpu.Registers.P &^= C
}

// Set the decimal mode flag to zero.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Set to 0
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Cld() {
	cpu.Registers.P &^= D
}

// Clears the interrupt disable flag allowing normal interrupt
// requests to be serviced.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Set to 0
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Cli() {
	cpu.Registers.P &^= I
}

// Clears the interrupt disable flag allowing normal interrupt
// requests to be serviced.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Set to 0
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Clv() {
	cpu.Registers.P &^= V
}

// Set the carry flag to one.
//
//         C 	Carry Flag 	  Set to 1
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Sec() {
	cpu.Registers.P |= C
}

// Set the decimal mode flag to one.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Set to 1
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Sed() {
	cpu.Registers.P |= D
}

// Set the interrupt disable flag to one.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Set to 1
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Sei() {
	cpu.Registers.P |= I
}

// The BRK instruction forces the generation of an interrupt
// request. The program counter and processor status are pushed on the
// stack then the IRQ interrupt vector at $FFFE/F is loaded into the
// PC and the break flag in the status set to one.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Set to 1
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Brk() {
	cpu.Registers.PC++

	cpu.push16(cpu.Registers.PC)
	cpu.push(uint8(cpu.Registers.P | B))

	cpu.Registers.P |= I

	low := cpu.Memory.Fetch(0xfffe)
	high := cpu.Memory.Fetch(0xffff)

	cpu.Registers.PC = (uint16(high) << 8) | uint16(low)
}

// The NOP instruction causes no changes to the processor other than
// the normal incrementing of the program counter to the next
// instruction.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) Nop() {
}

// Unofficial
//
// The NOP instruction causes no changes to the processor other than
// the normal incrementing of the program counter to the next
// instruction.
//
//         C 	Carry Flag 	  Not affected
//         Z 	Zero Flag 	  Not affected
//         I 	Interrupt Disable Not affected
//         D 	Decimal Mode Flag Not affected
//         B 	Break Command 	  Not affected
//         V 	Overflow Flag 	  Not affected
//         N 	Negative Flag 	  Not affected
func (cpu *M6502) NopAddress(address uint16) {
	if cpu.decode.enabled {
		value := cpu.Memory.Fetch(address)

		if !strings.HasPrefix(cpu.decode.decodedArgs, "#") &&
			!strings.HasSuffix(cpu.decode.decodedArgs, " = ") {
			cpu.decode.decodedArgs += fmt.Sprintf(" = ")
		}

		cpu.decode.decodedArgs += fmt.Sprintf("%02X", value)
	}
}

// The RTI instruction is used at the end of an interrupt processing
// routine. It pulls the processor flags from the stack followed by
// the program counter.
//
//         C 	Carry Flag 	  Set from stack
//         Z 	Zero Flag 	  Set from stack
//         I 	Interrupt Disable Set from stack
//         D 	Decimal Mode Flag Set from stack
//         B 	Break Command 	  Set from stack
//         V 	Overflow Flag 	  Set from stack
//         N 	Negative Flag 	  Set from stack
func (cpu *M6502) Rti() {
	cpu.Registers.P = Status(cpu.pull()) | U
	cpu.Registers.PC = cpu.pull16()
}
