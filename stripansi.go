package gossh

import "regexp"

// https://superuser.com/questions/380772/removing-ansi-color-codes-from-text-stream
// http://ascii-table.com/ansi-escape-sequences.php
// https://github.com/acarl005/stripansi/blob/master/stripansi.go
//const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
//var re = regexp.MustCompile(ansi)

// https://superuser.com/questions/380772/removing-ansi-color-codes-from-text-stream
// sed 's/\x1b\[[0-9;]*m//g'           # Remove color sequences only
// sed 's/\x1b\[[0-9;]*[a-zA-Z]//g'    # Remove all escape sequences
// sed 's/\x1b\[[0-9;]*[mGKH]//g'      # Remove color and move sequences

var re = regexp.MustCompile("\u001b\\[[0-9;]*[A-Ksu]") // nolint:gochecknoglobals

// StripAnsi strips the cursor, clears, and save positions escape code.
// https://github.com/pborman/ansi/blob/master/ansi.go
// https://www.lihaoyi.com/post/BuildyourownCommandLinewithANSIescapecodes.html
// Up: \u001b[{n}A moves cursor up by n
// Down: \u001b[{n}B moves cursor down by n
// Right: \u001b[{n}C moves cursor right by n
// Left: \u001b[{n}D moves cursor left by n
// Next Line: \u001b[{n}E moves cursor to beginning of line n lines down
// Prev Line: \u001b[{n}F moves cursor to beginning of line n lines down
// Set Column: \u001b[{n}G moves cursor to column n
// Set Position: \u001b[{n};{m}H moves cursor to row n column m
// Clear Screen: \u001b[{n}J clears the screen
// n=0 clears from cursor until end of screen,
// n=1 clears from cursor to beginning of screen
// n=2 clears entire screen
// Clear Line: \u001b[{n}K clears the current line
// n=0 clears from cursor to end of line
// n=1 clears from cursor to start of line
// n=2 clears entire line
// Save Position: \u001b[{s} saves the current cursor position
// Save Position: \u001b[{u} restores the cursor to the last saved position
func StripAnsi(str string) string {
	return re.ReplaceAllString(str, "")
}
