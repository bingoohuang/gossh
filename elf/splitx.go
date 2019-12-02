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

// IsSame tells the open and close is same of not.
func (o OpenClose) IsSame() bool {
	return o.Open == o.Close
}

type rememberOpenClose struct {
	OpenClose
	startPos int
}

// SplitX splits s by separate (not in (),[],{})
func SplitX(s string, separate string, ocs ...OpenClose) []string {
	ocs = setDefaultOpenCloses(ocs)

	subs := make([]string, 0)

	remembers := make([]rememberOpenClose, 0)
	pos := 0
	l := len(s)

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
		case len(remembers) > 0:
			last := remembers[len(remembers)-1]
			if yes, oc := isOpen(ch, ocs, last.OpenClose); yes {
				remembers = append(remembers, rememberOpenClose{OpenClose: oc, startPos: i})
			} else if ch == last.Close {
				remembers = remembers[0 : len(remembers)-1]
				if len(remembers) == 0 {
					remembers = make([]rememberOpenClose, 0)
				}
			}
		default:
			if yes, oc := isOpen(ch, ocs, OpenClose{}); yes {
				remembers = append(remembers, rememberOpenClose{OpenClose: oc, startPos: i})
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

func isOpen(s string, ocs []OpenClose, last OpenClose) (yes bool, oc OpenClose) {
	for _, oc := range ocs {
		if s == oc.Open {
			if !oc.IsSame() || s != last.Close {
				return true, oc
			}
		}
	}

	return false, OpenClose{}
}

func tryAddPart(subs []string, sub string) []string {
	s := strings.TrimSpace(sub)
	if s != "" {
		return append(subs, s)
	}

	return subs
}
