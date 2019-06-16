package output

import (
	"fmt"

	"github.com/logrusorgru/aurora"
)

const (
	ANSISuccess = "\x1b[1m\x1b[49m\x1b[32m"
	ANSIWarn    = "\x1b[1m\x1b[49m\x1b[33m"
	ANSIError   = "\x1b[1m\x1b[49m\x1b[31m"
	ANSIInfo    = "\x1b[1m\x1b[49m\x1b[34m"
)

var au aurora.Aurora

func init() {
	au = aurora.NewAurora(true)

}

func ColorDisable() {
	au = aurora.NewAurora(false)
}

func Italic(s string) string {
	return au.Italic(s).String()
}

func ColorFmt(ansiColor, format, suffix string) string {
	return fmt.Sprintf("%s%s\x1b[0m%s", ansiColor, format, suffix)
}
