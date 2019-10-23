package gossh

import "golang.org/x/crypto/ssh"

// NewSession connects to the remote SSH server, returns error if it couldn't establish a session to the SSH server
func NewSession(addr string, clientConfig *ssh.ClientConfig) (ssh.Conn, *ssh.Session, error) {
	client, err := DialTCP(addr, clientConfig)
	if err != nil {
		return nil, nil, err
	}

	sess, err := client.NewSession()

	return client.Conn, sess, err
}

// DialTCP connects to the remote SSH server
func DialTCP(addr string, clientConfig *ssh.ClientConfig) (*ssh.Client, error) {
	return ssh.Dial("tcp", addr, clientConfig)
}
