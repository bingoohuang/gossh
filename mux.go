package gossh

import (
	"fmt"
	"io"
	"strings"
)

// CmdChanClosed represents the cmd channel closed event.
type CmdChanClosed struct{}

func mux(cmdsChan chan string, executed chan interface{}, w io.Writer, r io.Reader) {
	var (
		ok      bool
		lastCmd string
		last    string
		buf     [65 * 1024]byte
	)

	for {
		t, err := r.Read(buf[:])
		if err != nil {
			fmt.Print(last)

			if err != io.EOF {
				fmt.Println(err.Error())
			}

			executed <- err

			return
		}

		sbuf, lastTwo := parseBuf(t, buf[:])
		switch lastTwo {
		case "$ ", "# ":
			preLines, curLine := GetLastLine(last + sbuf)
			if preLines != "" {
				fmt.Print(preLines)
			}

			if lastCmd != "" {
				executed <- lastCmd
			}

			last = curLine

			lastCmd, ok = <-cmdsChan
			if !ok {
				executed <- CmdChanClosed{}

				return
			}

			_, _ = w.Write([]byte(lastCmd + "\n"))
		default:
			last += sbuf
		}
	}
}

// GetLastLine gets the last line of s.
func GetLastLine(s string) (preLines, curLine string) {
	pos := strings.LastIndex(s, "\n")
	if pos < 0 || pos == len(s)-1 {
		curLine = s
	} else {
		preLines = s[:pos+1]
		curLine = s[pos+1:]
	}

	return preLines, curLine
}

// nolint gomnd
func parseBuf(t int, buf []byte) (sbuf, lastTwo string) {
	if t > 0 {
		sbuf = string(buf[:t])
	}

	if len(sbuf) > 2 {
		lastTwo = sbuf[t-2:]
	}

	return
}
