package main

import (
	"fmt"
	"strconv"
	"strings"
)

func Reg8(_ *Labels, _ *Definitions, param string) (uint16, error) {
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

func A(_ *Labels, _ *Definitions, param string) (uint16, error) {
	if param == "A" {
		return 0, nil
	}
	return 0, fmt.Errorf("Invalid A")
}

func HL(_ *Labels, _ *Definitions, param string) (uint16, error) {
	if param == "HL" {
		return 0, nil
	}
	return 0, fmt.Errorf("Invalid HL")
}

func SP(_ *Labels, _ *Definitions, param string) (uint16, error) {
	if param == "SP" {
		return 0, nil
	}
	return 0, fmt.Errorf("Invalid SP")
}

func IndirectC(_ *Labels, _ *Definitions, param string) (uint16, error) {
	if param == "(C)" {
		return 0, nil
	}
	return 0, fmt.Errorf("Invalid (C)")
}

func Reg16(_ *Labels, _ *Definitions, param string) (uint16, error) {
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

func Raw8(_ *Labels, defs *Definitions, param string) (uint16, error) {
	if strings.HasPrefix(param, "$") {
		param = strings.ToUpper(strings.TrimPrefix(param, "$"))
		if res, err := strconv.ParseUint(param, 16, 16); err == nil {
			if len(param) > 2 {
				return 0, fmt.Errorf("%s is > 8bit precision", param)
			}
			return uint16(res), nil
		}

		definition, ok := (*defs)[param]
		if !ok {
			return 0, fmt.Errorf("$%s is undefined", param)
		}

		res, ok := definition.(Raw8b)
		if !ok {
			return 0, fmt.Errorf("$%s is of type %T but Raw8b is expected", param, res)
		}
		return uint16(res), nil
	}
	if strings.HasPrefix(param, "0x") && len(param) > 4 {
		return 0, fmt.Errorf("%s is > 8bit precision", param)
	}
	res, err := strconv.ParseUint(param, 0, 8)
	return uint16(res), err
}

func Raw16(labels *Labels, defs *Definitions, param string) (uint16, error) {
	if strings.HasPrefix(param, "$") {
		param = strings.ToUpper(strings.TrimPrefix(param, "$"))
		if res, err := strconv.ParseUint(param, 16, 16); err == nil {
			return uint16(res), nil
		}

		definition, ok := (*defs)[param]
		if !ok {
			return 0, fmt.Errorf("$%s is undefined", param)
		}

		res, ok := definition.(Raw16b)
		if !ok {
			return 0, fmt.Errorf("$%s is of type %T but Raw16b is expected", param, res)
		}
		return uint16(res), nil
	}

	if strings.HasPrefix(param, "=") {
		var offset uint16 = 0
		labelWithoutOffset := param

		if strings.Contains(param, "+") {
			labelParts := strings.Split(param, "+")
			if len(labelParts) != 2 {
				return 0, fmt.Errorf(
					"Labels with offset should have exactly 1 offset (in \"%s\")",
					param,
				)
			}
			labelWithoutOffset = labelParts[0]
			o, err := strconv.ParseUint(labelParts[1], 0, 16)
			if err != nil {
				return 0, fmt.Errorf("Error while parsing label offset: %w", err)
			}
			offset = uint16(o)
		}

		if labels == nil {
			return 0, nil
		}

		label := strings.ToUpper(strings.TrimPrefix(labelWithoutOffset, "="))
		labelValue, ok := (*labels)[label]
		if !ok {
			return 0, fmt.Errorf("Label \"%s\" not found", label)
		}

		// TODO: Manage when multiple MBC
		if labelValue > 0x8000 {
			panic("Switchable ROM banks are not implemented yet")
		}

		return uint16(labelValue) + offset, nil
	}

	res, err := strconv.ParseUint(param, 0, 16)

	return uint16(res), err
}

func Raw16MacroRelativeLabel(labels *Labels, defs *Definitions, param string) (uint16, error) {
	if !strings.HasPrefix(param, "=$") {
		return 0, fmt.Errorf(
			"label \"%s\" is external to the macro",
			param,
		)
	}
	return Raw16(labels, defs, param)
}

func Reg16Indirect(_ *Labels, _ *Definitions, param string) (uint16, error) {
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

func Raw8Indirect(labels *Labels, defs *Definitions, param string) (uint16, error) {
	if strings.HasPrefix(param, "$") {
		param = strings.ToUpper(strings.TrimPrefix(param, "$"))

		definition, ok := (*defs)[param]
		if !ok {
			return 0, fmt.Errorf("$%s is undefined", param)
		}

		res, ok := definition.(Indirect8b)
		if !ok {
			return 0, fmt.Errorf("$%s is of type %T but Indirect8bb is expected", param, res)
		}
		return uint16(res), nil
	}

	if len(param) < 2 || param[0] != '(' || param[len(param)-1] != ')' {
		return 0, fmt.Errorf("Invalid raw8indirect")
	}

	res, err := Raw8(labels, defs, param[1:len(param)-1])
	if err == nil {
		return res, nil
	}
	return 0, err
}

func Raw16Indirect(labels *Labels, defs *Definitions, param string) (uint16, error) {
	if strings.HasPrefix(param, "$") {
		param = strings.ToUpper(strings.TrimPrefix(param, "$"))

		if labels == nil {
			return 0, nil
		}

		definition, ok := (*defs)[param]
		if !ok {
			return 0, fmt.Errorf("$%s is undefined", param)
		}

		res, ok := definition.(Indirect16b)
		if !ok {
			return 0, fmt.Errorf("$%s is of type %T but Indirect16b expected", param, res)
		}
		return uint16(res), nil
	}
	if len(param) < 2 || param[0] != '(' || param[len(param)-1] != ')' {
		return 0, fmt.Errorf("Invalid raw16indirect")
	}

	return Raw16(labels, defs, param[1:len(param)-1])
}

func Condition(_ *Labels, _ *Definitions, param string) (uint16, error) {
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

func BitOrdinal(_ *Labels, _ *Definitions, param string) (uint16, error) {
	res, err := strconv.ParseUint(param, 0, 3)
	return uint16(res), err
}
