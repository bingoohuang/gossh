package cmdtype_test

import (
	"testing"

	. "github.com/bingoohuang/gossh/pkg/cmdtype"
	"github.com/bingoohuang/gou/lang"
	"github.com/stretchr/testify/assert"
)

func TestParseResultVar(t *testing.T) {
	assert.Equal(t, lang.M2("date", "@abc"), lang.M2(ParseResultVar("date => @abc ")))
	assert.Equal(t, lang.M2("date", ""), lang.M2(ParseResultVar("date")))
}
