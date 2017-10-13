package log4go

import (
	"os"
	"fmt"
	"strings"
	"regexp"
	"strconv"
	"time"
)

type FileWriter struct {
	rec chan *Record
	rot chan bool
	
	filename string
	file     *os.File
	
	maxSize     int64 //文件允许最大大小 单位:bytes
	currentSize int64 //文件当前大小 单位:bytes
	
	date       bool
	openDate   string
	dateFormat string
	
	format string
	
	rotate    bool
	maxBackup int
}

func (w *FileWriter) Write(record *Record) {
	minLevel, find := strToLevel[strings.ToUpper(logger.Item("file.level"))]
	if find && int(record.Level) < int(minLevel) {
		return
	}
	w.rec <- record
}

func (w *FileWriter) Close() {
	close(w.rec)
	w.file.Sync()
}

func NewFileWriter() *FileWriter {
	format := logger.Item("file.format")
	if len(format) == 0 {
		format = logger.Item("log4go.format")
	}
	writer := new(FileWriter)
	writer.rec = make(chan *Record, logBufferLength)
	writer.rot = make(chan bool, 1)
	writer.format = format
	
	writer.filename = logger.Item("file.name")
	maxSize := logger.Item("file.size")
	match, _ := regexp.MatchString("^[0-9]+[KMGkmg]?$", maxSize)
	if match {
		size := []byte(maxSize)
		num := size[0:len(size)-1]
		unit := size[len(size)-1]
		factor1, _ := strconv.ParseInt(string(num), 10, 64)
		var factor2 int64
		switch unit {
		case 'K', 'k':
			factor2 = 1024
		case 'M', 'm':
			factor2 = 1024 * 1024
		case 'G', 'g':
			factor2 = 1024 * 1024 * 1024
		default:
			lastNum, _ := strconv.ParseInt(string(unit), 10, 64)
			if lastNum <= 0 {
				lastNum = 1
			}
			factor2 = 10 * lastNum
		}
		writer.maxSize = factor1 * factor2
	} else {
		writer.maxSize = 2 * 1024 * 1024 * 1024
	}
	writer.maxBackup, _ = strconv.Atoi(logger.Item("file.backup"))
	writer.dateFormat = logger.Item("file.date")
	if len(writer.dateFormat) > 0 {
		writer.date = true
	}
	if writer.maxBackup > 0 || writer.date {
		writer.rotate = true
	}
	writer.openDate = FormatTime(writer.dateFormat, time.Now())
	if err := writer.initFileLog(); err != nil {
		fmt.Fprintf(os.Stderr, "FileWriter(%q): %s\n", writer.filename, err)
		return nil
	}
	wg.Add(1)
	go writer.run()
	return writer
}

func (w *FileWriter) run() {
	defer func() {
		if w.file != nil {
			w.file.Close()
		}
		wg.Done()
	}()
	for {
		select {
		case <-w.rot:
			if err := w.initFileLog(); err != nil {
				fmt.Fprintf(os.Stderr, "FileWriter(%q): %s\n", w.filename, err)
				return
			}
		case rec, ok := <-w.rec:
			if !ok {
				return
			}
			if w.currentSize >= w.maxSize || (w.date && w.openDate != FormatTime(w.dateFormat, time.Now())) {
				if err := w.initFileLog(); err != nil {
					fmt.Fprintf(os.Stderr, "FileWriter(%q): %s\n", w.filename, err)
					return
				}
			}
			n, err := fmt.Fprintln(w.file, FormatLog(w.format, rec))
			if err != nil {
				fmt.Fprintf(os.Stderr, "FileWriter(%q): %s\n", w.filename, err)
				return
			}
			w.currentSize += int64(n)
		}
	}
}

func (w *FileWriter) initFileLog() error {
	if w.file != nil {
		w.file.Close()
	}
	if w.rotate {
		if FileExist(w.filename) {
			num := 1
			fName := ""
			
			if w.date { //包含日期
				num = w.maxBackup - 1
				for ; num >= 1; num-- { //最后一个丢弃
					fName = fmt.Sprintf("%s.%s.%d", w.filename, w.openDate, num)
					nfName := fmt.Sprintf("%s.%s.%d", w.filename, w.openDate, num+1)
					if FileExist(fName) {
						err := os.Rename(fName, nfName)
						if err != nil {
							return fmt.Errorf("Rotate: %s ", err)
						}
					}
				}
				
			} else { //不包含日期
				num = w.maxBackup - 1
				for ; num >= 1; num-- { //最后一个丢弃
					fName = fmt.Sprintf("%s.%d", w.filename, num)
					nfName := fmt.Sprintf("%s.%d", w.filename, num+1)
					if FileExist(fName) {
						err := os.Rename(fName, nfName)
						if err != nil {
							return fmt.Errorf("Rotate: %s ", err)
						}
					}
				}
			}
			err := os.Rename(w.filename, fName) //重命名
			if err != nil {
				return fmt.Errorf("Rotate: %s ", err)
			}
		}
	}
	
	fd, err := os.OpenFile(w.filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0664)
	if err != nil {
		return err
	}
	w.file = fd
	w.openDate = FormatTime(w.dateFormat, time.Now())
	w.currentSize = 0
	
	return nil
}
