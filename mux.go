package gossh

import (
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
)

// CmdChanClosed represents the cmd channel closed event.
type CmdChanClosed struct{}

func mux(cmdsChan chan string, executed chan interface{}, w io.Writer, r io.Reader) {
	uuidStr := uuid.New().String()
	testEcho := "echo " + uuidStr

	runner := &muxRunner{
		r:        r,
		w:        w,
		cmdsChan: cmdsChan,
		executed: executed,

		uuidStr:  uuidStr,
		testEcho: testEcho,
	}

	for {
		if err := runner.read(); err != nil {
			return
		}

		if !runner.exec(runner.parseBuf()) {
			return
		}
	}
}

type muxRunner struct {
	r        io.Reader
	w        io.Writer
	cmdsChan chan string
	executed chan interface{}

	lastCmd       string
	last          string
	buf           [65 * 1024]byte
	readN         int
	testEcho      string
	testEchoState EchoState
	uuidStr       string
}

// EchoState 回显状态
type EchoState int

const (
	// EchoStateInit 初始化，未知
	EchoStateInit EchoState = iota
	// EchoStateSent 已发送
	EchoStateSent
	// EchoStateFound 服务器回显
	EchoStateFound
	// EchoStateNotFound 服务器没有回显
	EchoStateNotFound
)

func (s *muxRunner) read() (err error) {
	s.readN, err = s.r.Read(s.buf[:])
	if err != nil {
		fmt.Print(s.last)

		if err != io.EOF {
			fmt.Println(err.Error())
		}

		s.executed <- err
	}

	return err
}

// nolint gomnd
func (s *muxRunner) parseBuf() (recv, lastTwo string) {
	if s.readN > 0 {
		recv = string(s.buf[:s.readN])
	}

	if len(recv) > 2 {
		lastTwo = recv[s.readN-2:]
	}

	return
}

func (s *muxRunner) exec(recv, lastTwo string) bool {
	if !isPrompt(lastTwo) {
		s.last += recv
		return true
	}

	if s.testEchoState == EchoStateSent {
		uuidCount := strings.Count(recv, s.uuidStr)
		if uuidCount == 2 { // nolint gomnd
			// 有回显，包括命令中的uuid和执行结果的uuid共2处
			s.testEchoState = EchoStateFound
		} else {
			s.testEchoState = EchoStateNotFound
		}
	} else {
		preLines, curLine := GetLastLine(s.last + recv)
		if preLines != "" {
			fmt.Print(preLines)
		}

		if s.lastCmd != "" {
			s.executed <- s.lastCmd
		}

		s.last = curLine

		if s.testEchoState == EchoStateInit {
			_, _ = s.w.Write([]byte(s.testEcho + "\n"))
			s.testEchoState = EchoStateSent

			return true
		}
	}

	ok := false

	s.lastCmd, ok = <-s.cmdsChan
	if !ok {
		s.executed <- CmdChanClosed{}

		return false
	}

	if s.testEchoState == EchoStateNotFound {
		fmt.Println(s.last + s.lastCmd)
		s.last = ""
	}

	_, _ = s.w.Write([]byte(s.lastCmd + "\n"))

	return true
}

func isPrompt(s string) bool {
	switch s {
	case "$ ", "# ":
		return true
	}

	return false
}

// GetLastLine gets the last line of s.
func GetLastLine(s string) (preLines, curLine string) {
	pos := strings.LastIndex(s, "\n")
	if pos < 0 || pos == len(s)-1 {
		return preLines, s
	}

	return s[:pos+1], s[pos+1:]
}
