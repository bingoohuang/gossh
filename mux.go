package gossh

import (
	"io"
	"log"
	"strings"
)

func mux(logger *log.Logger, cmd []string, w io.Writer, r io.Reader) {
	var buf [65 * 1024]byte

	lastCmd := ""
	last := ""

	for {
		t, err := r.Read(buf[:])
		if err != nil {
			logger.Print(last)

			if err != io.EOF {
				logger.Println(err)
			}

			return
		}

		sbuf, lastTwo := parseBuf(t, buf[:])
		switch lastTwo {
		case "$ ", "# ":
			if lastCmd == "" {
				a := GetLastLine(last + sbuf)
				logger.Print(a)
			} else {
				lastCmdOut := last + sbuf

				if !strings.Contains(lastCmdOut, lastCmd+"\r\n") {
					logger.Println(lastCmd)
				}

				logger.Print(lastCmdOut)
			}

			last = ""

			if len(cmd) == 0 {
				return
			}

			lastCmd = cmd[0]
			_, _ = w.Write([]byte(lastCmd + "\n"))
			cmd = cmd[1:]
		default:
			last += sbuf
		}
	}
}

// GetLastLine gets the last line of s.
func GetLastLine(s string) string {
	pos := strings.LastIndex(s, "\n")
	if pos < 0 || pos == len(s)-1 {
		return s
	}

	return s[pos+1:]
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
