package gossh

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/bingoohuang/gossh/pkg/cmdtype"
	"github.com/bingoohuang/gossh/pkg/gossh"
	"github.com/bingoohuang/gou/pbe"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"
)

// Config represents the structure of input toml file structure.
type Config struct {
	ReplaceQuote string `pflag:"replace for quote(\"). shorthand=q"`
	ReplaceBang  string `pflag:"replace for bang(!). shorthand=b"`

	Separator  string `pflag:"separator for hosts, cmds, default comma. shorthand=s"`
	NetTimeout string `pflag:"timeout(eg. 15s, 3m), empty for no timeout."`
	CmdTimeout string `pflag:"timeout(eg. 15s, 3m), default 15m."`

	Group    string `pflag:"group name."`
	CmdsFile string `pflag:"cmds file."`

	HostsFile string `pflag:"hosts file. shorthand=f"`
	Pass      string `pflag:"pass."`

	User string `pflag:"user. shorthand=u"`

	Passphrase string   `pflag:"passphrase for decrypt {PBE}Password. shorthand=p"`
	Cmds       []string `pflag:"commands to be executedChan. shorthand=C"`

	Hosts []string `pflag:"hosts. shorthand=H"`

	ExecMode int `pflag:"exec mode(0: cmd by cmd, 1 host by host). shorthand=e"`

	FirstConfirm bool

	Confirm bool `pflag:"conform to continue."`
	// 是否全局设置为远程shell命令
	GlobalRemote bool `pflag:"run as global remote ssh command(no need %host). shorthand=g"`
	PrintConfig  bool `pflag:"print config before running. shorthand=P"`

	SplitSSH bool `pflag:"split ssh commands by comma or not. shorthand=S"`
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

	ExecOption
}

func (c CmdWrap) String() string { return c.Cmd }

// Host represents the structure of remote host information for ssh.
type Host struct {
	w          io.WriteCloser
	r          io.Reader
	Properties map[string][]string

	sftpClient *sftp.Client

	groups map[string]int

	Proxy *Host

	client  *gossh.Connect
	session *ssh.Session

	resultVars map[string]string

	sftpSSHClient *ssh.Client

	cmdChan      chan CmdWrap
	executedChan chan interface{}

	Password string // empty when using public key
	Addr     string
	User     string
	ID       string

	localConnected bool
}

// globalVarsMap is the global map of result variable.
var globalVarsMap sync.Map

// cmdByCmd executes command one by one.
const cmdByCmd = "_CmdByCmd"

// NewExecModeCmdByCmd creates an exec mode command.
func NewExecModeCmdByCmd() *Host {
	return &Host{ID: cmdByCmd}
}

// IsExecModeCmdByCmd tests if this is mode of cmd one by one or not.
func (h *Host) IsExecModeCmdByCmd() bool { return h.ID == cmdByCmd }

