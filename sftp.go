package gossh

import (
	"fmt"

	"github.com/pkg/sftp"
)

// GetSftpClient get sftClient by host.
func (h *Host) GetSftpClient() (*sftp.Client, error) {
	if h.sftpClient != nil {
		return h.sftpClient, nil
	}

	gc, err := h.GetGosshConnect()
	if err != nil {
		return nil, err
	}

	sf, err := sftp.NewClient(gc.Client)
	if err != nil {
		return nil, fmt.Errorf("sftp.NewClient failed: %w", err)
	}

	h.sftpClient = sf
	h.sftpSSHClient = gc.Client

	return h.sftpClient, nil
}
