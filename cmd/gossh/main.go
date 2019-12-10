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
		fmt.Println("Version: v0.2.1")
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

	if len(gs.CmdGroups) == 0 {
		fmt.Println("There is nothing to do.")
	}

	for _, group := range gs.CmdGroups {
		group.Exec()
	}
}
