package elf

import (
	"strings"
	"unicode/utf8"
)

// OpenClose stands for open-close strings like ()[]{} and etc.
type OpenClose struct {
	Open  string
	Close string
}

// SplitX splits s by separate (not in (),[],{})
func SplitX(s string, separate string, ocs ...OpenClose) []string {
	ocs = setDefaultOpenCloses(ocs)

	subs := make([]string, 0)

	quotStart := -1
	pos := 0
	l := len(s)

	lastClose := ""

	for i := 0; i < l; {
		// w 当前字符宽度
		runeValue, w := utf8.DecodeRuneInString(s[i:])
		ch := string(runeValue)

		switch {
		case runeValue == '\\':
			if i+w < l {
				_, nextWidth := utf8.DecodeRuneInString(s[i+w:])
				i += nextWidth
			}
		case quotStart >= 0:
			if ch == lastClose {
				quotStart = -1
			}
		default:
			if yes, close := isOpen(ch, ocs); yes {
				quotStart = i
				lastClose = close
			} else if ch == separate {
				subs = tryAddPart(subs, s[pos:i])
				pos = i + w
			}
		}

		i += w
	}

	if pos < l {
		subs = tryAddPart(subs, s[pos:])
	}

	return subs
}

func setDefaultOpenCloses(ocs []OpenClose) []OpenClose {
	if len(ocs) == 0 {
		return []OpenClose{
			{"(", ")"},
			{"{", "}"},
			{"[", "]"},
			{"'", "'"},
		}
	}

	return ocs
}

func isOpen(open string, ocs []OpenClose) (yes bool, close string) {
	for _, oc := range ocs {
		if open == oc.Open {
			return true, oc.Close
		}
	}

	return false, ""
}

func tryAddPart(subs []string, sub string) []string {
	s := strings.TrimSpace(sub)
	if s != "" {
		return append(subs, s)
	}

	return subs
}
