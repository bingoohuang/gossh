package hostparse

import (
	"os"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
)

func ParseHostFile(hostsFile string) []Host {
	hosts := make([]Host, 0)

	f, _ := homedir.Expand(hostsFile)
	file, err := os.ReadFile(f)
	if err != nil {
		logrus.Warnf("failed to read hosts file %s: %v", hostsFile, err)
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
