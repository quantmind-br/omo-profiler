package config

// StripJSONC rewrites JSONC into plain JSON that encoding/json accepts:
// `//` line comments and `/* */` block comments become spaces, and commas
// that trail the last element of an object or array are blanked out.
//
// Comment bytes are replaced in place rather than removed so byte offsets of
// the surviving content — and therefore json error positions — stay accurate.
// Newlines inside block comments are preserved to keep line numbers correct.
func StripJSONC(src []byte) []byte {
	out := make([]byte, len(src))
	copy(out, src)

	inString := false
	escaped := false
	for i := 0; i < len(out); i++ {
		c := out[i]

		if inString {
			switch {
			case escaped:
				escaped = false
			case c == '\\':
				escaped = true
			case c == '"':
				inString = false
			}
			continue
		}

		switch {
		case c == '"':
			inString = true
		case c == '/' && i+1 < len(out) && out[i+1] == '/':
			for ; i < len(out) && out[i] != '\n'; i++ {
				out[i] = ' '
			}
		case c == '/' && i+1 < len(out) && out[i+1] == '*':
			out[i] = ' '
			out[i+1] = ' '
			i += 2
			for ; i < len(out); i++ {
				if out[i] == '*' && i+1 < len(out) && out[i+1] == '/' {
					out[i] = ' '
					out[i+1] = ' '
					i++
					break
				}
				if out[i] != '\n' {
					out[i] = ' '
				}
			}
		}
	}

	return blankTrailingCommas(out)
}

// blankTrailingCommas replaces a comma with a space when the next
// non-whitespace byte closes the enclosing object or array. Operates on
// comment-free input.
func blankTrailingCommas(buf []byte) []byte {
	inString := false
	escaped := false
	for i := range buf {
		c := buf[i]

		if inString {
			switch {
			case escaped:
				escaped = false
			case c == '\\':
				escaped = true
			case c == '"':
				inString = false
			}
			continue
		}

		if c == '"' {
			inString = true
			continue
		}
		if c != ',' {
			continue
		}

		for j := i + 1; j < len(buf); j++ {
			switch buf[j] {
			case ' ', '\t', '\r', '\n':
				continue
			case '}', ']':
				buf[i] = ' '
			}
			break
		}
	}
	return buf
}
