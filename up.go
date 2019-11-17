package gossh

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/bingoohuang/gossh/elf"
	"github.com/sirupsen/logrus"
)

// ExecInHosts executes uploading among hosts.
func (s *UlCmd) ExecInHosts(gs *GoSSH) error {
	switch s.localDirMode {
	case elf.SingleFileMode:
		s.singleSCP(gs)
	case elf.DirectoryMode:
		err2 := s.upSCP(gs)
		if err2 != nil {
			return err2
		}
	}

	return nil
}

func (s *UlCmd) upSCP(gs *GoSSH) error {
	if err := s.scpRecursively(s.remote, gs); err != nil {
		return err
	}

	return nil
}

func (s *UlCmd) scpRecursively(destBase string, gs *GoSSH) error {
	if err := filepath.Walk(s.local,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				destPath := strings.TrimPrefix(path, s.local)
				dest := filepath.Join(destBase, destPath)
				uploadFile(gs, path, dest)
			}

			return nil
		}); err != nil {
		logrus.Warnf(" filepath.Walk %s error %v", s.local, err)

		return err
	}

	return nil
}

func (s *UlCmd) singleSCP(gs *GoSSH) {
	baseFrom := filepath.Base(s.local)
	dest := s.remote
	baseDest := filepath.Base(dest)

	if baseDest != baseFrom {
		dest = filepath.Join(dest, baseFrom)
	}

	uploadFile(gs, s.local, dest)
}

func uploadFile(gs *GoSSH, src, dest string) {
	hostName := ""

	if submatchIndex := regexp.MustCompile(`%host(-\w+)?:`).
		FindStringSubmatchIndex(dest); len(submatchIndex) > 0 {
		if submatchIndex[2] > 0 {
			hostName = dest[submatchIndex[2]:submatchIndex[3]]
		}

		dest = dest[submatchIndex[1]:]
	}

	targetHosts := filterHosts(hostName, gs)
	if len(targetHosts) == 0 {
		logrus.Warnf("there is no host to upload %s", src)
	}

	fmt.Println("start to scp upload ", src, "to", dest, "on hosts", filterHostnames(targetHosts))

	var wg sync.WaitGroup

	wg.Add(len(targetHosts))

	for _, host := range targetHosts {
		go func(h Host, from, to string) {
			if err := upload(gs, h, from, to); err != nil {
				logrus.Warnf(" upload %s error %v", from, err)
			}
			wg.Done()
		}(*host, src, dest)
	}

	wg.Wait()
}

func upload(gs *GoSSH, h Host, from, to string) error {
	stat, err := os.Stat(from)
	if err != nil {
		return err
	}

	startTime := time.Now()

	if err = sftpUpload(gs, h, stat, from, to); err != nil {
		return err
	}

	logrus.Infof("scp upload %s to %s:%s cost %s, successfully!", from, h.Addr, to, time.Since(startTime).String())

	return nil
}
