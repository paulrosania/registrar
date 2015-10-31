package storage

import (
	"unicode"
)

func toSnake(s string) string {
	in := []rune(s)
	out := make([]rune, 0, len(in))
	for i, c := range in {
		if i > 0 && unicode.IsUpper(c) && ((i+1 < len(in) && unicode.IsLower(in[i+1])) || unicode.IsLower(in[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(c))
	}
	return string(out)
}
