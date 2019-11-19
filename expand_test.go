package gossh

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestMakeExpand(t *testing.T) {
	assert.Equal(t, MakeExpand("192.168.136.9:8022").MakeExpand(), []string{"192.168.136.9:8022"})
	assert.Equal(t, MakeExpand("192.168.136.(9 18):8022").MakeExpand(),
		[]string{"192.168.136.9:8022", "192.168.136.18:8022"})
	assert.Equal(t, MakeExpand("192.168.136.(9-10 18):8022").MakeExpand(),
		[]string{"192.168.136.9:8022", "192.168.136.10:8022", "192.168.136.18:8022"})
	assert.Equal(t, MakeExpand("192.168.136.(10-9 18)").MakeExpand(),
		[]string{"192.168.136.10", "192.168.136.9", "192.168.136.18"})
	assert.Equal(t, MakeExpand("(10-9 18)").MakeExpand(), []string{"10", "9", "18"})
	assert.Equal(t, MakeExpand("(10-9 18").MakeExpand(), []string{"(10-9 18"})
	assert.Equal(t, MakeExpand("(a-b 18)").MakeExpand(), []string{"a-b", "18"})
	assert.Equal(t, MakeExpand("(9-b 18)").MakeExpand(), []string{"9-b", "18"})
}
