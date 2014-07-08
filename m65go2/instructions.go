package m65go2

// Represents opcodes for the 6502 CPU
type OpCode uint8

// Represents an instruction for the 6502 CPU.  The Exec field
// implements the instruction and returns the total clock cycles to be
// consumed by the instruction.
type Instruction struct {
	Mneumonic string
	OpCode    OpCode
	Exec      func(*M6502) (status InstructionStatus)
}

// Stores instructions understood by the 6502 CPU, indexed by opcode.
type InstructionTable struct {
	opcodes         []*Instruction
	cycles          []uint16
	cyclesPageCross []uint16
}

type InstructionStatus uint16

const (
	PageCross InstructionStatus = 1 << iota
	Branched
)

// Returns a new, empty InstructionTable
func NewInstructionTable() InstructionTable {
	instructions := InstructionTable{
		opcodes: make([]*Instruction, 0x100),
		cycles: []uint16{
			7, 6, 0, 8, 3, 3, 5, 5, 3, 2, 2, 2, 4, 4, 6, 6,
			2, 5, 0, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7,
			6, 6, 0, 8, 3, 3, 5, 5, 4, 2, 2, 2, 4, 4, 6, 6,
			2, 5, 0, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7,
			6, 6, 0, 8, 3, 3, 5, 5, 3, 2, 2, 2, 3, 4, 6, 6,
			2, 5, 0, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7,
			6, 6, 0, 8, 3, 3, 5, 5, 4, 2, 2, 2, 5, 4, 6, 6,
			2, 5, 0, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7,
			2, 6, 2, 6, 3, 3, 3, 3, 2, 2, 2, 2, 4, 4, 4, 4,
			2, 6, 0, 6, 4, 4, 4, 4, 2, 5, 2, 5, 5, 5, 5, 5,
			2, 6, 2, 6, 3, 3, 3, 3, 2, 2, 2, 2, 4, 4, 4, 4,
			2, 5, 0, 5, 4, 4, 4, 4, 2, 4, 2, 4, 4, 4, 4, 4,
			2, 6, 2, 8, 3, 3, 5, 5, 2, 2, 2, 2, 4, 4, 6, 6,
			2, 5, 0, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7,
			2, 6, 2, 8, 3, 3, 5, 5, 2, 2, 2, 2, 4, 4, 6, 6,
			2, 5, 0, 8, 4, 4, 6, 6, 2, 4, 2, 7, 4, 4, 7, 7,
		},
		cyclesPageCross: []uint16{
			7, 6, 0, 8, 3, 3, 5, 5, 3, 2, 2, 2, 4, 4, 6, 6,
			3, 6, 0, 8, 4, 4, 6, 6, 2, 5, 2, 7, 5, 5, 7, 7,
			6, 6, 0, 8, 3, 3, 5, 5, 4, 2, 2, 2, 4, 4, 6, 6,
			3, 6, 0, 8, 4, 4, 6, 6, 2, 5, 2, 7, 5, 5, 7, 7,
			6, 6, 0, 8, 3, 3, 5, 5, 3, 2, 2, 2, 3, 4, 6, 6,
			3, 6, 0, 8, 4, 4, 6, 6, 2, 5, 2, 7, 5, 5, 7, 7,
			6, 6, 0, 8, 3, 3, 5, 5, 4, 2, 2, 2, 5, 4, 6, 6,
			3, 6, 0, 8, 4, 4, 6, 6, 2, 5, 2, 7, 5, 5, 7, 7,
			2, 6, 2, 6, 3, 3, 3, 3, 2, 2, 2, 2, 4, 4, 4, 4,
			3, 6, 0, 6, 4, 4, 4, 4, 2, 5, 2, 5, 5, 5, 5, 5,
			2, 6, 2, 6, 3, 3, 3, 3, 2, 2, 2, 2, 4, 4, 4, 4,
			3, 6, 0, 6, 4, 4, 4, 4, 2, 5, 2, 5, 5, 5, 5, 5,
			2, 6, 2, 8, 3, 3, 5, 5, 2, 2, 2, 2, 4, 4, 6, 6,
			3, 6, 0, 8, 4, 4, 6, 6, 2, 5, 2, 7, 5, 5, 7, 7,
			2, 6, 2, 8, 3, 3, 5, 5, 2, 2, 2, 2, 4, 4, 6, 6,
			3, 6, 0, 8, 4, 4, 6, 6, 2, 5, 2, 7, 5, 5, 7, 7,
		},
	}

	return instructions
}

