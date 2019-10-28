package gossh

import (
	"github.com/bingoohuang/gossh/cmdtype"
	"github.com/bingoohuang/gossh/pbe"
	"github.com/bingoohuang/gou/str"
	"github.com/spf13/viper"
)

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

// CmdExcResult means the detail exec result of cmd
type CmdExcResult struct {
}

type Cmd interface {
	Parse()
	ExecInHosts(hosts []*Host)
}

// CmdGroup represents the a group of structure of command line with same cmd type in config's cmds.
type CmdGroup struct {
	gs      *GoSSH
	Type    cmdtype.CmdType
	Cmds    []Cmd
	Results []CmdExcResult
}

func (g *CmdGroup) Parse() {
	for _, cmd := range g.Cmds {
		cmd.Parse()
	}
}

func (g *CmdGroup) Exec() {
	g.Results = make([]CmdExcResult, len(g.Cmds))
	switch g.Type {
	case cmdtype.Local:
		g.execLocal()
	case cmdtype.SSH:
		for _, cmd := range g.Cmds {
			cmd.ExecInHosts(g.gs.Hosts)
		}
	case cmdtype.SCP:
		for _, cmd := range g.Cmds {
			cmd.ExecInHosts(g.gs.Hosts)
		}
	}
}

type GoSSH struct {
	Vars      Vars
	Hosts     []*Host
	HostsMap  map[string]*Host
	CmdGroups []CmdGroup
}

func (c Config) Parse() GoSSH {
	gs := GoSSH{}

	gs.Vars = c.parseVars()
	gs.Hosts, gs.HostsMap = c.parseHosts()
	gs.CmdGroups = c.parseCmdGroups(&gs)

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

const passphrase = "passphrase"

func (c Config) parseVars() Vars {
	m := make(Vars)

	for _, v := range c.Vars {
		for k1, v1 := range str.SplitToMap(v, "=", ",") {
			m[k1] = v1
		}
	}

	if pp, ok := m[passphrase]; ok {
		viper.Set(pbe.PbePwd, pp)
	}

	return m
}
