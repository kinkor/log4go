package log4go

import (
	"fmt"
	"io"
	"os"
	"strings"
)

var stdout io.Writer = os.Stdout

type ConsoleWriter struct {
	rec    chan *Record
	format string
}

func NewConsoleWriter() *ConsoleWriter {
	format := logger.Item("console.format")
	if len(format) == 0 {
		format = logger.Item("log4go.format")
	}
	writer := new(ConsoleWriter)
	writer.rec = make(chan *Record, logBufferLength)
	writer.format = format
	wg.Add(1)
	go writer.run(stdout)
	return writer
}

func (w *ConsoleWriter) Write(r *Record) {
	minLevel, find := strToLevel[strings.ToUpper(logger.Item("console.level"))]
	if find && int(r.Level) < int(minLevel) {
		return
	}
	w.rec <- r
}

func (w *ConsoleWriter) Close() {
	close(w.rec)
}

func (w *ConsoleWriter) run(out io.Writer) {
	for rec := range w.rec {
		fmt.Fprintln(out, FormatLog(w.format, rec))
	}
	wg.Done()
}
