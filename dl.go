package gossh

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	errs "github.com/pkg/errors"

	"github.com/bingoohuang/gou/lang"
	"github.com/pkg/sftp"
)

// Exec execute in specified host.
func (s *DlCmd) Exec(gs *GoSSH, h *Host) error {
	sf, err := h.GetSftpClient()
	if err != nil {
		return errs.Wrapf(err, "GetSftpClient")
	}

	remotes, err := sf.Glob(s.remote)
	if err != nil {
		return errs.Wrapf(err, "Glob %s", s.remote)
	}

	if len(remotes) == 0 {
		return fmt.Errorf("no files to download for %s", s.remote)
	}

	for _, remote := range remotes {
		stat, err := sf.Stat(remote)
		if err != nil {
			return errs.Wrapf(err, "sftp.Stat %s", remote)
		}

		if err := download(gs.Vars.log, stat, h.Addr, s.local, remote, sf); err != nil {
			return err
		}
	}

	return nil
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

	defer lang.Closef(&err, localFile, "close file %s", to)

	writer := io.Writer(localFile)
	if _, err := io.Copy(writer, remoteFile); err != nil {
		return fmt.Errorf("io.Copy failed: %w", err)
	}

	_ = localFile.Sync()

	logger.Printf("downloaded %s:%s to %s cost %s, successfully!\n", host, from, to, time.Since(startTime).String())

	return nil
}
