package configuration

import (
	"io/ioutil"
	"strings"
)

type FileProvider struct {
	filename string
	buffer   *TreeBuffer
}

func NewFileProvider(filename string) *FileProvider {
	return &FileProvider{filename: filename}
}

func (f *FileProvider) loadFile() (*TreeBuffer, error) {
	buffer := NewTreeBuffer()
	bts, err := ioutil.ReadFile(f.filename)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(bts), "\n")
	for _, l := range lines {
		l = strings.TrimRight(l, "\r")
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "#") {
			continue
		}
		if strings.HasPrefix(l, "[") && strings.HasSuffix(l, "]") {
			continue
		}
		kvs := strings.SplitN(l, "=", 2)
		k := strings.ToLower(strings.TrimSpace(kvs[0]))
		v := ""
		if len(kvs) > 1 {
			v = strings.TrimSpace(kvs[1])
		}
		buffer.Set(k, v)
	}
	envBuffer, _ := NewEnvProvider().GetBuffer()
	buffer.MergeFrom(envBuffer, false)
	f.buffer = buffer
	return f.buffer, nil
}

func (f *FileProvider) GetBuffer() (*TreeBuffer, error) {
	if f.buffer == nil {
		return f.loadFile()
	} else {
		return f.buffer, nil
	}
}
