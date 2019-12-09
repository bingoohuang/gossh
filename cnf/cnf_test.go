package cnf

import (
	"net/http"
	"net/url"
	"regexp"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// CaptureConfig defines the config to capture a sub string from a string
type CaptureConfig struct {
	Matches      []string `pflag:"前置匹配(子串包含)"`
	Capture      string   `pflag:"匹配正则(优先级比锚点高)"`
	CaptureGroup int      `pflag:"捕获组序号"`

	AnchorStart string `pflag:"起始锚点(在capture为空时有效)"`
	AnchorEnd   string `pflag:"终止锚点(在capture为空时有效)"`

	_ *regexp.Regexp
}

// Post process line and then POST it out
type Post struct {
	PostURL string `pflag:"POST URL"`
	CaptureConfig

	_ *http.Client
	_ *url.URL
	_ url.Values
}

func TestDeclarePflagsByStruct(t *testing.T) {
	DeclarePflagsByStruct(Post{})

	plfagMap := make(map[string]*pflag.Flag)

	pflag.VisitAll(func(f *pflag.Flag) { plfagMap[f.Name] = f })

	assert.Contains(t, plfagMap, "postURL")
	assert.Contains(t, plfagMap, "matches")

	viper.Set("Matches", "a,b,c")
	viper.Set("PostURL", "https://a.b.c")

	p := &Post{}
	ViperToStruct(p)

	assert.Equal(t, []string{"a", "b", "c"}, p.Matches)
	assert.Equal(t, "https://a.b.c", p.PostURL)
}
