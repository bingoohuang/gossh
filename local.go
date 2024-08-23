package gossh

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/bingoohuang/gossh/pkg/cmdtype"
	"github.com/bingoohuang/ngg/ss"
	"github.com/gobars/cmd"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

// LocalCmd means local commands.
type LocalCmd struct {
	cmd       string
	resultVar string
}

func (g *GoSSH) buildLocalCmd(cmd string) HostsCmd {
	cmd = ss.ExpandHome(cmd)
	c, v := cmdtype.ParseResultVar(cmd)
	l := &LocalCmd{cmd: c, resultVar: v}

	return l
}

// LocalHost means the local host.
// nolint:gochecknoglobals
var LocalHost = &Host{ID: "localhost", Addr: "localhost", resultVars: make(map[string]string)}

// TargetHosts returns target hosts for the command.
func (LocalCmd) TargetHosts(string) Hosts { return []*Host{LocalHost} }

// RawCmd returns the original raw command.
func (l LocalCmd) RawCmd() string { return l.cmd }

// Exec execute in specified host.
func (l *LocalCmd) Exec(_ *GoSSH, h *Host, stdout io.Writer, _ ExecOption) error {
	localCmd, uuidStr := l.buildLocalCmd(h)
	timeout := viper.Get("CmdTimeout").(time.Duration)
	opts := cmd.Options{Buffered: true, Streaming: true, Timeout: timeout}
	echoCmd := h.SubstituteResultVars(l.cmd)

	p := cmd.NewCmdOptions(opts, "/bin/bash", "-c", localCmd)
	status := p.Start()
	uuidTimes := 0

	for {
		select {
		case so := <-p.Stdout:
			if so == uuidStr {
				if uuidTimes == 0 {
					_, _ = fmt.Fprintln(stdout, "$", echoCmd)
				}

				uuidTimes++
			} else {
				if uuidTimes == 2 {
					pwd, _ := os.Getwd()
					if pwd != so {
						_ = os.Chdir(so)
					}
				} else {
					h.SetResultVar(l.resultVar, so)
					_, _ = fmt.Fprintln(stdout, so)
				}
			}
		case se := <-p.Stderr:
			_, _ = fmt.Fprintln(stdout, se)
		case exitState := <-status:
			if exitState.Exit != 0 {
				_, _ = fmt.Fprintln(stdout, "exit status ", exitState.Exit)
			}
			return nil
		}
	}
}

// buildLocalCmd  把当前命令进行封装，为了更好地获得命令的输出，前后添加uuid的echo，并且最后打印当前目录，为了切换.
func (l *LocalCmd) buildLocalCmd(h *Host) (localCmdsStr string, uuidStr string) {
	uuidStr = uuid.New().String()

	return "echo " + uuidStr + ";" +
		h.SubstituteResultVars(l.cmd) +
		"; echo " + uuidStr + ";pwd", uuidStr
}
