package gossh

import (
	"bytes"
	"encoding/json"
)

// JSONPretty prettify the JSON encoding of data silently
func JSONPretty(data interface{}) string {
	p, _ := JSONPrettyE(data)
	return p
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
