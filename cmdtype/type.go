package cmdtype

import (
	"strings"

	"github.com/bingoohuang/gossh/elf"
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

// Parse parses the type of cmd,  returns CmdType, host part and real cmd part
func Parse(cmd string) (CmdType, string, string) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return Noop, "", ""
	}

	fields := elf.Fields(cmd, 2)

	if strings.HasPrefix(fields[0], "%host") {
		fields2 := elf.Fields(fields[1], 2)
		switch fields2[0] {
		case "%ul":
			return Ul, fields[0], fields2[1]
		case "%dl":
			return Dl, fields[0], fields2[1]
		}

		return SSH, fields[0], fields[1]
	}

	return Local, "", cmd
}
