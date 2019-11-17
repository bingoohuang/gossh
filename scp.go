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

func buildUlCmd(gs *GoSSH, fields []string, cmd string) *UlCmd {
	if len(fields) < 4 {
		logrus.Warnf("bad format for %s", cmd)
		return nil
	}

	local := fields[2]
	remote := fields[3]
	home, _ := homedir.Dir()

	local = strings.ReplaceAll(local, "~", home)
	dirMode, _ := elf.GetFileMode(local)

	return &UlCmd{UlDl{hosts: parseHosts(gs, fields[0]),
		cmd: cmd, local: local, localDirMode: dirMode, remote: remote}}
}

func buildDlCmd(gs *GoSSH, fields []string, cmd string) *DlCmd {
	if len(fields) < 4 {
		logrus.Warnf("bad format for %s", cmd)
		return nil
	}

	local := fields[3]
	remote := fields[2]
	home, _ := homedir.Dir()

	local = strings.ReplaceAll(local, "~", home)
	dirMode, _ := elf.GetFileMode(local)

	return &DlCmd{UlDl{hosts: parseHosts(gs, fields[0]),
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
