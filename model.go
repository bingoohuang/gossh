package gossh

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/bingoohuang/gossh/gossh"

	"github.com/sirupsen/logrus"

	"github.com/bingoohuang/gou/pbe"

	"github.com/bingoohuang/gossh/cmdtype"
	"github.com/spf13/viper"
)

// Config represents the structure of input toml file structure.
type Config struct {
	ReplaceQuote string `pflag:"replace for quote(\"). shorthand=q"`
	ReplaceBang  string `pflag:"replace for bang(!). shorthand=b"`

	Separator string `pflag:"separator for hosts, cmds, default comma. shorthand=s"`
	Timeout   string `pflag:"timeout(eg. 15s, 3m), empty for no timeout. shorthand=t"`

	SplitSSH    bool `pflag:"split ssh commands by comma or not. shorthand=S"`
	PrintConfig bool `pflag:"print config before running. shorthand=P"`

	Passphrase string   `pflag:"passphrase for decrypt {PBE}Password. shorthand=p"`
	Hosts      []string `pflag:"hosts. shorthand=H"`
	Cmds       []string `pflag:"commands to be executed. shorthand=C"`

	User      string `pflag:"user. shorthand=u"`
	Pass      string `pflag:"pass."`
	HostsFile string `pflag:"hosts file. shorthand=f"`
	CmdsFile  string `pflag:"cmds file."`
}

// GetSeparator get the separator
func (c Config) GetSeparator() string { return c.Separator }

// Host represents the structure of remote host information for ssh.
type Host struct {
	ID         string
	Addr       string
	User       string
	Password   string // empty when using public key
	Properties map[string]string

	Proxy *Host
}

// GetGosshConnect get gossh Connect
func (h Host) GetGosshConnect(timeout time.Duration) (*gossh.Connect, error) {
	gc := &gossh.Connect{}

	if h.Proxy != nil {
		pc, err := h.Proxy.GetGosshConnect(timeout)
		if err != nil {
			return nil, err
		}

		gc.ProxyDialer = pc.Client
	}

	if err := gc.CreateClient(h.Addr, gossh.PasswordKey(h.User, h.Password, timeout)); err != nil {
		return nil, fmt.Errorf("CreateClient(%s) failed: %w", h.Addr, err)
	}

	return gc, nil
}

const ignoreWarning = "-q -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"

// PrintSSH prints sshpass ssh commands
func (h Host) PrintSSH() {
	host, port, _ := net.SplitHostPort(h.Addr)

	sshCmd := fmt.Sprintf("sshpass -p %s ssh -p%s %s %s@%s", h.Password, port, ignoreWarning, h.User, host)
	fmt.Println(sshCmd)
}

// PrintSCP prints sshpass scp commands
func (h Host) PrintSCP() {
	host, port, _ := net.SplitHostPort(h.Addr)

	// sshpass -p xxx scp -P 9922 root@192.168.205.148:/root/xxx .
	scpCmd := fmt.Sprintf("sshpass -p %s scp -P%s %s %s@%s:. .", h.Password, port, ignoreWarning, h.User, host)
	fmt.Println(scpCmd)
}

// Prop finds property by name
func (h Host) Prop(name string) string {
	if v, ok := h.Properties[name]; ok {
		return v
	}

	return ""
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
	TargetHosts() Hosts

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

	if g.Type == cmdtype.Local {
		g.execLocal()
		return
	}

	for _, cmd := range g.Cmds {
		if len(cmd.TargetHosts()) == 0 {
			fmt.Printf("No target hosts for cmd %s to executed\n", cmd.RawCmd())
			continue
		}

		if err := cmd.ExecInHosts(g.gs); err != nil {
			fmt.Printf("ExecInHosts error %v", err)
		}
	}
}

// Hosts stands for slice of Host
type Hosts []*Host

// PrintSSH prints sshpass ssh commands for all hosts
func (hosts Hosts) PrintSSH() {
	for _, h := range hosts {
		h.PrintSSH()
	}
}

// FixHostID fix the host ID by sequence if it is blank.
func (hosts Hosts) FixHostID() {
	for i, h := range hosts {
		if h.ID == "" {
			h.ID = fmt.Sprintf("%d", i+1)
		}
	}
}

// PrintSCP prints sshpass scp commands for all hosts
func (hosts Hosts) PrintSCP() {
	for _, h := range hosts {
		h.PrintSCP()
	}
}

// FixProxy fix proxy
func (hosts Hosts) FixProxy() {
	m := make(map[string]*Host)
	for _, h := range hosts {
		m[h.ID] = h
	}

	for _, h := range hosts {
		if proxy := h.Prop("proxy"); proxy != "" && proxy != "-" {
			if proxyHost, ok := m[proxy]; ok {
				h.Proxy = proxyHost
			} else {
				logrus.Panicf("unable to fine proxy host by ID %s", proxy)
			}
		}
	}

	// 检测proxy的环
	for _, h := range hosts {
		if h.Proxy == nil {
			continue
		}

		m := make(map[string]bool)
		m[h.ID] = true

		h = h.Proxy
		i := 0

		for ; i < 10 && h != nil; i++ {
			if _, ok := m[h.ID]; ok {
				logrus.Errorf("proxy circled!")
				os.Exit(1)
			}

			m[h.ID] = true
			h = h.Proxy
		}

		if i == 10 {
			logrus.Errorf("proxy chain can not exceed 10!")
		}
	}
}

// GoSSH defines the structure of the whole cfg context.
type GoSSH struct {
	Vars      Config
	Hosts     Hosts
	CmdGroups []CmdGroup

	sftpClientMap *sftpClientMap
}

// Close closes gossh.
func (g *GoSSH) Close() {
	g.sftpClientMap.Close()
}

// Parse parses the flags or cnf files to GoSSH.
func (c *Config) Parse() GoSSH {
	gs := GoSSH{}

	c.parseVars()

	gs.Vars = *c
	gs.Hosts = c.parseHosts()
	gs.CmdGroups = c.parseCmdGroups(&gs)
	timeout := viper.Get("Timeout").(time.Duration)
	gs.sftpClientMap = makeSftpClientMap(timeout)

	return gs
}

func (c *Config) parseCmdGroups(gs *GoSSH) []CmdGroup {
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

func (c *Config) parseVars() {
	if c.Passphrase != "" {
		viper.Set(pbe.PbePwd, c.Passphrase)
	}

	duration, _ := time.ParseDuration(c.Timeout)
	viper.Set("Timeout", duration)

	if c.ReplaceQuote != "" {
		for i, cmd := range c.Cmds {
			c.Cmds[i] = strings.ReplaceAll(cmd, c.ReplaceQuote, `"`)
		}
	}

	if c.ReplaceBang != "" {
		for i, cmd := range c.Cmds {
			c.Cmds[i] = strings.ReplaceAll(cmd, c.ReplaceBang, `!`)
		}
	}
}
