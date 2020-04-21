package gossh

import (
	"fmt"

	"github.com/pkg/sftp"
)

func (h *Host) makeSftpClient() error {
	if h.sftpClient != nil {
		return nil
	}

	gc, err := h.GetGosshConnect()
	if err != nil {
		return err
	}

	sf, err := sftp.NewClient(gc.Client)
	if err != nil {
		return fmt.Errorf("sftp.NewClient failed: %w", err)
	}

	h.sftpClient = sf

	return nil
}

// GetClient get sftClient by host
func (h *Host) GetClient() (*sftp.Client, error) {
	err := h.makeSftpClient()
	if err != nil {
		return nil, err
	}

	return h.sftpClient, nil
}
