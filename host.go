package gossh

import (
	"fmt"
	"strings"

	"github.com/bingoohuang/gossh/elf"

	"github.com/bingoohuang/gossh/pbe"

	"github.com/bingoohuang/gou/str"
	"github.com/sirupsen/logrus"
)

func (c Config) parseHosts() Hosts {
	hosts := make(Hosts, 0)

	for _, host := range c.Hosts {
		fields := elf.FieldsX(host, "(", ")", -1)
		if len(fields) < 2 {
			logrus.Warnf("bad format for host %s", host)
			continue
		}

		_, addr := parseHostID(fields[0])
		user, pass := parseUserPass(fields[1])
		props := parseProps(fields)
		id := fixID(props, "")

		host := &Host{ID: id, Addr: addr, User: user, Password: pass, Properties: props}
		expanded := expandHost(host, len(hosts))
		hosts = append(hosts, expanded...)
	}

	return hosts
}

func expandHost(host *Host, hostsLen int) Hosts {
	ids := MakeExpand(host.ID).MakePart()
	addrs := MakeExpand(host.Addr).MakePart()
	users := MakeExpand(host.User).MakePart()
	passes := MakeExpand(host.Password).MakePart()

	maxExpands := elf.MaxInt(ids.Len(), addrs.Len(), users.Len(), passes.Len())

	expandedProps := make(map[string]Part)

	for k, v := range host.Properties {
		vv := MakeExpand(v).MakePart()
		expandedProps[k] = vv
		maxExpands = elf.MaxInt(maxExpands, vv.Len())
	}

	hosts := make(Hosts, maxExpands)

	for i := 0; i < maxExpands; i++ {
		props := make(map[string]string)

		for k, v := range expandedProps {
			props[k] = v.Part(i)
		}

		id := ids.Part(i)
		if id == "" {
			id = fmt.Sprintf("%d", i+hostsLen+1)
		}

		hosts[i] = &Host{
			ID:         id,
			Addr:       addrs.Part(i),
			User:       users.Part(i),
			Password:   passes.Part(i),
			Properties: props}
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
