package gossh

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
)

func (s SCPCmd) download(gs *GoSSH) error {
	re := regexp.MustCompile(`%host(-\w+)?:`)
	submatch := re.FindStringSubmatch(s.cmd)

	if len(submatch) >= 2 {
		group0 := submatch[0]
		name := submatch[1][1:]
		host := findHost(gs.Hosts, name)

		if host == nil {
			return fmt.Errorf("unable to find host %s in hosts", name)
		}

		source := s.source[len(group0):]

		return downloadHost(gs, *host, source, s.dest)
	}

	for _, host := range gs.Hosts {
		if err := downloadHost(nil, *host, s.source, s.dest); err != nil {
			return err
		}
	}

	return nil
}

func findHost(hosts []*Host, name string) *Host {
	for _, h := range hosts {
		if h.Name == name {
			return h
		}
	}

	return nil
}

func downloadHost(gs *GoSSH, h Host, from, to string) error {
	sf, err := gs.sftpClientMap.GetClient(h)
	if err != nil {
		return fmt.Errorf("gs.sftpClientMap.GetClient failed: %w", err)
	}

	stat, err := sf.Stat(from)
	if err != nil {
		return fmt.Errorf("sftp.Stat %s failed: %w", from, err)
	}

	if err := download(stat, h.Addr, to, from, sf); err != nil {
		return err
	}

	return nil
}

func download(stat os.FileInfo, host, to, from string, sf *sftp.Client) error {
	if stat.IsDir() {
		fileInfos, err := sf.ReadDir(from)
		if err != nil {
			return fmt.Errorf("sftp.ReadDir %s failed: %w", from, err)
		}

		for _, fi := range fileInfos {
			src := filepath.Join(from, fi.Name())
			if err := download(fi, host, to, src, sf); err != nil {
				return err
			}
		}

		return nil
	}

	dest := filepath.Join(to, from)

	if err := os.MkdirAll(filepath.Dir(dest), 0744); err != nil {
		return fmt.Errorf("MkdirAll %s failed: %w", to, err)
	}

	return downloadFile(sf, stat.Mode(), host, from, dest)
}

func downloadFile(sf *sftp.Client, perm os.FileMode, host, from, to string) error {
	startTime := time.Now()

	remoteFile, err := sf.Open(from)
	if err != nil {
		return fmt.Errorf("sftp.Open %s failed: %w", from, err)
	}

	localFile, err := os.OpenFile(to, os.O_RDWR|os.O_APPEND|os.O_CREATE, perm)
	if err != nil {
		return fmt.Errorf("os.OpenFile %s failed: %w", to, err)
	}

	defer localFile.Close()

	writer := io.Writer(localFile)
	if _, err := io.Copy(writer, remoteFile); err != nil {
		return fmt.Errorf("io.Copy failed: %w", err)
	}

	_ = localFile.Sync()

	logrus.Infof("scp download %s:%s to %s cost %s, successfully!", host, from, to, time.Since(startTime).String())

	return nil
}
