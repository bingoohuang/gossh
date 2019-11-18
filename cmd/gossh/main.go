package main

import (
	"fmt"

	"github.com/spf13/pflag"

	"github.com/sirupsen/logrus"

	"github.com/bingoohuang/gossh"
	"github.com/bingoohuang/gossh/cnf"
	"github.com/bingoohuang/gossh/pbe"
)

func main() {
	logrus.SetLevel(logrus.InfoLevel)

	pbe.DeclarePflags()
	defer pbe.DealPflag()

	ver := pflag.BoolP("version", "v", false, "show version")

	cnf.DeclarePflags()
	cnf.DeclarePflagsByStruct(gossh.Config{})

	if err := cnf.ParsePflags("GOSSH"); err != nil {
		panic(err)
	}

	if *ver {
		fmt.Println("Version: v0.1.1")
		return
	}

	var config gossh.Config

	cnf.LoadByPflag(&config)

	if config.PrintConfig {
		fmt.Printf("Config%s\n", gossh.JSONPretty(config))
	}

	gs := config.Parse()

	for _, group := range gs.CmdGroups {
		group.Exec()
	}
}
