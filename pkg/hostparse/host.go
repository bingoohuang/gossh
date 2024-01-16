package hostparse

import (
	"log"
	"net/url"
	"strings"

	"github.com/bingoohuang/gg/pkg/codec/b64"
	"github.com/bingoohuang/gou/mat"
	"github.com/bingoohuang/gou/str"
)

// Host represents the structure of remote host information for ssh.
type Host struct {
	ID       string
	Addr     string
	Port     string
	User     string
	Password string // empty when using public key
	Props    map[string][]string

	Raw string // register the raw template line like `user:pass@host:port`
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
		sc, _ = ParseDirectServer(f0)
		sc.Props = ParseProps(fields[1:])
	} else {
		sc.Addr, sc.Port = SplitHostPort(f0)
		if len(fields) > 1 {
			eqPos := strings.IndexByte(fields[1], '=')
			upPos := strings.IndexFunc(fields[1], func(r rune) bool {
				return r == '/' || r == ':'
			})

			// 只有 = 在 /: 之前，才认为是 props，否则认为是 user/pass
			if eqPos > 0 && (upPos < 0 || eqPos < upPos) {
				sc.Props = ParseProps(fields[1:])
			} else if upPos > 0 {
				sc.User, sc.Pass = fields[1][:upPos], fields[1][upPos+1:]
			}
		} else {
			sc.Props = ParseProps(fields[1:])
		}
	}

	var passEncodedAlgo string
	passEncodedAlgo, sc.Pass = EvalPass(sc.Pass)

	t := Host{ID: sc.GetProp("id"), Addr: sc.Addr, Port: sc.Port, User: sc.User, Password: sc.Pass, Props: sc.Props}
	hosts = append(hosts, t.Expands(passEncodedAlgo != "")...)

	for i := range hosts {
		hosts[i].Raw = tmpl
	}

	return hosts
}

func EvalPass(pass string) (passEncodedAlgo, evaluated string) {
	if strings.HasPrefix(pass, "{URL}") {
		if p, err := url.QueryUnescape(pass[5:]); err != nil {
			log.Fatalf("failed to url decode %s, error: %v", pass, err)
		} else {
			return "URL", p
		}
	} else if strings.HasPrefix(pass, "{BASE64}") {
		s := strings.TrimRight(pass[8:], "=")
		if p, err := b64.DecodeString(s); err != nil {
			log.Fatalf("failed to url decode %s, error: %v", pass, err)
		} else {
			return "BASE64", p
		}
	}

	return "", pass
}

func (c Host) Expands(passEncodedAlgo bool) []Host {
	hosts := str.MakeExpand(c.Addr).MakePart()
	ports := str.MakeExpand(c.Port).MakePart()
	users := str.MakeExpand(c.User).MakePart()
	var passes str.Part
	if passEncodedAlgo {
		passes = str.MakePart([]string{c.Password})
	} else {
		passes = str.MakeExpand(c.Password).MakePart()
	}
	ids := str.MakeExpand(c.ID).MakePart()
	maxExpands := mat.MaxInt(hosts.Len(), ports.Len(), users.Len(), passes.Len(), ids.Len())

	propsExpands := make(map[string]str.Part)
	for k, v := range c.Props {
		expandV := str.MakeExpand(v[0])
		propsExpands[k] = expandV.MakePart()
		if l := expandV.MaxLen(); maxExpands < l {
			maxExpands = l
		}
	}

	partPropsFn := func(i int) map[string][]string {
		m := make(map[string][]string)
		for k, v := range propsExpands {
			m[k] = append(m[k], v.Part(i))
		}
		return m
	}

	tmpls := make([]Host, maxExpands)

	for i := 0; i < maxExpands; i++ {
		props := partPropsFn(i)
		for k, vv := range props {
			for i, v := range vv {
				vv[i] = SubstituteProps(v, props)
			}
			props[k] = vv
		}

		tmpls[i] = Host{
			ID:       ids.Part(i),
			Addr:     hosts.Part(i),
			Port:     ports.Part(i),
			User:     users.Part(i),
			Password: passes.Part(i),
			Props:    props,
		}
	}

	return tmpls
}

func SubstituteProps(s string, props map[string][]string) string {
	if s == "" {
		return s
	}

	for k, vv := range props {
		v := ""
		if len(vv) > 0 {
			v = vv[len(vv)-1]
		}
		s = strings.ReplaceAll(s, "{"+k+"}", v)
	}

	return s
}

func SplitHostPort(addr string) (string, string) {
	if !strings.Contains(addr, ":") {
		return addr, "22"
	}

	pos := strings.Index(addr, ":")

	return addr[0:pos], addr[pos+1:]
}

func ParseProps(fields []string) map[string][]string {
	props := make(map[string][]string)

	for i := 0; i < len(fields); i++ {
		k, v := str.Split2(fields[i], "=", true, true)
		props[k] = append(props[k], v)
	}

	return props
}
