package gossh

// Config represents the structure of input toml file structure.
type Config struct {
	Vars  []string
	Hosts []string
	Cmds  []string
}

// Vars alias the map[string]string.
type Vars map[string]string

// Host represents the structure of remote host information for ssh.
type Host struct {
	Name       string
	Addr       string
	User       string
	Password   string // empty when using public key
	Properties Vars
}

// CmdType represents the cmd types.
type CmdType int

const (
	// LocalCmd means the commands will be executed locally.
	// https://github.com/uber-go/guide/blob/master/style.md#start-enums-at-one
	LocalCmd CmdType = iota + 1
	// ScpCmd means the commands will scp some files to remote hosts.
	ScpCmd
	// SSHCmd means the ssh commands will executed by ssh in remote hosts.
	SSHCmd
)

// Cmd represents the interface of command to be executed.
type Cmd interface {
}

// CmdLine represents the structure of command line in config's cmd.
type CmdLine struct {
	Type CmdType
	Cmd  Cmd
}
