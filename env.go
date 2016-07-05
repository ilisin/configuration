package configuration

import (
	"os"
	"strings"
)

type EnvProvider struct {
	buffer *TreeBuffer
}

func NewEnvProvider() *EnvProvider {
	return &EnvProvider{}
}

func (f *EnvProvider) loadEnv() *TreeBuffer {
	f.buffer = NewTreeBuffer()
	envs := os.Environ()
	for _, env := range envs {
		kvs := strings.SplitN(env, "=", 2)
		if len(kvs) == 2 {
			f.buffer.Set(kvs[0], kvs[1])
		} else {
			f.buffer.Set(kvs[0], "")
		}
	}
	return f.buffer
}

func (f *EnvProvider) GetBuffer() (*TreeBuffer, error) {
	if f.buffer == nil {
		return f.loadEnv(), nil
	} else {
		return f.buffer, nil
	}
}
