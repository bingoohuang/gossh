package elf

import (
	"regexp"
	"strings"
)

// Tag represents tag of the struct field
type Tag struct {
	Raw  string
	Main string
	Opts map[string]string
}

// GetOpt gets opt's value by its name
func (t Tag) GetOpt(optName string) string {
	if opt, ok := t.Opts[optName]; ok && opt != "" {
		return opt
	}

	return ""
}

// DecodeTag decode tag values
func DecodeTag(rawTag string) Tag {
	opts := make(map[string]string)
	mainPart := ""

	re := regexp.MustCompile(`(\w+)\s*=\s*(\w+)`)
	submatchIndex := re.FindAllStringSubmatchIndex(rawTag, -1)

	if submatchIndex == nil {
		mainPart = rawTag
	} else {
		for i, g := range submatchIndex {
			if i == 0 {
				mainPart = strings.TrimSpace(rawTag[:g[0]])
			}

			k := rawTag[g[2]:g[3]]
			v := rawTag[g[4]:g[5]]
			opts[k] = v
		}
	}

	return Tag{Raw: rawTag, Main: mainPart, Opts: opts}
}