// Executes an instruction in the InstructionTable, returns number of
// cycles taken to execute
func (instructions InstructionTable) Execute(cpu *M6502, opcode OpCode) (cycles uint16) {
	inst := instructions.opcodes[opcode]

	if inst == nil {
		return
	}

	status := inst.Exec(cpu)

	if status&PageCross == 0 {
		cycles = cpu.Instructions.cycles[opcode]
	} else {
		cycles = cpu.Instructions.cyclesPageCross[opcode]
	}

	if status&Branched != 0 {
		cycles++
	}

	return
}

// Adds an instruction to the InstructionTable
func (instructions InstructionTable) AddInstruction(inst *Instruction) {
	instructions.opcodes[inst.OpCode] = inst
}

// Removes any instruction with the given opcode
func (instructions InstructionTable) RemoveInstruction(opcode OpCode) {
	instructions.opcodes[opcode] = nil
}

// Adds the 6502 CPU's instruction set to the InstructionTable.
func (instructions InstructionTable) InitInstructions() {
	// LDA

	for _, o := range []OpCode{0xa1, 0xa5, 0xa9, 0xad, 0xb1, 0xb5, 0xb9, 0xbd} {
		opcode := o

		instructions.AddInstruction(&Instruction{
			Mneumonic: "LDA",
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				cpu.Lda(cpu.aluAddress(opcode, &status))
				return
			}})
	}

	// LDX

	for _, o := range []OpCode{0xa2, 0xa6, 0xae, 0xb6, 0xbe} {
		opcode := o

		instructions.AddInstruction(&Instruction{
			Mneumonic: "LDX",
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				cpu.Ldx(cpu.rmwAddress(opcode, &status))
				return
			}})
	}

	// LDY

	for _, o := range []OpCode{0xa0, 0xa4, 0xac, 0xb4, 0xbc} {
		opcode := o

		instructions.AddInstruction(&Instruction{
			Mneumonic: "LDY",
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				cpu.Ldy(cpu.controlAddress(opcode, &status))
				return
			}})
	}

	// STA

	for _, o := range []OpCode{0x81, 0x85, 0x8d, 0x91, 0x95, 0x99, 0x9d} {
		opcode := o

		instructions.AddInstruction(&Instruction{
			Mneumonic: "STA",
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				cpu.Sta(cpu.aluAddress(opcode, &status))
				return
			}})
	}

	// STX

	for _, o := range []OpCode{0x86, 0x8e, 0x96} {
		opcode := o

		instructions.AddInstruction(&Instruction{
			Mneumonic: "STX",
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				cpu.Stx(cpu.rmwAddress(opcode, &status))
				return
			}})
	}

	// STY

	for _, o := range []OpCode{0x84, 0x8c, 0x94} {
		opcode := o

		instructions.AddInstruction(&Instruction{
			Mneumonic: "STY",
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				cpu.Sty(cpu.controlAddress(opcode, &status))
				return
			}})
	}

	// TAX

	//     Implied
	instructions.AddInstruction(&Instruction{
		Mneumonic: "TAX",
		OpCode:    0xaa,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Tax()
			return
		}})

	// TAY

	//     Implied
	instructions.AddInstruction(&Instruction{
		Mneumonic: "TAY",
		OpCode:    0xa8,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Tay()
			return
		}})

	// TXA

	//     Implied
	instructions.AddInstruction(&Instruction{
		Mneumonic: "TXA",
		OpCode:    0x8a,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Txa()
			return
		}})

	// TYA

	//     Implied
	instructions.AddInstruction(&Instruction{
		Mneumonic: "TYA",
		OpCode:    0x98,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Tya()
			return
		}})

	// TSX

	//     Implied
	instructions.AddInstruction(&Instruction{
		Mneumonic: "TSX",
		OpCode:    0xba,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Tsx()
			return
		}})

	// TXS

	//     Implied
	instructions.AddInstruction(&Instruction{
		Mneumonic: "TXS",
		OpCode:    0x9a,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Txs()
			return
		}})

	// PHA

	//     Implied
	instructions.AddInstruction(&Instruction{
		Mneumonic: "PHA",
		OpCode:    0x48,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Pha()
			return
		}})

	// PHP

	//     Implied
	instructions.AddInstruction(&Instruction{
		Mneumonic: "PHP",
		OpCode:    0x08,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Php()
			return
		}})

	// PLA

	//     Implied
	instructions.AddInstruction(&Instruction{
		Mneumonic: "PLA",
		OpCode:    0x68,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Pla()
			return
		}})

	// PLP

	//     Implied
	instructions.AddInstruction(&Instruction{
		Mneumonic: "PLP",
		OpCode:    0x28,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Plp()
			return
		}})

	// AND

	for _, o := range []OpCode{0x21, 0x25, 0x29, 0x2d, 0x31, 0x35, 0x39, 0x3d} {
		opcode := o

		instructions.AddInstruction(&Instruction{
			Mneumonic: "AND",
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				cpu.And(cpu.aluAddress(opcode, &status))
				return
			}})
	}

	// EOR

	for _, o := range []OpCode{0x41, 0x45, 0x49, 0x4d, 0x51, 0x55, 0x59, 0x5d} {
		opcode := o

		instructions.AddInstruction(&Instruction{
			Mneumonic: "EOR",
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				cpu.Eor(cpu.aluAddress(opcode, &status))
				return
			}})
	}

	// ORA

	for _, o := range []OpCode{0x01, 0x05, 0x09, 0x0d, 0x11, 0x15, 0x19, 0x1d} {
		opcode := o

		instructions.AddInstruction(&Instruction{
			Mneumonic: "ORA",
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				cpu.Ora(cpu.aluAddress(opcode, &status))
				return
			}})
	}

	// BIT

	for _, o := range []OpCode{0x24, 0x2c} {
		opcode := o

		instructions.AddInstruction(&Instruction{
			Mneumonic: "BIT",
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				cpu.Bit(cpu.controlAddress(opcode, &status))
				return
			}})
	}

	// ADC

	for _, o := range []OpCode{0x61, 0x65, 0x69, 0x6d, 0x71, 0x75, 0x79, 0x7d} {
		opcode := o

		instructions.AddInstruction(&Instruction{
			Mneumonic: "ADC",
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				cpu.Adc(cpu.aluAddress(opcode, &status))
				return
			}})
	}

	// SBC

	for _, o := range []OpCode{0xe1, 0xe5, 0xeb, 0xe9, 0xed, 0xf1, 0xf5, 0xf9, 0xfd} {
		opcode := o
		mneumonic := ""

		if opcode == 0xeb {
			mneumonic = "*"
		}

		mneumonic += "SBC"

		instructions.AddInstruction(&Instruction{
			Mneumonic: mneumonic,
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				cpu.Sbc(cpu.aluAddress(opcode, &status))
				return
			}})
	}

	// DCP

	for _, o := range []OpCode{0xc3, 0xc7, 0xcf, 0xd3, 0xd7, 0xdb, 0xdf} {
		opcode := o

		instructions.AddInstruction(&Instruction{
			Mneumonic: "*DCP",
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				cpu.Dcp(cpu.unofficialAddress(opcode, &status))
				return
			}})
	}

	// ISB

	for _, o := range []OpCode{0xe3, 0xe7, 0xef, 0xf3, 0xf7, 0xfb, 0xff} {
		opcode := o

		instructions.AddInstruction(&Instruction{
			Mneumonic: "*ISB",
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				cpu.Isb(cpu.unofficialAddress(opcode, &status))
				return
			}})
	}

	// SLO

	for _, o := range []OpCode{0x03, 0x07, 0x0f, 0x13, 0x17, 0x1b, 0x1f} {
		opcode := o

		instructions.AddInstruction(&Instruction{
			Mneumonic: "*SLO",
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				cpu.Slo(cpu.unofficialAddress(opcode, &status))
				return
			}})
	}

	// RLA

	for _, o := range []OpCode{0x23, 0x27, 0x2f, 0x33, 0x37, 0x3b, 0x3f} {
		opcode := o

		instructions.AddInstruction(&Instruction{
			Mneumonic: "*RLA",
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				cpu.Rla(cpu.unofficialAddress(opcode, &status))
				return
			}})
	}

	// SRE

	for _, o := range []OpCode{0x43, 0x47, 0x4f, 0x53, 0x57, 0x5b, 0x5f} {
		opcode := o

		instructions.AddInstruction(&Instruction{
			Mneumonic: "*SRE",
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				cpu.Sre(cpu.unofficialAddress(opcode, &status))
				return
			}})
	}

	// RRA

	for _, o := range []OpCode{0x63, 0x67, 0x6f, 0x73, 0x77, 0x7b, 0x7f} {
		opcode := o

		instructions.AddInstruction(&Instruction{
			Mneumonic: "*RRA",
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				cpu.Rra(cpu.unofficialAddress(opcode, &status))
				return
			}})
	}

	// CMP

	for _, o := range []OpCode{0xc1, 0xc5, 0xc9, 0xcd, 0xd1, 0xd5, 0xd9, 0xdd} {
		opcode := o

		instructions.AddInstruction(&Instruction{
			Mneumonic: "CMP",
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				cpu.Cmp(cpu.aluAddress(opcode, &status))
				return
			}})
	}

	// CPX

	for _, o := range []OpCode{0xe0, 0xe4, 0xec} {
		opcode := o

		instructions.AddInstruction(&Instruction{
			Mneumonic: "CPX",
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				cpu.Cpx(cpu.controlAddress(opcode, &status))
				return
			}})
	}

	// CPY

	for _, o := range []OpCode{0xc0, 0xc4, 0xcc} {
		opcode := o

		instructions.AddInstruction(&Instruction{
			Mneumonic: "CPY",
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				cpu.Cpy(cpu.controlAddress(opcode, &status))
				return
			}})
	}

	// INC

	//     Zero Page
	instructions.AddInstruction(&Instruction{
		Mneumonic: "INC",
		OpCode:    0xe6,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Inc(cpu.zeroPageAddress())
			return
		}})

	//     Zero Page,X
	instructions.AddInstruction(&Instruction{
		Mneumonic: "INC",
		OpCode:    0xf6,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Inc(cpu.zeroPageIndexedAddress(X))
			return
		}})

	//     Absolute
	instructions.AddInstruction(&Instruction{
		Mneumonic: "INC",
		OpCode:    0xee,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Inc(cpu.absoluteAddress())
			return
		}})

	//     Absolute,X
	instructions.AddInstruction(&Instruction{
		Mneumonic: "INC",
		OpCode:    0xfe,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Inc(cpu.absoluteIndexedAddress(X, &status))
			return
		}})

	// INX

	//     Implied
	instructions.AddInstruction(&Instruction{
		Mneumonic: "INX",
		OpCode:    0xe8,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Inx()
			return
		}})

	// INY

	//     Implied
	instructions.AddInstruction(&Instruction{
		Mneumonic: "INY",
		OpCode:    0xc8,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Iny()
			return
		}})

	// DEC

	//     Zero Page
	instructions.AddInstruction(&Instruction{
		Mneumonic: "DEC",
		OpCode:    0xc6,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Dec(cpu.zeroPageAddress())
			return
		}})

	//     Zero Page,X
	instructions.AddInstruction(&Instruction{
		Mneumonic: "DEC",
		OpCode:    0xd6,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Dec(cpu.zeroPageIndexedAddress(X))
			return
		}})

	//     Absolute
	instructions.AddInstruction(&Instruction{
		Mneumonic: "DEC",
		OpCode:    0xce,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Dec(cpu.absoluteAddress())
			return
		}})

	//     Absolute,X
	instructions.AddInstruction(&Instruction{
		Mneumonic: "DEC",
		OpCode:    0xde,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Dec(cpu.absoluteIndexedAddress(X, &status))
			return
		}})

	// DEX

	//     Implied
	instructions.AddInstruction(&Instruction{
		Mneumonic: "DEX",
		OpCode:    0xca,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Dex()
			return
		}})

	// DEY

	//     Implied
	instructions.AddInstruction(&Instruction{
		Mneumonic: "DEY",
		OpCode:    0x88,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Dey()
			return
		}})

	// ASL

	//     Accumulator
	instructions.AddInstruction(&Instruction{
		Mneumonic: "ASL",
		OpCode:    0x0a,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.AslA()
			return
		}})

	//     Zero Page
	instructions.AddInstruction(&Instruction{
		Mneumonic: "ASL",
		OpCode:    0x06,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Asl(cpu.zeroPageAddress())
			return
		}})

	//     Zero Page,X
	instructions.AddInstruction(&Instruction{
		Mneumonic: "ASL",
		OpCode:    0x16,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Asl(cpu.zeroPageIndexedAddress(X))
			return
		}})

	//     Absolute
	instructions.AddInstruction(&Instruction{
		Mneumonic: "ASL",
		OpCode:    0x0e,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Asl(cpu.absoluteAddress())
			return
		}})

	//     Absolute,X
	instructions.AddInstruction(&Instruction{
		Mneumonic: "ASL",
		OpCode:    0x1e,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Asl(cpu.absoluteIndexedAddress(X, &status))
			return
		}})

	// LSR

	//     Accumulator
	instructions.AddInstruction(&Instruction{
		Mneumonic: "LSR",
		OpCode:    0x4a,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.LsrA()
			return
		}})

	//     Zero Page
	instructions.AddInstruction(&Instruction{
		Mneumonic: "LSR",
		OpCode:    0x46,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Lsr(cpu.zeroPageAddress())
			return
		}})

	//     Zero Page,X
	instructions.AddInstruction(&Instruction{
		Mneumonic: "LSR",
		OpCode:    0x56,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Lsr(cpu.zeroPageIndexedAddress(X))
			return
		}})

	//     Absolute
	instructions.AddInstruction(&Instruction{
		Mneumonic: "LSR",
		OpCode:    0x4e,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Lsr(cpu.absoluteAddress())
			return
		}})

	//     Absolute,X
	instructions.AddInstruction(&Instruction{
		Mneumonic: "LSR",
		OpCode:    0x5e,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Lsr(cpu.absoluteIndexedAddress(X, &status))
			return
		}})

	// ROL

	//     Accumulator
	instructions.AddInstruction(&Instruction{
		Mneumonic: "ROL",
		OpCode:    0x2a,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.RolA()
			return
		}})

	//     Zero Page
	instructions.AddInstruction(&Instruction{
		Mneumonic: "ROL",
		OpCode:    0x26,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Rol(cpu.zeroPageAddress())
			return
		}})

	//     Zero Page,X
	instructions.AddInstruction(&Instruction{
		Mneumonic: "ROL",
		OpCode:    0x36,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Rol(cpu.zeroPageIndexedAddress(X))
			return
		}})

	//     Absolute
	instructions.AddInstruction(&Instruction{
		Mneumonic: "ROL",
		OpCode:    0x2e,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Rol(cpu.absoluteAddress())
			return
		}})

	//     Absolute,X
	instructions.AddInstruction(&Instruction{
		Mneumonic: "ROL",
		OpCode:    0x3e,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Rol(cpu.absoluteIndexedAddress(X, &status))
			return
		}})

	// ROR

	//     Accumulator
	instructions.AddInstruction(&Instruction{
		Mneumonic: "ROR",
		OpCode:    0x6a,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.RorA()
			return
		}})

	//     Zero Page
	instructions.AddInstruction(&Instruction{
		Mneumonic: "ROR",
		OpCode:    0x66,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Ror(cpu.zeroPageAddress())
			return
		}})

	//     Zero Page,X
	instructions.AddInstruction(&Instruction{
		Mneumonic: "ROR",
		OpCode:    0x76,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Ror(cpu.zeroPageIndexedAddress(X))
			return
		}})

	//     Absolute
	instructions.AddInstruction(&Instruction{
		Mneumonic: "ROR",
		OpCode:    0x6e,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Ror(cpu.absoluteAddress())
			return
		}})

	//     Absolute,X
	instructions.AddInstruction(&Instruction{
		Mneumonic: "ROR",
		OpCode:    0x7e,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Ror(cpu.absoluteIndexedAddress(X, &status))
			return
		}})

	// JMP

	//     Absolute
	instructions.AddInstruction(&Instruction{
		Mneumonic: "JMP",
		OpCode:    0x4c,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Jmp(cpu.absoluteAddress())
			return
		}})

	//     Indirect
	instructions.AddInstruction(&Instruction{
		Mneumonic: "JMP",
		OpCode:    0x6c,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Jmp(cpu.indirectAddress())
			return
		}})

	// JSR

	//     Absolute
	instructions.AddInstruction(&Instruction{
		Mneumonic: "JSR",
		OpCode:    0x20,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Jsr(cpu.absoluteAddress())
			return
		}})

	// RTS

	//     Implied
	instructions.AddInstruction(&Instruction{
		Mneumonic: "RTS",
		OpCode:    0x60,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Rts()
			return
		}})

	// BCC

	//     Relative
	instructions.AddInstruction(&Instruction{
		Mneumonic: "BCC",
		OpCode:    0x90,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Bcc(cpu.controlAddress(0x90, &status), &status)
			return
		}})

	// BCS

	//     Relative
	instructions.AddInstruction(&Instruction{
		Mneumonic: "BCS",
		OpCode:    0xb0,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Bcs(cpu.controlAddress(0xb0, &status), &status)
			return
		}})

	// BEQ

	//     Relative
	instructions.AddInstruction(&Instruction{
		Mneumonic: "BEQ",
		OpCode:    0xf0,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Beq(cpu.controlAddress(0xf0, &status), &status)
			return
		}})

	// BMI

	//     Relative
	instructions.AddInstruction(&Instruction{
		Mneumonic: "BMI",
		OpCode:    0x30,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Bmi(cpu.controlAddress(0x30, &status), &status)
			return
		}})

	// BNE

	//     Relative
	instructions.AddInstruction(&Instruction{
		Mneumonic: "BNE",
		OpCode:    0xd0,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Bne(cpu.controlAddress(0xd0, &status), &status)
			return
		}})

	// BPL

	//     Relative
	instructions.AddInstruction(&Instruction{
		Mneumonic: "BPL",
		OpCode:    0x10,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Bpl(cpu.controlAddress(0x10, &status), &status)
			return
		}})

	// BVC

	//     Relative
	instructions.AddInstruction(&Instruction{
		Mneumonic: "BVC",
		OpCode:    0x50,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Bvc(cpu.controlAddress(0x50, &status), &status)
			return
		}})

	// BVS

	//     Relative
	instructions.AddInstruction(&Instruction{
		Mneumonic: "BVS",
		OpCode:    0x70,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Bvs(cpu.controlAddress(0x70, &status), &status)
			return
		}})

	// CLC

	//     Implied
	instructions.AddInstruction(&Instruction{
		Mneumonic: "CLC",
		OpCode:    0x18,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Clc()
			return
		}})

	// CLD

	//     Implied
	instructions.AddInstruction(&Instruction{
		Mneumonic: "CLD",
		OpCode:    0xd8,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Cld()
			return
		}})

	// CLI

	//     Implied
	instructions.AddInstruction(&Instruction{
		Mneumonic: "CLI",
		OpCode:    0x58,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Cli()
			return
		}})

	// CLV

	//     Implied
	instructions.AddInstruction(&Instruction{
		Mneumonic: "CLV",
		OpCode:    0xb8,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Clv()
			return
		}})

	// SEC

	//     Implied
	instructions.AddInstruction(&Instruction{
		Mneumonic: "SEC",
		OpCode:    0x38,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Sec()
			return
		}})

	// SED

	//     Implied
	instructions.AddInstruction(&Instruction{
		Mneumonic: "SED",
		OpCode:    0xf8,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Sed()
			return
		}})

	// SEI

	//     Implied
	instructions.AddInstruction(&Instruction{
		Mneumonic: "SEI",
		OpCode:    0x78,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Sei()
			return
		}})

	// BRK

	//     Implied
	instructions.AddInstruction(&Instruction{
		Mneumonic: "BRK",
		OpCode:    0x00,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Brk()
			return
		}})

	// NOP

	//     Implied
	instructions.AddInstruction(&Instruction{
		Mneumonic: "NOP",
		OpCode:    0xea,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Nop()
			return
		}})

	//     Unofficial

	for _, o := range []OpCode{0x1a, 0x3a, 0x5a, 0x7a, 0xda, 0xfa} {
		opcode := o

		instructions.AddInstruction(&Instruction{
			Mneumonic: "*NOP",
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				cpu.Nop()
				return
			}})
	}

	for _, o := range []OpCode{0x04, 0x14, 0x34, 0x44, 0x54, 0x64, 0x74, 0xd4, 0xf4, 0x80, 0x82, 0x89, 0xc2, 0xe2} {
		opcode := o

		instructions.AddInstruction(&Instruction{
			Mneumonic: "*NOP",
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				var address uint16

				switch {
				case opcode == 0x80, opcode == 0x82, opcode == 0x89,
					opcode == 0xc2, opcode == 0xe2:
					address = cpu.immediateAddress()
				case (opcode>>4)&0x01 == 0:
					address = cpu.zeroPageAddress()
				default:
					address = cpu.zeroPageIndexedAddress(X)
				}

				cpu.NopAddress(address)
				return
			}})
	}

	for _, o := range []OpCode{0x0c, 0x1c, 0x3c, 0x5c, 0x7c, 0xdc, 0xfc} {
		opcode := o

		instructions.AddInstruction(&Instruction{
			Mneumonic: "*NOP",
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				var address uint16

				if (opcode>>4)&0x01 == 0 {
					address = cpu.absoluteAddress()
				} else {
					address = cpu.absoluteIndexedAddress(X, &status)
				}

				cpu.NopAddress(address)
				return
			}})
	}

	// LAX

	//     Unofficial

	for _, o := range []OpCode{0xa3, 0xa7, 0xaf, 0xb3, 0xb7, 0xbf, 0xab} {
		opcode := o

		instructions.AddInstruction(&Instruction{
			Mneumonic: "*LAX",
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				cpu.Lax(cpu.unofficialAddress(opcode, &status))
				return
			}})
	}

	// SAX

	//     Unofficial

	for _, o := range []OpCode{0x83, 0x87, 0x8f, 0x97} {
		opcode := o

		instructions.AddInstruction(&Instruction{
			Mneumonic: "*SAX",
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				cpu.Sax(cpu.unofficialAddress(opcode, &status))
				return
			}})
	}

	// ANC

	//     Unofficial

	for _, o := range []OpCode{0x0b, 0x2b} {
		opcode := o

		instructions.AddInstruction(&Instruction{
			Mneumonic: "*ANC",
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				cpu.Anc(cpu.unofficialAddress(opcode, &status))
				return
			}})
	}

	// ALR

	//     Unofficial

	for _, o := range []OpCode{0x4b} {
		opcode := o

		instructions.AddInstruction(&Instruction{
			Mneumonic: "*ALR",
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				cpu.Alr(cpu.unofficialAddress(opcode, &status))
				return
			}})
	}

	// ARR

	//     Unofficial

	for _, o := range []OpCode{0x6b} {
		opcode := o

		instructions.AddInstruction(&Instruction{
			Mneumonic: "*ARR",
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				cpu.Arr(cpu.unofficialAddress(opcode, &status))
				return
			}})
	}

	// AXS

	//     Unofficial

	for _, o := range []OpCode{0xcb} {
		opcode := o

		instructions.AddInstruction(&Instruction{
			Mneumonic: "*AXS",
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				cpu.Axs(cpu.unofficialAddress(opcode, &status))
				return
			}})
	}

	// SHY

	//     Unofficial

	for _, o := range []OpCode{0x9c} {
		opcode := o

		instructions.AddInstruction(&Instruction{
			Mneumonic: "*SHY",
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				cpu.Shy(cpu.unofficialAddress(opcode, &status))
				return
			}})
	}

	// SHX

	//     Unofficial

	for _, o := range []OpCode{0x9e} {
		opcode := o

		instructions.AddInstruction(&Instruction{
			Mneumonic: "*SHX",
			OpCode:    opcode,
			Exec: func(cpu *M6502) (status InstructionStatus) {
				cpu.Shx(cpu.unofficialAddress(opcode, &status))
				return
			}})
	}

	// RTI

	//     Implied
	instructions.AddInstruction(&Instruction{
		Mneumonic: "RTI",
		OpCode:    0x40,
		Exec: func(cpu *M6502) (status InstructionStatus) {
			cpu.Rti()
			return
		}})
}
