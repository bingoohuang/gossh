package gossh

import (
	"strings"

	"github.com/bingoohuang/gossh/pbe"

	"github.com/bingoohuang/gou/str"
	"github.com/sirupsen/logrus"
)

func (c Config) parseHosts() []*Host {
	hosts := make([]*Host, 0)

	for _, host := range c.Hosts {
		fields := strings.Fields(host)
		if len(fields) < 2 {
			logrus.Warnf("bad format for host %s", host)
			continue
		}

		id, addr := parseHostID(fields[0])
		user, pass := parseUserPass(fields[1])
		props := parseProps(fields)
		id = fixID(props, id)

		host := &Host{ID: id, Addr: addr, User: user, Password: pass, Properties: props}
		hosts = append(hosts, host)
	}

	return hosts
}

func fixID(props map[string]string, id string) string {
	if customID, ok := props["id"]; ok && customID != "" {
		return customID
	}

	return id
}

func parseProps(fields []string) map[string]string {
	props := make(map[string]string)

	for i := 2; i < len(fields); i++ {
		k, v := str.Split2(fields[i], "=", true, true)
		props[k] = v
	}

	return props
}

func parseUserPass(userpass string) (string, string) {
	user, pass := str.Split2(userpass, "/", false, false)
	if pass != "" {
		var err error
		if pass, err = pbe.Ebp(pass); err != nil {
			panic(err)
		}
	}

	return user, pass
}

func parseHostID(addr string) (string, string) {
	if !strings.Contains(addr, ":") {
		return addr, addr + ":22"
	}

	return addr[0:strings.Index(addr, ":")], addr
}
