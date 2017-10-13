package log4go

import (
	"net"
	"strings"
	"strconv"
	"fmt"
	"os"
)

type NetWriter struct {
	rec chan *Record
	
	host string
	port int
	
	sock net.Conn
}

func (w *NetWriter) Write(r *Record) {
	minLevel, find := strToLevel[strings.ToUpper(logger.Item("console.level"))]
	if find && int(r.Level) < int(minLevel) {
		return
	}
	w.rec <- r
}

func (w *NetWriter) Close() {
	close(w.rec)
}

func NewNetWriter() *NetWriter {
	format := logger.Item("console.format")
	if len(format) == 0 {
		format = logger.Item("log4go.format")
	}
	writer := new(NetWriter)
	writer.rec = make(chan *Record, logBufferLength)
	writer.host = logger.Item("console.host")
	writer.port, _ = strconv.Atoi(logger.Item("console.port"))
	address := writer.host + string(writer.port)
	conn, err := net.Dial("tcp", address)
	writer.sock = conn
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewNetWriter(%q): %s\n", address, err)
	}
	wg.Add(1)
	go writer.run()
	return writer
}

func (w *NetWriter) run() {

}