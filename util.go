package log4go

import (
	"os"
)

func FileExist(file string) bool {
	_, e := os.Stat(file)
	return e == nil || os.IsExist(e)
}
