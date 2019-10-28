package elf

import (
	"fmt"
	"os"
)

// FileExists 检查文件是否存在，并且不是目录
func FileExists(name string) error {
	if fi, err := os.Stat(name); err != nil {
		return err
	} else if fi.IsDir() {
		return fmt.Errorf("file %s is a directory", name)
	}

	return nil
}

type DirMode int

const (
	UnknownDirMode DirMode = iota
	DirectoryMode
	SingleFileMode
)

// GetDirMode tells the name is a directory or not
func GetDirMode(name string) (DirMode, error) {
	if fi, err := os.Stat(name); err != nil {
		return UnknownDirMode, err
	} else if fi.IsDir() {
		return DirectoryMode, nil
	}

	return SingleFileMode, nil
}
