package gossh

import (
	"strings"

	homedir "github.com/mitchellh/go-homedir"

	"github.com/bingoohuang/gossh/elf"
	"github.com/sirupsen/logrus"
)

// Direction means scp upload or download
type Direction int

const (
	// UnknownDirection means unknown direction.
	UnknownDirection Direction = iota
	// UploadDirection means upload direction.
	UploadDirection
	// DownloadDirection means download direction.
	DownloadDirection
)

// SCPCmd means commands for scp.
type SCPCmd struct {
	direction Direction
	source    string
	sourceDir elf.DirMode
	dest      string
	destDir   elf.DirMode
	cmd       string
}

// Parse parses SCPCmd.
func (SCPCmd) Parse() {}

func buildSCPCmd(cmd string) *SCPCmd {
	fields := strings.Fields(cmd)
	if len(fields) < 3 {
		logrus.Warnf("bad format for %s", cmd)
		return nil
	}

	from := fields[1]
	dest := fields[2]

	direction := parseDirection(from, dest, cmd)
	if direction == UnknownDirection {
		return nil
	}

	home, _ := homedir.Dir()

	switch direction {
	case UploadDirection:
		from = strings.ReplaceAll(from, "~", home)
		dirMode, _ := elf.GetFileMode(from)

		return &SCPCmd{
			direction: UploadDirection,
			source:    from,
			sourceDir: dirMode,
			dest:      dest,
			destDir:   elf.UnknownDirMode,
			cmd:       cmd,
		}
	case DownloadDirection:
		dest = strings.ReplaceAll(dest, "~", home)
		dirMode, _ := elf.GetFileMode(dest)

		return &SCPCmd{
			direction: DownloadDirection,
			source:    from,
			sourceDir: elf.UnknownDirMode,
			dest:      dest,
			destDir:   dirMode,
			cmd:       cmd,
		}
	}

	return nil
}

func parseDirection(from, dest, cmd string) Direction {
	switch {
	case strings.Contains(from, "%host"):
		return DownloadDirection
	case strings.Contains(dest, "%host"):
		return UploadDirection
	default:
		logrus.Warnf("unknown direction for %s", cmd)
	}

	return UnknownDirection
}

// ExecInHosts execute in specified hosts.
func (s *SCPCmd) ExecInHosts(gs *GoSSH) error {
	switch s.direction {
	case UploadDirection:
		return s.upload(gs)
	case DownloadDirection:
		return s.download(gs)
	}

	return nil
}
