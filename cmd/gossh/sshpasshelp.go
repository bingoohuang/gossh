package main

import (
	"os"

	"github.com/bingoohuang/gossh"
	"github.com/spf13/pflag"
)

type sshpassHelp struct {
	sshPlag *bool
	scpPlag *bool
}

func (s *sshpassHelp) declarePlags() {
	s.sshPlag = pflag.BoolP("ssh", "", false, "create sshpassHelp ssh for hosts")
	s.scpPlag = pflag.BoolP("scp", "s", false, "create sshpassHelp scp for hosts")
}

func (s sshpassHelp) do(gs gossh.GoSSH) {
	if !*s.sshPlag && !*s.scpPlag {
		return
	}

	if *s.sshPlag {
		gs.Hosts.PrintSSH()
	}

	if *s.scpPlag {
		gs.Hosts.PrintSCP()
	}

	os.Exit(0)
}
