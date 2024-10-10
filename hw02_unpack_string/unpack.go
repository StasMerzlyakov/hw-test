package hw02unpackstring

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(input string) (string, error) {
	var result strings.Builder
	var saved rune

	const (
		StateBase = iota
		StateChar
		StateSlash
	)

	state := StateBase
	for pos, r := range input {
		switch state {
		case StateBase:
			switch r {
			case '\\':
				state = StateSlash
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				return "", fmt.Errorf("%w - unexpected rune %s pos %d", ErrInvalidString, string(r), pos)
			default:
				saved = r
				state = StateChar
			}
		case StateChar:
			switch r {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				runeStr := string(r)
				if n, err := strconv.Atoi(runeStr); err != nil {
					return "", fmt.Errorf("%w - convert rune (%s pos %d) to int error %s",
						ErrInvalidString, runeStr, pos, err.Error())
				} else {
					result.WriteString(strings.Repeat(string(saved), n))
					state = StateBase
				}
			case '\\':
				result.WriteRune(saved)
				state = StateSlash
			default:
				result.WriteRune(saved)
				saved = r
			}
		case StateSlash:
			saved = r
			state = StateChar
		default:
			return "", errors.New("implementation error")
		}
	}
	switch state {
	case StateChar:
		result.WriteRune(saved)
	case StateSlash:
		return "", fmt.Errorf("%w - unexpected '\\' at the end of string", ErrInvalidString)
	}

	return result.String(), nil
}
