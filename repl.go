package gossh

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/pkg/errors"
)

// Repl execute in specified hosts.
func Repl(gs *GoSSH, hosts []*Host, stdout io.Writer, hostGroup string) {
	green := color.New(color.FgGreen).SprintfFunc()
	rl, err := readline.NewEx(&readline.Config{
		Prompt:      green(">>> "),
		HistoryFile: "/tmp/gossh-histories",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not create prompt: %v", err)
		os.Exit(1)
	}

	if err := repl(gs, hosts, stdout, rl, (ExecOption{Repl: true}), hostGroup); err != nil {
		fmt.Fprintf(os.Stderr, "could not create prompt: %v", err)
		os.Exit(1)
	}

	defer rl.Close()
}

func repl(gs *GoSSH, hosts []*Host, stdout io.Writer, rl *readline.Instance, eo ExecOption, hostGroup string) error {
	lastErrInterrupt := time.Time{}
	hosts = append(hosts, LocalHost)
	for {
		line, err := rl.Readline()
		if errors.Is(err, readline.ErrInterrupt) {
			if lastErrInterrupt.IsZero() {
				lastErrInterrupt = time.Now()
				continue
			}

			if time.Since(lastErrInterrupt) < 5*time.Second {
				return nil
			}

			lastErrInterrupt = time.Now()
			continue
		}

		if err != nil {
			return err
		}

		lastErrInterrupt = time.Time{}
		if len(line) == 0 {
			continue
		}

		if line == "exit" || line == "quit" {
			return nil
		}

		executeReplCmd(gs, hosts, stdout, line, eo, hostGroup)
	}
}

func executeReplCmd(gs *GoSSH, hosts []*Host, w io.Writer, line string, eo ExecOption, hostGroup string) {
	if line == "%hosts" {
		for _, h := range hosts {
			fmt.Fprintf(w, "ID:%s addr:%s note:%s\n", h.ID, h.Addr, h.Prop("note"))
		}
		return
	}

	cmd, err := gs.Config.parseCmd(gs, line)
	if err != nil {
		fmt.Fprintf(w, "failed to parse cmd: %s, error: %v", line, err)
		return
	}

	if cmd == nil {
		return
	}

	for _, host := range hosts {
		if err := ExecInHosts(gs, host, cmd, w, eo, hostGroup); err != nil {
			fmt.Fprintf(w, "ExecInHosts error %v\n", err)
		}
	}
}
