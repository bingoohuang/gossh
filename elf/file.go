package elf

import (
	"fmt"
	"os"
)

// SingleFileExists 检查文件是否存在，并且不是目录
func SingleFileExists(name string) error {
	mode, err := GetFileMode(name)
	if err != nil {
		return err
	}

	if mode == SingleFileMode {
		return nil
	}

	return fmt.Errorf("file %s is a directory", name)
}

// DirMode means file is single file or directory.
type DirMode int

const (
	// UnknownDirMode means unknown file or directory.
	UnknownDirMode DirMode = iota
	// DirectoryMode means directory.
	DirectoryMode
	// SingleFileMode means single file.
	SingleFileMode
)

// GetFileMode tells the name is a directory or not
func GetFileMode(name string) (DirMode, error) {
	if fi, err := os.Stat(name); err != nil {
		return UnknownDirMode, err
	} else if fi.IsDir() {
		return DirectoryMode, nil
	}

	return SingleFileMode, nil
}
