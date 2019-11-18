package gossh

import (
	"strings"

	homedir "github.com/mitchellh/go-homedir"

	"github.com/bingoohuang/gossh/elf"
	"github.com/sirupsen/logrus"
)

// UlDl scp...
type UlDl struct {
	hosts        []*Host
	cmd          string
	remote       string
	local        string
	localDirMode elf.DirMode
}

// UlCmd upload cmd structure.
type UlCmd struct {
	UlDl
}

// DlCmd download cmd structure.
type DlCmd struct {
	UlDl
}

// Parse parses UlCmd.
func (UlDl) Parse() {}

func buildUlCmd(gs *GoSSH, hostPart, realCmd, cmd string) *UlCmd {
	fields := elf.Fields(realCmd, 2)
	if len(fields) < 2 {
		logrus.Warnf("bad format for %s", cmd)
		return nil
	}

	local := fields[0]
	remote := fields[1]
	home, _ := homedir.Dir()

	local = strings.ReplaceAll(local, "~", home)
	dirMode, _ := elf.GetFileMode(local)

	return &UlCmd{UlDl{hosts: parseHosts(gs, hostPart),
		cmd: cmd, local: local, localDirMode: dirMode, remote: remote}}
}

func buildDlCmd(gs *GoSSH, hostPart, realCmd, cmd string) *DlCmd {
	fields := elf.Fields(realCmd, 2)
	if len(fields) < 2 {
		logrus.Warnf("bad format for %s", cmd)
		return nil
	}

	remote := fields[0]
	local := fields[1]
	home, _ := homedir.Dir()

	local = strings.ReplaceAll(local, "~", home)
	dirMode, _ := elf.GetFileMode(local)

	return &DlCmd{UlDl{hosts: parseHosts(gs, hostPart),
		cmd: cmd, local: local, localDirMode: dirMode, remote: remote}}
}

func parseHosts(gs *GoSSH, hostTag string) []*Host {
	host := hostTag[len(`%host`):]

	if host == "" {
		return gs.Hosts
	}

	host = strings.TrimPrefix(host, "(")
	host = strings.TrimPrefix(host, "-")
	host = strings.TrimSuffix(host, ")")

	found := findHost(gs.Hosts, host)
	if found == nil {
		return nil
	}

	return []*Host{found}
}
