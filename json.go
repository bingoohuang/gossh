package gossh

import (
	"bytes"
	"encoding/json"

	"github.com/bingoohuang/gossh/elf"
)

// JSONPretty prettify the JSON encoding of data silently
func JSONPretty(data interface{}) string {
	return elf.PickFirst(JSONPrettyE(data))
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

// JSONCompact compact the JSON encoding of data silently
func JSONCompact(data interface{}) string {
	return elf.PickFirst(JSONCompactE(data))
}

// JSONCompactE compact the JSON encoding of data
func JSONCompactE(data interface{}) (string, error) {
	switch v := data.(type) {
	case string:
		buffer := new(bytes.Buffer)
		if err := json.Compact(buffer, []byte(v)); err != nil {
			return "", err
		}

		return buffer.String(), nil
	case []byte:
		buffer := new(bytes.Buffer)
		if err := json.Compact(buffer, v); err != nil {
			return "", err
		}

		return buffer.String(), nil
	default:
		b, err := json.Marshal(data)
		if err != nil {
			return "", err
		}

		return string(b), nil
	}
}
