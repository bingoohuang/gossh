package gossh

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"

	"github.com/gobars/cmd"
	"github.com/google/uuid"
	homedir "github.com/mitchellh/go-homedir"
)

// LocalCmd means local commands.
type LocalCmd struct {
	cmd string
}

// TargetHosts returns target hosts for the command
func (LocalCmd) TargetHosts() Hosts { return nil }

// RawCmd returns the original raw command
func (l LocalCmd) RawCmd() string { return l.cmd }

// Parse parses the local cmd
func (l *LocalCmd) Parse() {
	home, _ := homedir.Dir()
	l.cmd = strings.ReplaceAll(l.cmd, "~", home)
}

// ExecInHosts execute in specified hosts.
func (l *LocalCmd) ExecInHosts(_ *GoSSH, target *Host) error {
	if target != nil && target.ID != "localhost" {
		return nil
	}

	localCmd, uuidStr := l.buildLocalCmd()

	timeout := viper.Get("CmdTimeout").(time.Duration)

	opts := cmd.Options{Buffered: true, Streaming: true, Timeout: timeout}
	p := cmd.NewCmdOptions(opts, "/bin/bash", "-c", localCmd)
	status := p.Start()

	for {
		select {
		case so := <-p.Stdout:
			if so == uuidStr {
				fmt.Println("$", l.cmd)
			} else {
				fmt.Println(so)
			}
		case se := <-p.Stderr:
			_, _ = fmt.Fprintln(os.Stderr, se)
		case exitState := <-status:
			fmt.Println("exit status ", exitState.Exit)
			return nil
		}
	}
}

func (l *LocalCmd) buildLocalCmd() (localCmdsStr string, uuidStr string) {
	uuidStr = uuid.New().String()

	return "echo " + uuidStr + ";" + l.cmd, uuidStr
}
