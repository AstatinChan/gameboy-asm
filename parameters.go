package main

import (
	"fmt"
	"strconv"
	"strings"
)

func Reg8(_ *ProgramState, param string) (uint16, error) {
	switch param {
	case "A":
		return 7, nil
	case "B":
		return 0, nil
	case "C":
		return 1, nil
	case "D":
		return 2, nil
	case "E":
		return 3, nil
	case "H":
		return 4, nil
	case "L":
		return 5, nil
	case "(HL)":
		return 6, nil
	}
	return 0, fmt.Errorf("Invalid reg8")
}

func A(_ *ProgramState, param string) (uint16, error) {
	if param == "A" {
		return 0, nil
	}
	return 0, fmt.Errorf("Invalid A")
}

func HL(_ *ProgramState, param string) (uint16, error) {
	if param == "HL" {
		return 0, nil
	}
	return 0, fmt.Errorf("Invalid HL")
}

func SP(_ *ProgramState, param string) (uint16, error) {
	if param == "SP" {
		return 0, nil
	}
	return 0, fmt.Errorf("Invalid SP")
}

func IndirectC(_ *ProgramState, param string) (uint16, error) {
	if param == "(C)" {
		return 0, nil
	}
	return 0, fmt.Errorf("Invalid (C)")
}

func Reg16(_ *ProgramState, param string) (uint16, error) {
	switch param {
	case "BC":
		return 0, nil
	case "DE":
		return 1, nil
	case "HL":
		return 2, nil
	case "AF":
		return 3, nil
	// TODO Split in two different for push and not push instructions
	case "SP":
		return 3, nil
	}
	return 0, fmt.Errorf("Invalid reg16")
}

func Raw8(_ *ProgramState, param string) (uint16, error) {
	res, err := strconv.ParseInt(param, 0, 8)
	return uint16(res), err
}

func Raw16(state *ProgramState, param string) (uint16, error) {
	res, err := strconv.ParseInt(param, 0, 16)

	if strings.HasPrefix(param, "=") {
		if state == nil {
			return 0, nil
		}

		label := strings.ToUpper(strings.TrimPrefix(param, "="))
		labelValue, ok := state.Labels[label]
		if !ok {
			return 0, fmt.Errorf("Label \"%s\" not found", label)
		}

		// TODO: Manage when multiple MBC
		if labelValue > 0x8000 {
			panic("Switchable ROM banks are not implemented yet")
		}

		return uint16(labelValue), nil
	}

	return uint16(res), err
}

func Reg16Indirect(_ *ProgramState, param string) (uint16, error) {
	switch param {
	case "(BC)":
		return 0, nil
	case "(DE)":
		return 1, nil
	case "(HL+)":
		return 2, nil
	case "(HL-)":
		return 3, nil
	}
	return 0, fmt.Errorf("Invalid reg16 indirect")
}

func Raw8Indirect(_ *ProgramState, param string) (uint16, error) {
	if len(param) < 2 || param[0] != '(' || param[len(param)-1] != ')' {
		return 0, fmt.Errorf("Invalid raw8indirect")
	}

	res, err := strconv.ParseInt(param[1:len(param)-1], 0, 8)
	return uint16(res), err
}

func Raw16Indirect(_ *ProgramState, param string) (uint16, error) {
	if len(param) < 2 || param[0] != '(' || param[len(param)-1] != ')' {
		return 0, fmt.Errorf("Invalid raw16indirect")
	}

	res, err := strconv.ParseInt(param[1:len(param)-1], 0, 16)
	return uint16(res), err
}

func Condition(_ *ProgramState, param string) (uint16, error) {
	switch param {
	case "NZ":
		return 0, nil
	case "Z":
		return 1, nil
	case "NC":
		return 2, nil
	case "C":
		return 3, nil
	}
	return 0, fmt.Errorf("Invalid condition")
}

func BitOrdinal(_ *ProgramState, param string) (uint16, error) {
	res, err := strconv.ParseInt(param, 0, 3)
	return uint16(res), err
}
