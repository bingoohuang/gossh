package gossh

import (
	"fmt"

	"github.com/bingoohuang/gossh/cmdtype"
	"github.com/bingoohuang/gossh/pbe"
	"github.com/spf13/viper"
)

// Config represents the structure of input toml file structure.
type Config struct {
	PrintConfig bool
	Passphrase  string
	Hosts       []string
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

// Cmd means the executable interface
type Cmd interface {
	// Parse parses the command.
	Parse()
	// ExecInHosts execute in specified hosts.
	ExecInHosts(gs *GoSSH) error
}

// CmdGroup represents the a group of structure of command line with same cmd type in config's cmds.
type CmdGroup struct {
	gs      *GoSSH
	Type    cmdtype.CmdType
	Cmds    []Cmd
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
	g.Results = make([]CmdExcResult, len(g.Cmds))
	switch g.Type {
	case cmdtype.Local:
		g.execLocal()
	case cmdtype.SSH:
		for _, cmd := range g.Cmds {
			if err := cmd.ExecInHosts(g.gs); err != nil {
				fmt.Printf("ExecInHosts error %v", err)
			}
		}
	case cmdtype.SCP:
		for _, cmd := range g.Cmds {
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
	HostsMap  map[string]*Host
	CmdGroups []CmdGroup

	sftpClientMap sftpClientMap
}

// Close closes gossh.
func (g *GoSSH) Close() {
	g.sftpClientMap.Close()
}

// Parse parses the flags or cnf files to GoSSH.
func (c Config) Parse() GoSSH {
	gs := GoSSH{}

	_ = c.parseVars()
	gs.Hosts, gs.HostsMap = c.parseHosts()
	gs.CmdGroups = c.parseCmdGroups(&gs)
	gs.sftpClientMap = make(sftpClientMap)

	return gs
}

func (c Config) parseCmdGroups(gs *GoSSH) []CmdGroup {
	lastCmdType := cmdtype.Noop

	var group *CmdGroup

	groups := make([]*CmdGroup, 0)

	for _, cmd := range c.Cmds {
		cmdType, cmd := cmdtype.Parse(cmd)
		if cmdType == cmdtype.Noop {
			continue
		}

		if lastCmdType != cmdType {
			lastCmdType = cmdType
			group = &CmdGroup{
				gs:   gs,
				Type: cmdType,
				Cmds: make([]Cmd, 0),
			}
			groups = append(groups, group)
		}

		switch cmdType {
		case cmdtype.Local:
			group.Cmds = append(group.Cmds, &LocalCmd{cmd: cmd})
		case cmdtype.SCP:
			group.Cmds = append(group.Cmds, buildSCPCmd(cmd))
		case cmdtype.SSH:
			group.Cmds = append(group.Cmds, buildSSHCmd(cmd))
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
