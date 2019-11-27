package gossh

import (
	"path/filepath"
	"strings"

	"github.com/bingoohuang/gossh/doublestar"

	homedir "github.com/mitchellh/go-homedir"

	"github.com/bingoohuang/gossh/elf"
	"github.com/sirupsen/logrus"
)

// UlDl scp...
type UlDl struct {
	hosts  Hosts
	cmd    string
	remote string
	local  string
}

// UlCmd upload cmd structure.
type UlCmd struct {
	UlDl
	basedir    string
	localFiles []string
}

// DlCmd download cmd structure.
type DlCmd struct {
	UlDl
}

// Parse parses UlCmd.
func (UlDl) Parse() {}

// TargetHosts returns target hosts for the command
func (u UlDl) TargetHosts() Hosts { return u.hosts }

// RawCmd returns the original raw command
func (u UlDl) RawCmd() string { return u.cmd }

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
	localFiles, basedir, err := doublestar.Glob(local)

	if err != nil {
		logrus.Fatalf("doublestar.Glob(%s) error %v", local, err)
	}

	if len(localFiles) == 0 {
		logrus.Fatalf("there is no file matched for pattern %s to upload", local)
	}

	if len(localFiles) == 1 {
		basedir = filepath.Dir(localFiles[0])
	}

	hosts := parseHosts(gs, hostPart)

	return &UlCmd{
		UlDl:       UlDl{hosts: hosts, cmd: cmd, local: local, remote: remote},
		localFiles: localFiles,
		basedir:    basedir,
	}
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

	hosts := parseHosts(gs, hostPart)

	return &DlCmd{UlDl: UlDl{hosts: hosts, cmd: cmd, local: local, remote: remote}}
}

func parseHosts(gs *GoSSH, hostTag string) Hosts {
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

	return Hosts{found}
}

func findHost(hosts Hosts, name string) *Host {
	for _, h := range hosts {
		if h.ID == name {
			return h
		}
	}

	return nil
}
