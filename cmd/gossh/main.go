package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/bingoohuang/gg/pkg/v"
	"github.com/bingoohuang/gossh"
	"github.com/bingoohuang/gossh/pkg/cnf"
	"github.com/bingoohuang/gou/enc"
	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	logrus.SetLevel(logrus.InfoLevel)

	DeclarePbePflags()

	var ssh sshpassHelp

	ver := pflag.BoolP("version", "v", false, "show version")
	repl := pflag.BoolP("repl", "", false, "repl mode")
	tag := pflag.StringP("tag", "t", "", "command prefix tag")

	ssh.declarePlags()
	cnf.DeclarePflags()
	cnf.DeclarePflagsByStruct(gossh.Config{})

	if err := cnf.ParsePflags("GOSSH"); err != nil {
		panic(err)
	}

	if *ver {
		fmt.Println(v.Version())
		return
	}

	if DealPbePflag() {
		return
	}

	var config gossh.Config
	LoadByPflag(*tag, &config)

	if config.Group == "" {
		config.Group = "default"
	}

	if config.PrintConfig {
		fmt.Printf("Config%s\n", enc.JSONPretty(config))
	}

	gs := config.Parse()
	ssh.do(gs)

	if len(gs.Cmds) == 0 {
		*repl = true
	}

	logsDir, _ := homedir.Expand("~/.gossh/logs/")
	_ = os.MkdirAll(logsDir, os.ModePerm)
	cnfFile := filepath.Base(viper.GetString("cnf"))

	if cnfFile != "" {
		cnfFile += "-"
	}

	logFn := filepath.Join(logsDir, cnfFile+time.Now().Format("20060102150304")+".log")
	logFile, err := os.Create(logFn)

	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create log file %s, error:%v\n", logFn, err)
	} else {
		fmt.Fprintf(os.Stdout, "log file %s created\n", logFn)
		fmt.Fprintf(logFile, "started at %s\n", time.Now().UTC().Format("2006-01-02 15:03:04"))
	}

	start := time.Now()

	var stdout io.Writer = os.Stdout

	if logFile != nil {
		stdout = io.MultiWriter(os.Stdout, logFile)

		defer func() {
			fmt.Fprintf(logFile, "finished at %s\n", time.Now().UTC().Format("2006-01-02 15:03:04"))
			fmt.Fprintf(logFile, "cost %s\n", time.Since(start))
			fmt.Fprintf(os.Stdout, "log file %s recorded\n", logFn)

			logFile.Close()
		}()
	}

	eo := gossh.ExecOption{}
	switch gs.Config.ExecMode {
	case gossh.ExecModeCmdByCmd:
		gossh.ExecCmds(&gs, gossh.NewExecModeCmdByCmd(), stdout, eo, config.Group)
	case gossh.ExecModeHostByHost:
		hosts := append([]*gossh.Host{gossh.LocalHost}, gs.Hosts...)
		for _, host := range hosts {
			gossh.ExecCmds(&gs, host, stdout, eo, config.Group)
		}
	}

	if *repl {
		gs.Config.GlobalRemote = true
		gossh.Repl(&gs, gs.Hosts, stdout, config.Group)
	}

	_ = gs.Close()
}
