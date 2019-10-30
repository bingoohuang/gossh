package gossh

import (
	"fmt"

	"github.com/bingoohuang/gossh/gossh"
	"github.com/pkg/sftp"
)

func makeSftpClient(h Host) (*sftp.Client, error) {
	conn, err := gossh.DialTCP(h.Addr, gossh.PasswordKey(h.User, h.Password))
	if err != nil {
		return nil, fmt.Errorf("ssh.Dial(%q) failed: %w", h.Addr, err)
	}

	sf, err := sftp.NewClient(conn)
	if err != nil {
		return nil, fmt.Errorf("sftp.NewClient failed: %w", err)
	}

	return sf, nil
}

type sftpClientMap map[string]*sftp.Client

// GetClient get sftClient by host
func (m *sftpClientMap) GetClient(h Host) (*sftp.Client, error) {
	if c, ok := (*m)[h.Addr]; ok {
		return c, nil
	}

	c, err := makeSftpClient(h)
	if err != nil {
		return nil, err
	}

	(*m)[h.Addr] = c

	return c, nil
}

// Close closes all the sftpClients in map.
func (m *sftpClientMap) Close() {
	for _, v := range *m {
		v.Close()
	}
}
