package gossh

import (
	"fmt"
	"time"

	"github.com/bingoohuang/gou/str"

	"github.com/spf13/viper"

	"golang.org/x/crypto/ssh"

	"github.com/sirupsen/logrus"
)

// SSHCmd means SSH command.
type SSHCmd struct {
	cmd   string
	hosts Hosts
}

// Parse parses command.
func (SSHCmd) Parse() {}

// TargetHosts returns target hosts for the command
func (s SSHCmd) TargetHosts() Hosts { return s.hosts }

// RawCmd returns the original raw command
func (s SSHCmd) RawCmd() string { return s.cmd }

// ExecInHosts execute in specified hosts.
func (s SSHCmd) ExecInHosts(gs *GoSSH) error {
	timeout := viper.Get("Timeout").(time.Duration)

	for _, host := range s.hosts {
		if err := func(h Host, cmd string) error {
			cmds := []string{cmd}
			if gs.Vars.SplitSSH {
				cmds = str.SplitX(cmd, ";")
			}

			if err := h.SSH(cmds, timeout); err != nil {
				logrus.Warnf("ssh in host %s error %v", h.Addr, err)
				return err
			}
			return nil
		}(*host, s.cmd); err != nil {
			return err
		}
	}

	return nil
}

func buildSSHCmd(gs *GoSSH, hostPart, realCmd, _ string) *SSHCmd {
	return &SSHCmd{hosts: parseHosts(gs, hostPart), cmd: realCmd}
}

// SSH executes ssh commands  on remote host h.
// http://networkbit.ch/golang-ssh-client/
func (h Host) SSH(cmd []string, timeout time.Duration) error {
	fmt.Println()
	fmt.Println("---", h.Addr, "---")

	gc, err := h.GetGosshConnect(timeout)
	if err != nil {
		return err
	}

	defer gc.Close()

	if err := sshScripts(gc.Client, cmd); err != nil {
		return fmt.Errorf("exec cmd %s failed: %w", cmd, err)
	}

	return nil
}

// nolint gomnd
func sshScripts(client *ssh.Client, cmd []string) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}

	defer session.Close()

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

	mux(cmd, w, r)

	return nil
}
