package gossh

import (
	"io/ioutil"
	"time"

	"github.com/spf13/viper"

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

	auth := []ssh.AuthMethod{ssh.PublicKeys(signer)}

	return MakeClientConfig(username, auth), nil
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

	auth := []ssh.AuthMethod{ssh.PublicKeys(signer)}

	return MakeClientConfig(username, auth), nil
}

// PasswordKey returns the ssh.ClientConfig based on specified username and password.
func PasswordKey(username, password string) *ssh.ClientConfig {
	auth := []ssh.AuthMethod{ssh.Password(password)}

	return MakeClientConfig(username, auth)
}

// MakeClientConfig makes a new ssh.ClientConfig.
func MakeClientConfig(username string, auth []ssh.AuthMethod) *ssh.ClientConfig {
	timeout := viper.Get("NetTimeout").(time.Duration)

	return &ssh.ClientConfig{
		User:            username,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // nolint:G106
		Timeout:         timeout,
	}
}
