package gossh

import (
	"io/ioutil"
	"strings"

	homedir "github.com/mitchellh/go-homedir"

	"github.com/sirupsen/logrus"

	"github.com/bingoohuang/gou/mat"
	"github.com/bingoohuang/gou/pbe"

	"github.com/bingoohuang/gou/str"
)

func (c Config) parseHosts() Hosts {
	hosts := make(Hosts, 0)

	for _, host := range c.Hosts {
		hosts = append(hosts, c.parseHost(host)...)
	}

	if c.HostsFile != "" {
		hosts = append(hosts, c.parseHostFile()...)
	}

	hosts.FixHostID()
	hosts.FixProxy()

	return hosts
}

func (c Config) parseHostFile() Hosts {
	hosts := make([]*Host, 0)

	hostsFile, _ := homedir.Expand(c.HostsFile)
	file, err := ioutil.ReadFile(hostsFile)

	if err != nil {
		logrus.Warnf("failed to read hosts file %s: %v", c.HostsFile, err)
		return hosts
	}

	for _, line := range strings.Split(string(file), "\n") {
		hostLine := strings.TrimSpace(line)
		if hostLine != "" && !strings.HasPrefix(hostLine, "#") {
			hosts = append(hosts, c.parseHost(hostLine)...)
		}
	}

	return hosts
}

func (c Config) parseHost(host string) Hosts {
	host = strings.TrimSpace(host)
	if host == "" {
		return Hosts{}
	}

	fields := str.FieldsX(host, "(", ")", -1)
	//if len(fields) < 2 && {
	//	logrus.Warnf("bad format for host %s", host)
	//	continue
	//}

	addr := parseHostID(fields[0])
	user, pass := parseUserPass(fields, 1)
	props := parseProps(fields)
	id := fixID(props)

	return c.expandHost(&Host{ID: id, Addr: addr, User: user, Password: pass, Properties: props})
}

func (c Config) expandHost(host *Host) Hosts {
	ids := str.MakeExpand(host.ID).MakePart()
	addrs := str.MakeExpand(host.Addr).MakePart()
	users := str.MakeExpand(host.User).MakePart()
	passes := str.MakeExpand(host.Password).MakePart()
	maxExpands := mat.MaxInt(ids.Len(), addrs.Len(), users.Len(), passes.Len())
	expandedProps := make(map[string]str.Part)

	for k, v := range host.Properties {
		vv := str.MakeExpand(v).MakePart()
		expandedProps[k] = vv
		maxExpands = mat.MaxInt(maxExpands, vv.Len())
	}

	hosts := make(Hosts, maxExpands)

	for i := 0; i < maxExpands; i++ {
		props := make(map[string]string)

		for k, v := range expandedProps {
			props[k] = v.Part(i)
		}

		hosts[i] = &Host{
			ID: ids.Part(i), Addr: addrs.Part(i),
			User: users.Part(i), Password: passes.Part(i),
			Properties: props}

		if hosts[i].User == "" {
			hosts[i].User = c.User
		}

		if hosts[i].Password == "" {
			hosts[i].Password = c.Pass
		}
	}

	return hosts
}

func fixID(props map[string]string) string {
	if v, ok := props["id"]; ok {
		return v
	}

	return ""
}

func parseProps(fields []string) map[string]string {
	props := make(map[string]string)

	for i := 2; i < len(fields); i++ {
		k, v := str.Split2(fields[i], "=", true, true)
		props[k] = v
	}

	return props
}

func parseUserPass(fields []string, index int) (string, string) {
	if index >= len(fields) {
		return "", ""
	}

	userpass := fields[index]
	user, pass := str.Split2(userpass, "/", false, false)

	if pass != "" {
		var err error
		if pass, err = pbe.Ebp(pass); err != nil {
			panic(err)
		}
	}

	return user, pass
}

func parseHostID(addr string) string {
	if strings.Contains(addr, ":") {
		return addr
	}

	return addr + ":22"
}
