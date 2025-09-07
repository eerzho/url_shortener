package base62

import "strings"

const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func Encode(num int) string {
	if num == 0 {
		return "0"
	}

	charsLen := len(chars)

	var result strings.Builder
	for num > 0 {
		result.WriteByte(chars[num%charsLen])
		num /= charsLen
	}

	encoded := result.String()
	runes := []rune(encoded)

	i := 0
	j := len(runes) - 1
	for i < j {
		runes[i], runes[j] = runes[j], runes[i]
		i++
		j--
	}

	return string(runes)
}

func Decode(encoded string) int {
	charsLen := len(chars)

	result := 0
	power := 1
	for i := len(encoded) - 1; i >= 0; i-- {
		char := encoded[i]

		pos := strings.IndexByte(chars, char)
		if pos == -1 {
			return 0
		}

		result += pos * power
		power *= charsLen
	}

	return result
}
