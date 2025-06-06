package main

import (
	"fmt"
	"strconv"
	"strings"
)

func parseOffset(param string) (string, uint16, error) {
	if strings.Contains(param, "+") {
		labelParts := strings.Split(param, "+")
		if len(labelParts) != 2 {
			return "", 0, fmt.Errorf(
				"Labels with offset should have exactly 1 offset (in \"%s\")",
				param,
			)
		}
		labelWithoutOffset := labelParts[0]
		o, err := strconv.ParseUint(labelParts[1], 0, 16)
		if err != nil {
			return "", 0, fmt.Errorf("Error while parsing label offset: %w", err)
		}
		offset := uint16(o)
		return labelWithoutOffset, offset, nil
	}
	return param, 0, nil
}

func Reg8(_ *Labels, lastAbsoluteLabel string, _ *Definitions, param string) (uint16, error) {
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

func A(_ *Labels, lastAbsoluteLabel string, _ *Definitions, param string) (uint16, error) {
	if param == "A" {
		return 0, nil
	}
	return 0, fmt.Errorf("Invalid A")
}

func HL(_ *Labels, lastAbsoluteLabel string, _ *Definitions, param string) (uint16, error) {
	if param == "HL" {
		return 0, nil
	}
	return 0, fmt.Errorf("Invalid HL")
}

func SP(_ *Labels, lastAbsoluteLabel string, _ *Definitions, param string) (uint16, error) {
	if param == "SP" {
		return 0, nil
	}
	return 0, fmt.Errorf("Invalid SP")
}

func IndirectC(_ *Labels, lastAbsoluteLabel string, _ *Definitions, param string) (uint16, error) {
	if param == "(C)" {
		return 0, nil
	}
	return 0, fmt.Errorf("Invalid (C)")
}

func Reg16(_ *Labels, lastAbsoluteLabel string, _ *Definitions, param string) (uint16, error) {
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

func Raw8(
	labels *Labels,
	lastAbsoluteLabel string,
	defs *Definitions,
	param string,
) (uint16, error) {
	if strings.HasPrefix(param, "high(") && strings.HasSuffix(param, ")") {
		v, err := Raw16(labels, lastAbsoluteLabel, defs, param[5:len(param)-1])
		if err != nil {
			return 0, err
		}
		return uint16(v >> 8), nil
	}
	if strings.HasPrefix(param, "low(") && strings.HasSuffix(param, ")") {
		v, err := Raw16(labels, lastAbsoluteLabel, defs, param[4:len(param)-1])
		if err != nil {
			return 0, err
		}
		return uint16(v & 0xff), nil
	}
	if strings.HasPrefix(param, "inv(") && strings.HasSuffix(param, ")") {
		v, err := Raw8(labels, lastAbsoluteLabel, defs, param[4:len(param)-1])
		if err != nil {
			return 0, err
		}
		return uint16((256 / v) & 0xff), nil
	}
	if strings.HasPrefix(param, "$") {
		param = strings.ToUpper(strings.TrimPrefix(param, "$"))
		if res, err := strconv.ParseUint(param, 16, 16); err == nil {
			if len(param) > 2 {
				return 0, fmt.Errorf("%s is > 8bit", param)
			}
			return uint16(res), nil
		}

		varWithoutOffset, offset, err := parseOffset(param)
		if err != nil {
			return 0, err
		}

		definition, ok := (*defs)[varWithoutOffset]
		if !ok {
			return 0, fmt.Errorf("$%s is undefined", varWithoutOffset)
		}

		res, ok := definition.(Raw8b)
		if !ok {
			return 0, fmt.Errorf(
				"$%s is of type %T but Raw8b is expected",
				varWithoutOffset,
				definition,
			)
		}

		if uint16(res)+offset > 0xff {
			return 0, fmt.Errorf(
				"overflow: $%s (0x%02x) + 0x%02x exceeds 0xff",
				varWithoutOffset,
				uint16(res),
				offset,
			)
		}

		return uint16(res) + offset, nil
	}
	if strings.HasPrefix(param, "0x") && len(param) > 4 {
		return 0, fmt.Errorf("%s is > 8bit", param)
	}
	res, err := strconv.ParseUint(param, 0, 8)
	return uint16(res), err
}

func Raw16(
	labels *Labels,
	lastAbsoluteLabel string,
	defs *Definitions,
	param string,
) (uint16, error) {
	if strings.Contains(param, "-") {
		spl := strings.Split(param, "-")

		v, err := Raw16(labels, lastAbsoluteLabel, defs, spl[0])
		if err != nil {
			return 0, err
		}
		result := v

		for _, arg := range spl[1:] {
			v, err := Raw16(labels, lastAbsoluteLabel, defs, arg)
			if err != nil {
				return 0, err
			}
			result -= v
		}

		return result, nil
	}

	if strings.HasPrefix(param, "$") {
		param = strings.ToUpper(strings.TrimPrefix(param, "$"))
		if res, err := strconv.ParseUint(param, 16, 16); err == nil {
			return uint16(res), nil
		}

		varWithoutOffset, offset, err := parseOffset(param)
		if err != nil {
			return 0, err
		}

		definition, ok := (*defs)[varWithoutOffset]
		if !ok {
			return 0, fmt.Errorf("$%s is undefined", varWithoutOffset)
		}

		res, ok := definition.(Raw16b)
		if !ok {
			return 0, fmt.Errorf(
				"$%s is of type %T but Raw16b is expected",
				varWithoutOffset,
				definition,
			)
		}

		if uint32(res)+uint32(offset) > 0xffff {
			return 0, fmt.Errorf(
				"overflow: $%s (0x%04x) + 0x%04x exceeds 0xffff",
				varWithoutOffset,
				uint16(res),
				offset,
			)
		}

		return uint16(res) + offset, nil
	}

	if strings.HasPrefix(param, "=") {
		labelWithoutOffset, offset, err := parseOffset(param[1:])
		if err != nil {
			return 0, err
		}

		if labels == nil {
			return 0, nil
		}

		if strings.HasPrefix(labelWithoutOffset, ".") {
			if lastAbsoluteLabel == "" {
				return 0, fmt.Errorf(
					"Relative label \"%s\" referenced outside of parent",
					labelWithoutOffset,
				)
			}
			labelWithoutOffset = lastAbsoluteLabel + labelWithoutOffset
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

func Raw16MacroRelativeLabel(
	labels *Labels,
	lastAbsoluteLabel string,
	defs *Definitions,
	param string,
) (uint16, error) {
	if !strings.HasPrefix(param, "=$") {
		return 0, fmt.Errorf(
			"label \"%s\" is external to the macro",
			param,
		)
	}
	return Raw16(labels, lastAbsoluteLabel, defs, param)
}

func Reg16Indirect(
	_ *Labels,
	lastAbsoluteLabel string,
	_ *Definitions,
	param string,
) (uint16, error) {
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

func Raw8Indirect(
	labels *Labels,
	lastAbsoluteLabel string,
	defs *Definitions,
	param string,
) (uint16, error) {
	if strings.HasPrefix(param, "$") {
		param = strings.ToUpper(strings.TrimPrefix(param, "$"))

		definition, ok := (*defs)[param]
		if !ok {
			return 0, fmt.Errorf("$%s is undefined", param)
		}

		res, ok := definition.(Indirect8b)
		if !ok {
			return 0, fmt.Errorf("$%s is of type %T but Indirect8bb is expected", param, definition)
		}
		return uint16(res), nil
	}

	if len(param) < 2 || param[0] != '(' || param[len(param)-1] != ')' {
		return 0, fmt.Errorf("Invalid raw8indirect")
	}

	res, err := Raw8(labels, lastAbsoluteLabel, defs, param[1:len(param)-1])
	if err == nil {
		return res, nil
	}
	return 0, err
}

func Raw16Indirect(
	labels *Labels,
	lastAbsoluteLabel string,
	defs *Definitions,
	param string,
) (uint16, error) {
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
			return 0, fmt.Errorf("$%s is of type %T but Indirect16b expected", param, definition)
		}
		return uint16(res), nil
	}
	if len(param) < 2 || param[0] != '(' || param[len(param)-1] != ')' {
		return 0, fmt.Errorf("Invalid raw16indirect")
	}

	return Raw16(labels, lastAbsoluteLabel, defs, param[1:len(param)-1])
}

func Condition(_ *Labels, lastAbsoluteLabel string, _ *Definitions, param string) (uint16, error) {
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

func BitOrdinal(_ *Labels, lastAbsoluteLabel string, _ *Definitions, param string) (uint16, error) {
	res, err := strconv.ParseUint(param, 0, 3)
	return uint16(res), err
}
