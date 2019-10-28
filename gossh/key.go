package gossh

import (
	"io/ioutil"

	"golang.org/x/crypto/ssh"
)

// PrivateKey loads a public key from "path" and returns a SSH ClientConfig to authenticate with the server.
func PrivateKey(username, path string) (*ssh.ClientConfig, error) {
	privateKey, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	return &ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // nolint G106
	}, nil
}

// PrivateKeyPassphrase returns the ssh.ClientConfig based on specified username, passphrase and path.
func PrivateKeyPassphrase(username, passphrase, path string) (*ssh.ClientConfig, error) {
	privateKey, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKeyWithPassphrase(privateKey, []byte(passphrase))
	if err != nil {
		return nil, err
	}

	return &ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // nolint G106
	}, nil
}

// PasswordKey returns the ssh.ClientConfig based on specified username and password.
func PasswordKey(username, password string) *ssh.ClientConfig {
	return &ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // nolint G106
	}
}
