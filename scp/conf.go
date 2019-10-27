package scp

import (
	"time"
)

// ClientConf is a struct containing all the configuration options used by an scp client.
type ClientConf struct {
	Timeout      time.Duration
	RemoteBinary string
}

// NewConf creates a new client configurer.
// It takes the required parameters: the addr and the ssh.ClientConfig and
// returns a configurer populated with the default values for the optional
// parameters.
//
// These optional parameters can be set by using the methods provided on the
// ClientConf struct.
func NewConf() *ClientConf {
	return &ClientConf{
		Timeout:      time.Minute,
		RemoteBinary: "scp",
	}
}

// CreateClient builds a client with the configuration stored within the ClientConf
func (c *ClientConf) CreateClient() Client {
	return Client{
		Timeout:      c.Timeout,
		RemoteBinary: c.RemoteBinary,
	}
}
