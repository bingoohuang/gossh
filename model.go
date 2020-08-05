package gossh

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"

	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"

	"github.com/bingoohuang/gossh/gossh"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/sftp"

	"github.com/sirupsen/logrus"

	"github.com/bingoohuang/gou/pbe"

	"github.com/bingoohuang/gossh/cmdtype"
	"github.com/spf13/viper"
)

// Config represents the structure of input toml file structure.
type Config struct {
	ReplaceQuote string `pflag:"replace for quote(\"). shorthand=q"`
	ReplaceBang  string `pflag:"replace for bang(!). shorthand=b"`

	Separator  string `pflag:"separator for hosts, cmds, default comma. shorthand=s"`
	NetTimeout string `pflag:"timeout(eg. 15s, 3m), empty for no timeout."`
	CmdTimeout string `pflag:"timeout(eg. 15s, 3m), default 15m."`

	SplitSSH    bool `pflag:"split ssh commands by comma or not. shorthand=S"`
	PrintConfig bool `pflag:"print config before running. shorthand=P"`
	// 是否全局设置为远程shell命令
	GlobalRemote bool `pflag:"run as global remote ssh command(no need %host). shorthand=g"`

	Passphrase string   `pflag:"passphrase for decrypt {PBE}Password. shorthand=p"`
	Hosts      []string `pflag:"hosts. shorthand=H"`
	Cmds       []string `pflag:"commands to be executedChan. shorthand=C"`

	User      string `pflag:"user. shorthand=u"`
	Pass      string `pflag:"pass."`
	HostsFile string `pflag:"hosts file. shorthand=f"`
	CmdsFile  string `pflag:"cmds file."`

	ExecMode int `pflag:"exec mode(0: cmd by cmd, 1 host by host). shorthand=e"`
	log      *log.Logger
}

const (
	// ExecModeCmdByCmd means execute a command in all relative hosts and then continue to next command
	// eg. cmd1: host1,host2, cmd2:host1, host2
	ExecModeCmdByCmd int = iota
	// ExecModeHostByHost means execute a host relative commands and the continue to next host.
	// eg .host1: cmd1,cmd2, host2:cmd1, cmd2
	ExecModeHostByHost
)

// GetSeparator get the separator.
func (c Config) GetSeparator() string { return c.Separator }

// CmdWrap wraps a command with result variable name.
type CmdWrap struct {
	Cmd       string
	ResultVar string
}

// Host represents the structure of remote host information for ssh.
type Host struct {
	ID         string
	Addr       string
	User       string
	Password   string // empty when using public key
	Properties map[string]string

	Proxy *Host

	client       *gossh.Connect
	session      *ssh.Session
	w            io.WriteCloser
	r            io.Reader
	cmdChan      chan CmdWrap
	executedChan chan interface{}

	sftpClient    *sftp.Client
	sftpSSHClient *ssh.Client

	resultVars map[string]string
}

// resultVarsMap is the global map of result variable.
var resultVarsMap sync.Map // nolint:gochecknoglobals

// SubstituteResultVars substitutes the variables in the command line string.
func (h *Host) SubstituteResultVars(cmd string) string {
	for k, v := range h.resultVars {
		cmd = strings.ReplaceAll(cmd, k, v)
	}

	resultVarsMap.Range(func(k, v interface{}) bool {
		cmd = strings.ReplaceAll(cmd, k.(string), v.(string))
		return true
	})

	return cmd
}

// SetResultVar sets the value of result variable.
func (h *Host) SetResultVar(varName, varValue string) {
	if varName == "" {
		return
	}

	if len(varValue) > 1 && IsCapitalized(varName[1:]) {
		resultVarsMap.Store(varName, varValue)
	} else {
		h.resultVars[varName] = varValue
	}
}

// IsCapitalized test a string is a capitalized one.
func IsCapitalized(str string) bool {
	for _, v := range str {
		return unicode.IsUpper(v)
	}

	return false
}

// Close closes the resource associated to the host.
func (h *Host) Close() error {
	var g errgroup.Group

	if h.cmdChan != nil {
		close(h.cmdChan)

		h.cmdChan = nil
	}

	if s := h.session; s != nil {
		h.session = nil

		g.Go(s.Close)
	}

	if c := h.client; c != nil {
		h.client = nil

		g.Go(c.Close)
	}

	if c := h.sftpClient; c != nil {
		h.sftpClient = nil

		g.Go(c.Close)
	}

	if c := h.sftpSSHClient; c != nil {
		h.sftpSSHClient = nil

		g.Go(c.Close)
	}

	return g.Wait()
}

// GetGosshConnect get gossh Connect.
func (h *Host) GetGosshConnect() (*gossh.Connect, error) {
	gc := &gossh.Connect{}

	if h.Proxy != nil {
		pc, err := h.Proxy.GetGosshConnect()
		if err != nil {
			return nil, err
		}

		gc.ProxyDialer = pc.Client
	}

	if err := gc.CreateClient(h.Addr, gossh.PasswordKey(h.User, h.Password)); err != nil {
		return nil, fmt.Errorf("CreateClient(%s) failed: %w", h.Addr, err)
	}

	return gc, nil
}

const ignoreWarning = "-q -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"

// PrintSSH prints sshpass ssh commands.
func (h *Host) PrintSSH() {
	host, port, _ := net.SplitHostPort(h.Addr)

	sshCmd := fmt.Sprintf("sshpass -p %s ssh -p%s %s %s@%s", h.Password, port, ignoreWarning, h.User, host)
	fmt.Println(sshCmd)
}

