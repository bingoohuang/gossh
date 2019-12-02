package elf

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestBaseDir(t *testing.T) {
	assert.Equal(t, "/a", BaseDir([]string{"/a/b", "/a/c"}))
	assert.Equal(t, "/", BaseDir([]string{"/a/b", "/b/c"}))
}
