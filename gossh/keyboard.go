package gossh

import (
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var KeyText = map[string]rune{
	"CtrlA": KeyCtrlA, "CtrlB": KeyCtrlB, "CtrlC": KeyCtrlC, "CtrlD": KeyCtrlD, "CtrlE": KeyCtrlE, "CtrlF": KeyCtrlF,
	"CtrlG": KeyCtrlG, "CtrlH": KeyCtrlH, "CtrlI": KeyCtrlI, "CtrlJ": KeyCtrlJ, "CtrlK": KeyCtrlK, "CtrlL": KeyCtrlL,
	"CtrlM": KeyCtrlM, "CtrlN": KeyCtrlN, "CtrlO": KeyCtrlO, "CtrlP": KeyCtrlP, "CtrlQ": KeyCtrlQ, "CtrlR": KeyCtrlR,
	"CtrlS": KeyCtrlS, "CtrlT": KeyCtrlT, "CtrlU": KeyCtrlU, "CtrlV": KeyCtrlV, "CtrlW": KeyCtrlW, "CtrlX": KeyCtrlX,
	"CtrlY": KeyCtrlY, "CtrlZ": KeyCtrlZ, "Escape": KeyEscape, "LeftBracket": KeyLeftBracket,
	"RightBracket": KeyRightBracket, "Enter": KeyEnter, "N": KeyEnter, "Backspace": KeyBackspace,
	"Unknown": KeyUnknown, "Up": KeyUp, "Down": KeyDown, "Left": KeyLeft, "Right": KeyRight,
	"Home": KeyHome, "End": KeyEnd, "PasteStart": KeyPasteStart, "PasteEnd": KeyPasteEnd, "Insert": KeyInsert,
	"Del": KeyDelete, "PgUp": KeyPgUp, "PgDn": KeyPgDn, "Pause": KeyPause,
	"F1": KeyF1, "F2": KeyF2, "F3": KeyF3, "F4": KeyF4, "F5": KeyF5, "F6": KeyF6, "F7": KeyF7,
	"F8": KeyF8, "F9": KeyF9, "F10": KeyF10, "F11": KeyF11, "F12": KeyF12,
}

func ExecuteInitialCmd(initialCmd string, w io.Writer, r *TryReader) {
	for _, v := range ConvertKeys(initialCmd) {
		time.Sleep(100 * time.Millisecond)
		_, _ = w.Write(v)
	}
}

var numReg = regexp.MustCompile(`^\d+`)

func ConvertKeys(s string) [][]byte {
	groups := make([][]byte, 0)

	for s != "" {
		start := strings.Index(s, "{")
		end := strings.Index(s, "}")
		if start < 0 || end < 0 || start > end {
			groups = append(groups, []byte(s))
			break
		}

		if start > 0 {
			groups = append(groups, []byte(s[:start]))
		}

		key := strings.TrimSpace(s[start+1 : end])
		num := 1
		if numStr := numReg.FindString(key); numStr != "" {
			num, _ = strconv.Atoi(numStr)
			key = key[len(numStr):]
		}
		for k, v := range KeyText {
			if strings.EqualFold(k, key) {
				vbytes := []byte(string([]rune{v}))
				for i := 0; i < num; i++ {
					groups = append(groups, vbytes)
				}
				break
			}
		}

		s = s[end+1:]
	}

	return groups
}

// Giant list of key constants.  Everything above KeyUnknown matches an actual
// ASCII key value.  After that, we have various pseudo-keys in order to
// represent complex byte sequences that correspond to keys like Page up, Right
// arrow, etc.
const (
	KeyCtrlA = 1 + iota
	KeyCtrlB
	KeyCtrlC
	KeyCtrlD
	KeyCtrlE
	KeyCtrlF
	KeyCtrlG
	KeyCtrlH
	KeyCtrlI
	KeyCtrlJ
	KeyCtrlK
	KeyCtrlL
	KeyCtrlM
	KeyCtrlN
	KeyCtrlO
	KeyCtrlP
	KeyCtrlQ
	KeyCtrlR
	KeyCtrlS
	KeyCtrlT
	KeyCtrlU
	KeyCtrlV
	KeyCtrlW
	KeyCtrlX
	KeyCtrlY
	KeyCtrlZ
	KeyEscape
	KeyLeftBracket  = '['
	KeyRightBracket = ']'
	KeyEnter        = '\n'
	KeyBackspace    = 127
	KeyUnknown      = 0xd800 /* UTF-16 surrogate area */ + iota
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyHome
	KeyEnd
	KeyPasteStart
	KeyPasteEnd
	KeyInsert
	KeyDelete
	KeyPgUp
	KeyPgDn
	KeyPause
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
)
