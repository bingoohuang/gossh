package gossh

import (
	"io/ioutil"

	"golang.org/x/crypto/ssh"
)

//PrivateKey Loads a private and public key from "path" and returns a SSH ClientConfig to authenticate with the server
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

func PrivateKeyWithPassphrase(username string, passpharase []byte, path string) (*ssh.ClientConfig, error) {
	privateKey, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKeyWithPassphrase(privateKey, passpharase)
	if err != nil {
		return nil, err
	}

	return &ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // nolint G106
	}, nil
}

func PasswordKey(username, password string) (*ssh.ClientConfig, error) {
	return &ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // nolint G106
	}, nil
}
