package decode

import (
	"fmt"
	"strconv"
	"unicode"
)

func isDigit(c byte) bool {
	return unicode.IsDigit(rune(c))
}

func decodeOneFrom(str string, start int) (interface{}, int, error) {
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
			decoded, endIdx, err := decodeOneFrom(str, curr)
			if err != nil {
				return [](interface{}){}, 0, err
			}

			res = append(res, decoded)
			curr = endIdx + 1
		}

		return res, curr, nil
	} else if first_char == 'd' {
		res := map[string](interface{}){}

		curr := start + 1
		for str[curr] != 'e' {
			decoded, endIdx, err := decodeOneFrom(str, curr)
			decoded_key, ok := decoded.(string)
			if !ok || err != nil {
				return [](interface{}){}, 0, err
			}

			decoded_val, endIdx, err := decodeOneFrom(str, endIdx+1)
			if err != nil {
				return [](interface{}){}, 0, err
			}

			res[decoded_key] = decoded_val
			curr = endIdx + 1
		}

		return res, curr, nil
	}

	return nil, 0, fmt.Errorf("unexpected case?")
}

func Decode(str string) (interface{}, error) {
	res, endIdx, err := decodeOneFrom(str, 0)

	if endIdx != len(str)-1 {
		return [](interface{}){}, fmt.Errorf("didn't consume entire string?")
	}

	return res, err
}
