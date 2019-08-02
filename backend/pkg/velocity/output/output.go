package output

import (
	"fmt"

	"github.com/logrusorgru/aurora"
)

var (
	ANSISuccess = aurora.GreenFg
	ANSIWarn    = aurora.YellowFg
	ANSIError   = aurora.RedFg
	ANSIInfo    = aurora.BlueFg
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

func ColorFmt(ansiColor aurora.Color, format, suffix string) string {
	return fmt.Sprintf("%s%s", au.Colorize(format, ansiColor), suffix)
}
