package gossh

import (
	"fmt"
	"path/filepath"
	"strings"

	errs "github.com/pkg/errors"

	"github.com/bingoohuang/gou/file"
	"github.com/bingoohuang/gou/str"

	"github.com/bmatcuk/doublestar"

	homedir "github.com/mitchellh/go-homedir"

	"github.com/sirupsen/logrus"
)

// UlDl scp...
type UlDl struct {
	hosts  Hosts
	remote string
	local  string
}

// UlCmd upload cmd structure.
type UlCmd struct {
	UlDl
	basedir    string
	localFiles []string
}

func (s *UlCmd) init(h *Host) error {
	if len(s.localFiles) > 0 {
		return nil
	}

	s.local = h.SubstituteResultVars(s.local)
	localFiles, err := doublestar.Glob(s.local)
	basedir := file.BaseDir(localFiles)

	if err != nil {
		return errs.Wrapf(err, "doublestar.Glob(%s)", s.local)
	}

	if len(localFiles) == 0 {
		return errs.Wrapf(err, "there is no file matched for pattern %s to upload", s.local)
	}

	if len(localFiles) == 1 {
		basedir = filepath.Dir(localFiles[0])
	}

	s.localFiles = localFiles
	s.basedir = basedir

	return nil
}

// DlCmd download cmd structure.
type DlCmd struct {
	UlDl
}

// TargetHosts returns target hosts for the command.
func (u *UlDl) TargetHosts(hostGroup string) Hosts {
	hosts := make(Hosts, 0, len(u.hosts))

	for _, h := range u.hosts {
		if h.groups[hostGroup] == 1 {
			hosts = append(hosts, h)
		}
	}

	return hosts
}

// nolint:gomnd
func (g *GoSSH) buildUlCmd(hostPart, realCmd string) (HostsCmd, error) {
	fields := str.Fields(realCmd, 2)
	if len(fields) < 2 {
		return nil, fmt.Errorf("bad format for %s", realCmd) // nolint:goerr113
	}

	return &UlCmd{UlDl: UlDl{
		hosts:  g.parseHosts(hostPart),
		local:  strings.ReplaceAll(fields[0], "~", str.PickFirst(homedir.Dir())),
		remote: fields[1],
	}}, nil
}

// nolint:gomnd
func (g *GoSSH) buildDlCmd(hostPart, realCmd string) (HostsCmd, error) {
	fields := str.Fields(realCmd, 2)
	if len(fields) < 2 {
		return nil, fmt.Errorf("bad format for %s", realCmd) // nolint:goerr113
	}

	return &DlCmd{UlDl: UlDl{
		hosts:  g.parseHosts(hostPart),
		local:  strings.ReplaceAll(fields[1], "~", str.PickFirst(homedir.Dir())),
		remote: fields[0],
	}}, nil
}

func (g *GoSSH) parseHosts(hostTag string) Hosts {
	host := hostTag[len(`%host`):]
	if host == "" {
		return g.Hosts
	}

	if host = strings.TrimPrefix(host, "-"); host == "" {
		return g.Hosts
	}

	return g.findHost(host)
}

func (g *GoSSH) findHost(name string) Hosts {
	targetHosts := make(Hosts, 0)
	tm := make(map[string]bool)

	m := make(map[string]*Host)
	for _, h := range g.Hosts {
		m[h.ID] = h
	}

	for _, id := range str.MakeExpand(name).MakeExpand() {
		if _, yes := tm[id]; yes {
			logrus.Warnf("ignored duplicate host ID %s", id)
			continue
		}

		if v, ok := m[id]; ok {
			targetHosts = append(targetHosts, v)
			tm[id] = true
		} else {
			logrus.Warnf("unknown host ID %s", id)
		}
	}

	return targetHosts
}