// PrintSCP prints sshpass scp commands.
func (h Host) PrintSCP() {
	host, port, _ := net.SplitHostPort(h.Addr)

	// sshpass -p xxx scp -P 9922 root@192.168.205.148:/root/xxx .
	scpCmd := fmt.Sprintf("sshpass -p %s scp -P%s %s %s@%s:. .", h.Password, port, ignoreWarning, h.User, host)
	fmt.Println(scpCmd)
}

// Prop finds property by name.
func (h *Host) Prop(name string) string {
	if v, ok := h.Properties[name]; ok {
		return v
	}

	return ""
}

// IsConnected tells if host is connected by ssh or sftp.
func (h *Host) IsConnected() bool {
	return h.client != nil || h.sftpClient != nil
}

// CmdExcResult means the detail exec result of cmd.
type CmdExcResult struct {
}

// HostsCmd means the executable interface.
type HostsCmd interface {
	// Parse parses the command.
	Parse()
	// Exec execute in specified host.
	Exec(gs *GoSSH, host *Host) error
	// TargetHosts returns target hosts for the command
	TargetHosts() Hosts
}

// ExecInHosts execute in specified hosts.
func ExecInHosts(gs *GoSSH, target *Host, hostsCmd HostsCmd) error {
	for _, host := range hostsCmd.TargetHosts() {
		if target == nil || target == host {
			if target == nil || !host.IsConnected() {
				fmt.Printf("\n---> %s <--- \n\n", host.Addr)
			}

			if err := hostsCmd.Exec(gs, host); err != nil {
				fmt.Printf("Error occurred %v\n", err)
			}
		}
	}

	return nil
}

// Hosts stands for slice of Host.
type Hosts []*Host

// Close closes all the host related resources.
func (hosts Hosts) Close() error {
	var g errgroup.Group

	for _, host := range hosts {
		g.Go(host.Close)
	}

	return g.Wait()
}

// PrintSSH prints sshpass ssh commands for all hosts.
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

// PrintSCP prints sshpass scp commands for all hosts.
func (hosts Hosts) PrintSCP() {
	for _, h := range hosts {
		h.PrintSCP()
	}
}

// FixProxy fix proxy.
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

		if i == 10 { // nolint:gomnd
			logrus.Errorf("proxy chain can not exceed 10!")
		}
	}
}

// GoSSH defines the structure of the whole cfg context.
type GoSSH struct {
	Vars  *Config
	Hosts Hosts

	Cmds []HostsCmd
}

// Close closes gossh.
func (g *GoSSH) Close() error {
	return g.Hosts.Close()
}

// LogPrintf calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (g *GoSSH) LogPrintf(format string, v ...interface{}) {
	g.Vars.log.Printf(format, v...)
}

// Parse parses the flags or cnf files to GoSSH.
func (c *Config) Parse() GoSSH {
	gs := GoSSH{}

	c.parseCmdsFile()
	c.parseVars()

	c.fixPass()

	c.log = log.New(os.Stdout, "", 0)

	gs.Vars = c
	gs.Hosts = c.parseHosts()
	gs.Cmds = c.parseCmdGroups(&gs)

	return gs
}

func (c *Config) fixPass() {
	if c.Pass == "" {
		return
	}

	var err error

	if c.Pass, err = pbe.Ebp(c.Pass); err != nil {
		panic(err)
	}
}

func (c *Config) parseCmdGroups(gs *GoSSH) []HostsCmd {
	cmds := make([]HostsCmd, 0)

	for _, cmd := range c.Cmds {
		cmdType, hostPart, realCmd := cmdtype.Parse(c.GlobalRemote, cmd)
		if cmdType == cmdtype.Noop {
			continue
		}

		var (
			err     error
			hostCmd HostsCmd
		)

		switch cmdType {
		case cmdtype.Local:
			hostCmd = gs.buildLocalCmd(cmd)
		case cmdtype.Ul:
			hostCmd, err = gs.buildUlCmd(hostPart, realCmd, cmd)
		case cmdtype.Dl:
			hostCmd, err = gs.buildDlCmd(hostPart, realCmd, cmd)
		case cmdtype.SSH:
			hostCmd, err = gs.buildSSHCmd(hostPart, realCmd, cmd)
		case cmdtype.Noop:
			continue
		default:
			continue
		}

		if err != nil {
			logrus.Fatalf("failed to build ul command %v", err)
		}

		hostCmd.Parse()
		cmds = append(cmds, hostCmd)
	}

	return cmds
}

func (c *Config) parseCmdsFile() {
	if c.CmdsFile == "" {
		return
	}

	cmdsFile, _ := homedir.Expand(c.CmdsFile)
	file, err := ioutil.ReadFile(cmdsFile)

	if err != nil {
		logrus.Warnf("failed to read cmds file %s: %v", c.CmdsFile, err)
		return
	}

	for _, line := range strings.Split(string(file), "\n") {
		if l := strings.TrimSpace(line); l != "" && !strings.HasPrefix(l, "#") {
			c.Cmds = append(c.Cmds, l)
		}
	}
}

func (c *Config) parseVars() {
	if c.Passphrase != "" {
		viper.Set(pbe.PbePwd, c.Passphrase)
	}

	netTimeout, _ := time.ParseDuration(c.NetTimeout)
	viper.Set("NetTimeout", netTimeout)

	cmdTimeout, _ := time.ParseDuration(c.CmdTimeout)
	if cmdTimeout == 0 {
		cmdTimeout = 15 * time.Minute // nolint
	}

	viper.Set("CmdTimeout", cmdTimeout)

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
