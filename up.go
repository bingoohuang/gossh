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

func (s *SCPCmd) upload(gs *GoSSH) error {
	var err error
	if s.sourceDir == elf.UnknownDirMode {
		if s.sourceDir, err = elf.GetFileMode(s.source); err != nil {
			logrus.Warnf("error %v", err)
			return err
		}
	}

	switch s.sourceDir {
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

func (s *SCPCmd) upSCP(gs *GoSSH) error {
	destBase := strings.TrimPrefix(s.dest, "%host:")
	if err := s.scpRecursively(destBase, gs); err != nil {
		return err
	}

	return nil
}

func (s *SCPCmd) scpRecursively(destBase string, gs *GoSSH) error {
	if err := filepath.Walk(s.source,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				dest := filepath.Join(destBase, path)
				uploadFile(gs, path, dest)
			}

			return nil
		}); err != nil {
		logrus.Warnf(" filepath.Walk %s error %v", s.source, err)

		return err
	}

	return nil
}

func (s *SCPCmd) singleSCP(gs *GoSSH) {
	baseFrom := filepath.Base(s.source)
	dest := s.dest
	baseDest := filepath.Base(dest)

	if baseDest != baseFrom {
		dest = filepath.Join(dest, baseFrom)
	}

	dest = strings.TrimPrefix(dest, "%host:")
	uploadFile(gs, s.source, dest)
}

func uploadFile(gs *GoSSH, src, dest string) {
	var wg sync.WaitGroup

	wg.Add(len(gs.Hosts))

	for _, host := range gs.Hosts {
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
