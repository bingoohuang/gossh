package gossh

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/bingoohuang/gossh/gossh"
	"golang.org/x/crypto/ssh"

	"github.com/sirupsen/logrus"
)

// SSHCmd means SSH command.
type SSHCmd struct {
	cmd  string
	host string
}

// Parse parses command.
func (SSHCmd) Parse() {}

// ExecInHosts execute in specified hosts.
func (s SSHCmd) ExecInHosts(gs *GoSSH) error {
	hostName := ""

	if submatchIndex := regexp.MustCompile(`%host(-\w+)?`).
		FindStringSubmatchIndex(s.host); len(submatchIndex) > 0 {
		if submatchIndex[2] > 0 {
			hostName = s.host[submatchIndex[2]:submatchIndex[3]]
		}
	}

	targetHosts := filterHosts(hostName, gs)
	if len(targetHosts) == 0 {
		logrus.Warnf("there is no host to ssh %s", s.cmd)
	}

	printCmds(targetHosts, s)

	for _, host := range targetHosts {
		if err := func(h Host, cmd string) error {
			if err := sshInHost(*host, cmd); err != nil {
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

func printCmds(targetHosts []*Host, s SSHCmd) {
	targetHostnames := filterHostnames(targetHosts)

	fmt.Println("executing ", s.cmd, "on hosts", targetHostnames)
}

func filterHostnames(targetHosts []*Host) []string {
	targetHostnames := make([]string, len(targetHosts))
	for i, t := range targetHosts {
		targetHostnames[i] = t.Addr
	}

	return targetHostnames
}

func filterHosts(hostName string, gs *GoSSH) []*Host {
	hostName = strings.TrimPrefix(hostName, "-")

	targetHosts := make([]*Host, 0, len(gs.Hosts))

	for _, host := range gs.Hosts {
		if hostName == "" || host.Name == hostName {
			targetHosts = append(targetHosts, host)
		}
	}

	return targetHosts
}

func buildSSHCmd(cmd string) *SSHCmd {
	parts := strings.SplitN(cmd, " ", 3)
	if len(parts) != 3 {
		logrus.Warnf("bad format for %s", cmd)
		return nil
	}

	return &SSHCmd{host: parts[1], cmd: parts[2]}
}

// http://networkbit.ch/golang-ssh-client/
func sshInHost(h Host, cmd string) error {
	sshClt, err := gossh.DialTCP(h.Addr, gossh.PasswordKey(h.User, h.Password))
	if err != nil {
		return fmt.Errorf("ssh.Dial(%q) failed: %w", h.Addr, err)
	}

	defer sshClt.Close()

	if err := sshScripts(sshClt, []string{cmd}); err != nil {
		return fmt.Errorf("exec cmd %s failed: %w", cmd, err)
	}

	return nil
}

func sshScripts(client *ssh.Client, scripts []string) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}

	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := session.RequestPty("vt100", 800, 400, modes); err != nil {
		return err
	}

	w, err := session.StdinPipe()
	if err != nil {
		return err
	}

	session.Stdout = os.Stdout
	session.Stderr = os.Stdin

	if err := session.Shell(); err != nil {
		return err
	}

	for _, cmd := range scripts {
		_, _ = w.Write([]byte(cmd + "\n"))
	}

	_, _ = w.Write([]byte("exit\n"))

	if err := session.Wait(); err != nil {
		return err
	}

	return err
}
