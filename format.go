package log4go

import (
	"time"
	"bytes"
	"fmt"
	"regexp"
)

/*
%T{Y-m-d H:i:s} 时间 后面是时间格式 可写可不写  不写就是默认Y-m-d H:i:s
%L  日志级别
%M  日志内容
%S  日志来源
*/
func FormatLog(format string, record *Record) string {
	pieces := bytes.Split([]byte(format), []byte{'%'})
	var buffer bytes.Buffer
	for i, piece := range pieces {
		if i > 0 && len(piece) > 0 {
			start := 1
			switch piece[0] {
			case 'T':
				reg := regexp.MustCompile("{(.*)?}")
				loc := reg.FindIndex(piece)
				format := "Y-m-d H:i:s"
				if len(loc) > 0 {
					format = string(piece[loc[0]+1:loc[1]-1])
					start = loc[1]
				}
				buffer.WriteString(FormatTime(format, record.Time))
			case 'L':
				buffer.WriteString(fmt.Sprintf("%-5s", levelStrings[record.Level]))
			case 'M':
				buffer.WriteString(record.Content)
			case 'S':
				buffer.WriteString(record.Source)
			default:
				buffer.Write([]byte{'%', piece[0]})
			}
			buffer.Write(piece[start:])
		} else {
			buffer.Write(piece)
		}
	}
	return buffer.String()
}

/*
Y 年份
m 月份
d 日期
H 24小时
i 分钟
s 秒钟
*/
func FormatTime(format string, t time.Time) string {
	year, month, day, hour, minute, second := t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second()
	var buffer bytes.Buffer
	chars := []byte(format)
	for _, piece := range chars {
		switch piece {
		case 'Y':
			buffer.WriteString(fmt.Sprintf("%4d", year))
		case 'm':
			buffer.WriteString(fmt.Sprintf("%02d", month))
		case 'd':
			buffer.WriteString(fmt.Sprintf("%02d", day))
		case 'H':
			buffer.WriteString(fmt.Sprintf("%02d", hour))
		case 'i':
			buffer.WriteString(fmt.Sprintf("%02d", minute))
		case 's':
			buffer.WriteString(fmt.Sprintf("%02d", second))
		default:
			buffer.WriteByte(piece)
		}
	}
	return buffer.String()
}
