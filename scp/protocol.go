package scp

import (
	"bufio"
	"io"
)

// ResponseType represents the type of the response.
type ResponseType = uint8

const (
	// OK represents everything is OK
	OK ResponseType = iota + 1
	// Warning represents there is warning.
	Warning
	// Error means there is error.
	Error
)

// Response has tree types of responses that the remote can send back:
// ok, warning and error
//
// The difference between warning and error is that the connection is not closed by the remote,
// however, a warning can indicate a file transfer failure (such as invalid destination directory)
// and such be handled as such.
//
// All responses except for the `OK` type always have a message (although these can be empty)
//
// The remote sends a confirmation after every SCP command, because a failure can occur after every
// command, the response should be read and checked after sending them.
type Response struct {
	Type    ResponseType
	Message string
}

// ParseResponse reads from the given reader (assuming it is the output of the remote)
// and parses it into a Response structure
func ParseResponse(reader io.Reader) (rsp Response, err error) {
	buffer := make([]uint8, 1)
	if _, err = reader.Read(buffer); err != nil {
		return
	}

	rsp.Type = buffer[0]

	if rsp.Type > 0 {
		bufferedReader := bufio.NewReader(reader)
		if rsp.Message, err = bufferedReader.ReadString('\n'); err != nil {
			return
		}
	}

	return
}

// IsOK tells the response is OK or not.
func (r *Response) IsOK() bool { return r.Type == OK }

// IsWarning tells the response has warning or not.
func (r *Response) IsWarning() bool { return r.Type == Warning }

// IsError tells if there is error in the response.
func (r *Response) IsError() bool { return r.Type == Error }

// IsFailure tells if the response if none OK or not.
func (r *Response) IsFailure() bool { return r.Type > 0 }

// GetMessage returns the detailed message.
func (r *Response) GetMessage() string { return r.Message }
