package task9

import (
	"errors"
	"unicode"
)

// Unpack performs primitive string unpacking like "a4bc2d5e" -> "aaaabccddddde".
// Rules:
// - A digit after a letter repeats that letter N times total.
// - Digits cannot start the string; strings like "45" are invalid.
// - Empty input returns empty output and no error.
// - Escape sequences are NOT supported (per requirements).
func Unpack(input string) (string, error) {
	if input == "" {
		return "", nil
	}

	runes := []rune(input)
	// Invalid if first rune is a digit
	if unicode.IsDigit(runes[0]) {
		return "", errors.New("invalid string: starts with digit")
	}

	var result []rune
	var prev rune
	var hasPrev bool

	appendPrev := func(count int) {
		if !hasPrev {
			return
		}
		for i := 0; i < count; i++ {
			result = append(result, prev)
		}
	}

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if unicode.IsDigit(r) {
			// Repeat previous rune count-1 more times (we already plan to append once later)
			if !hasPrev {
				return "", errors.New("invalid string: digit without preceding char")
			}
			count := int(r - '0')
			if count == 0 {
				// Zero means remove previous pending append
				hasPrev = false
				continue
			}
			appendPrev(count)
			hasPrev = false
			continue
		}

		// Flush pending previous single append when encountering a new letter
		if hasPrev {
			appendPrev(1)
		}
		prev = r
		hasPrev = true
	}

	// Append last pending single if any
	if hasPrev {
		appendPrev(1)
	}

	return string(result), nil
}


