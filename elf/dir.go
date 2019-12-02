package elf

import (
	"path/filepath"
	"strings"
)

// BaseDir returns the common directory for a slice of directories.
func BaseDir(dirs []string) string {
	baseDir := ""

	for _, dir := range dirs {
		d := filepath.Dir(dir)

		if baseDir == "" {
			baseDir = d
		} else {
			for !strings.HasPrefix(d, baseDir) {
				baseDir = filepath.Dir(baseDir)
			}
		}

		if baseDir == "/" {
			break
		}
	}

	if baseDir == "" {
		baseDir = "/"
	}

	return baseDir
}
