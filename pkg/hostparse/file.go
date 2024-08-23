package hostparse

import (
	"log"
	"os"
	"strings"

	"github.com/bingoohuang/ngg/ss"
)

func ParseHostFile(hostsFile string) []Host {
	hosts := make([]Host, 0)

	f := ss.ExpandHome(hostsFile)
	file, err := os.ReadFile(f)
	if err != nil {
		log.Printf("W! failed to read hosts file %s: %v", hostsFile, err)
		return nil
	}

	for _, line := range strings.Split(string(file), "\n") {
		hostLine := strings.TrimSpace(line)
		if hostLine != "" && !strings.HasPrefix(hostLine, "#") {
			hosts = append(hosts, Parse(hostLine)...)
		}
	}

	return hosts
}
