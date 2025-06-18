package main

import (
	"fmt"
	"strings"
)

type ParamType func(labels *Labels, lastAbsoluteLabel string, defs *Definitions, currentAddr uint32, param string) (uint32, error)

type InstructionParams struct {
	Types            []ParamType
	Assembler        func(currentAddress uint32, args []uint32) ([]uint8, error)
	Wildcard         bool
	MacroForbidden   bool
	LabelsBeforeOnly bool
	SkipFirstPass    bool
}

type InstructionSet map[string][]InstructionParams

var Instructions = InstructionSetNew()

func absoluteJPValueToRelative(baseAddress uint32, absoluteAddress uint32) (uint8, error) {
	newAddress := (int32(absoluteAddress) - int32(baseAddress) - 2)
	if newAddress < -127 || newAddress > 128 {
		return 0, fmt.Errorf(
			"Address 0x%04x and 0x%04x are too far apart to use JR. Please use JP instead",
			baseAddress,
			absoluteAddress,
		)
	}
	return uint8(newAddress & 0xff), nil
}

func InstructionSetNew() InstructionSet {
	result := make(InstructionSet)

	result["LD"] = []InstructionParams{
		{
			Types: []ParamType{Reg8, Reg8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) {
				return []byte{
					0b01000000 | (uint8(uint8(args[0])) << 3) | uint8(uint8(args[1])),
				}, nil
			},
		},
		// {
		// 	Types:     []ParamType{HL, Raw8Indirect},
		// 	Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b11111000, uint8(args[1])}, nil },
		// },
		{
			Types: []ParamType{Reg8, Raw8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) {
				return []byte{0b00000110 | (uint8(args[0]) << 3), uint8(args[1])}, nil
			},
		},
		{
			Types:     []ParamType{A, Raw8Indirect},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b11110000, uint8(args[1])}, nil },
		},
		{
			Types:     []ParamType{Raw8Indirect, A},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b11100000, uint8(args[0])}, nil },
		},
		{
			Types:     []ParamType{A, Reg16Indirect},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b00001010 | uint8(args[1])<<4}, nil },
		},
		{
			Types:     []ParamType{Reg16Indirect, A},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b00000010 | uint8(args[0])<<4}, nil },
		},
		{
			Types: []ParamType{A, Raw16Indirect},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) {
				return []byte{0b11111010, uint8(args[1]) & 0xff, uint8(args[1] >> 8)}, nil
			},
		},
		{
			Types: []ParamType{Raw16Indirect, A},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) {
				return []byte{0b11101010, uint8(args[0]) & 0xff, uint8(args[0] >> 8)}, nil
			},
		},
		{
			Types:     []ParamType{A, IndirectC},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b11110010}, nil },
		},
		{
			Types:     []ParamType{IndirectC, A},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b11100010}, nil },
		},
		{
			Types: []ParamType{Reg16, Raw16},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) {
				return []byte{
					0b00000001 | (uint8(args[0]) << 4),
					uint8(args[1]) & 0xff,
					uint8(args[1] >> 8),
				}, nil
			},
		},
		{
			Types: []ParamType{Raw16Indirect, SP},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) {
				return []byte{0b00001000, uint8(args[0]) & 0xff, uint8(args[0] >> 8)}, nil
			},
		},
		{
			Types:     []ParamType{SP, HL},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b11111001}, nil },
		},
	}
	result["PUSH"] = []InstructionParams{
		{
			Types:     []ParamType{Reg16},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b11000101 | (uint8(args[0]) << 4)}, nil },
		},
	}
	result["POP"] = []InstructionParams{
		{
			Types:     []ParamType{Reg16},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b11000001 | (uint8(args[0]) << 4)}, nil },
		},
	}
	result["ADD"] = []InstructionParams{
		{
			Types:     []ParamType{Reg8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b10000000 | (uint8(args[0]))}, nil },
		},
		{
			Types:     []ParamType{Raw8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b11000110, uint8(args[0])}, nil },
		},
		{
			Types:     []ParamType{SP, Raw8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b11101000, uint8(args[1])}, nil },
		},
	}
	result["ADC"] = []InstructionParams{
		{
			Types:     []ParamType{Reg8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b10001000 | uint8(args[0])}, nil },
		},
		{
			Types:     []ParamType{Raw8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b11001110, uint8(args[0])}, nil },
		},
	}
	result["SUB"] = []InstructionParams{
		{
			Types:     []ParamType{Reg8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b10010000 | (uint8(args[0]))}, nil },
		},
		{
			Types:     []ParamType{Raw8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b11010110, uint8(args[0])}, nil },
		},
	}
	result["SBC"] = []InstructionParams{
		{
			Types:     []ParamType{Reg8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b10011000 | uint8(args[0])}, nil },
		},
		{
			Types:     []ParamType{Raw8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b11011110, uint8(args[0])}, nil },
		},
	}
	result["CP"] = []InstructionParams{
		{
			Types:     []ParamType{Reg8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b10111000 | (uint8(args[0]))}, nil },
		},
		{
			Types:     []ParamType{Raw8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b11111110, uint8(args[0])}, nil },
		},
	}
	result["INC"] = []InstructionParams{
		{
			Types:     []ParamType{Reg8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b00000100 | (uint8(args[0]) << 3)}, nil },
		},

		{
			Types:     []ParamType{Reg16},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b00000011 | (uint8(args[0]) << 4)}, nil },
		},
	}
	result["DEC"] = []InstructionParams{
		{
			Types:     []ParamType{Reg8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b00000101 | (uint8(args[0]) << 3)}, nil },
		},

		{
			Types:     []ParamType{Reg16},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b00001011 | (uint8(args[0]) << 4)}, nil },
		},
	}
	result["AND"] = []InstructionParams{
		{
			Types:     []ParamType{Reg8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b10100000 | (uint8(args[0]))}, nil },
		},
		{
			Types:     []ParamType{Raw8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b11100110, uint8(args[0])}, nil },
		},
	}
	result["OR"] = []InstructionParams{
		{
			Types:     []ParamType{Reg8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b10110000 | (uint8(args[0]))}, nil },
		},
		{
			Types:     []ParamType{Raw8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b11110110, uint8(args[0])}, nil },
		},
	}
	result["XOR"] = []InstructionParams{
		{
			Types:     []ParamType{Reg8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b10101000 | (uint8(args[0]))}, nil },
		},
		{
			Types:     []ParamType{Raw8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b11101110, uint8(args[0])}, nil },
		},
	}
	result["CCF"] = []InstructionParams{
		{
			Types:     []ParamType{},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b00111111}, nil },
		},
	}
	result["SCF"] = []InstructionParams{
		{
			Types:     []ParamType{},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b00110111}, nil },
		},
	}
	result["DAA"] = []InstructionParams{
		{
			Types:     []ParamType{},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b00100111}, nil },
		},
	}
	result["CPL"] = []InstructionParams{
		{
			Types:     []ParamType{},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b00101111}, nil },
		},
	}
	result["JP"] = []InstructionParams{
		{
			Types: []ParamType{Raw16},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) {
				return []byte{0b11000011, uint8(args[0]) & 0xff, uint8(args[0] >> 8)}, nil
			},
		},
		{
			Types:     []ParamType{HL},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b11101001}, nil },
		},
		{
			Types: []ParamType{Condition, Raw16},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) {
				return []byte{
					0b11000010 | (uint8(args[0]) << 3),
					uint8(args[1]) & 0xff,
					uint8(args[1] >> 8),
				}, nil
			},
		},
	}
	result["JR"] = []InstructionParams{
		{
			Types:     []ParamType{Raw8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b00011000, uint8(args[0])}, nil },
		},
		{
			Types: []ParamType{Condition, Raw8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) {
				return []byte{0b00100000 | (uint8(args[0]) << 3), uint8(args[1])}, nil
			},
		},
		{
			Types: []ParamType{Raw16},
			Assembler: func(currentAddress uint32, args []uint32) ([]byte, error) {
				relativeAddress, err := absoluteJPValueToRelative(currentAddress, args[0])
				if err != nil {
					return nil, err
				}
				return []byte{0b00011000, relativeAddress}, nil
			},
		},
		{
			Types: []ParamType{Condition, Raw16},
			Assembler: func(currentAddress uint32, args []uint32) ([]byte, error) {
				relativeAddress, err := absoluteJPValueToRelative(currentAddress, args[1])
				if err != nil {
					return nil, err
				}
				return []byte{0b00100000 | (uint8(args[0]) << 3), relativeAddress}, nil
			},
		},
	}
	result["CALL"] = []InstructionParams{
		{
			Types: []ParamType{Raw16},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) {
				return []byte{0b11001101, uint8(args[0]) & 0xff, uint8(args[0] >> 8)}, nil
			},
		},
		{
			Types: []ParamType{Condition, Raw16},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) {
				return []byte{
					0b11000100 | (uint8(args[0]) << 3),
					uint8(args[1]) & 0xff,
					uint8(args[1] >> 8),
				}, nil
			},
		},
	}
	result["RET"] = []InstructionParams{
		{
			Types:     []ParamType{},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b11001001}, nil },
		},
		{
			Types:     []ParamType{Condition},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b11000000 | (uint8(args[0]) << 3)}, nil },
		},
	}
	result["RETI"] = []InstructionParams{
		{
			Types:     []ParamType{},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b11011001}, nil },
		},
	}
	result["RST"] = []InstructionParams{
		{
			Types:     []ParamType{BitOrdinal},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b11000111 | (uint8(args[0]) << 3)}, nil },
		},
	}
	result["DI"] = []InstructionParams{
		{
			Types:     []ParamType{},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b11110011}, nil },
		},
	}
	result["EI"] = []InstructionParams{
		{
			Types:     []ParamType{},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b11111011}, nil },
		},
	}
	result["NOP"] = []InstructionParams{
		{
			Types:     []ParamType{},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b00000000}, nil },
		},
	}
	result["HALT"] = []InstructionParams{
		{
			Types:     []ParamType{},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b01110110}, nil },
		},
	}
	result["STOP"] = []InstructionParams{
		{
			Types:     []ParamType{},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b00010000, 0b00000000}, nil },
		},
	}
	result["RLCA"] = []InstructionParams{
		{
			Types:     []ParamType{},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b00000111}, nil },
		},
	}
	result["RLA"] = []InstructionParams{
		{
			Types:     []ParamType{},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b00010111}, nil },
		},
	}
	result["RRCA"] = []InstructionParams{
		{
			Types:     []ParamType{},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b00001111}, nil },
		},
	}
	result["RRA"] = []InstructionParams{
		{
			Types:     []ParamType{},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) { return []byte{0b00011111}, nil },
		},
	}
	result["BIT"] = []InstructionParams{
		{
			Types: []ParamType{BitOrdinal, Reg8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) {
				return []byte{0b11001011, 0b01000000 | (uint8(args[0]) << 3) | uint8(args[1])}, nil
			},
		},
	}
	result["SET"] = []InstructionParams{
		{
			Types: []ParamType{BitOrdinal, Reg8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) {
				return []byte{0b11001011, 0b11000000 | (uint8(args[0]) << 3) | uint8(args[1])}, nil
			},
		},
	}
	result["RES"] = []InstructionParams{
		{
			Types: []ParamType{BitOrdinal, Reg8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) {
				return []byte{0b11001011, 0b10000000 | (uint8(args[0]) << 3) | uint8(args[1])}, nil
			},
		},
	}
	result["RLC"] = []InstructionParams{
		{
			Types: []ParamType{Reg8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) {
				return []byte{0b11001011, 0b00000000 | uint8(args[0])}, nil
			},
		},
	}
	result["RL"] = []InstructionParams{
		{
			Types: []ParamType{Reg8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) {
				return []byte{0b11001011, 0b00010000 | uint8(args[0])}, nil
			},
		},
	}
	result["RRC"] = []InstructionParams{
		{
			Types: []ParamType{Reg8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) {
				return []byte{0b11001011, 0b00001000 | uint8(args[0])}, nil
			},
		},
	}
	result["RR"] = []InstructionParams{
		{
			Types: []ParamType{Reg8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) {
				return []byte{0b11001011, 0b00011000 | uint8(args[0])}, nil
			},
		},
	}
	result["SLA"] = []InstructionParams{
		{
			Types: []ParamType{Reg8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) {
				return []byte{0b11001011, 0b00100000 | uint8(args[0])}, nil
			},
		},
	}
	result["SWAP"] = []InstructionParams{
		{
			Types: []ParamType{Reg8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) {
				return []byte{0b11001011, 0b00110000 | uint8(args[0])}, nil
			},
		},
	}
	result["SRA"] = []InstructionParams{
		{
			Types: []ParamType{Reg8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) {
				return []byte{0b11001011, 0b00101000 | uint8(args[0])}, nil
			},
		},
	}
	result["SRL"] = []InstructionParams{
		{
			Types: []ParamType{Reg8},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) {
				return []byte{0b11001011, 0b00111000 | uint8(args[0])}, nil
			},
		},
	}
	result["DBG"] = []InstructionParams{
		{
			Types: []ParamType{},
			Assembler: func(_ uint32, args []uint32) ([]byte, error) {
				return []byte{0b11010011}, nil
			},
		},
	}

	return result
}

