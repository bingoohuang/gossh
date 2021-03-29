package gossh

import (
	"github.com/bingoohuang/gg/pkg/ss"
	"github.com/bingoohuang/gossh/pkg/hostparse"
	"net"
)

func (c Config) parseHosts() Hosts {
	hosts := make(Hosts, 0)

	for _, host := range c.Hosts {
		hosts = append(hosts, c.parseHost(host)...)
	}
	if c.HostsFile != "" {
		hosts = append(hosts, c.parseHostFile()...)
	}

	hosts.FixHost()
	hosts.FixProxy()
	return hosts
}

func (c Config) parseHost(host string) Hosts {
	return convertHosts(hostparse.Parse(host))
}
func (c Config) parseHostFile() Hosts {
	return convertHosts(hostparse.ParseHostFile(c.HostsFile))
}

func convertHosts(parsed []hostparse.Host) Hosts {
	hosts := make(Hosts, len(parsed))
	for i, p := range parsed {
		addr := net.JoinHostPort(p.Addr, ss.Or(p.Port, "22"))
		hosts[i] = &Host{ID: p.ID, Addr: addr, User: p.User, Password: p.Password, Properties: p.Props}
	}
	return hosts
}
