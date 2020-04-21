package gossh

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/sftp"
)

// ExecInHosts executes downloading among hosts.
func (s *DlCmd) ExecInHosts(gs *GoSSH, target *Host) error {
	for _, host := range s.hosts {
		if target == nil || target == host {
			s.do(gs, *host)
		}
	}

	return nil
}

func (s *DlCmd) do(gs *GoSSH, h Host) {
	if err := s.downloadHost(gs, h); err != nil {
		gs.Vars.log.Printf("download %s error %v\n", s.remote, err)
	}
}

func (s *DlCmd) downloadHost(gs *GoSSH, h Host) error {
	sf, err := h.GetClient()
	if err != nil {
		return fmt.Errorf("gs.sftpClientMap.GetClient failed: %w", err)
	}

	stat, err := sf.Stat(s.remote)
	if err != nil {
		return fmt.Errorf("sftp.Stat %s failed: %w", s.remote, err)
	}

	return download(gs.Vars.log, stat, h.Addr, s.local, s.remote, sf)
}

func download(logger *log.Logger, stat os.FileInfo, host, to, from string, sf *sftp.Client) error {
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

			if err := download(nil, fi, host, to1, src, sf); err != nil {
				return err
			}
		}

		return nil
	}

	dest := filepath.Join(to, filepath.Base(from))

	if err := os.MkdirAll(to, 0744); err != nil {
		return fmt.Errorf("MkdirAll %s failed: %w", to, err)
	}

	return downloadFile(logger, sf, stat.Mode(), host, from, dest)
}

func downloadFile(logger *log.Logger, sf *sftp.Client, perm os.FileMode, host, from, to string) error {
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

	logger.Printf("downloaded %s:%s to %s cost %s, successfully!\n", host, from, to, time.Since(startTime).String())

	return nil
}
