package main

import (
	"fmt"
	"strconv"
	"strings"
)

func parseOffset(param string) (string, uint32, error) {
	if strings.Contains(param, "+") {
		labelParts := strings.Split(param, "+")
		if len(labelParts) != 2 {
			return "", 0, fmt.Errorf(
				"Labels with offset should have exactly 1 offset (in \"%s\")",
				param,
			)
		}
		labelWithoutOffset := labelParts[0]
		o, err := strconv.ParseUint(labelParts[1], 0, 32)
		if err != nil {
			return "", 0, fmt.Errorf("Error while parsing label offset: %w", err)
		}
		offset := uint32(o)
		return labelWithoutOffset, offset, nil
	}
	return param, 0, nil
}

func Reg8(
	_ *Labels,
	lastAbsoluteLabel string,
	_ *Definitions,
	_ uint32,
	param string,
) (uint32, error) {
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

func A(
	_ *Labels,
	lastAbsoluteLabel string,
	_ *Definitions,
	_ uint32,
	param string,
) (uint32, error) {
	if param == "A" {
		return 0, nil
	}
	return 0, fmt.Errorf("Invalid A")
}

func HL(
	_ *Labels,
	lastAbsoluteLabel string,
	_ *Definitions,
	_ uint32,
	param string,
) (uint32, error) {
	if param == "HL" {
		return 0, nil
	}
	return 0, fmt.Errorf("Invalid HL")
}

func SP(
	_ *Labels,
	lastAbsoluteLabel string,
	_ *Definitions,
	_ uint32,
	param string,
) (uint32, error) {
	if param == "SP" {
		return 0, nil
	}
	return 0, fmt.Errorf("Invalid SP")
}

func IndirectC(
	_ *Labels,
	lastAbsoluteLabel string,
	_ *Definitions,
	_ uint32,
	param string,
) (uint32, error) {
	if param == "(C)" {
		return 0, nil
	}
	return 0, fmt.Errorf("Invalid (C)")
}

func Reg16(
	_ *Labels,
	lastAbsoluteLabel string,
	_ *Definitions,
	_ uint32,
	param string,
) (uint32, error) {
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
	currentAddress uint32,
	param string,
) (uint32, error) {
	if strings.HasPrefix(param, "high(") && strings.HasSuffix(param, ")") {
		v, err := Raw16(labels, lastAbsoluteLabel, defs, currentAddress, param[5:len(param)-1])
		if err != nil {
			return 0, err
		}
		return uint32(v >> 8), nil
	}
	if strings.HasPrefix(param, "low(") && strings.HasSuffix(param, ")") {
		v, err := Raw16(labels, lastAbsoluteLabel, defs, currentAddress, param[4:len(param)-1])
		if err != nil {
			return 0, err
		}
		return uint32(v & 0xff), nil
	}
	if strings.HasPrefix(param, "inv(") && strings.HasSuffix(param, ")") {
		v, err := Raw8(labels, lastAbsoluteLabel, defs, currentAddress, param[4:len(param)-1])
		if err != nil {
			return 0, err
		}
		return uint32((256 / v) & 0xff), nil
	}
	if strings.HasPrefix(param, "bank(") && strings.HasSuffix(param, ")") {
		v, err := ROMAddress(labels, lastAbsoluteLabel, defs, currentAddress, param[5:len(param)-1])
		if err != nil {
			return 0, err
		}
		return uint32((v / 0x4000) & 0xff), nil
	}
	if strings.HasPrefix(param, "$") {
		param = strings.ToUpper(strings.TrimPrefix(param, "$"))
		if res, err := strconv.ParseUint(param, 16, 16); err == nil {
			if len(param) > 2 {
				return 0, fmt.Errorf("%s is > 8bit", param)
			}
			return uint32(res), nil
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

		if uint32(res)+offset > 0xff {
			return 0, fmt.Errorf(
				"overflow: $%s (0x%02x) + 0x%02x exceeds 0xff",
				varWithoutOffset,
				uint32(res),
				offset,
			)
		}

		return uint32(res) + offset, nil
	}
	if strings.HasPrefix(param, "0x") && len(param) > 4 {
		return 0, fmt.Errorf("%s is > 8bit", param)
	}
	res, err := strconv.ParseUint(param, 0, 8)
	return uint32(res), err
}

func Raw16(
	labels *Labels,
	lastAbsoluteLabel string,
	defs *Definitions,
	currentAddress uint32,
	param string,
) (uint32, error) {
	if strings.HasPrefix(param, "&") {
		return Raw16Indirect(labels, lastAbsoluteLabel, defs, currentAddress, param[1:])
	}
	if strings.HasPrefix(param, "ptr(") && strings.HasSuffix(param, ")") {
		v, err := ROMAddress(labels, lastAbsoluteLabel, defs, currentAddress, param[4:len(param)-1])
		if err != nil {
			return 0, err
		}

		bank := v / 0x4000
		if bank == 0 {
			return v, nil
		}

		return v - bank*0x4000 + 0x4000, nil
	}

	if strings.HasPrefix(param, "$") {
		param = strings.ToUpper(strings.TrimPrefix(param, "$"))
		if res, err := strconv.ParseUint(param, 16, 16); err == nil {
			return uint32(res & 0xffff), nil
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
				uint32(res),
				offset,
			)
		}

		return uint32(res) + offset, nil
	}

	if strings.Contains(param, "+") {
		spl := strings.Split(param, "+")

		v, err := ROMAddress(labels, lastAbsoluteLabel, defs, currentAddress, spl[0])
		if err != nil {
			return 0, err
		}

		result := v
		bank := v / 0x4000
		for _, arg := range spl[1:] {
			v, err := Raw16(labels, lastAbsoluteLabel, defs, currentAddress, arg)
			if err != nil {
				return 0, err
			}

			result += v
		}

		if (result / 0x4000) != bank {
			return 0, fmt.Errorf(
				"Cannot add an offset to a ROMAddress that would overflow the current bank (bank(%s) == %v != bank(%s) == %v",
				spl[0],
				bank,
				param,
				result/0x4000,
			)
		}

		return result & 0xffff, nil
	}

	if strings.Contains(param, "-") {
		spl := strings.Split(param, "-")

		v, err := ROMAddress(labels, lastAbsoluteLabel, defs, currentAddress, spl[0])
		if err != nil {
			return 0, err
		}

		bank := v / 0x4000
		result := v

		for i, arg := range spl[1:] {
			v, err := ROMAddress(labels, lastAbsoluteLabel, defs, currentAddress, arg)
			if err == nil {
				otherBank := v / 0x4000
				if bank != otherBank {
					return 0, fmt.Errorf(
						"Cannot get distance between rom addresses in different banks (%s is in bank %v, %s is in bank %v)",
						spl[0],
						bank,
						arg,
						otherBank,
					)
				}
				result -= v
				continue
			}

			v, err = Raw16(labels, lastAbsoluteLabel, defs, currentAddress, arg)
			if err != nil {
				return 0, err
			}
			if result/0x4000 != (result-v)/0x4000 && labels != nil {
				return 0, fmt.Errorf(
					"Cannot change the bank of a rom address by substracting an offset (bank(%s (=%v)) == %v != bank(%s (=%v)) == %v",
					strings.Join(spl[:i+1], "-"),
					result,
					result/0x4000,
					strings.Join(spl[:i+2], "-"),
					result-v,
					(result-v)/0x4000,
				)
			}
			result -= v
		}

		return result & 0xffff, nil
	}

	romAddr, err := ROMAddress(labels, lastAbsoluteLabel, defs, currentAddress, param)
	if err == nil {
		currentBank := currentAddress / 0x4000
		romAddrBank := romAddr / 0x4000

		// TODO: Forbidding calls from other banks to bank 0
		if romAddrBank == 0 {
			return romAddr, nil
		}

		if currentBank != romAddrBank {
			return 0, fmt.Errorf(
				"Cannot use an address from another bank (or bank 0). Please change the bank using bank(x) and get the raw bankless ptr using ptr(x). %s in bank %v, but current address is %v",
				param,
				romAddrBank,
				currentBank,
			)
		}

		return romAddr - romAddrBank*0x4000 + 0x4000, nil
	}

	res, err := strconv.ParseUint(param, 0, 16)

	return uint32(res), err
}

func Raw16MacroRelativeLabel(
	labels *Labels,
	lastAbsoluteLabel string,
	defs *Definitions,
	currentAddress uint32,
	param string,
) (uint32, error) {
	if !strings.HasPrefix(param, "=$") {
		return 0, fmt.Errorf(
			"label \"%s\" is external to the macro",
			param,
		)
	}
	return Raw16(labels, lastAbsoluteLabel, defs, currentAddress, param)
}

func Reg16Indirect(
	_ *Labels,
	lastAbsoluteLabel string,
	_ *Definitions,
	_ uint32,
	param string,
) (uint32, error) {
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
	currentAddress uint32,
	param string,
) (uint32, error) {
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
		return uint32(res), nil
	}

	if len(param) < 2 || param[0] != '(' || param[len(param)-1] != ')' {
		return 0, fmt.Errorf("Invalid raw8indirect")
	}

	res, err := Raw8(labels, lastAbsoluteLabel, defs, currentAddress, param[1:len(param)-1])
	if err == nil {
		return res, nil
	}
	return 0, err
}

