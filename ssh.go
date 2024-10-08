package gossh

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/bingoohuang/gossh/pkg/cmdtype"
	"github.com/bingoohuang/gossh/pkg/gossh"
	"github.com/bingoohuang/ngg/ss"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

// SSHCmd means SSH command.
type SSHCmd struct {
	cmd       string
	resultVar string
	hosts     Hosts
}

// TargetHosts returns target hosts for the command.
func (s *SSHCmd) TargetHosts(hostGroup string) Hosts {
	hosts := make(Hosts, 0, len(s.hosts))

	for _, h := range s.hosts {
		if h.groups[hostGroup] == 1 {
			hosts = append(hosts, h)
		}
	}

	return hosts
}

// Exec execute in specified host.
func (s *SSHCmd) Exec(gs *GoSSH, h *Host, stdout io.Writer, eo ExecOption) error {
	cmds := []string{s.cmd}
	if gs.Config.SplitSSH {
		cmds = ss.SplitX(s.cmd, ";")
	}

	return h.SSH(cmds, s.resultVar, stdout, eo)
}

// nolint:unparam
func (g *GoSSH) buildSSHCmd(hostPart, realCmd string) (*SSHCmd, error) {
	c, v := cmdtype.ParseResultVar(realCmd)

	return &SSHCmd{hosts: g.parseHosts(hostPart), cmd: c, resultVar: v}, nil
}

// SSH executes ssh commands  on remote host h.
// http://networkbit.ch/golang-ssh-client/
func (h *Host) SSH(cmds []string, resultVar string, stdout io.Writer, eo ExecOption) (err error) {
	if h.client == nil {
		if h.client, err = h.GetGosshConnect(); err != nil {
			return err
		}
	}

	if err := h.setupSession(stdout); err != nil {
		return errors.Wrapf(err, "setupSession")
	}

	for _, cmd := range cmds {
		extra := parseExtra(resultVar)
		if extra != nil {
			resultVar = ""
		}

		wrap := CmdWrap{Cmd: h.SubstituteResultVars(cmd), ResultVar: resultVar, ExecOption: eo}
		h.cmdChan <- wrap
		h.waitCmdExecuted(wrap)

		if extra != nil {
			extra.DoExtra()
		}
	}

	return nil
}

type Extra struct {
	Dur time.Duration
}

func (e *Extra) DoExtra() {
	if e.Dur > 0 {
		time.Sleep(e.Dur)
	}
}

func parseExtra(resultVar string) *Extra {
	var sleepDur time.Duration
	if strings.HasPrefix(resultVar, "@sleep") {
		dur := resultVar[6:]
		if dur == "" {
			dur = "3s"
		}
		var err error
		if sleepDur, err = time.ParseDuration(dur); err != nil {
			sleepDur = 3 * time.Second
		}
		return &Extra{Dur: sleepDur}
	}

	return nil
}

func (h *Host) waitCmdExecuted(cmd CmdWrap) {
	timeout := viper.Get("CmdTimeout").(time.Duration)
	ticker := time.NewTicker(timeout)
	defer ticker.Stop()

	for {
		select {
		case executed := <-h.executedChan:
			if s, ok := executed.(CmdWrap); ok && s == cmd {
				return
			}
		case <-ticker.C:
			_ = h.Close()

			time.AfterFunc(1*time.Second, func() { close(h.executedChan) })
			for executed := range h.executedChan {
				if _, ok := executed.(error); ok {
					break
				}
			}

			fmt.Printf("[%s] TIMOUT IN %v\n", cmd, timeout)
			return
		}
	}
}

func (h *Host) setupSession(stdout io.Writer) error {
	if h.session != nil {
		return nil
	}

	session, err := h.client.Client.NewSession()
	if err != nil {
		return err
	}

	// disable echoing input/output speed = 14.4kbaud
	modes := ssh.TerminalModes{ssh.ECHO: 0, ssh.TTY_OP_ISPEED: 14400, ssh.TTY_OP_OSPEED: 14400}

	term := os.Getenv("TERM")
	if term == "" {
		term = "xterm-256color" // alternative to vt100
	}

	if err := session.RequestPty(term, 800, 400, modes); err != nil {
		return err
	}

	w, err := session.StdinPipe()
	if err != nil {
		return err
	}

	r, err := session.StdoutPipe()
	if err != nil {
		return err
	}

	if err := session.Shell(); err != nil {
		return err
	}

	tryReader := gossh.NewTryReader(r)
	if v := h.Prop("initial_cmd"); v != "" {
		ExecuteInitialCmd(v, w)
	}

	h.session = session
	h.w = w
	h.r = tryReader
	h.cmdChan = make(chan CmdWrap, 1)
	h.executedChan = make(chan interface{}, 1)

	go mux(h.cmdChan, h.executedChan, h.w, h.r, h, stdout)

	return nil
}

// ExecuteInitialCmd executes initial command.
func ExecuteInitialCmd(initialCmd string, w io.Writer) {
	for _, v := range gossh.ConvertKeys(initialCmd) {
		time.Sleep(100 * time.Millisecond)
		_, _ = w.Write(v)
	}
}
