package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var MacroInstructions = NewInstructionSetMacros()

func InlineMacroAssembler(b []byte) []InstructionParams {
	return []InstructionParams{
		{
			Types: []ParamType{},
			Assembler: func(currentAddress uint32, args []uint32) ([]uint8, error) {
				return b, nil
			},
		},
	}
}

func NewInstructionSetMacros() InstructionSet {
	result := make(InstructionSet)

	result[".PADTO"] = []InstructionParams{
		{
			Types: []ParamType{Raw16},
			Assembler: func(currentAddress uint32, args []uint32) ([]byte, error) {
				return make([]byte, args[0]-currentAddress), nil
			},
			MacroForbidden:   true,
			LabelsBeforeOnly: true,
		},
		{
			Types: []ParamType{Raw16MacroRelativeLabel},
			Assembler: func(currentAddress uint32, args []uint32) ([]byte, error) {
				return make([]byte, args[0]-currentAddress), nil
			},
			MacroForbidden:   false,
			LabelsBeforeOnly: true,
		},
	}

	result[".DB"] = []InstructionParams{
		{
			Types: []ParamType{Raw8},
			Assembler: func(currentAddress uint32, args []uint32) ([]byte, error) {
				result := make([]byte, len(args))
				for i := range args {
					result[i] = uint8(args[i])
				}
				return result, nil
			},
			Wildcard: true,
		},
		{
			Types: []ParamType{Raw16},
			Assembler: func(currentAddress uint32, args []uint32) ([]byte, error) {
				result := make([]byte, len(args)*2)
				for i := range args {
					result[i*2] = uint8(args[i] >> 8)
					result[i*2+1] = uint8(args[i] & 0xff)
				}
				return result, nil
			},
			Wildcard: true,
		},
	}

	return result
}

type (
	Indirect8b    uint32
	Indirect16b   uint32
	Raw8b         uint32
	Raw16b        uint32
	ROMAddressPtr uint32
)

