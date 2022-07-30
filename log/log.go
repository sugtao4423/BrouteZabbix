package log

import (
	"fmt"
	"log"
	"strings"
)

const (
	ColorInfo  = "\033[0;36m"
	ColorWarn  = "\033[0;33m"
	ColorError = "\033[0;31m"
	ColorClear = "\033[0m"
)

func content(v []any) string {
	return strings.TrimSuffix(fmt.Sprintln(v...), "\n")
}

func Debug(v ...any) {
	log.Println("Debug:", content(v))
}

func Info(v ...any) {
	msg := ColorInfo + "Info: " + content(v) + ColorClear
	log.Println(msg)
}

func Warn(v ...any) {
	msg := ColorWarn + "Warn: " + content(v) + ColorClear
	log.Println(msg)
}

func Error(v ...any) {
	msg := ColorError + "Error: " + content(v) + ColorClear
	log.Println(msg)
}

func Infof(format string, v ...any) {
	msg := ColorInfo + "Info: " + fmt.Sprintf(format, v...) + ColorClear
	log.Println(msg)
}
