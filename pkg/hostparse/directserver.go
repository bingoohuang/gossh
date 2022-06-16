package hostparse

import (
	"strings"
)

// IsDirectServer tells that the server is a direct server address like user:pass@host:port.
func IsDirectServer(server string) bool {
	return strings.Index(server, "@") > 0
}

type ServerConfig struct {
	User  string
	Pass  string
	Addr  string
	Port  string
	Props map[string]string
}

// ParseDirectServer parses a direct server address.
func ParseDirectServer(server string) (ServerConfig, bool) {
	// LastIndex of "@" will allow that password contains "@"
	atPos := strings.LastIndex(server, "@")
	sc := ServerConfig{}

	if atPos < 0 {
		return sc, false
	}

	left := server[:atPos]
	right := server[atPos+1:]

	sc.User, sc.Pass, _ = Split2BySeps(left, ":", "/")
	commaPos := strings.Index(right, ":")
	_, sc.Pass = EvalPass(sc.Pass)

	if commaPos == -1 {
		sc.Addr = right
		sc.Port = "22"
	} else {
		sc.Addr = right[:commaPos]
		sc.Port = right[commaPos+1:]
	}

	return sc, true
}

func Split2BySeps(s string, seps ...string) (s1, s2, sep string) {
	sepIndex := make(map[string]int)
	for _, sep := range seps {
		if v := strings.Index(s, sep); v > 0 {
			sepIndex[sep] = v
		}
	}

	minSepIndex := len(s)
	minSep := ""
	for k, v := range sepIndex {
		if v < minSepIndex {
			minSepIndex = v
			minSep = k
		}
	}

	if minSep == "" {
		return s, "", ""
	}

	return s[:minSepIndex], s[minSepIndex+1:], minSep
}
