package hostparse

import (
	"io/ioutil"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
)

func ParseHostFile(hostsFile string) []Host {
	hosts := make([]Host, 0)

	f, _ := homedir.Expand(hostsFile)
	file, err := ioutil.ReadFile(f)
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
