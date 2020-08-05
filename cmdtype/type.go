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

	if globalRemote {
		return parseRemote(cmd)
	}

	if f := str.Fields(cmd, 2); strings.HasPrefix(f[0], hostTag) {
		return parseRemote(f[1])
	}

	return Local, "", cmd
}

const hostTag = "%host"

func parseRemote(cmd string) (CmdType, string, string) {
	fields2 := str.Fields(cmd, 2)
	switch fields2[0] {
	case "%ul":
		return Ul, hostTag, fields2[1]
	case "%dl":
		return Dl, hostTag, fields2[1]
	}

	return SSH, hostTag, cmd
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
