package cmdtype

import (
	"regexp"
	"strings"

	"github.com/bingoohuang/gou/str"
)

// CmdType represents the cmd types.
type CmdType int

const (
	// Noop means no operation cmd, its a placeholder for some purpose.
	Noop CmdType = iota
	// Local means the commands will be executed locally.
	Local
	// Ul uploads.
	Ul
	// Dl downloads.
	Dl
	// SSH means the ssh commands will executed by ssh in remote hosts.
	SSH
)

// Parse parses the type of cmd,  returns CmdType, host part and real cmd part.
func Parse(globalRemote bool, cmd string) (CmdType, string, string) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return Noop, "", ""
	}

	f := str.Fields(cmd, 2)
	if strings.HasPrefix(f[0], hostTag) {
		return parseRemote(f[1], f[0])
	}
	if strings.HasPrefix(f[0], localTag) {
		return Local, "", cmd
	}

	if globalRemote {
		return parseRemote(cmd, hostTag)
	}

	return Local, "", cmd
}

const (
	hostTag  = "%host"
	localTag = "%local"
)

func parseRemote(cmd, hostPart string) (CmdType, string, string) {
	fields2 := str.Fields(cmd, 2)
	switch fields2[0] {
	case "%ul":
		return Ul, hostPart, fields2[1]
	case "%dl":
		return Dl, hostPart, fields2[1]
	}

	return SSH, hostPart, cmd
}

// nolint
var resultVarPattern = regexp.MustCompile(`(.*?)\s*=>\s*(@\w+)\s*$`)

// ParseResultVar parses the result variable from the end of command like ... => @xyz.
func ParseResultVar(s string) (cmd, resultVar string) {
	subs := resultVarPattern.FindAllStringSubmatch(s, -1)
	if len(subs) == 0 {
		return s, ""
	}

	return subs[0][1], subs[0][2]
}
