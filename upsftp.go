package gossh

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func sftpUpload(gs *GoSSH, h Host, fromStat os.FileInfo, from string, to string) error {
	fromFile, _ := os.Open(from)
	defer fromFile.Close()

	sf, err := gs.sftpClientMap.GetClient(h)
	if err != nil {
		return fmt.Errorf("gs.sftpClientMap.GetClient failed: %w", err)
	}

	dest := to
	stat, _ := sf.Stat(dest)

	if stat != nil && stat.IsDir() {
		dest = filepath.Join(dest, filepath.Base(from))
	} else if err := sf.MkdirAll(filepath.Dir(dest)); err != nil {
		return fmt.Errorf("sftp MkdirAll %s error %w", dest, err)
	}

	f, err := sf.Create(dest)
	if err != nil {
		return fmt.Errorf("sftp Create %s error %w", dest, err)
	}

	if _, err := io.Copy(f, fromFile); err != nil {
		return fmt.Errorf("io.Copy failed: %w", err)
	}

	if err := sf.Chmod(dest, fromStat.Mode()); err != nil {
		return fmt.Errorf("sf.Chmo %s failed: %w", dest, err)
	}

	return nil
}
