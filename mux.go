package gossh

import (
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
)

// CmdChanClosed represents the cmd channel closed event.
type CmdChanClosed struct{}

func mux(cmdsCh chan CmdWrap, executedCh chan interface{}, w io.Writer, r io.Reader, h *Host, stdout io.Writer) {
	uuidStr := uuid.New().String()
	testEcho := "echo " + uuidStr

	runner := &muxRunner{
		r:          r,
		w:          w,
		cmdsCh:     cmdsCh,
		executedCh: executedCh,

		uuidStr:  uuidStr,
		testEcho: testEcho,

		host: h,
		out:  stdout,
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
	w          io.Writer
	out        io.Writer
	r          io.Reader
	cmdsCh     chan CmdWrap
	executedCh chan interface{}

	host     *Host
	uuidStr  string
	last     string
	testEcho string

	lastCmd       CmdWrap
	testEchoState EchoState
	readN         int
	buf           [65 * 1024]byte
}

// EchoState 回显状态.
type EchoState int

const (
	// EchoStateInit 初始化，未知.
	EchoStateInit EchoState = iota
	// EchoStateSent 已发送.
	EchoStateSent
	// EchoStateFound 服务器回显.
	EchoStateFound
	// EchoStateNotFound 服务器没有回显.
	EchoStateNotFound
)

func (s *muxRunner) read() (err error) {
	s.readN, err = s.r.Read(s.buf[:])
	if err != nil {
		fmt.Print(s.last)

		if err != io.EOF {
			fmt.Println(err.Error())
		}

		s.executedCh <- err
	}

	return err
}

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

	newFound := false
	if s.testEchoState == EchoStateSent {
		uuidCount := strings.Count(s.last+recv, s.uuidStr)
		newFound = uuidCount >= 1
		if uuidCount >= 2 {
			// 有回显，包括命令中的uuid和执行结果的uuid共2处
			s.testEchoState = EchoStateFound
		} else {
			s.testEchoState = EchoStateNotFound
		}
	}

	s.last = s.setResult(recv, newFound)

	if s.testEchoState == EchoStateInit {
		_, _ = s.w.Write([]byte(s.testEcho + "\n"))
		s.testEchoState = EchoStateSent
		return true
	}

	ok := false
	s.lastCmd, ok = <-s.cmdsCh
	if !ok {
		s.executedCh <- CmdChanClosed{}
		return false
	}

	if s.testEchoState == EchoStateNotFound {
		_, _ = fmt.Fprintln(s.out, StripAnsi(s.last+s.lastCmd.Cmd))
		s.last = ""
	}

	_, _ = s.w.Write([]byte(s.lastCmd.Cmd + "\n"))

	return true
}

func (s *muxRunner) setResult(recv string, newFound bool) string {
	preLines, curLine := GetLastLine(s.last + recv)
	if preLines != "" && !newFound {
		if p := strings.Index(preLines, "Last login:"); p > 0 {
			preLines = preLines[p:]
		}
		_, _ = fmt.Fprint(s.out, StripAnsi(preLines))
	}

	if s.lastCmd.Cmd != "" {
		s.executedCh <- s.lastCmd
		_, result := GetLastLine(strings.TrimSpace(preLines))
		s.host.SetResultVar(s.lastCmd.ResultVar, result)
	}
	return curLine
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
