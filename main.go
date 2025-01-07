package main

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

type (
	Labels      map[string]uint
	Definitions map[string]any
)

type ProgramState struct {
	Labels  Labels
	Defs    Definitions
	IsMacro bool
}

func parseFile(input_file_name string, input []byte, offset uint) ([]byte, error) {
	state := ProgramState{
		Labels:  make(map[string]uint),
		Defs:    make(map[string]any),
		IsMacro: false,
	}

	_, err := firstPass(input_file_name, input, offset, &state)
	if err != nil {
		return nil, err
	}
	return secondPass(input_file_name, input, offset, state)
}

func firstPass(
	input_file_name string,
	input []byte,
	offset uint,
	state *ProgramState,
) ([]byte, error) {
	lines := strings.Split(string(input), "\n")

	line_nb := 0
	result := []byte{}
	lastAbsoluteLabel := ""
	for line_nb < len(lines) {
		line := lines[line_nb]
		line_parts := strings.Split(line, ";")
		line = line_parts[0]
		is_label_defined := strings.Contains(line, ":")

		if is_label_defined {
			parts := strings.Split(line, ":")
			for _, label := range parts[:len(parts)-1] {
				label = strings.TrimSpace(strings.ToUpper(label))
				isCharsetAllowed := regexp.MustCompile(`^[a-zA-Z0-9_.$-]*$`).MatchString(label)
				if !isCharsetAllowed {
					return nil, fmt.Errorf(
						"File %s, line %d:\nLabel \"%s\" contains special characters. Only alphanumeric, dashes and underscores are allowed",
						input_file_name,
						line_nb+1,
						label,
					)
				}

				if strings.HasPrefix(label, ".") {
					if lastAbsoluteLabel == "" {
						return nil, fmt.Errorf(
							"Relative label \"%s\" found without a parent",
							label,
						)
					}

					label = lastAbsoluteLabel + label
				} else {
					labelParts := strings.Split(label, ".")
					if len(labelParts) < 1 {
						return nil, fmt.Errorf("Unknown issue while retrieving label absolute part ! (label: \"%s\")", label)
					}

					lastAbsoluteLabel = labelParts[0]
				}

				if _, ok := state.Labels[label]; ok {
					return nil, fmt.Errorf(
						"File %s, line %d:\nLabel %s is already defined",
						input_file_name,
						line_nb+1,
						label,
					)
				}

				if label[0] == '$' && !state.IsMacro {
					return nil, fmt.Errorf("Labels starting with $ can only be used inside macros")
				}
				if label[0] != '$' && state.IsMacro {
					return nil, fmt.Errorf("Labels inside a macro must start with $")
				}

				state.Labels[label] = uint(len(result)) + offset
			}

			line = parts[len(parts)-1]
		}

		line = strings.TrimSpace(line)

		// nil sets all the labels and defintion to 0 & thus, to not crash JR, the currentAddress should also be 0
		if strings.HasPrefix(line, ".") {
			err := MacroParse(line, lines, &result, state, &line_nb, true, offset)
			if err != nil {
				return nil, fmt.Errorf(
					"File %s, line %d (1st pass|macro):\n%w",
					input_file_name,
					line_nb+1,
					err,
				)
			}
		} else {
			next_instruction, err := Instructions.Parse(&state.Labels, &state.Defs, state.IsMacro, true, 0, line)
			if err != nil {
				return nil, fmt.Errorf(
					"File %s, line %d (1st pass):\n%w",
					input_file_name,
					line_nb+1,
					err,
				)
			}

			result = append(result, next_instruction...)
		}
		line_nb += 1
	}

	return result, nil
}

func secondPass(
	input_file_name string,
	input []byte,
	offset uint,
	state ProgramState,
) ([]byte, error) {
	lines := strings.Split(string(input), "\n")

	line_nb := 0
	result := []byte{}
	for line_nb < len(lines) {
		line := lines[line_nb]
		line_parts := strings.Split(line, ";")
		line = line_parts[0]
		is_label_defined := strings.Contains(line, ":")

		if is_label_defined {
			parts := strings.Split(line, ":")

			line = parts[len(parts)-1]
		}

		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, ".") {
			err := MacroParse(line, lines, &result, &state, &line_nb, false, offset)
			if err != nil {
				return nil, fmt.Errorf(
					"File %s, line %d (2nd pass|macro):\n%w",
					input_file_name,
					line_nb+1,
					err,
				)
			}
		} else {
			next_instruction, err := Instructions.Parse(&state.Labels, &state.Defs, state.IsMacro, false, uint16(uint(len(result))+offset), line)
			if err != nil {
				return nil, fmt.Errorf(
					"File %s, line %d (2nd pass): %w",
					input_file_name,
					line_nb+1,
					err,
				)
			}

			result = append(result, next_instruction...)
		}
		line_nb += 1
	}

	return result, nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: gbasm [input_file] [output_file]\n")
		os.Exit(1)
	}

	input_file_name := os.Args[1]
	output_file_name := os.Args[2]

	input_file, err := os.Open(input_file_name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while opening input file: %s\n", err.Error())
		os.Exit(1)
	}

	input, err := io.ReadAll(input_file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading input file: %s\n", err.Error())
		os.Exit(1)
	}

	result, err := parseFile(input_file_name, input, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}

	output_file, err := os.Create(output_file_name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while opening output file: %s\n", err.Error())
		os.Exit(1)
	}

	_, err = output_file.Write(result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing to output file: %s\n", err.Error())
		os.Exit(1)
	}
}
