package gossh

import (
	"strconv"
	"strings"

	"github.com/bingoohuang/gossh/elf"
)

// Expandable abstract a thing that can be expanded to parts.
type Expandable interface {
	// MakePart returns i'th item.
	Part(i int) string
	// Len returns the length of part items.
	Len() int
}

// Part as a part of something.
type Part struct {
	p []string
}

// MakePart make a direct part by s.
func MakePart(s []string) Part { return Part{p: s} }

// MakeFixedPart make a fixed part by s.
func MakeFixedPart(s string) Part { return Part{p: []string{s}} }

// Len returns the length of part items.
func (f Part) Len() int { return len(f.p) }

// Part returns i'th item.
func (f Part) Part(i int) string {
	l := len(f.p)

	if i >= l {
		return f.p[l-1]
	}

	return f.p[i]
}

// MakeExpandPart makes an expanded part.
func MakeExpandPart(s string) Part {
	expanded := make([]string, 0)
	fs := elf.Fields(s, -1)

	for _, f := range fs {
		items := expandRange(f)
		expanded = append(expanded, items...)
	}

	return Part{p: expanded}
}

func expandRange(f string) []string {
	hyphenPos := strings.Index(f, "-")
	if hyphenPos <= 0 || hyphenPos == len(f)-1 {
		return []string{f}
	}

	from := strings.TrimSpace(f[0:hyphenPos])
	to := strings.TrimSpace(f[hyphenPos+1:])

	fromI := 0
	toI := 0

	var err error

	if fromI, err = strconv.Atoi(from); err != nil {
		return []string{f}
	}

	if toI, err = strconv.Atoi(to); err != nil {
		return []string{f}
	}

	parts := make([]string, 0)

	if fromI < toI {
		for i := fromI; i <= toI; i++ {
			parts = append(parts, strconv.Itoa(i))
		}
	} else {
		for i := fromI; i >= toI; i-- {
			parts = append(parts, strconv.Itoa(i))
		}
	}

	return parts
}

// Expand structured a expandable unit.
type Expand struct {
	raw   string
	parts []Expandable
}

// MaxLen returns the max length among the inner parts.
func (f Expand) MaxLen() int {
	maxLen := 0

	for _, p := range f.parts {
		l := p.Len()
		if l > maxLen {
			maxLen = l
		}
	}

	return maxLen
}

// MakePart makes a part of expand.
func (f Expand) MakePart() Part {
	return MakePart(f.MakeExpand())
}

// MakeExpand makes a expanded string slice of expand.
func (f Expand) MakeExpand() []string {
	ml := f.MaxLen()
	parts := make([]string, ml)

	for i := 0; i < ml; i++ {
		part := ""

		for _, p := range f.parts {
			part += p.Part(i)
		}

		parts[i] = part
	}

	return parts
}

// MakeExpand  makes an expand by s.
func MakeExpand(s string) Expand {
	parts := make([]Expandable, 0)

	for {
		l := strings.Index(s, "(")
		if l < 0 {
			parts = append(parts, MakeFixedPart(s))
			break
		}

		r := strings.Index(s[l:], ")")
		if r < 0 {
			parts = append(parts, MakeFixedPart(s))
			break
		}

		if lp := s[0:l]; lp != "" {
			parts = append(parts, MakeFixedPart(lp))
		}

		parts = append(parts, MakeExpandPart(s[l+1:l+r]))

		if l+r+1 == len(s) {
			break
		}

		s = s[l+r+1:]
	}

	return Expand{raw: s, parts: parts}
}
