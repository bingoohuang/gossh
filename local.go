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

// LocalHost means the local host.
// nolint gochecknoglobals
var LocalHost = &Host{ID: "localhost", Addr: "localhost"}

// TargetHosts returns target hosts for the command
func (LocalCmd) TargetHosts() Hosts { return []*Host{LocalHost} }

// RawCmd returns the original raw command
func (l LocalCmd) RawCmd() string { return l.cmd }

// Parse parses the local cmd
func (l *LocalCmd) Parse() {
	home, _ := homedir.Dir()
	l.cmd = strings.ReplaceAll(l.cmd, "~", home)
}

// Exec execute in specified host.
func (l *LocalCmd) Exec(_ *GoSSH, _ *Host) error {
	localCmd, uuidStr := l.buildLocalCmd()
	timeout := viper.Get("CmdTimeout").(time.Duration)
	opts := cmd.Options{Buffered: true, Streaming: true, Timeout: timeout}
	p := cmd.NewCmdOptions(opts, "/bin/bash", "-c", localCmd)
	status := p.Start()
	uuidTimes := 0

	for {
		select {
		case so := <-p.Stdout:
			if so == uuidStr {
				if uuidTimes == 0 {
					fmt.Println("$", l.cmd)
				}

				uuidTimes++
			} else {
				if uuidTimes == 2 { // nolint gomnd
					pwd, _ := os.Getwd()
					if pwd != so {
						_ = os.Chdir(so)
					}
				} else {
					fmt.Println(so)
				}
			}
		case se := <-p.Stderr:
			_, _ = fmt.Fprintln(os.Stderr, se)
		case exitState := <-status:
			fmt.Println("exit status ", exitState.Exit)
			return nil
		}
	}
}

// buildLocalCmd  把当前命令进行封装，为了更好地获得命令的输出，前后添加uuid的echo，并且最后打印当前目录，为了切换。
func (l *LocalCmd) buildLocalCmd() (localCmdsStr string, uuidStr string) {
	uuidStr = uuid.New().String()

	return "echo " + uuidStr + ";" + l.cmd + "; echo " + uuidStr + ";pwd", uuidStr
}
