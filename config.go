package log4go

import (
	"os"
	"bufio"
	"io"
	"strings"
	"bytes"
)

type Config struct {
	Mapping map[string]string
}

func (c *Config) InitConfig(path string) {
	c.Mapping = make(map[string]string)
	
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	
	var section string
	
	r := bufio.NewReader(f)
	for {
		b, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		s := strings.TrimSpace(string(b))
		
		if strings.HasPrefix(s, "#") || strings.HasPrefix(s, ";") {
			continue
		}
		
		var key bytes.Buffer
		
		if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
			section = strings.TrimSpace(s[1:len(s)-1])
		}
		key.WriteString(section)
		
		index := strings.Index(s, "=")
		
		if index <= 0 {
			continue
		}
		
		first := strings.TrimSpace(s[:index])
		key.WriteString(".")
		key.WriteString(first)
		
		second := strings.TrimSpace(s[index+1:])
		
		c.Mapping[key.String()] = second
	}
}

func (c Config) Item(key string) string {
	return c.Mapping[key]
}