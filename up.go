package gossh

import (
	"os"
	"path/filepath"
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
				s.uploadFile(gs, path, dest)
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

	s.uploadFile(gs, s.local, dest)
}

func (s *UlCmd) uploadFile(gs *GoSSH, src, dest string) {
	targetHosts := s.hosts
	if len(targetHosts) == 0 {
		logrus.Warnf("there is no host to upload %s", src)
	}

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

	logrus.Infof("uploaded %s to %s:%s cost %s, successfully!", from, h.Addr, to, time.Since(startTime).String())

	return nil
}
