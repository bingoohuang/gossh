package hostparse

import (
	"log"
	"net/url"
	"strings"

	"github.com/bingoohuang/ngg/ss"
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

	note := ""
	if noteIndex := strings.Index(tmpl, "#"); noteIndex > 0 {
		note = strings.TrimSpace(tmpl[noteIndex+1:])
		tmpl = strings.TrimSpace(tmpl[:noteIndex])
	}

	fields := ss.FieldsX(tmpl, "(", ")", -1)
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
			if upPos < 0 && eqPos < 0 {
				sc.User = fields[1]
			} else if eqPos > 0 && (upPos < 0 || eqPos < upPos) {
				// 只有 = 在 /: 之前，才认为是 props，否则认为是 user/pass
				sc.Props = ParseProps(fields[1:])
			} else if upPos > 0 {
				sc.User, sc.Pass = fields[1][:upPos], fields[1][upPos+1:]
				sc.Props = ParseProps(fields[2:])
			}
		} else {
			sc.Props = ParseProps(fields[1:])
		}
	}

	var passEncodedAlgo string
	passEncodedAlgo, sc.Pass = EvalPass(sc.Pass)

	if note != "" {
		if len(sc.Props["note"]) > 0 {
			sc.Props["note"][0] = sc.Props["note"][0] + " " + note
		} else {
			sc.Props["note"] = []string{note}
		}
	}

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
		if p, err := ss.Base64().Decode(s); err != nil {
			log.Fatalf("failed to url decode %s, error: %v", pass, err)
		} else {
			return "BASE64", p.String()
		}
	}

	return "", pass
}

func (c Host) Expands(passEncodedAlgo bool) []Host {
	hosts := ss.MakeExpand(c.Addr).MakePart()
	ports := ss.MakeExpand(c.Port).MakePart()
	users := ss.MakeExpand(c.User).MakePart()
	var passes ss.ExpandPart
	if passEncodedAlgo {
		passes = ss.MakePart([]string{c.Password})
	} else {
		passes = ss.MakeExpand(c.Password).MakePart()
	}
	ids := ss.MakeExpand(c.ID).MakePart()
	maxExpands := max(hosts.Len(), ports.Len(), users.Len(), passes.Len(), ids.Len())

	propsExpands := make(map[string]ss.ExpandPart)
	for k, v := range c.Props {
		expandV := ss.MakeExpand(v[0])
		propsExpands[k] = expandV.MakePart()
		if l := expandV.MaxLen(); maxExpands < l {
			maxExpands = l
		}
	}

	partPropsFn := func(i int) map[string][]string {
		m := make(map[string][]string)
		for k, v := range propsExpands {
			m[k] = append(m[k], v.ExpandPart(i))
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
			ID:       ids.ExpandPart(i),
			Addr:     hosts.ExpandPart(i),
			Port:     ports.ExpandPart(i),
			User:     users.ExpandPart(i),
			Password: passes.ExpandPart(i),
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
		k, v := ss.Split2(fields[i], "=")
		props[k] = append(props[k], v)
	}

	return props
}
