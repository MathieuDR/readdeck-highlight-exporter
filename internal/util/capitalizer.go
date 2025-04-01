package util

import (
	"unicode"
	"unicode/utf8"
)

func Capitalize(input string) string {
	r, size := utf8.DecodeRuneInString(input)
	if r == utf8.RuneError {
		return input
	}

	return string(unicode.ToUpper(r)) + input[size:]
}
