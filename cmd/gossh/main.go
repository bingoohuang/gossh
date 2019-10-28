package main

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/bingoohuang/gossh"
	"github.com/bingoohuang/gossh/cnf"
	"github.com/bingoohuang/gossh/pbe"
)

const (
	timeout = 10 * time.Second
)

func main() {
	logrus.SetLevel(logrus.InfoLevel)

	pbe.DeclarePflags()
	defer pbe.DealPflag()

	cnf.DeclarePflags()
	cnf.DeclarePflagsByStruct(gossh.Config{})
	cnf.ParsePflags("GOSSH")

	var config gossh.Config
	cnf.LoadByPflag(&config)

	fmt.Printf("Config%s\n", gossh.JSONPretty(config))

	gs := config.Parse()
	for _, group := range gs.CmdGroups {
		group.Exec()
	}

}
