package gossh

import (
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/proxy"
)

// Connect structure to store contents about ssh connection.
type Connect struct {
	// Client *ssh.Client
	Client *ssh.Client

	// Session
	Session *ssh.Session

	// ProxyDialer
	ProxyDialer proxy.Dialer
}

// CreateClient connects to the remote SSH server, returns error if it couldn't establish a session to the SSH server
func (c *Connect) CreateClient(addr string, clientConfig *ssh.ClientConfig) error {
	// check Dialer
	if c.ProxyDialer == nil {
		c.ProxyDialer = proxy.Direct
	}

	// Dial to host:port
	netConn, err := c.ProxyDialer.Dial("tcp", addr)
	if err != nil {
		return err
	}

	// Create new ssh connect
	sshCon, channel, req, err := ssh.NewClientConn(netConn, addr, clientConfig)
	if err != nil {
		return err
	}

	// Create *ssh.Client
	c.Client = ssh.NewClient(sshCon, channel, req)

	sess, err := c.Client.NewSession()
	if err != nil {
		return err
	}

	c.Session = sess

	return nil
}

// Close closes the ssh client.
func (c *Connect) Close() error {
	client := c.Client
	c.Client = nil

	if client != nil {
		return client.Close()
	}

	return nil
}

// CreateSession create a session.
func (c *Connect) CreateSession() (*ssh.Session, error) {
	return c.Client.NewSession()
}
