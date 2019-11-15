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

		addr := fields[0]
		userpass := fields[1]
		id := addr

		if !strings.Contains(addr, ":") {
			addr += ":22"
		} else {
			id = id[0:strings.Index(id, ":")]
		}

		user, pass := str.Split2(userpass, "/", false, false)
		if pass != "" {
			if pass, err = pbe.Ebp(pass); err != nil {
				panic(err)
			}
		}

		props := make(map[string]string)

		if len(fields) > 2 {
			for i := 2; i < len(fields); i++ {
				k, v := str.Split2(fields[i], "=", true, true)
				props[k] = v
			}
		}

		if customID, ok := props["id"]; ok && customID != "" {
			id = customID
		}

		host := &Host{
			ID:         id,
			Addr:       addr,
			User:       user,
			Password:   pass,
			Properties: props,
		}
		hosts = append(hosts, host)
		m[id] = host
	}

	return hosts, m
}
