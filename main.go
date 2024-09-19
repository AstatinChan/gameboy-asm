package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: gbasm [input_file] [output_file]")
		os.Exit(1)
	}

	input_file_name := os.Args[1]
	output_file_name := os.Args[2]

	input_file, err := os.Open(input_file_name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while opening input file: %s", err.Error())
		os.Exit(1)
	}

	input, err := io.ReadAll(input_file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading input file: %s", err.Error())
		os.Exit(1)
	}

	lines := strings.Split(string(input), "\n")

	result := []byte{}
	for _, line := range lines {
		next_instruction, err := Instructions.Parse(line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err.Error())
			os.Exit(1)
		}

		result = append(result, next_instruction...)
	}

	output_file, err := os.Create(output_file_name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while opening output file: %s", err.Error())
		os.Exit(1)
	}

	_, err = output_file.Write(result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while writing to output file: %s", err.Error())
		os.Exit(1)
	}
}
