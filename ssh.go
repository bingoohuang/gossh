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
			if err := h.SSH([]string{cmd}, timeout); err != nil {
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

	sshClt, err := gossh.DialTCP(h.Addr, gossh.PasswordKey(h.User, h.Password, timeout))
	if err != nil {
		return fmt.Errorf("ssh.Dial(%q) failed: %w", h.Addr, err)
	}

	defer sshClt.Close()

	if err := sshScripts(sshClt, cmd); err != nil {
		return fmt.Errorf("exec cmd %s failed: %w", cmd, err)
	}

	return nil
}

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

	session.Stdout = os.Stdout
	session.Stderr = os.Stdin

	in, out := MuxShell(w, r)

	if err := session.Shell(); err != nil {
		return err
	}

	mux(cmd, out, in)

	return nil
}

func mux(cmd []string, out <-chan string, in chan<- string) {
	fmt.Print(GetLastLine(<-out))

	for i, cmd := range cmd {
		in <- cmd

		rout := <-out

		if i == len(cmd)-1 {
			rout = StripLastLine(rout)
		}

		fmt.Print(rout)
	}
}

// StripLastLine strips the last line of s.
func StripLastLine(s string) string {
	pos := strings.LastIndex(s, "\n")
	if pos < 0 || pos == len(s)-1 {
		return s
	}

	return s[0 : pos+1]
}

// GetLastLine gets the last line of s.
func GetLastLine(s string) string {
	pos := strings.LastIndex(s, "\n")
	if pos < 0 || pos == len(s)-1 {
		return s
	}

	return s[pos+1:]
}

// MuxShell ...
func MuxShell(w io.Writer, r io.Reader) (chan<- string, <-chan string) {
	in, out := make(chan string, 1), make(chan string, 1)

	var wg sync.WaitGroup

	wg.Add(1) //for the shell itself

	go func() {
		for cmd := range in {
			fmt.Println(cmd)
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
