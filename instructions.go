package main

import (
	"fmt"
	"strings"
)

type ParamType func(state *ProgramState, param string) (uint16, error)

type InstructionParams struct {
	Types     []ParamType
	Assembler func(args []uint16) []uint8
}

type InstructionSet map[string][]InstructionParams

var Instructions = InstructionSetNew()

func InstructionSetNew() InstructionSet {
	result := make(InstructionSet)

	result["LD"] = []InstructionParams{
		{
			Types: []ParamType{Reg8, Reg8},
			Assembler: func(args []uint16) []byte {
				return []byte{0b01000000 | (uint8(uint8(args[0])) << 3) | uint8(uint8(args[1]))}
			},
		},
		{
			Types:     []ParamType{Reg8, Raw8},
			Assembler: func(args []uint16) []byte { return []byte{0b00000110 | (uint8(args[0]) << 3), uint8(args[1])} },
		},
		{
			Types:     []ParamType{A, Reg16Indirect},
			Assembler: func(args []uint16) []byte { return []byte{0b00001010 | uint8(args[1])<<4} },
		},
		{
			Types:     []ParamType{Reg16Indirect, A},
			Assembler: func(args []uint16) []byte { return []byte{0b00000010 | uint8(args[1])<<4} },
		},
		{
			Types:     []ParamType{A, Raw16Indirect},
			Assembler: func(args []uint16) []byte { return []byte{0b11111010, uint8(args[1]) & 0xff, uint8(args[1] >> 8)} },
		},
		{
			Types:     []ParamType{Raw16Indirect, A},
			Assembler: func(args []uint16) []byte { return []byte{0b11101010, uint8(args[0]) & 0xff, uint8(args[0] >> 8)} },
		},
		{
			Types:     []ParamType{A, IndirectC},
			Assembler: func(args []uint16) []byte { return []byte{0b11110010} },
		},
		{
			Types:     []ParamType{IndirectC, A},
			Assembler: func(args []uint16) []byte { return []byte{0b11100010} },
		},
		{
			Types:     []ParamType{A, Raw8Indirect},
			Assembler: func(args []uint16) []byte { return []byte{0b11110000, uint8(args[1])} },
		},
		{
			Types:     []ParamType{Raw8Indirect, A},
			Assembler: func(args []uint16) []byte { return []byte{0b11100000, uint8(args[0])} },
		},
		{
			Types: []ParamType{Reg16, Raw16},
			Assembler: func(args []uint16) []byte {
				return []byte{
					0b00000001 | (uint8(args[0]) << 4),
					uint8(args[1]) & 0xff,
					uint8(args[1] >> 8),
				}
			},
		},
		{
			Types:     []ParamType{Raw16Indirect, SP},
			Assembler: func(args []uint16) []byte { return []byte{0b00001000, uint8(args[0]) & 0xff, uint8(args[0] >> 8)} },
		},
		{
			Types:     []ParamType{SP, HL},
			Assembler: func(args []uint16) []byte { return []byte{0b11111001} },
		},
		{
			Types:     []ParamType{HL, Raw8},
			Assembler: func(args []uint16) []byte { return []byte{0b11111000, uint8(args[1])} },
		},
	}
	result["PUSH"] = []InstructionParams{
		{
			Types:     []ParamType{Reg16},
			Assembler: func(args []uint16) []byte { return []byte{0b11000101 | (uint8(args[0]) << 4)} },
		},
	}
	result["POP"] = []InstructionParams{
		{
			Types:     []ParamType{Reg16},
			Assembler: func(args []uint16) []byte { return []byte{0b11000001 | (uint8(args[0]) << 4)} },
		},
	}
	result["ADD"] = []InstructionParams{
		{
			Types:     []ParamType{Reg8},
			Assembler: func(args []uint16) []byte { return []byte{0b10000000 | (uint8(args[0]))} },
		},
		{
			Types:     []ParamType{Raw8},
			Assembler: func(args []uint16) []byte { return []byte{0b11000110, uint8(args[0])} },
		},
		{
			Types:     []ParamType{SP, Raw8},
			Assembler: func(args []uint16) []byte { return []byte{0b11101000, uint8(args[1])} },
		},
	}
	result["ADC"] = []InstructionParams{
		{
			Types:     []ParamType{Reg8},
			Assembler: func(args []uint16) []byte { return []byte{0b10001000 | uint8(args[0])} },
		},
		{
			Types:     []ParamType{Raw8},
			Assembler: func(args []uint16) []byte { return []byte{0b11001110, uint8(args[0])} },
		},
	}
	result["SUB"] = []InstructionParams{
		{
			Types:     []ParamType{Reg8},
			Assembler: func(args []uint16) []byte { return []byte{0b10010000 | (uint8(args[0]))} },
		},
		{
			Types:     []ParamType{Raw8},
			Assembler: func(args []uint16) []byte { return []byte{0b11010110, uint8(args[0])} },
		},
	}
	result["SBC"] = []InstructionParams{
		{
			Types:     []ParamType{Reg8},
			Assembler: func(args []uint16) []byte { return []byte{0b10011000 | uint8(args[0])} },
		},
		{
			Types:     []ParamType{Raw8},
			Assembler: func(args []uint16) []byte { return []byte{0b11011110, uint8(args[0])} },
		},
	}
	result["CP"] = []InstructionParams{
		{
			Types:     []ParamType{Reg8},
			Assembler: func(args []uint16) []byte { return []byte{0b10111000 | (uint8(args[0]))} },
		},
		{
			Types:     []ParamType{Raw8},
			Assembler: func(args []uint16) []byte { return []byte{0b11111110, uint8(args[0])} },
		},
	}
	result["INC"] = []InstructionParams{
		{
			Types:     []ParamType{Reg8},
			Assembler: func(args []uint16) []byte { return []byte{0b00000100 | (uint8(args[0]) << 3)} },
		},

		{
			Types:     []ParamType{Reg16},
			Assembler: func(args []uint16) []byte { return []byte{0b00000011 | (uint8(args[0]) << 4)} },
		},
	}
	result["DEC"] = []InstructionParams{
		{
			Types:     []ParamType{Reg8},
			Assembler: func(args []uint16) []byte { return []byte{0b00000101 | (uint8(args[0]) << 3)} },
		},

		{
			Types:     []ParamType{Reg16},
			Assembler: func(args []uint16) []byte { return []byte{0b00001011 | (uint8(args[0]) << 4)} },
		},
	}
	result["AND"] = []InstructionParams{
		{
			Types:     []ParamType{Reg8},
			Assembler: func(args []uint16) []byte { return []byte{0b10100000 | (uint8(args[0]))} },
		},
		{
			Types:     []ParamType{Raw8},
			Assembler: func(args []uint16) []byte { return []byte{0b11100110, uint8(args[0])} },
		},
	}
	result["OR"] = []InstructionParams{
		{
			Types:     []ParamType{Reg8},
			Assembler: func(args []uint16) []byte { return []byte{0b10110000 | (uint8(args[0]))} },
		},
		{
			Types:     []ParamType{Raw8},
			Assembler: func(args []uint16) []byte { return []byte{0b11110110, uint8(args[0])} },
		},
	}
	result["XOR"] = []InstructionParams{
		{
			Types:     []ParamType{Reg8},
			Assembler: func(args []uint16) []byte { return []byte{0b10101000 | (uint8(args[0]))} },
		},
		{
			Types:     []ParamType{Raw8},
			Assembler: func(args []uint16) []byte { return []byte{0b11101110, uint8(args[0])} },
		},
	}
	result["CCF"] = []InstructionParams{
		{Types: []ParamType{}, Assembler: func(args []uint16) []byte { return []byte{0b00111111} }},
	}
	result["SCF"] = []InstructionParams{
		{Types: []ParamType{}, Assembler: func(args []uint16) []byte { return []byte{0b00110111} }},
	}
	result["DAA"] = []InstructionParams{
		{Types: []ParamType{}, Assembler: func(args []uint16) []byte { return []byte{0b00100111} }},
	}
	result["CPL"] = []InstructionParams{
		{Types: []ParamType{}, Assembler: func(args []uint16) []byte { return []byte{0b00101111} }},
	}
	result["JP"] = []InstructionParams{
		{
			Types:     []ParamType{Raw16},
			Assembler: func(args []uint16) []byte { return []byte{0b11000011, uint8(args[0]) & 0xff, uint8(args[0] >> 8)} },
		},
		{
			Types:     []ParamType{HL},
			Assembler: func(args []uint16) []byte { return []byte{0b11101001} },
		},
		{
			Types: []ParamType{Condition, Raw16},
			Assembler: func(args []uint16) []byte {
				return []byte{
					0b11000010 | (uint8(args[0]) << 3),
					uint8(args[1]) & 0xff,
					uint8(args[1] >> 8),
				}
			},
		},
	}
	result["JR"] = []InstructionParams{
		{
			Types:     []ParamType{Raw8},
			Assembler: func(args []uint16) []byte { return []byte{0b00011000, uint8(args[0])} },
		},
		// TODO: Add the relative thingies somehow
		{
			Types:     []ParamType{Raw16},
			Assembler: func(args []uint16) []byte { return []byte{0b00011000, uint8(args[0])} },
		},
		{
			Types:     []ParamType{Condition, Raw8},
			Assembler: func(args []uint16) []byte { return []byte{0b00100000 | (uint8(args[0]) << 3), uint8(args[1])} },
		},
		// {
		// 	Types:     []ParamType{Condition, Raw16},
		// 	Assembler: func(args []uint16) []byte { return []byte{0b00100000 | (uint8(args[0]) << 3), uint8(args[1])} },
		// },
	}
	result["CALL"] = []InstructionParams{
		{
			Types:     []ParamType{Raw16},
			Assembler: func(args []uint16) []byte { return []byte{0b11001101, uint8(args[0]) & 0xff, uint8(args[0] >> 8)} },
		},
		{
			Types: []ParamType{Condition, Raw16},
			Assembler: func(args []uint16) []byte {
				return []byte{
					0b11000100 | (uint8(args[0]) << 3),
					uint8(args[1]) & 0xff,
					uint8(args[1] >> 8),
				}
			},
		},
	}
	result["RET"] = []InstructionParams{
		{Types: []ParamType{}, Assembler: func(args []uint16) []byte { return []byte{0b11001001} }},
		{
			Types:     []ParamType{Condition},
			Assembler: func(args []uint16) []byte { return []byte{0b11000000 | (uint8(args[0]) << 3)} },
		},
	}
	result["RETI"] = []InstructionParams{
		{Types: []ParamType{}, Assembler: func(args []uint16) []byte { return []byte{0b11011001} }},
	}
	result["RST"] = []InstructionParams{
		{
			Types:     []ParamType{BitOrdinal},
			Assembler: func(args []uint16) []byte { return []byte{0b11000111 | (uint8(args[0]) << 3)} },
		},
	}
	result["DI"] = []InstructionParams{
		{Types: []ParamType{}, Assembler: func(args []uint16) []byte { return []byte{0b11110011} }},
	}
	result["EI"] = []InstructionParams{
		{Types: []ParamType{}, Assembler: func(args []uint16) []byte { return []byte{0b11111011} }},
	}
	result["NOP"] = []InstructionParams{
		{Types: []ParamType{}, Assembler: func(args []uint16) []byte { return []byte{0b00000000} }},
	}
	result["HALT"] = []InstructionParams{
		{Types: []ParamType{}, Assembler: func(args []uint16) []byte { return []byte{0b01110110} }},
	}
	result["STOP"] = []InstructionParams{
		{
			Types:     []ParamType{},
			Assembler: func(args []uint16) []byte { return []byte{0b00010000, 0b00000000} },
		},
	}
	result["RLCA"] = []InstructionParams{
		{Types: []ParamType{}, Assembler: func(args []uint16) []byte { return []byte{0b00000111} }},
	}
	result["RLA"] = []InstructionParams{
		{Types: []ParamType{}, Assembler: func(args []uint16) []byte { return []byte{0b00010111} }},
	}
	result["RRCA"] = []InstructionParams{
		{Types: []ParamType{}, Assembler: func(args []uint16) []byte { return []byte{0b00001111} }},
	}
	result["RRA"] = []InstructionParams{
		{Types: []ParamType{}, Assembler: func(args []uint16) []byte { return []byte{0b00011111} }},
	}
	result["BIT"] = []InstructionParams{
		{
			Types: []ParamType{BitOrdinal, Reg8},
			Assembler: func(args []uint16) []byte {
				return []byte{0b11001011, 0b01000000 | (uint8(args[0]) << 3) | uint8(args[1])}
			},
		},
	}
	result["SET"] = []InstructionParams{
		{
			Types: []ParamType{BitOrdinal, Reg8},
			Assembler: func(args []uint16) []byte {
				return []byte{0b11001011, 0b11000000 | (uint8(args[0]) << 3) | uint8(args[1])}
			},
		},
	}
	result["RES"] = []InstructionParams{
		{
			Types: []ParamType{BitOrdinal, Reg8},
			Assembler: func(args []uint16) []byte {
				return []byte{0b11001011, 0b10000000 | (uint8(args[0]) << 3) | uint8(args[1])}
			},
		},
	}
	result["RLC"] = []InstructionParams{
		{
			Types:     []ParamType{Reg8},
			Assembler: func(args []uint16) []byte { return []byte{0b11001011, 0b00000000 | uint8(args[0])} },
		},
	}
	result["RL"] = []InstructionParams{
		{
			Types:     []ParamType{Reg8},
			Assembler: func(args []uint16) []byte { return []byte{0b11001011, 0b00010000 | uint8(args[0])} },
		},
	}
	result["RRC"] = []InstructionParams{
		{
			Types:     []ParamType{Reg8},
			Assembler: func(args []uint16) []byte { return []byte{0b11001011, 0b00001000 | uint8(args[0])} },
		},
	}
	result["RR"] = []InstructionParams{
		{
			Types:     []ParamType{Reg8},
			Assembler: func(args []uint16) []byte { return []byte{0b11001011, 0b00011000 | uint8(args[0])} },
		},
	}
	result["SLA"] = []InstructionParams{
		{
			Types:     []ParamType{Reg8},
			Assembler: func(args []uint16) []byte { return []byte{0b11001011, 0b00100000 | uint8(args[0])} },
		},
	}
	result["SWAP"] = []InstructionParams{
		{
			Types:     []ParamType{Reg8},
			Assembler: func(args []uint16) []byte { return []byte{0b11001011, 0b00110000 | uint8(args[0])} },
		},
	}
	result["SRA"] = []InstructionParams{
		{
			Types:     []ParamType{Reg8},
			Assembler: func(args []uint16) []byte { return []byte{0b11001011, 0b00101000 | uint8(args[0])} },
		},
	}
	result["SRL"] = []InstructionParams{
		{
			Types:     []ParamType{Reg8},
			Assembler: func(args []uint16) []byte { return []byte{0b11001011, 0b00111000 | uint8(args[0])} },
		},
	}

	return result
}

func (set InstructionSet) Parse(
	state *ProgramState,
	line string,
) ([]byte, error) {
	words := strings.Fields(strings.ReplaceAll(strings.Trim(line, " \t\n"), ",", " "))

	if len(words) < 1 {
		return []uint8{}, nil
	}

	instruction, ok := set[words[0]]
	if !ok {
		return nil, fmt.Errorf("Unknown instruction \"%s\"", words[0])
	}

	params := words[1:]

instruction_param_loop:
	for _, instrParam := range instruction {
		if len(instrParam.Types) != len(params) {
			continue
		}

		parsed_params := make([]uint16, len(params))
		for i, paramType := range instrParam.Types {
			parsed, err := paramType(state, params[i])
			if err != nil {
				continue instruction_param_loop
			}

			parsed_params[i] = parsed
		}

		return instrParam.Assembler(parsed_params), nil
	}
	return nil, fmt.Errorf(
		"Instruction \"%s\" doesn't have a parameter set that can parse \"%s\"",
		words[0],
		line,
	)
}
