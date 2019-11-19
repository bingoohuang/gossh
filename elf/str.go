package elf

import (
	"unicode"
	"unicode/utf8"
)

// Fields splits the string s around each instance of one or more consecutive white space
// characters, as defined by unicode.IsSpace, returning a slice of substrings of s or an
// empty slice if s contains only white space.
// The count determines the number of substrings to return:
//   n > 0: at most n substrings; the last substring will be the unsplit remainder.
//   n == 0: the result is nil (zero substrings)
//   n < 0: all substrings
// Code are copy from strings.Fields and add count parameter to control the max fields.
func Fields(s string, count int) []string {
	if count == 0 {
		return nil
	}

	// First count the fields.
	// This is an exact count if s is ASCII, otherwise it is an approximation.
	n, setBits := countFields(s, count)

	if setBits >= utf8.RuneSelf {
		// Some runes in the input string are not ASCII.
		return FieldsFunc(s, count, unicode.IsSpace)
	}

	// ASCII fast path
	a := make([]string, n)
	na := 0
	fieldStart := 0
	i := 0

	// Skip spaces in the front of the input.
	for i < len(s) && asciiSpace[s[i]] != 0 {
		i++
	}

	fieldStart = i

	for i < len(s) && (count < 0 || na < count) {
		if asciiSpace[s[i]] == 0 {
			i++
			continue
		}

		if na == count-1 {
			a[na] = s[fieldStart:]
		} else {
			a[na] = s[fieldStart:i]
		}

		na++
		i++

		// Skip spaces in between fields.
		for i < len(s) && asciiSpace[s[i]] != 0 {
			i++
		}

		fieldStart = i
	}

	if fieldStart < len(s) && (count < 0 || na < count) { // Last field might end at EOF.
		a[na] = s[fieldStart:]
	}

	return fixLastField(a)
}

func fixLastField(a []string) []string {
	lastIndex := len(a) - 1
	last := a[lastIndex]
	stopPos := 0

	for i := 0; i < len(last); i++ {
		isSep := asciiSpace[last[i]] == 1
		if isSep {
			if stopPos == 0 {
				stopPos = i
			}
		} else {
			stopPos = 0
		}
	}

	if stopPos > 0 {
		a[lastIndex] = last[0:stopPos]
	}

	return a
}

func countFields(s string, count int) (int, uint8) {
	// setBits is used to track which bits are set in the bytes of s.
	setBits := uint8(0)
	n := 0
	wasSpace := 1

	for i := 0; i < len(s); i++ {
		r := s[i]
		setBits |= r
		isSpace := int(asciiSpace[r])
		n += wasSpace & ^isSpace
		wasSpace = isSpace
	}

	if count < 0 || n < count {
		return n, setBits
	}

	return count, setBits
}

var asciiSpace = [256]uint8{'\t': 1, '\n': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1} // nolint gochecknoglobals

// FieldsFunc splits the string s at each run of Unicode code points c satisfying f(c)
// and returns an array of slices of s. If all code points in s satisfy f(c) or the
// string is empty, an empty slice is returned.
// FieldsFunc makes no guarantees about the order in which it calls f(c).
// If f does not return consistent results for a given c, FieldsFunc may crash.
func FieldsFunc(s string, count int, f func(rune) bool) []string {
	// A span is used to record a slice of s of the form s[start:end].
	// The start index is inclusive and the end index is exclusive.
	type span struct {
		start int
		end   int
	}

	spans := make([]span, 0, 32)

	// Find the field start and end indices.
	wasField := false
	fromIndex := 0
	ending := false

	for i, rune := range s {
		isSep := f(rune)

		if isSep {
			if wasField {
				spans = append(spans, span{start: fromIndex, end: i})
				wasField = false

				if count > 0 && len(spans) == count-1 {
					ending = true
				}
			}

			continue
		}

		if ending {
			wasField = true
			fromIndex = i

			break
		}

		if !wasField {
			wasField = true
			fromIndex = i

			if count == 1 {
				break
			}
		}
	}

	// Last field might end at EOF.
	if wasField {
		spans = append(spans, span{fromIndex, len(s)})
	}

	// Create strings from recorded field indices.
	a := make([]string, len(spans))
	for i, span := range spans {
		a[i] = s[span.start:span.end]
	}

	return fixLastFieldFunc(a, f)
}

func fixLastFieldFunc(a []string, f func(rune) bool) []string {
	lastIndex := len(a) - 1
	last := a[lastIndex]
	stopPos := 0

	for i, rune := range last {
		isSep := f(rune)
		if isSep {
			if stopPos == 0 {
				stopPos = i
			}
		} else {
			stopPos = 0
		}
	}

	if stopPos > 0 {
		a[lastIndex] = last[0:stopPos]
	}

	return a
}