// SubstituteResultVars substitutes the variables in the command line string.
func (h *Host) SubstituteResultVars(cmd string) string {
	for k, v := range h.Properties {
		cmd = strings.ReplaceAll(cmd, "@"+k, v[0])
	}
	for k, v := range h.resultVars {
		cmd = strings.ReplaceAll(cmd, k, v)
	}

	globalVarsMap.Range(func(k, v interface{}) bool {
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

	if IsCapitalized(varName[1:]) {
		globalVarsMap.Store(varName, varValue)
	} else {
		if h.resultVars == nil {
			h.resultVars = make(map[string]string)
		}
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

	key := gossh.PasswordKey(h.User, h.Password)
	if err := gc.CreateClient(h.Addr, key); err != nil {
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
	if v := h.Properties[name]; len(v) > 0 {
		return v[0]
	}

	return ""
}

// IsConnected tells if host is connected by ssh or sftp.
func (h *Host) IsConnected() bool {
	if h.ID == "localhost" {
		if h.localConnected {
			return true
		}

		h.localConnected = true
		return false
	}

	return h.client != nil || h.sftpClient != nil
}

// CmdExcResult means the detail exec result of cmd.
type CmdExcResult struct{}

// HostsCmd means the executable interface.
type HostsCmd interface {
	// Exec execute in specified host.
	Exec(gs *GoSSH, host *Host, stdout io.Writer, eo ExecOption) error
	// TargetHosts returns target hosts for the command
	TargetHosts(hostGroup string) Hosts
}

// ExecCmds executes commands.
func ExecCmds(gs *GoSSH, host *Host, stdout io.Writer, eo ExecOption, hostGroup string) {
	for _, cmd := range gs.Cmds {
		if err := ExecInHosts(gs, host, cmd, stdout, eo, hostGroup); err != nil {
			fmt.Fprintf(stdout, "ExecInHosts error %v\n", err)
		}
	}
}

// ExecOption defines the options of execute.
type ExecOption struct {
	Repl bool
}

// ExecInHosts execute in specified hosts.
func ExecInHosts(gs *GoSSH, target *Host, hostsCmd HostsCmd, stdout io.Writer, eo ExecOption, hostGroup string) error {
	for _, host := range hostsCmd.TargetHosts(hostGroup) {
		if target.IsExecModeCmdByCmd() || target == host {
			if target.IsExecModeCmdByCmd() {
				if eo.Repl || target.Addr != host.Addr {
					_, _ = fmt.Fprintf(stdout, "\n---> %s <---\n", host.Addr)
					target.Addr = host.Addr
				}
			} else if eo.Repl || !host.IsConnected() {
				_, _ = fmt.Fprintf(stdout, "\n---> %s <---\n", host.Addr)
			}

			if gs.Config.Confirm {
				if !gs.Config.FirstConfirm {
					gs.Config.FirstConfirm = true
				} else {
					fmt.Print("Press Enter to go on:")
					reader := bufio.NewReader(os.Stdin)
					_, _ = reader.ReadString('\n')
				}
			}

			if err := hostsCmd.Exec(gs, host, stdout, eo); err != nil {
				_, _ = fmt.Fprintf(stdout, "Error occurred %v\n", err)
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

// FixHost fix the host ID by sequence if it is blank.
func (hosts Hosts) FixHost() {
	for i, h := range hosts {
		if h.ID == "" {
			h.ID = fmt.Sprintf("%d", i+1)
		}
		if v, err := pbe.Ebp(h.Password); err != nil {
			panic(err)
		} else {
			h.Password = v
		}

		h.groups = make(map[string]int)
		groups := h.Properties["groups"]
		if len(groups) == 0 {
			groups = h.Properties["group"]
		}
		if len(groups) > 0 {
			for _, group := range strings.Split(groups[0], "/") {
				if group != "" {
					h.groups[group] = 1
				}
			}
		}
		if len(h.groups) == 0 {
			h.groups["default"] = 1
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

		if i == 10 {
			logrus.Errorf("proxy chain can not exceed 10!")
		}
	}
}

// GoSSH defines the structure of the whole cfg context.
type GoSSH struct {
	Config *Config
	Hosts  Hosts
	Cmds   []HostsCmd
}

// Close closes gossh.
func (g *GoSSH) Close() error {
	return g.Hosts.Close()
}

// Parse parses the flags or cnf files to GoSSH.
func (c *Config) Parse() GoSSH {
	gs := GoSSH{}

	c.parseCmdsFile()
	c.parseVars()

	c.fixPass()

	gs.Config = c
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
		hostCmd, err := c.parseCmd(gs, cmd)
		if err != nil {
			logrus.Fatalf("failed to parse cmd: %s error: %v", cmd, err)
		}
		if hostCmd != nil {
			cmds = append(cmds, hostCmd)
		}
	}

	return cmds
}

func (c *Config) parseCmd(gs *GoSSH, cmd string) (hostCmd HostsCmd, err error) {
	switch cmdType, hostPart, realCmd := cmdtype.Parse(c.GlobalRemote, cmd); cmdType {
	case cmdtype.Local:
		hostCmd = gs.buildLocalCmd(realCmd)
	case cmdtype.Ul:
		hostCmd, err = gs.buildUlCmd(hostPart, realCmd)
	case cmdtype.Dl:
		hostCmd, err = gs.buildDlCmd(hostPart, realCmd)
	case cmdtype.SSH:
		hostCmd, err = gs.buildSSHCmd(hostPart, realCmd)
	}

	return hostCmd, err
}

func (c *Config) parseCmdsFile() {
	if c.CmdsFile == "" {
		return
	}

	cmdsFile, _ := homedir.Expand(c.CmdsFile)
	file, err := os.ReadFile(cmdsFile)
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
	DecryptPassphrase(c.Passphrase)

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

func DecryptPassphrase(passphrase string) {
	if passphrase == "" {
		passphrase = os.Getenv("PASS")
	}
	if passphrase != "" {
		if strings.HasPrefix(passphrase, "{PBE}") {
			// 身无彩凤双飞翼，心有灵犀一点通
			viper.Set(pbe.PbePwd, "S!cfsf1*Ylx1.t")
			if p, err := pbe.Ebp(passphrase); err == nil {
				passphrase = p
			}
			viper.Set(pbe.PbePwd, "")
		}

		viper.Set(pbe.PbePwd, passphrase)
	}
}
