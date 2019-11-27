package gossh

import (
	"bytes"
	"encoding/json"

	"github.com/bingoohuang/gossh/elf"
)

// JSONPretty prettify the JSON encoding of data silently
func JSONPretty(data interface{}) string {
	return elf.IgnoreError(JSONPrettyE(data))
}

// JSONPrettyE prettify the JSON encoding of data
func JSONPrettyE(data interface{}) (string, error) {
	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	encoder.SetIndent("", "\t")

	err := encoder.Encode(data)
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}
