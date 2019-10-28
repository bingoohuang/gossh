package gossh

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gobars/cmd"
	"github.com/google/uuid"
	"github.com/mitchellh/go-homedir"
)

type LocalCmd struct {
	cmd string
}

func (l *LocalCmd) ExecInHosts(_ []*Host) error {
	return nil
}

// Parse parses the local cmd
func (l *LocalCmd) Parse() {
	home, _ := homedir.Dir()
	l.cmd = strings.ReplaceAll(l.cmd, "~", home)
}

// execLocal executes local shells
func (g CmdGroup) execLocal() {
	localCmds, uuids := g.buildLocalCmds()

	opts := cmd.Options{Buffered: true, Streaming: true, Timeout: 10 * time.Second}
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
