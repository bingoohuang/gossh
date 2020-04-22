package gossh

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/bingoohuang/gou/str"

	"github.com/spf13/viper"

	"golang.org/x/crypto/ssh"
)

// SSHCmd means SSH command.
type SSHCmd struct {
	cmd   string
	hosts Hosts
}

// Parse parses command.
func (*SSHCmd) Parse() {}

// TargetHosts returns target hosts for the command
func (s *SSHCmd) TargetHosts() Hosts { return s.hosts }

func (s *SSHCmd) Exec(gs *GoSSH, h *Host) error {
	cmds := []string{s.cmd}
	if gs.Vars.SplitSSH {
		cmds = str.SplitX(s.cmd, ";")
	}

	return h.SSH(cmds)
}

func buildSSHCmd(gs *GoSSH, hostPart, realCmd, _ string) *SSHCmd {
	return &SSHCmd{hosts: parseHosts(gs, hostPart), cmd: realCmd}
}

// SSH executes ssh commands  on remote host h.
// http://networkbit.ch/golang-ssh-client/
func (h *Host) SSH(cmds []string) error {
	if h.client == nil {
		gc, err := h.GetGosshConnect()
		if err != nil {
			return err
		}

		h.client = gc
	}

	if err := h.setupSession(); err != nil {
		return errors.Wrapf(err, "setupSession")
	}

	for _, cmd := range cmds {
		h.cmdChan <- cmd
		h.waitCmdExecuted(cmd)
	}

	return nil
}

func (h *Host) waitCmdExecuted(cmd string) {
	timeout := viper.Get("CmdTimeout").(time.Duration)
	ticker := time.NewTicker(timeout)

	defer ticker.Stop()

	for {
		select {
		case executed := <-h.executedChan:
			if s, ok := executed.(string); ok && s == cmd {
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

// nolint gomnd
func (h *Host) setupSession() error {
	if h.session == nil {
		session, err := h.client.Client.NewSession()
		if err != nil {
			return err
		}

		// disable echoing input/output speed = 14.4kbaud
		modes := ssh.TerminalModes{ssh.ECHO: 0, ssh.TTY_OP_ISPEED: 14400, ssh.TTY_OP_OSPEED: 14400}
		if err := session.RequestPty("vt100", 800, 400, modes); err != nil {
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

		h.session = session
		h.w = w
		h.r = r
		h.cmdChan = make(chan string, 1)
		h.executedChan = make(chan interface{}, 1)

		go mux(h.cmdChan, h.executedChan, h.w, h.r)
	}

	return nil
}
