package gossh

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
)

// ExecInHosts executes downloading among hosts.
func (s *DlCmd) ExecInHosts(gs *GoSSH, wg *sync.WaitGroup) error {
	if gs.Vars.Goroutines == Off {
		for _, host := range s.hosts {
			s.do(gs, *host, nil)
		}

		return nil
	}

	if gs.Vars.Goroutines == CmdScope {
		wg = &sync.WaitGroup{}
	}

	wg.Add(len(s.hosts))

	for _, host := range s.hosts {
		go s.do(gs, *host, wg)
	}

	if gs.Vars.Goroutines == CmdScope {
		wg.Wait()
	}

	return nil
}

func (s *DlCmd) do(gs *GoSSH, h Host, wg *sync.WaitGroup) {
	if err := s.downloadHost(gs, h); err != nil {
		logrus.Warnf("download %s error %v", s.remote, err)
	}

	if wg != nil {
		wg.Done()
	}
}

func (s *DlCmd) downloadHost(gs *GoSSH, h Host) error {
	sf, err := gs.sftpClientMap.GetClient(h)
	if err != nil {
		return fmt.Errorf("gs.sftpClientMap.GetClient failed: %w", err)
	}

	stat, err := sf.Stat(s.remote)
	if err != nil {
		return fmt.Errorf("sftp.Stat %s failed: %w", s.remote, err)
	}

	return download(stat, h.Addr, s.local, s.remote, sf)
}

func download(stat os.FileInfo, host, to, from string, sf *sftp.Client) error {
	if stat.IsDir() {
		fileInfos, err := sf.ReadDir(from)
		if err != nil {
			return fmt.Errorf("sftp.ReadDir %s failed: %w", from, err)
		}

		for _, fi := range fileInfos {
			src := filepath.Join(from, fi.Name())
			to1 := to

			if fi.IsDir() {
				to1 = filepath.Join(to, fi.Name())
			}

			if err := download(fi, host, to1, src, sf); err != nil {
				return err
			}
		}

		return nil
	}

	dest := filepath.Join(to, filepath.Base(from))

	if err := os.MkdirAll(to, 0744); err != nil {
		return fmt.Errorf("MkdirAll %s failed: %w", to, err)
	}

	return downloadFile(sf, stat.Mode(), host, from, dest)
}

func downloadFile(sf *sftp.Client, perm os.FileMode, host, from, to string) error {
	startTime := time.Now()

	remoteFile, err := sf.Open(from)
	if err != nil {
		return fmt.Errorf("sftp.Open %s failed: %w", from, err)
	}

	localFile, err := os.OpenFile(to, os.O_RDWR|os.O_APPEND|os.O_CREATE, perm)
	if err != nil {
		return fmt.Errorf("os.OpenFile %s failed: %w", to, err)
	}

	defer localFile.Close()

	writer := io.Writer(localFile)
	if _, err := io.Copy(writer, remoteFile); err != nil {
		return fmt.Errorf("io.Copy failed: %w", err)
	}

	_ = localFile.Sync()

	logrus.Infof("downloaded %s:%s to %s cost %s, successfully!", host, from, to, time.Since(startTime).String())

	return nil
}
