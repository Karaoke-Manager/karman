package mediatype

import "strings"

// isTSpecial reports whether rune is in 'tspecials' as defined by RFC 1521 and RFC 2045.
func isTSpecial(r rune) bool {
	return strings.ContainsRune(`()<>@,;:\"/[]?=`, r)
}

// isToken reports whether s is a 'token' as defined by RFC 1521 and RFC 2045.
func isToken(s string) bool {
	if s == "" {
		return false
	}
	return strings.IndexFunc(s, isNotTokenChar) < 0
}

// isTokenChar reports whether rune is in 'token' as defined by RFC 1521 and RFC 2045.
func isTokenChar(r rune) bool {
	// token := 1*<any (US-ASCII) CHAR except SPACE, CTLs,
	//             or tspecials>
	return r > 0x20 && r < 0x7f && !isTSpecial(r)
}

// isNotTokenChar reports whether rune is not in 'token' as defined by RFC 1521 and RFC 2045.
func isNotTokenChar(r rune) bool {
	return !isTokenChar(r)
}
