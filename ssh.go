package gossh

import (
	"fmt"
	"github.com/bingoohuang/gossh/gossh"
	"io"
	"os"
	"time"

	"github.com/pkg/errors"

	"github.com/bingoohuang/gossh/cmdtype"
	"github.com/bingoohuang/gou/str"

	"github.com/spf13/viper"

	"golang.org/x/crypto/ssh"
)

// SSHCmd means SSH command.
type SSHCmd struct {
	cmd       string
	resultVar string
	hosts     Hosts
}

// Parse parses command.
func (*SSHCmd) Parse() {}

// TargetHosts returns target hosts for the command.
func (s *SSHCmd) TargetHosts() Hosts { return s.hosts }

// Exec execute in specified host.
func (s *SSHCmd) Exec(gs *GoSSH, h *Host, stdout io.Writer) error {
	cmds := []string{s.cmd}
	if gs.Vars.SplitSSH {
		cmds = str.SplitX(s.cmd, ";")
	}

	return h.SSH(cmds, s.resultVar, stdout)
}

// nolint:unparam
func (g *GoSSH) buildSSHCmd(hostPart, realCmd string) (*SSHCmd, error) {
	c, v := cmdtype.ParseResultVar(realCmd)

	return &SSHCmd{hosts: g.parseHosts(hostPart), cmd: c, resultVar: v}, nil
}

// SSH executes ssh commands  on remote host h.
// http://networkbit.ch/golang-ssh-client/
func (h *Host) SSH(cmds []string, resultVar string, stdout io.Writer) error {
	if h.client == nil {
		gc, err := h.GetGosshConnect()
		if err != nil {
			return err
		}

		h.client = gc
	}

	if err := h.setupSession(stdout); err != nil {
		return errors.Wrapf(err, "setupSession")
	}

	for _, cmd := range cmds {
		wrap := CmdWrap{Cmd: h.SubstituteResultVars(cmd), ResultVar: resultVar}
		h.cmdChan <- wrap
		h.waitCmdExecuted(wrap)
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

// nolint:gomnd
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

func ExecuteInitialCmd(initialCmd string, w io.Writer) {
	for _, v := range gossh.ConvertKeys(initialCmd) {
		time.Sleep(100 * time.Millisecond)
		_, _ = w.Write(v)
	}
}
