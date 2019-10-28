package gossh

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/bingoohuang/gossh/gossh"

	"github.com/bingoohuang/gossh/elf"
	"github.com/bingoohuang/gossh/scp"
	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
)

type Direction int

const (
	UnknownDir Direction = iota
	UploadDir
	DownloadDir
)

type SCPCmd struct {
	direction Direction
	source    string
	sourceDir elf.DirMode
	dest      string
	destDir   elf.DirMode
	cmd       string
}

func (SCPCmd) Parse() {

}

func buildSCPCmd(cmd string) *SCPCmd {
	fields := strings.Fields(cmd)
	if len(fields) < 3 {
		logrus.Warnf("bad format for %s", cmd)
		return nil
	}

	from := fields[1]
	dest := fields[2]
	direction := UnknownDir

	if strings.Contains(from, "%host") {
		direction = DownloadDir
	} else if strings.Contains(dest, "%host") {
		direction = UploadDir
	} else {
		logrus.Warnf("unknown direction for %s", cmd)
		return nil
	}
	home, _ := homedir.Dir()
	from = strings.ReplaceAll(from, "~", home)

	dirMode, _ := elf.GetDirMode(from)
	if direction == UploadDir {
		return &SCPCmd{
			direction: UploadDir,
			source:    from,
			sourceDir: dirMode,
			dest:      dest,
			destDir:   elf.UnknownDirMode,
			cmd:       cmd,
		}
	}

	return nil
}

func (s *SCPCmd) ExecInHosts(hosts []*Host) {
	if s.direction == UploadDir {
		s.Upload(hosts)
	}
}

func (s *SCPCmd) Upload(hosts []*Host) {
	var err error
	if s.sourceDir == elf.UnknownDirMode {
		if s.sourceDir, err = elf.GetDirMode(s.source); err != nil {
			logrus.Warnf("error %v", err)
			return
		}
	}

	if s.sourceDir == elf.SingleFileMode {
		baseFrom := filepath.Base(s.source)
		dest := s.dest
		baseDest := filepath.Base(dest)
		if baseDest != baseFrom {
			dest = filepath.Join(dest, baseFrom)
		}

		var wg sync.WaitGroup
		wg.Add(len(hosts))

		for _, host := range hosts {
			go func(h Host, from, to string) {
				ScpUpload(*host, from, to)
				wg.Done()
			}(*host, s.source, strings.TrimPrefix(dest, "%host:"))

		}

		wg.Wait()
	}
}

func ScpUpload(h Host, from, to string) error {
	startTime := time.Now()
	scpClient := scp.NewConf().CreateClient()

	if err := scpClient.Connect(h.Addr, gossh.PasswordKey(h.User, h.Password)); err != nil {
		return fmt.Errorf("couldn't establish a connection to the remote server %w", err)
	}

	defer scpClient.Close()

	f, _ := os.Open(from)
	defer f.Close()

	stat, _ := os.Stat(from)

	mod := fmt.Sprintf("0%o", stat.Mode())
	if err := scpClient.CopyFile(f, to, mod); err != nil {
		return fmt.Errorf("error while copying file %s to %s:%s %w", from, h.Addr, to, err)
	}

	logrus.Infof("scp upload %s to %s:%s cost %s, successfully!", from, h.Addr, to, time.Since(startTime).String())

	return nil
}
