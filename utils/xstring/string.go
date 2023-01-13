package xstring

import (
	"unicode"
	"unicode/utf8"
)

// FirstIsUpper 首字母是否大写
func FirstIsUpper(s string) bool {
	r, _ := utf8.DecodeRuneInString(s)
	return r != utf8.RuneError && unicode.IsUpper(r)
}

// FirstIsIsLower 首字母是否小写
func FirstIsIsLower(s string) bool {
	r, _ := utf8.DecodeRuneInString(s)
	return r != utf8.RuneError && unicode.IsLower(r)
}
