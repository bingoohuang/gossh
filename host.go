package gossh

import (
	"strings"

	"github.com/bingoohuang/gossh/pbe"

	"github.com/bingoohuang/gou/str"
	"github.com/sirupsen/logrus"
)

func (c Config) parseHosts() ([]*Host, map[string]*Host) {
	m := make(map[string]*Host)
	hosts := make([]*Host, 0)
	var err error

	for _, host := range c.Hosts {
		fields := strings.Fields(host)
		if len(fields) < 3 {
			logrus.Warnf("bad format for host %s", host)
			continue
		}

		name := fields[0]
		addr := fields[1]
		userpass := fields[2]

		if !strings.Contains(addr, ":") {
			addr += ":22"
		}

		user, pass := str.Split2(userpass, "/", false, false)
		if pass != "" {
			if pass, err = pbe.PbeDecrypt(pass); err != nil {
				panic(err)
			}
		}

		props := make(map[string]string)

		if len(fields) > 3 {
			props = str.SplitToMap(fields[3], "=", ",")
		}

		host := &Host{
			Name:       name,
			Addr:       addr,
			User:       user,
			Password:   pass,
			Properties: props,
		}
		hosts = append(hosts, host)
		m[name] = host
	}

	return hosts, m
}
