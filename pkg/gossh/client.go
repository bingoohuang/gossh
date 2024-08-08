package gossh

import (
	"io"
	"time"

	"github.com/bingoohuang/gonet"
	"github.com/bingoohuang/gossh/pkg/brg"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/proxy"
)

// Connect structure to store contents about ssh connection.
type Connect struct {
	Client      *ssh.Client
	ProxyDialer proxy.Dialer
}

// CreateClient connects to the remote SSH server, returns error if it couldn't establish a session to the SSH server.
func (c *Connect) CreateClient(addr string, cc *ssh.ClientConfig) error {
	dialer := c.ProxyDialer
	if dialer == nil {
		dialer = gonet.DialerTimeoutBean{ConnTimeout: cc.Timeout, ReadWriteTimeout: cc.Timeout}
	}

	targetInfo, addr := brg.CreateTargetInfo(addr)

	netConn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return err
	}

	for i, target := range targetInfo {
		if i > 0 {
			time.Sleep(100 * time.Millisecond)
		}
		if _, err := netConn.Write([]byte(target)); err != nil {
			return err
		}
	}

	sshCon, channel, req, err := ssh.NewClientConn(netConn, addr, cc)
	if err != nil {
		return err
	}

	c.Client = ssh.NewClient(sshCon, channel, req)
	return nil
}

// Close closes the ssh client.
func (c *Connect) Close() error {
	client := c.Client
	c.Client = nil

	if client != nil {
		return client.Close()
	}

	if c.ProxyDialer != nil {
		if closer, ok := c.ProxyDialer.(io.Closer); ok {
			_ = closer.Close()
		}
	}

	return nil
}