func (set InstructionSet) Parse(
	labels *Labels,
	defs *Definitions,
	isMacro bool,
	isFirstPass bool,
	currentAddress uint32,
	lastAbsoluteLabel string,
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

	var rejectedErrors error
instruction_param_loop:
	for _, instrParam := range instruction {
		if instrParam.SkipFirstPass && isFirstPass {
			return []byte{}, nil
		}

		if !instrParam.Wildcard && len(instrParam.Types) != len(params) {
			continue
		}

		parsed_params := make([]uint32, len(params))
		for i := range parsed_params {
			var paramType ParamType
			if instrParam.Wildcard {
				paramType = instrParam.Types[0]
			} else {
				paramType = instrParam.Types[i]
			}

			accessibleLabels := labels

			if isFirstPass && !instrParam.LabelsBeforeOnly {
				accessibleLabels = nil
			}
			parsed, err := paramType(accessibleLabels, lastAbsoluteLabel, defs, currentAddress, params[i])
			if err != nil {
				rejectedError := fmt.Errorf("\t[Rejected] Param Type %v: %w\n", paramType, err)
				if rejectedErrors == nil {
					rejectedErrors = rejectedError
				} else {
					rejectedErrors = fmt.Errorf("%w%w", rejectedErrors, rejectedError)
				}
				continue instruction_param_loop
			}

			parsed_params[i] = parsed
		}

		if instrParam.MacroForbidden && isMacro {
			rejectedError := fmt.Errorf("\t[Rejected] This instruction cannot be used with this set of params inside of a macro\n")
			if rejectedErrors == nil {
				rejectedErrors = rejectedError
			} else {
				rejectedErrors = fmt.Errorf("%w%w", rejectedErrors, rejectedError)
			}
			continue
			// return nil, fmt.Errorf("")
		}

		return instrParam.Assembler(currentAddress, parsed_params)
	}
	return nil, fmt.Errorf(
		"Instruction \"%s\" doesn't have a parameter set that can parse \"%s\"\n%w",
		words[0],
		line,
		rejectedErrors,
	)
}
