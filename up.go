package gossh

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bingoohuang/gou/lang"
	errs "github.com/pkg/errors"

	"github.com/pkg/sftp"

	"github.com/cheggaaa/pb/v3"
)

// Exec execute in specified host.
func (s *UlCmd) Exec(gs *GoSSH, h *Host) error {
	if err := s.init(); err != nil {
		return err
	}

	startTime := time.Now()

	if err := s.sftpUpload(gs, h); err != nil {
		return err
	}

	gs.Vars.log.Printf("uploaded %s to %s:%s cost %s, successfully!\n",
		s.local, h.Addr, s.remote, time.Since(startTime).String())

	return nil
}

func (s *UlCmd) sftpUpload(gs *GoSSH, h *Host) error {
	sf, err := h.GetSftpClient()
	if err != nil {
		return fmt.Errorf("gs.sftpClientMap.GetSftpClient failed: %w", err)
	}

	remote := h.SubstituteResultVars(s.remote)
	stat, err := sf.Stat(remote)

	overrideSingleFile := false
	isDir := false

	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return errs.Wrapf(err, "stat remote %s", remote)
		}

		isDir = true

		if filepath.Base(s.localFiles[0]) == filepath.Base(remote) {
			overrideSingleFile = true
		}
	} else if !stat.IsDir() {
		overrideSingleFile = true
	}

	if len(s.localFiles) > 1 && overrideSingleFile {
		return fmt.Errorf("unable to upload multiple files %s to remote single file %s", s.local, s.remote)
	}

	if isDir || stat.IsDir() {
		localDirs := extractDirs(s.localFiles)
		for _, localDir := range localDirs {
			localDir := h.SubstituteResultVars(localDir)
			relativePart := strings.TrimPrefix(localDir, s.basedir)
			remoteDir := filepath.Join(remote, relativePart)

			if err := sf.MkdirAll(remoteDir); err != nil {
				return errs.Wrapf(err, "sftp MkdirAll %s", remoteDir)
			}
		}
	}

	for _, localFile := range s.localFiles {
		localFile := h.SubstituteResultVars(localFile)
		if err := uploadSingle(gs.Vars.log, sf, s.basedir, localFile, remote, overrideSingleFile); err != nil {
			return errs.Wrapf(err, "uploadSingle %s to %s", localFile, remote)
		}
	}

	return nil
}

func uploadSingle(l *log.Logger, sf *sftp.Client, basedir, local, remote string, overrideSingle bool) (err error) {
	fromFile, _ := os.Open(local)
	defer lang.Closef(&err, fromFile, "close local %s", local)

	dest := remote

	if !overrideSingle {
		dest = filepath.Join(remote, strings.TrimPrefix(local, basedir))
	}

	f, err := sf.Create(dest)
	if err != nil {
		return errs.Wrapf(err, "sftp Create %s", dest)
	}

	defer lang.Closef(&err, f, "close dest %s", dest)

	fromStat, err := fromFile.Stat()
	if err != nil {
		return errs.Wrapf(err, "stat file %s", local)
	}

	l.Printf("start to upload %s to %s\n", local, dest)

	start := time.Now()
	bar := pb.StartNew(int(fromStat.Size()))

	if _, err := io.Copy(bar.NewProxyWriter(f), fromFile); err != nil {
		return errs.Wrapf(err, "io.Copy %s to %s", local, dest)
	}

	bar.Finish()

	l.Printf("complete to upload %s to %s, cost %v\n", local, dest, time.Since(start))

	if err := sf.Chmod(dest, fromStat.Mode()); err != nil {
		return errs.Wrapf(err, "sf.Chmod %s", dest)
	}

	return err
}

func extractDirs(files []string) []string {
	dirs := make([]string, 0)

	for _, f := range files {
		d := filepath.Dir(f)

		if !inDirs(dirs, d) {
			dirs = append(dirs, d)
		}
	}

	merged := make([]string, 0)

	for i, d := range dirs {
		if !inDirs(dirs[i+1:], d) {
			merged = append(merged, d)
		}
	}

	return merged
}

func inDirs(dirs []string, d string) bool {
	for _, dir := range dirs {
		if strings.HasPrefix(dir, d) {
			return true
		}
	}

	return false
}
