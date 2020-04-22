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

	"github.com/pkg/sftp"

	"github.com/cheggaaa/pb/v3"
)

// Exec execute in specified host.
func (s *UlCmd) Exec(gs *GoSSH, h *Host) error {
	startTime := time.Now()

	if err := s.sftpUpload(gs, h); err != nil {
		return err
	}

	gs.Vars.log.Printf("uploaded %s to %s:%s cost %s, successfully!\n",
		s.local, h.Addr, s.remote, time.Since(startTime).String())

	return nil
}

func (s *UlCmd) sftpUpload(gs *GoSSH, h *Host) error {
	sf, err := h.GetClient()
	if err != nil {
		return fmt.Errorf("gs.sftpClientMap.GetClient failed: %w", err)
	}

	remote := s.remote
	stat, err := sf.Stat(remote)

	overrideSingleFile := false
	isDir := false

	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("stat remote %s error %w", remote, err)
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
			relativePart := strings.TrimPrefix(localDir, s.basedir)
			remoteDir := filepath.Join(remote, relativePart)

			if err := sf.MkdirAll(remoteDir); err != nil {
				return fmt.Errorf("sftp MkdirAll %s error %w", remoteDir, err)
			}
		}
	}

	for _, localFile := range s.localFiles {
		if err := uploadSingleOne(gs.Vars.log, sf, s.basedir, localFile, remote, overrideSingleFile); err != nil {
			return fmt.Errorf("uploadSingleOne %s to %s error %w", localFile, remote, err)
		}
	}

	return nil
}

func uploadSingleOne(logger *log.Logger, sf *sftp.Client,
	basedir, localFile, remote string, overrideSingleFile bool) error {
	fromFile, _ := os.Open(localFile)
	defer fromFile.Close()

	dest := remote

	if !overrideSingleFile {
		dest = filepath.Join(remote, strings.TrimPrefix(localFile, basedir))
	}

	f, err := sf.Create(dest)
	if err != nil {
		return fmt.Errorf("sftp Create %s error %w", dest, err)
	}

	defer f.Close()

	fromStat, err := fromFile.Stat()
	if err != nil {
		return fmt.Errorf("stat file %s error %w", localFile, err)
	}

	logger.Printf("start to upload %s to %s\n", localFile, dest)

	start := time.Now()
	bar := pb.StartNew(int(fromStat.Size()))

	if _, err := io.Copy(bar.NewProxyWriter(f), fromFile); err != nil {
		return fmt.Errorf("io.Copy failed: %w", err)
	}

	bar.Finish()

	logger.Printf("complete to upload %s to %s, cost %v\n", localFile, dest, time.Since(start))

	if err := sf.Chmod(dest, fromStat.Mode()); err != nil {
		return fmt.Errorf("sf.Chmo %s failed: %w", dest, err)
	}

	return nil
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
