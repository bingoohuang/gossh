package gossh

import (
	"fmt"

	"github.com/bingoohuang/gossh/cmdtype"
	"github.com/bingoohuang/gossh/pbe"
	"github.com/spf13/viper"
)

// Config represents the structure of input toml file structure.
type Config struct {
	PrintConfig bool     `pflag:"print config before running"`
	Passphrase  string   `pflag:"passphrase for decrypt {PBE}Password,shorthand=p"`
	Hosts       []string `pflag:"shorthand=h"`
	Cmds        []string
}

// Host represents the structure of remote host information for ssh.
type Host struct {
	ID         string
	Addr       string
	User       string
	Password   string // empty when using public key
	Properties map[string]string
}

// CmdExcResult means the detail exec result of cmd
type CmdExcResult struct {
}

// HostsCmd means the executable interface
type HostsCmd interface {
	// Parse parses the command.
	Parse()
	// ExecInHosts execute in specified hosts.
	ExecInHosts(gs *GoSSH) error

	// TargetHosts returns target hosts for the command
	TargetHosts() []*Host

	// RawCmd returns the original raw command
	RawCmd() string
}

// CmdGroup represents the a group of structure of command line with same cmd type in config's cmds.
type CmdGroup struct {
	gs      *GoSSH
	Type    cmdtype.CmdType
	Cmds    []HostsCmd
	Results []CmdExcResult
}

// Parse parses the CmdGroup's data.
func (g *CmdGroup) Parse() {
	for _, cmd := range g.Cmds {
		cmd.Parse()
	}
}

// Exec executes the CmdGroup.
func (g *CmdGroup) Exec() {
	cmdsCount := len(g.Cmds)
	if cmdsCount == 0 {
		fmt.Println("There is not commands to execute")
		return
	}

	g.Results = make([]CmdExcResult, cmdsCount)
	switch g.Type {
	case cmdtype.Local:
		g.execLocal()
	default:
		for _, cmd := range g.Cmds {
			if len(cmd.TargetHosts()) == 0 {
				fmt.Printf("There is no target hosts for cmd %s to executed\n", cmd.RawCmd())
				continue
			}

			if err := cmd.ExecInHosts(g.gs); err != nil {
				fmt.Printf("ExecInHosts error %v", err)
			}
		}
	}
}

// GoSSH defines the structure of the whole cfg context.
type GoSSH struct {
	Vars      Config
	Hosts     []*Host
	CmdGroups []CmdGroup

	sftpClientMap *sftpClientMap
}

// Close closes gossh.
func (g *GoSSH) Close() {
	g.sftpClientMap.Close()
}

// Parse parses the flags or cnf files to GoSSH.
func (c Config) Parse() GoSSH {
	gs := GoSSH{}

	_ = c.parseVars()
	gs.Hosts = c.parseHosts()
	gs.CmdGroups = c.parseCmdGroups(&gs)
	gs.sftpClientMap = makeSftpClientMap()

	return gs
}

func (c Config) parseCmdGroups(gs *GoSSH) []CmdGroup {
	lastCmdType := cmdtype.Noop

	var group *CmdGroup

	groups := make([]*CmdGroup, 0)

	for _, cmd := range c.Cmds {
		cmdType, hostPart, realCmd := cmdtype.Parse(cmd)
		if cmdType == cmdtype.Noop {
			continue
		}

		if lastCmdType != cmdType {
			lastCmdType = cmdType
			group = &CmdGroup{gs: gs, Type: cmdType, Cmds: make([]HostsCmd, 0)}
			groups = append(groups, group)
		}

		switch cmdType {
		case cmdtype.Local:
			group.Cmds = append(group.Cmds, &LocalCmd{cmd: cmd})
		case cmdtype.Ul:
			group.Cmds = append(group.Cmds, buildUlCmd(gs, hostPart, realCmd, cmd))
		case cmdtype.Dl:
			group.Cmds = append(group.Cmds, buildDlCmd(gs, hostPart, realCmd, cmd))
		case cmdtype.SSH:
			group.Cmds = append(group.Cmds, buildSSHCmd(gs, hostPart, realCmd, cmd))
		}
	}

	returnGroups := make([]CmdGroup, len(groups))

	for i, group := range groups {
		group.Parse()
		returnGroups[i] = *group
	}

	return returnGroups
}

func (c Config) parseVars() Config {
	if c.Passphrase != "" {
		viper.Set(pbe.PbePwd, c.Passphrase)
	}

	return c
}
