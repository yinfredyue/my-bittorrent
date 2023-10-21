package decode

import (
	"fmt"
	"strconv"
	"unicode"
)

func isDigit(c byte) bool {
	return unicode.IsDigit(rune(c))
}

func decodeFrom(str string, start int) (interface{}, int, error) {
	first_char := str[start]

	if isDigit(first_char) {
		// 5:hello -> hello
		var colonIndex int

		for i := start + 1; i < len(str); i++ {
			if str[i] == ':' {
				colonIndex = i
				break
			}
		}

		length, err := strconv.Atoi(str[start:colonIndex])
		if err != nil {
			return "", 0, err
		}

		return str[colonIndex+1 : colonIndex+1+length], colonIndex + length, nil
	} else if first_char == 'i' {
		// i52e -> 52
		var nonDigitIndex int

		for i := start + 1; i < len(str); i++ {
			if str[i] == 'e' {
				nonDigitIndex = i
				break
			}
		}

		integer, err := strconv.Atoi(str[start+1 : nonDigitIndex])
		return integer, nonDigitIndex, err
	} else if first_char == 'l' {
		// l5:helloi52ee -> ["hello", 52]
		// l5:helloi52elee -> ["hello", 52, []]
		// l5:helloi52el1:s2:ssi32eee -> ["hello", 52, ["s", "ss", 32]]
		curr := start + 1
		res := [](interface{}){}

		for str[curr] != 'e' {
			decoded, endIdx, err := decodeFrom(str, curr)
			if err != nil {
				return [](interface{}){}, 0, err
			}

			res = append(res, decoded)
			curr = endIdx + 1
		}

		return res, curr, nil
	}

	return nil, 0, fmt.Errorf("unexpected case?")
}

func Decode(str string) ([]interface{}, error) {
	res := [](interface{}){}

	curr := 0
	for curr < len(str) {
		decoded, endIdx, err := decodeFrom(str, curr)
		if err != nil {
			return [](interface{}){}, err
		}

		curr = endIdx + 1
		res = append(res, decoded)
	}

	return res, nil
}