func Raw16Indirect(
	labels *Labels,
	lastAbsoluteLabel string,
	defs *Definitions,
	currentAddress uint32,
	param string,
) (uint32, error) {
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
		return uint32(res), nil
	}
	if len(param) < 2 || param[0] != '(' || param[len(param)-1] != ')' {
		return 0, fmt.Errorf("Invalid raw16indirect")
	}

	return Raw16(labels, lastAbsoluteLabel, defs, currentAddress, param[1:len(param)-1])
}

func ROMAddress(
	labels *Labels,
	lastAbsoluteLabel string,
	defs *Definitions,
	currentAddress uint32,
	param string,
) (uint32, error) {
	if param == "." {
		return currentAddress, nil
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

		return uint32(labelValue) + offset, nil
	}

	if len(param) != 7 || param[2] != ':' {
		return 0, fmt.Errorf("Couldn't parse \"%s\" as a ROM addr", param)
	}

	bank, err := strconv.ParseUint(param[0:2], 16, 8)
	if err != nil {
		return 0, fmt.Errorf(
			"Couldn't parse bank number in \"%s\"", param,
		)
	}

	addr, err := strconv.ParseUint(param[3:], 16, 8)
	if err != nil {
		return 0, fmt.Errorf(
			"Couldn't parse address in \"%s\"", param,
		)
	}

	return uint32(bank*0x4000 + addr), nil
}

func Condition(
	_ *Labels,
	lastAbsoluteLabel string,
	_ *Definitions,
	_ uint32,
	param string,
) (uint32, error) {
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

func BitOrdinal(
	_ *Labels,
	lastAbsoluteLabel string,
	_ *Definitions,
	_ uint32,
	param string,
) (uint32, error) {
	res, err := strconv.ParseUint(param, 0, 3)
	return uint32(res), err
}
