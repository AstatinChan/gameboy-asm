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

func printSymbols(labels map[string]uint) {
	for key, value := range labels {
		if value < 0x4000 {
			fmt.Printf("00:%04x %s\n", value, key)
		} else {
			bank := value / 0x4000
			addr := value%0x4000 + 0x4000
			fmt.Printf("%02x:%04x %s\n", bank, addr, key)
		}
	}
}

func parseFile(inputFileName string, input []byte, offset uint) ([]byte, error) {
	state := ProgramState{
		Labels:  make(map[string]uint),
		Defs:    make(map[string]any),
		IsMacro: false,
	}

	_, err := firstPass(inputFileName, input, offset, &state)
	if err != nil {
		return nil, err
	}
	printSymbols(state.Labels)
	return secondPass(inputFileName, input, offset, state)
}

func firstPass(
	inputFileName string,
	input []byte,
	offset uint,
	state *ProgramState,
) ([]byte, error) {
	lines := strings.Split(string(input), "\n")

	lineNb := 0
	result := []byte{}
	lastAbsoluteLabel := ""
	for lineNb < len(lines) {
		line := lines[lineNb]
		lineParts := strings.Split(line, ";")
		line = lineParts[0]
		isLabelDefined := strings.Contains(line, ":")

		if isLabelDefined {
			parts := strings.Split(line, ":")
			for _, label := range parts[:len(parts)-1] {
				label = strings.TrimSpace(strings.ToUpper(label))
				isCharsetAllowed := regexp.MustCompile(`^[a-zA-Z0-9_.$-]*$`).MatchString(label)
				if !isCharsetAllowed {
					return nil, fmt.Errorf(
						"File %s, line %d:\nLabel \"%s\" contains special characters. Only alphanumeric, dashes and underscores are allowed",
						inputFileName,
						lineNb+1,
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
						inputFileName,
						lineNb+1,
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
			err := MacroParse(line, lines, &result, state, &lineNb, true, offset, lastAbsoluteLabel)
			if err != nil {
				return nil, fmt.Errorf(
					"File %s, line %d (1st pass|macro):\n%w",
					inputFileName,
					lineNb+1,
					err,
				)
			}
		} else {
			nextInstruction, err := Instructions.Parse(&state.Labels, &state.Defs, state.IsMacro, true, 0, lastAbsoluteLabel, line)
			if err != nil {
				return nil, fmt.Errorf(
					"File %s, line %d (1st pass):\n%w",
					inputFileName,
					lineNb+1,
					err,
				)
			}

			result = append(result, nextInstruction...)
		}
		lineNb += 1
	}

	return result, nil
}

func secondPass(
	inputFileName string,
	input []byte,
	offset uint,
	state ProgramState,
) ([]byte, error) {
	lines := strings.Split(string(input), "\n")

	lineNb := 0
	result := []byte{}
	lastAbsoluteLabel := ""
	for lineNb < len(lines) {
		line := lines[lineNb]
		lineParts := strings.Split(line, ";")
		line = lineParts[0]
		isLabelDefined := strings.Contains(line, ":")

		if isLabelDefined {
			parts := strings.Split(line, ":")

			line = parts[len(parts)-1]

			for _, label := range parts[:len(parts)-1] {
				label = strings.TrimSpace(strings.ToUpper(label))
				if !strings.HasPrefix(label, ".") {
					labelParts := strings.Split(label, ".")
					if len(labelParts) < 1 {
						return nil, fmt.Errorf(
							"Unknown issue while retrieving label absolute part ! (label: \"%s\")",
							label,
						)
					}

					lastAbsoluteLabel = labelParts[0]
				}
			}
		}

		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, ".") {
			err := MacroParse(
				line,
				lines,
				&result,
				&state,
				&lineNb,
				false,
				offset,
				lastAbsoluteLabel,
			)
			if err != nil {
				return nil, fmt.Errorf(
					"File %s, line %d (2nd pass|macro):\n%w",
					inputFileName,
					lineNb+1,
					err,
				)
			}
		} else {
			nextInstruction, err := Instructions.Parse(&state.Labels, &state.Defs, state.IsMacro, false, uint16(uint(len(result))+offset), lastAbsoluteLabel, line)
			if err != nil {
				return nil, fmt.Errorf(
					"File %s, line %d (2nd pass): %w",
					inputFileName,
					lineNb+1,
					err,
				)
			}

			result = append(result, nextInstruction...)
		}
		lineNb += 1
	}

	return result, nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: gbasm [input_file] [output_file]\n")
		os.Exit(1)
	}

	inputFileName := os.Args[1]
	outputFileName := os.Args[2]

	inputFile, err := os.Open(inputFileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while opening input file: %s\n", err.Error())
		os.Exit(1)
	}

	input, err := io.ReadAll(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading input file: %s\n", err.Error())
		os.Exit(1)
	}

	result, err := parseFile(inputFileName, input, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}

	outputFile, err := os.Create(outputFileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while opening output file: %s\n", err.Error())
		os.Exit(1)
	}

	_, err = outputFile.Write(result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing to output file: %s\n", err.Error())
		os.Exit(1)
	}
}
