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

// ExecInHosts execute in specified hosts.
func (LocalCmd) ExecInHosts(_ *GoSSH) error { return nil }

// RawCmd returns the original raw command
func (l LocalCmd) RawCmd() string { return l.cmd }

// Parse parses the local cmd
func (l *LocalCmd) Parse() {
	home, _ := homedir.Dir()
	l.cmd = strings.ReplaceAll(l.cmd, "~", home)
}

// execLocal executes local shells
func (g CmdGroup) execLocal() {
	localCmds, uuids := g.buildLocalCmds()

	timeout := viper.Get("Timeout").(time.Duration)

	opts := cmd.Options{Buffered: true, Streaming: true, Timeout: timeout}
	p := cmd.NewCmdOptions(opts, "/bin/bash", "-c", localCmds)
	status := p.Start()

	uuidIndex := 0
	cmdIndex := 0

	for {
		select {
		case so := <-p.Stdout:
			if so == uuids[uuidIndex] {
				if cmdIndex < len(g.Cmds) {
					fmt.Println("$", g.Cmds[cmdIndex].(*LocalCmd).cmd)
				}

				uuidIndex++
				cmdIndex++
			} else {
				fmt.Println(so)
			}
		case se := <-p.Stderr:
			_, _ = fmt.Fprintln(os.Stderr, se)
		case exitState := <-status:
			fmt.Println("exit status ", exitState.Exit)
			return
		}
	}
}

func (g CmdGroup) buildLocalCmds() (localCmdsStr string, uuids []string) {
	uuids = make([]string, 00)
	localCmds := make([]string, 0)

	uuidStr := uuid.New().String()
	uuids = append(uuids, uuidStr)
	localCmds = append(localCmds, "echo "+uuidStr)

	for _, localCmd := range g.Cmds {
		lc := localCmd.(*LocalCmd)
		localCmds = append(localCmds, lc.cmd)

		uuidStr := uuid.New().String()
		uuids = append(uuids, uuidStr)

		localCmds = append(localCmds, "echo "+uuidStr)
	}

	return strings.Join(localCmds, ";"), uuids
}
