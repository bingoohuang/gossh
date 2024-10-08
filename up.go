package gossh

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bingoohuang/ngg/ss"
	"github.com/bingoohuang/pb"
	errs "github.com/pkg/errors"
	"github.com/pkg/sftp"
)

// Exec execute in specified host.
func (s *UlCmd) Exec(_ *GoSSH, h *Host, stdout io.Writer, _ ExecOption) error {
	if err := s.init(h); err != nil {
		return err
	}

	startTime := time.Now()

	if err := s.sftpUpload(stdout, h); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(stdout, "uploaded %s to %s:%s cost %s, successfully!\n",
		s.local, h.Addr, s.remote, time.Since(startTime).String())

	return nil
}

func (s *UlCmd) sftpUpload(stdout io.Writer, h *Host) error {
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
		// nolint:goerr113
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
		if err := uploadSingle(stdout, sf, s.basedir, localFile, remote, overrideSingleFile); err != nil {
			return errs.Wrapf(err, "uploadSingle %s to %s", localFile, remote)
		}
	}

	return nil
}

func uploadSingle(stdout io.Writer, sf *sftp.Client, basedir, local, remote string, overrideSingle bool) (err error) {
	fromFile, _ := os.Open(local)
	defer ss.Close(fromFile)

	dest := remote

	if !overrideSingle {
		dest = filepath.Join(remote, strings.TrimPrefix(local, basedir))
	}

	f, err := sf.Create(dest)
	if err != nil {
		return errs.Wrapf(err, "sftp Create %s", dest)
	}

	defer ss.Close(f)

	fromStat, err := fromFile.Stat()
	if err != nil {
		return errs.Wrapf(err, "stat file %s", local)
	}

	fmt.Fprintf(stdout, "start to upload %s to %s\n", local, dest)

	start := time.Now()
	bar := pb.Start64(fromStat.Size())

	if _, err := io.Copy(bar.NewProxyWriter(f), fromFile); err != nil {
		return errs.Wrapf(err, "io.Copy %s to %s", local, dest)
	}

	bar.Finish()

	fmt.Fprintf(stdout, "complete to upload %s to %s, cost %v\n", local, dest, time.Since(start))

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
