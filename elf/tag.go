package elf

import (
	"strings"
)

// Tag represents tag of the struct field
type Tag struct {
	Raw  string
	Main string
	Opts map[string]string
}

// DecodeTag decode tag values
func DecodeTag(rawTag string) Tag {
	tags := strings.Split(rawTag, ",")
	opts := make(map[string]string)
	mainTag := ""

	for i, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}

		kvs := strings.SplitN(tag, "=", 2)
		k := kvs[0]

		if len(kvs) == 2 {
			opts[k] = kvs[1]
			continue
		}

		if i == 0 {
			mainTag = k
		} else {
			opts[k] = ""
		}
	}

	return Tag{Raw: rawTag, Main: mainTag, Opts: opts}
}
