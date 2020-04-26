package main

import (
	"fmt"

	"github.com/bingoohuang/gou/enc"

	"github.com/spf13/pflag"

	"github.com/sirupsen/logrus"

	"github.com/bingoohuang/gossh"
	"github.com/bingoohuang/gou/cnf"
	"github.com/bingoohuang/gou/pbe"
)

func main() {
	logrus.SetLevel(logrus.InfoLevel)

	pbe.DeclarePflags()

	var ssh sshpassHelp

	ver := pflag.BoolP("version", "v", false, "show version")

	ssh.declarePlags()
	cnf.DeclarePflags()
	cnf.DeclarePflagsByStruct(gossh.Config{})

	if err := cnf.ParsePflags("GOSSH"); err != nil {
		panic(err)
	}

	if *ver {
		fmt.Println("Version: v1.0.0")
		return
	}

	var config gossh.Config

	cnf.LoadByPflag(&config)

	if config.PrintConfig {
		fmt.Printf("Config%s\n", enc.JSONPretty(config))
	}

	gs := config.Parse()

	if pbe.DealPflag() {
		return
	}

	ssh.do(gs)

	if len(gs.Cmds) == 0 {
		fmt.Println("There is nothing to do.")
	}

	hosts := append([]*gossh.Host{gossh.LocalHost}, gs.Hosts...)

	switch gs.Vars.ExecMode {
	case gossh.ExecModeCmdByCmd:
		execCmds(gs, nil)
	case gossh.ExecModeHostByHost:
		for _, host := range hosts {
			execCmds(gs, host)
		}
	}

	_ = gs.Close()
}

func execCmds(gs gossh.GoSSH, host *gossh.Host) {
	for _, cmd := range gs.Cmds {
		if err := gossh.ExecInHosts(&gs, host, cmd); err != nil {
			gs.LogPrintf("ExecInHosts error %v\n", err)
		}
	}
}
