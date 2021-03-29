package gossh

import (
	"fmt"
	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"io"
	"os"
	"time"
)

// Repl execute in specified hosts.
func Repl(gs *GoSSH, hosts []*Host, stdout io.Writer) {
	green := color.New(color.FgGreen).SprintfFunc()
	rl, err := readline.NewEx(&readline.Config{
		Prompt:      green(">>> "),
		HistoryFile: "/tmp/gossh-histories",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not create prompt: %v", err)
		os.Exit(1)
	}

	eo := ExecOption{Repl: true}
	if err := repl(gs, hosts, stdout, rl, eo); err != nil {
		fmt.Fprintf(os.Stderr, "could not create prompt: %v", err)
		os.Exit(1)
	}

	defer rl.Close()
}

func repl(gs *GoSSH, hosts []*Host, stdout io.Writer, rl *readline.Instance, eo ExecOption) error {
	lastErrInterrupt := time.Time{}
	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
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

		executeReplCmd(gs, hosts, stdout, line, eo)
	}

	return nil
}

func executeReplCmd(gs *GoSSH, hosts []*Host, w io.Writer, line string, eo ExecOption) {
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
		if err := ExecInHosts(gs, host, cmd, w, eo); err != nil {
			fmt.Fprintf(w, "ExecInHosts error %v\n", err)
		}
	}
}
