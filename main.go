package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type ProgramState struct {
	Labels map[string]uint
	Defs   map[string]any
}

func parseFile(input_file_name string, input []byte) ([]byte, error) {
	lines := strings.Split(string(input), "\n")

	state := ProgramState{
		Labels: make(map[string]uint),
		Defs:   make(map[string]any),
	}

	label_line_number := uint(0)
	for line_nb, line := range lines {
		is_label_defined := strings.Contains(line, ":")

		if is_label_defined {
			parts := strings.Split(line, ":")
			for _, label := range parts[:len(parts)-1] {
				label = strings.ToUpper(label)
				if _, ok := state.Labels[label]; ok {
					fmt.Fprintf(
						os.Stderr,
						"File %s, line %d:\nLabel %s is already defined",
						input_file_name,
						line_nb,
						label,
					)
					os.Exit(1)
				}
				state.Labels[label] = label_line_number
			}

			line = parts[len(parts)-1]
		}

		line = strings.TrimSpace(line)

		next_instruction, err := Instructions.Parse(nil, line)
		if err != nil {
			return nil, fmt.Errorf(
				"File %s, line %d (1st pass):\n%w",
				input_file_name,
				line_nb,
				err,
			)
		}

		// TODO: Handle the case of program bigger than MBC (or maybe do it directly in the parameters)
		label_line_number += uint(len(next_instruction))
	}

	result := []byte{}
	for line_nb, line := range lines {
		is_label_defined := strings.Contains(line, ":")

		if is_label_defined {
			parts := strings.Split(line, ":")

			line = parts[len(parts)-1]
		}

		next_instruction, err := Instructions.Parse(&state, line)
		if err != nil {
			return nil, fmt.Errorf(
				"File %s, line %d (2nd pass): %w",
				input_file_name,
				line_nb,
				err,
			)
		}

		result = append(result, next_instruction...)
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

	result, err := parseFile(input_file_name, input)
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
