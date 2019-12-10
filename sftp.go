package gossh

import (
	"fmt"
	"sync"
	"time"

	"github.com/pkg/sftp"
)

func makeSftpClient(h Host, timeout time.Duration) (*sftp.Client, error) {
	gc, err := h.GetGosshConnect(timeout)
	if err != nil {
		return nil, err
	}

	sf, err := sftp.NewClient(gc.Client)
	if err != nil {
		return nil, fmt.Errorf("sftp.NewClient failed: %w", err)
	}

	return sf, nil
}

type sftpClientMap struct {
	m map[string]*sftp.Client
	sync.Mutex
	timeout time.Duration
}

func makeSftpClientMap(timeout time.Duration) *sftpClientMap {
	return &sftpClientMap{
		m:       make(map[string]*sftp.Client),
		timeout: timeout,
	}
}

// GetClient get sftClient by host
func (m *sftpClientMap) GetClient(h Host) (*sftp.Client, error) {
	m.Lock()
	defer m.Unlock()

	if c, ok := m.m[h.Addr]; ok {
		return c, nil
	}

	c, err := makeSftpClient(h, m.timeout)
	if err != nil {
		return nil, err
	}

	m.m[h.Addr] = c

	return c, nil
}

// Close closes all the sftpClients in map.
func (m *sftpClientMap) Close() {
	m.Lock()
	defer m.Unlock()

	for _, v := range m.m {
		v.Close()
	}
}
