package gossh

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func sftpUpload(gs *GoSSH, h Host, _ os.FileInfo, from string, to string) error {
	fromFile, _ := os.Open(from)
	defer fromFile.Close()

	sf, err := gs.sftpClientMap.GetClient(h)
	if err != nil {
		return fmt.Errorf("gs.sftpClientMap.GetClient failed: %w", err)
	}

	dest := to
	if err := sf.MkdirAll(filepath.Dir(dest)); err != nil {
		return fmt.Errorf("sftp MkdirAll %s error %w", dest, err)
	}

	f, err := sf.Create(dest)
	if err != nil {
		return fmt.Errorf("sftp Create %s error %w", dest, err)
	}

	if _, err := io.Copy(f, fromFile); err != nil {
		return fmt.Errorf("io.Copy failed: %w", err)
	}

	return nil
}
