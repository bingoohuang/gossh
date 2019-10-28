package cmdtype

import "strings"

// CmdType represents the cmd types.
type CmdType int

const (
	// Noop means no operation cmd, its a placeholder for some purpose.
	Noop CmdType = iota
	// Local means the commands will be executed locally.
	Local
	// SCP means the commands will scp some files to remote hosts.
	SCP
	// SSH means the ssh commands will executed by ssh in remote hosts.
	SSH
)

// Parse parses the type of cmd
func Parse(cmd string) (CmdType, string) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return Noop, ""
	}

	if strings.HasPrefix(cmd, "scp") && strings.Contains(cmd, "%host") {
		return SCP, cmd
	}

	if strings.HasPrefix(cmd, "ssh") && strings.Contains(cmd, "%host") {
		return SSH, cmd
	}

	return Local, cmd
}
