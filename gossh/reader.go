package gossh

import (
	"io"
)

type TryReader struct {
	data chan []byte
	err  error
	r    io.Reader
}

func (c *TryReader) begin() {
	buf := make([]byte, 1024)
	for {
		n, err := c.r.Read(buf)
		if n > 0 {
			tmp := make([]byte, n)
			copy(tmp, buf[:n])
			c.data <- tmp
		}
		if err != nil {
			c.err = err
			close(c.data)
			return
		}
	}
}

func (c *TryReader) Read(p []byte) (int, error) {
	d, ok := <-c.data
	if !ok {
		return 0, c.err
	}
	copy(p, d)
	return len(d), nil
}

func (c *TryReader) TryRead(p []byte) (int, error) {
	select {
	case d, ok := <-c.data:
		if !ok {
			return 0, c.err
		}
		copy(p, d)
		return len(d), nil
	default:
		return 0, nil
	}
}

func NewTryReader(r io.Reader) *TryReader {
	c := &TryReader{
		r:    r,
		data: make(chan []byte),
	}
	go c.begin()
	return c
}