func MacroParse(
	line string,
	lines []string,
	result *[]byte,
	state *ProgramState,
	lineNb *int,
	isFirstPass bool,
	offset uint,
	LastAbsoluteLabel string,
) error {
	words := strings.Split(line, " ")
	if len(words) == 0 {
		return fmt.Errorf("Macro parsed doesn't accept empty lines")
	}

	macroName := words[0]

	if _, ok := MacroInstructions[macroName]; ok {
		new_instruction, err := MacroInstructions.Parse(
			&state.Labels,
			&state.Defs,
			state.IsMacro,
			isFirstPass,
			uint32(uint(len(*result))+offset),
			LastAbsoluteLabel,
			line,
		)
		if err != nil {
			return fmt.Errorf("Macro instruction parsing failed %w", err)
		}

		*result = append(*result, new_instruction...)
		return nil
	} else if macroName == ".INCLUDE" && !state.IsMacro {
		filePath := strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, ".INCLUDE")), "\"'")

		input_file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("Error while opening file %s", filePath)
		}

		input, err := io.ReadAll(input_file)
		if err != nil {
			return fmt.Errorf("Error while reading file %s", filePath)
		}

		fileStartOffset := uint(len(*result)) + offset
		if isFirstPass {
			included, err := firstPass(filePath, input, fileStartOffset, state)
			if err != nil {
				return err
			}
			*result = append(*result, included...)
		} else {
			included, err := secondPass(filePath, input, fileStartOffset, *state)
			if err != nil {
				return err
			}
			*result = append(*result, included...)
		}
	} else if macroName == ".DEFINE" && !state.IsMacro {
		if len(words) != 3 {
			return fmt.Errorf(".DEFINE must have 2 arguments (%v)", words)
		}

		name := strings.ToUpper(words[1])
		_, err := strconv.ParseUint(name, 16, 16)
		if err == nil {
			return fmt.Errorf("Defined variable \"%s\" is also valid hexadecimal", name)
		}

		var definedValue any
		if v, err := Raw8Indirect(&state.Labels, LastAbsoluteLabel, &state.Defs, 0xffffffff, words[2]); err == nil {
			definedValue = Indirect8b(v)
		} else if v, err := Raw16Indirect(&state.Labels, LastAbsoluteLabel, &state.Defs, 0xffffffff, words[2]); err == nil {
			definedValue = Indirect16b(v)
		} else if v, err := Raw8(&state.Labels, LastAbsoluteLabel, &state.Defs, 0xffffffff, words[2]); err == nil {
			definedValue = Raw8b(v)
		} else if v, err := Raw16(&state.Labels, LastAbsoluteLabel, &state.Defs, 0xffffffff, words[2]); err == nil {
			definedValue = Raw16b(v)
		} else {
			return fmt.Errorf("\"%s\" could not be parsed as a .DEFINE argument", words[2])
		}

		state.Defs[name] = definedValue
	} else if macroName == ".MACRODEF" && !state.IsMacro {
		if len(words) < 2 {
			return fmt.Errorf(".MACRODEF should have at least one argument, followed by the definition")
		}
		definedMacroName := strings.ToUpper(words[1])
		definedMacroArguments := words[2:]
		for i := range definedMacroArguments {
			definedMacroArguments[i] = strings.ToUpper(definedMacroArguments[i])
		}
		(*lineNb) += 1
		macroContent := []byte{}
		for *lineNb < len(lines) && strings.TrimSpace(strings.Split(lines[*lineNb], ";")[0]) != ".END" {
			macroContent = append(macroContent, (lines[*lineNb] + "\n")...)
			(*lineNb) += 1
		}

		parameterTypes := make([]ParamType, len(definedMacroArguments))
		for i, v := range definedMacroArguments {
			if len(v) != 0 && v[0] == '=' {
				parameterTypes[i] = ROMAddress
			} else {
				parameterTypes[i] = Raw16
			}
		}

		if isFirstPass {
			if _, ok := MacroInstructions[definedMacroName]; ok {
				return fmt.Errorf("Macro %s is already defined", definedMacroName)
			}

			MacroInstructions["."+definedMacroName] = []InstructionParams{
				{
					Types: parameterTypes,
					Assembler: func(currentAddress uint32, args []uint32) ([]uint8, error) {
						definitions := Clone(state.Defs)
						labels := Clone(state.Labels)
						for i, macroArg := range definedMacroArguments {
							if len(macroArg) != 0 && macroArg[0] == '=' {
								labels[macroArg[1:]] = uint(args[i])
							} else {
								definitions[macroArg] = Raw16b(args[i])
							}
						}
						state := ProgramState{
							Labels:  labels,
							Defs:    definitions,
							IsMacro: true,
						}
						new_instructions, err := firstPass("MACRO$"+definedMacroName, macroContent, 0, &state)
						if err != nil {
							return nil, err
						}
						return new_instructions, nil
					},
				},
			}
		} else {
			MacroInstructions["."+definedMacroName] = []InstructionParams{
				{
					Types: parameterTypes,
					Assembler: func(currentAddress uint32, args []uint32) ([]uint8, error) {
						definitions := Clone(state.Defs)
						labels := Clone(state.Labels)
						for i, macroArg := range definedMacroArguments {
							if len(macroArg) != 0 && macroArg[0] == '=' {
								labels[macroArg[1:]] = uint(args[i])
							} else {
								definitions[macroArg] = Raw16b(args[i])
							}
						}
						state := ProgramState{
							Labels:  labels,
							Defs:    definitions,
							IsMacro: true,
						}
						_, err := firstPass("MACRO$"+definedMacroName, macroContent, uint(currentAddress), &state)
						if err != nil {
							return nil, err
						}
						new_instructions, err := secondPass("MACRO$"+definedMacroName, macroContent, uint(currentAddress), state)
						if err != nil {
							return nil, err
						}

						return new_instructions, nil
					},
				},
			}
		}
	} else {
		return fmt.Errorf("Unknown macro \"%s\"", macroName)
	}
	return nil
}

func Clone[K comparable, V any](arg map[K]V) map[K]V {
	result := make(map[K]V)
	for k, v := range arg {
		result[k] = v
	}
	return result
}
