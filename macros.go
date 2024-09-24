package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var MacroInstructions = NewInstructionSetMacros()

func NewInstructionSetMacros() InstructionSet {
	result := make(InstructionSet)

	result[".PADTO"] = []InstructionParams{
		{
			Types: []ParamType{Raw16},
			Assembler: func(currentAddress uint16, args []uint16) ([]byte, error) {
				fmt.Printf(
					".PADTO 0x%04x, currentAddress: 0x%04x, inserting: 0x%04x\n",
					args[0],
					currentAddress,
					args[0]-currentAddress,
				)
				return make([]byte, args[0]-currentAddress), nil
			},
		},
	}

	result[".DB"] = []InstructionParams{
		{
			Types: []ParamType{Raw8},
			Assembler: func(currentAddress uint16, args []uint16) ([]byte, error) {
				result := make([]byte, len(args))
				for i := range args {
					result[i] = uint8(args[i])
				}
				return result, nil
			},
			Wildcard: true,
		},
	}

	return result
}

type (
	Indirect8b  uint16
	Indirect16b uint16
	Raw8b       uint16
	Raw16b      uint16
)

func MacroParse(
	line string,
	result *[]byte,
	state *ProgramState,
	_ *int, // line_nb
	is_first_pass bool,
	offset uint,
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
			uint16(uint(len(*result))+offset),
			line,
		)
		if err != nil {
			return fmt.Errorf("Macro instruction parsing failed %w", err)
		}

		*result = append(*result, new_instruction...)
		return nil
	} else if macroName == ".INCLUDE" {
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
		if is_first_pass {
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
	} else if macroName == ".DEFINE" {
		if len(words) != 3 {
			return fmt.Errorf(".DEFINE must have 2 arguments (%v)", words)
		}

		name := strings.ToUpper(words[1])
		_, err := strconv.ParseUint(name, 16, 16)
		if err == nil {
			return fmt.Errorf("Defined variable \"%s\" is also valid hexadecimal", name)
		}

		var definedValue any
		if v, err := Raw8Indirect(&state.Labels, &state.Defs, words[2]); err == nil {
			definedValue = Indirect8b(v)
		} else if v, err := Raw16Indirect(&state.Labels, &state.Defs, words[2]); err == nil {
			definedValue = Indirect16b(v)
		} else if v, err := Raw8(&state.Labels, &state.Defs, words[2]); err == nil {
			definedValue = Raw8b(v)
		} else if v, err := Raw16(&state.Labels, &state.Defs, words[2]); err == nil {
			definedValue = Raw16b(v)
		} else {
			return fmt.Errorf("\"%s\" could not be parsed as a .DEFINE argument", words[2])
		}

		state.Defs[name] = definedValue
	} else {
		return fmt.Errorf("Unknown macro \"%s\"", macroName)
	}
	return nil
}
