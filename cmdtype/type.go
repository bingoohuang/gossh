package cmdtype

import "strings"

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

// Parse parses the type of cmd
func Parse(cmd string) (CmdType, []string) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return Noop, nil
	}

	fields := strings.Fields(cmd)

	if strings.HasPrefix(fields[0], "%host") {
		switch fields[1] {
		case "%ul":
			return Ul, fields
		case "%dl":
			return Dl, fields
		}

		return SSH, fields
	}

	return Local, fields
}
