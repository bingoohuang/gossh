package hostparse

import (
	"github.com/bingoohuang/gou/mat"
	"github.com/bingoohuang/gou/str"
	"strings"
)

// Host represents the structure of remote host information for ssh.
type Host struct {
	ID       string
	Addr     string
	Port     string
	User     string
	Password string // empty when using public key
	Props    map[string]string
}

// Parse parses the host tmpl configuration.
func Parse(tmpl string) []Host {
	hosts := make([]Host, 0)

	fields := str.FieldsX(tmpl, "(", ")", -1)
	if len(fields) == 0 {
		return hosts
	}

	sc := ServerConfig{}

	f0 := fields[0]
	if IsDirectServer(f0) {
		sc, _ = ParseDirectServer(tmpl)
		sc.Props = ParseProps(fields[1:])
	} else {
		sc.Addr, sc.Port = SplitHostPort(f0)
		sc.User, sc.Pass, _ = Split2BySeps(fields[1], ":", "/")
		sc.Props = ParseProps(fields[2:])
	}

	t := Host{ID: sc.Props["id"], Addr: sc.Addr, Port: sc.Port, User: sc.User, Password: sc.Pass, Props: sc.Props}
	hosts = append(hosts, t.Expands()...)

	return hosts
}

func (c Host) Expands() []Host {
	hosts := str.MakeExpand(c.Addr).MakePart()
	ports := str.MakeExpand(c.Port).MakePart()
	users := str.MakeExpand(c.User).MakePart()
	passes := str.MakeExpand(c.Password).MakePart()
	ids := str.MakeExpand(c.ID).MakePart()
	maxExpands := mat.MaxInt(hosts.Len(), ports.Len(), users.Len(), passes.Len(), ids.Len())

	propsExpands := make(map[string]str.Part)
	for k, v := range c.Props {
		expandV := str.MakeExpand(v)
		propsExpands[k] = expandV.MakePart()
		if l := expandV.MaxLen(); maxExpands < l {
			maxExpands = l
		}
	}

	partPropsFn := func(i int) map[string]string {
		m := make(map[string]string)
		for k, v := range propsExpands {
			m[k] = v.Part(i)
		}
		return m
	}

	tmpls := make([]Host, maxExpands)

	for i := 0; i < maxExpands; i++ {
		tmpls[i] = Host{
			ID:       ids.Part(i),
			Addr:     hosts.Part(i),
			Port:     ports.Part(i),
			User:     users.Part(i),
			Password: passes.Part(i),
			Props:    partPropsFn(i),
		}
	}

	return tmpls
}

func SplitHostPort(addr string) (string, string) {
	if !strings.Contains(addr, ":") {
		return addr, "22"
	}

	pos := strings.Index(addr, ":")

	return addr[0:pos], addr[pos+1:]
}

func ParseProps(fields []string) map[string]string {
	props := make(map[string]string)

	for i := 0; i < len(fields); i++ {
		k, v := str.Split2(fields[i], "=", true, true)
		props[k] = v
	}

	return props
}
