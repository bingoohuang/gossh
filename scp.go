package gossh

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/bingoohuang/ngg/ss"
	"github.com/bmatcuk/doublestar"
	errs "github.com/pkg/errors"
)

// UlDl scp...
type UlDl struct {
	remote string
	local  string
	hosts  Hosts
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
	basedir := ss.CommonDir(localFiles)

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

func (g *GoSSH) buildUlCmd(hostPart, realCmd string) (HostsCmd, error) {
	fields := ss.Fields(realCmd, 2)
	if len(fields) < 2 {
		return nil, fmt.Errorf("bad format for %s", realCmd) // nolint:goerr113
	}

	return &UlCmd{UlDl: UlDl{
		hosts:  g.parseHosts(hostPart),
		local:  ss.ExpandHome(fields[0]),
		remote: fields[1],
	}}, nil
}

func (g *GoSSH) buildDlCmd(hostPart, realCmd string) (HostsCmd, error) {
	fields := ss.Fields(realCmd, 2)
	if len(fields) < 2 {
		return nil, fmt.Errorf("bad format for %s", realCmd) // nolint:goerr113
	}

	return &DlCmd{UlDl: UlDl{
		hosts:  g.parseHosts(hostPart),
		local:  ss.ExpandHome(fields[1]),
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

	for _, id := range ss.MakeExpand(name).MakeExpand() {
		if _, yes := tm[id]; yes {
			log.Printf("W! ignored duplicate host ID %s", id)
			continue
		}

		if v, ok := m[id]; ok {
			targetHosts = append(targetHosts, v)
			tm[id] = true
		} else {
			log.Printf("W! unknown host ID %s", id)
		}
	}

	return targetHosts
}
