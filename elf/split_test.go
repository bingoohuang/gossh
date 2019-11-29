package elf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFieldsX(t *testing.T) {
	assert.Nil(t, FieldsX("(a b) c", "(", ")", 0))
	assert.Equal(t, []string{"(a b)  c"}, FieldsX("(a b)  c ", "(", ")", 1))
	assert.Equal(t, []string{"(a b)", "c"}, FieldsX("(a b)  c", "(", ")", 2))
	assert.Equal(t, []string{"(a b)", "c  d e"}, FieldsX("(a b)  c  d e ", "(", ")", 2))
	assert.Equal(t, []string{"(a b)", "c"}, FieldsX("(a b) c", "(", ")", -1))
	assert.Equal(t, []string{"(a b)", "(c d)"}, FieldsX(" (a b) (c d) ", "(", ")", -1))
	assert.Equal(t, []string{"(中 华) (人 民)"}, FieldsX("(中 华) (人 民)  ", "(", ")", 1))
	assert.Equal(t, []string{"(中 华)", "(人 民)"}, FieldsX(" (中 华) (人 民)  ", "(", ")", -1))
	assert.Equal(t, []string{"(中 华)", "(人 民)  共和国"}, FieldsX(" (中 华) (人 民)  共和国", "(", ")", 2))
}
