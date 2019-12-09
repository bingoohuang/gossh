package gossh

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONCompact(t *testing.T) {
	assert.Equal(t, `{"Name":"bingoo","Value":1000}`, JSONCompact(struct {
		Name  string
		Value int
	}{
		Name:  "bingoo",
		Value: 1000,
	}))

	assert.Equal(t, `{"spacedValue":"spaced value"}`, JSONCompact(`{ "spacedValue": "spaced value" }`))
}
