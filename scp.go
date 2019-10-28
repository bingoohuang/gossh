package gossh

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/bingoohuang/gossh/gossh"
	homedir "github.com/mitchellh/go-homedir"

	"github.com/bingoohuang/gossh/elf"
	"github.com/bingoohuang/gossh/scp"
	"github.com/sirupsen/logrus"
)

// Direction means scp upload or download
type Direction int

const (
	// UploadDirection means upload direction.
	UploadDirection Direction = iota + 1
	// DownloadDirection means download direction.
	DownloadDirection
)

// SCPCmd means commands for scp.
type SCPCmd struct {
	direction Direction
	source    string
	sourceDir elf.DirMode
	dest      string
	destDir   elf.DirMode
	cmd       string
}

// Parse parses SCPCmd.
func (SCPCmd) Parse() {}

func buildSCPCmd(cmd string) *SCPCmd {
	fields := strings.Fields(cmd)
	if len(fields) < 3 {
		logrus.Warnf("bad format for %s", cmd)
		return nil
	}

	from := fields[1]
	dest := fields[2]

	var direction Direction

	switch {
	case strings.Contains(from, "%host"):
		direction = DownloadDirection
	case strings.Contains(dest, "%host"):
		direction = UploadDirection
	default:
		logrus.Warnf("unknown direction for %s", cmd)
		return nil
	}

	home, _ := homedir.Dir()
	from = strings.ReplaceAll(from, "~", home)

	dirMode, _ := elf.GetFileMode(from)

	if direction == UploadDirection {
		return &SCPCmd{
			direction: UploadDirection,
			source:    from,
			sourceDir: dirMode,
			dest:      dest,
			destDir:   elf.UnknownDirMode,
			cmd:       cmd,
		}
	}

	return nil
}

// ExecInHosts execute in specified hosts.
func (s *SCPCmd) ExecInHosts(hosts []*Host) error {
	if s.direction == UploadDirection {
		return s.upload(hosts)
	}

	return nil
}

func (s *SCPCmd) upload(hosts []*Host) error {
	var err error
	if s.sourceDir == elf.UnknownDirMode {
		if s.sourceDir, err = elf.GetFileMode(s.source); err != nil {
			logrus.Warnf("error %v", err)
			return err
		}
	}

	switch s.sourceDir {
	case elf.SingleFileMode:
		s.singleSCP(hosts)
	case elf.DirectoryMode:
		err2 := s.dirSCP(hosts)
		if err2 != nil {
			return err2
		}
	}

	return nil
}

func (s *SCPCmd) dirSCP(hosts []*Host) error {
	destBase := strings.TrimPrefix(s.dest, "%host:")

	remotePaths, err := s.buildRemotePaths(destBase)
	if err != nil {
		return err
	}

	mkdirs := SSHCmd{cmd: "mkdir -p " + strings.Join(remotePaths, " ")}
	if err := mkdirs.ExecInHosts(hosts); err != nil {
		return err
	}

	if err := s.scpRecursively(destBase, hosts); err != nil {
		return err
	}

	return nil
}

func (s *SCPCmd) scpRecursively(destBase string, hosts []*Host) error {
	if err := filepath.Walk(s.source,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				dest := filepath.Join(destBase, path)
				uploadFile(hosts, path, dest)
			}

			return nil
		}); err != nil {
		logrus.Warnf(" filepath.Walk %s error %v", s.source, err)

		return err
	}

	return nil
}

func (s *SCPCmd) buildRemotePaths(destBase string) ([]string, error) {
	remotePaths := make([]string, 0)

	if err := filepath.Walk(s.source,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				dest := filepath.Join(destBase, path)
				remotePaths = append(remotePaths, dest)
			}

			return nil
		}); err != nil {
		logrus.Warnf(" filepath.Walk %s error %v", s.source, err)
		return nil, err
	}

	return remotePaths, nil
}

func (s *SCPCmd) singleSCP(hosts []*Host) {
	baseFrom := filepath.Base(s.source)
	dest := s.dest
	baseDest := filepath.Base(dest)

	if baseDest != baseFrom {
		dest = filepath.Join(dest, baseFrom)
	}

	dest = strings.TrimPrefix(dest, "%host:")
	uploadFile(hosts, s.source, dest)
}

func uploadFile(hosts []*Host, src, dest string) {
	var wg sync.WaitGroup

	wg.Add(len(hosts))

	for _, host := range hosts {
		go func(h Host, from, to string) {
			if err := scpUpload(h, from, to); err != nil {
				logrus.Warnf(" scpUpload %s error %v", from, err)
			}
			wg.Done()
		}(*host, src, dest)
	}

	wg.Wait()
}

func scpUpload(h Host, from, to string) error {
	stat, err := os.Stat(from)
	if err != nil {
		return err
	}

	startTime := time.Now()
	scpClient := scp.NewConf().CreateClient()

	if err := scpClient.Connect(h.Addr, gossh.PasswordKey(h.User, h.Password)); err != nil {
		return fmt.Errorf("couldn't establish a connection to the remote server %w", err)
	}

	defer scpClient.Close()

	f, _ := os.Open(from)
	defer f.Close()

	mod := fmt.Sprintf("0%o", stat.Mode())
	if err := scpClient.CopyFile(f, to, mod); err != nil {
		return fmt.Errorf("error while copying file %s to %s:%s %w", from, h.Addr, to, err)
	}

	logrus.Infof("scp upload %s to %s:%s cost %s, successfully!", from, h.Addr, to, time.Since(startTime).String())

	return nil
}
