package internal

import (
	"fmt"
	"os"

	"github.com/pressly/goose/v3"
)

type gooseLogger struct{}

func NewGooseLogger() goose.Logger {
	return &gooseLogger{}
}

func (l *gooseLogger) Fatalf(format string, v ...any) {
	l.Printf(format, v...)
	os.Exit(1)
}

func (*gooseLogger) Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}
