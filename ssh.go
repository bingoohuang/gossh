package gossh

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"

	"github.com/bingoohuang/gossh/gossh"
	"golang.org/x/crypto/ssh"

	"github.com/sirupsen/logrus"
)

// SSHCmd means SSH command.
type SSHCmd struct {
	cmd   string
	hosts []*Host
}

// Parse parses command.
func (SSHCmd) Parse() {}

// TargetHosts returns target hosts for the command
func (s SSHCmd) TargetHosts() []*Host { return s.hosts }

// RawCmd returns the original raw command
func (s SSHCmd) RawCmd() string { return s.cmd }

// ExecInHosts execute in specified hosts.
func (s SSHCmd) ExecInHosts(gs *GoSSH) error {
	timeout := viper.Get("Timeout").(time.Duration)

	for _, host := range s.hosts {
		if err := func(h Host, cmd string) error {
			if err := sshInHost(h, cmd, timeout); err != nil {
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

// http://networkbit.ch/golang-ssh-client/
func sshInHost(h Host, cmd string, timeout time.Duration) error {
	fmt.Println()
	fmt.Println("---", h.Addr, "---")
	fmt.Println()

	sshClt, err := gossh.DialTCP(h.Addr, gossh.PasswordKey(h.User, h.Password, timeout))
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

	session.Stdout = os.Stdout
	session.Stderr = os.Stdin

	in, out := MuxShell(w, r)

	if err := session.Shell(); err != nil {
		return err
	}

	fmt.Print(lastLine(<-out))

	for i, cmd := range scripts {
		in <- cmd

		rout := <-out

		if i == len(scripts)-1 {
			rout = nonLastLine(rout)
		}

		fmt.Print(rout)
	}

	in <- "exit"

	return err
}

func nonLastLine(s string) string {
	pos := strings.LastIndex(s, "\n")
	if pos < 0 {
		return s
	}

	if pos == len(s)-1 {
		return s
	}

	return s[0 : pos+1]
}

func lastLine(s string) string {
	pos := strings.LastIndex(s, "\n")
	if pos < 0 {
		return s
	}

	if pos == len(s)-1 {
		return s
	}

	return s[pos+1:]
}

// MuxShell ...
func MuxShell(w io.Writer, r io.Reader) (chan<- string, <-chan string) {
	in := make(chan string, 1)
	out := make(chan string, 1)

	var wg sync.WaitGroup

	wg.Add(1) //for the shell itself

	go func() {
		for cmd := range in {
			if cmd != "exit" {
				fmt.Println(cmd)
			}

			wg.Add(1)
			_, _ = w.Write([]byte(cmd + "\n"))
			wg.Wait()
		}
	}()

	go func() {
		var buf [65 * 1024]byte
		var t int

		for {
			n, err := r.Read(buf[t:])
			if err != nil {
				close(in)
				close(out)
				return
			}

			t += n
			last := buf[t-2]
			if last == '$' || last == '#' { //assuming the $PS1 == 'sh-4.3$ '
				out <- string(buf[:t])
				t = 0
				wg.Done()
			}
		}
	}()

	return in, out
}
