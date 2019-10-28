package main

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/bingoohuang/gossh"
	"github.com/bingoohuang/gossh/cnf"
	"github.com/bingoohuang/gossh/pbe"
)

func main() {
	logrus.SetLevel(logrus.InfoLevel)

	pbe.DeclarePflags()
	defer pbe.DealPflag()

	cnf.DeclarePflags()
	cnf.DeclarePflagsByStruct(gossh.Config{})

	if err := cnf.ParsePflags("GOSSH"); err != nil {
		panic(err)
	}

	var config gossh.Config

	cnf.LoadByPflag(&config)

	fmt.Printf("Config%s\n", gossh.JSONPretty(config))

	gs := config.Parse()

	for _, group := range gs.CmdGroups {
		group.Exec()
	}
}
