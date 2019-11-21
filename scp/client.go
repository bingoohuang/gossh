package scp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"

	"github.com/bingoohuang/gossh/gossh"

	"golang.org/x/crypto/ssh"
)

// Client represents a client of scp session.
type Client struct {
	// the client config to use
	ClientConfig *ssh.ClientConfig
	// stores the SSH session while the connection is running
	Session *ssh.Session
	// stores the SSH connection itself in order to close it after transfer
	Conn ssh.Conn

	// the clients waits for the given timeout until given up the connection
	Timeout time.Duration
	// the absolute path to the remote SCP binary
	RemoteBinary string
}

// Connect connects to the remote SSH server, returns error if it couldn't establish a session to the SSH server
func (a *Client) Connect(addr string, clientConfig *ssh.ClientConfig) (err error) {
	a.Conn, a.Session, err = gossh.NewSession(addr, clientConfig)

	return
}

// CopyFromFile copies the contents of an os.File to a remote location,
// it will get the length of the file by looking it up from the filesystem
func (a *Client) CopyFromFile(file os.File, remotePath string, permissions string) error {
	stat, err := file.Stat()
	if err != nil {
		return err
	}

	return a.Copy(&file, remotePath, permissions, stat.Size())
}

// CopyFile copies the contents of an io.Reader to a remote location,
// the length is determined by reading the io.Reader until EOF
// if the file length in know in advance please use "Copy" instead
func (a *Client) CopyFile(fileReader io.Reader, remotePath string, permissions string) error {
	contentsBytes, err := ioutil.ReadAll(fileReader)
	if err != nil {
		return err
	}

	bytesReader := bytes.NewReader(contentsBytes)

	return a.Copy(bytesReader, remotePath, permissions, int64(len(contentsBytes)))
}

// waitTimeout waits for the waitgroup for the specified max timeout.
// Returns true if waiting timed out.
func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})

	go func() {
		defer close(c)
		wg.Wait()
	}()

	if timeout > 0 {
		<-c
		return false // completed normally
	}

	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}

// checkResponse checks the response it reads from the remote, and will return a single error in case of failure
func checkResponse(r io.Reader) error {
	if rsp, err := ParseResponse(r); err != nil {
		return err
	} else if rsp.IsFailure() {
		return errors.New(rsp.GetMessage())
	}

	return nil
}

// Copy copies the contents of an io.Reader to a remote location
func (a *Client) Copy(r io.Reader, remotePath, permissions string, size int64) error {
	wg := sync.WaitGroup{}
	wg.Add(2)

	errCh := make(chan error, 2)

	go func() {
		defer wg.Done()
		w, err := a.Session.StdinPipe()
		if err != nil {
			errCh <- err
			return
		}

		defer w.Close()

		stdout, err := a.Session.StdoutPipe()

		if err != nil {
			errCh <- err
			return
		}

		if err := a.copy(w, r, path.Base(remotePath), permissions, size, stdout); err != nil {
			errCh <- err
			return
		}
	}()

	go func() {
		defer wg.Done()
		if err := a.Session.Run(fmt.Sprintf("%s -qt %s", a.RemoteBinary, remotePath)); err != nil {
			errCh <- err
			return
		}
	}()

	if waitTimeout(&wg, a.Timeout) {
		return errors.New("timeout when upload files")
	}

	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Client) copy(w io.Writer, r io.Reader, filename, permissions string, size int64, stdout io.Reader) error {
	if _, err := fmt.Fprintln(w, "C"+permissions, size, filename); err != nil {
		return err
	}

	if err := checkResponse(stdout); err != nil {
		return err
	}

	if _, err := io.Copy(w, r); err != nil {
		return err
	}

	if _, err := fmt.Fprint(w, "\x00"); err != nil {
		return err
	}

	if err := checkResponse(stdout); err != nil {
		return err
	}

	return nil
}

// Close closes the client.
func (a *Client) Close() error {
	if err := a.Session.Close(); err != nil {
		return err
	}

	if err := a.Conn.Close(); err != nil {
		return err
	}

	return nil
}
