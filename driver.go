package configuration

import (
	"fmt"
	"strings"
)

const DefaultProvider string = "file::./config.ini"

type Driver struct {
	Type         ConfigType
	ContextParam string
	Provider     Provider
}

var confTypeM = map[ConfigType]string{
	CTFileConf: "file",
	//CTEtcd:     "etcd",
	CTEnv:      "env",
}

// parse provider to config type and fetch provider's param
func (this *Driver) ParseProvider(provider string) (err error) {
	this.Type = CTUnkown
	for key, value := range confTypeM {
		if strings.HasPrefix(provider, value+"::") {
			this.Type = key
			this.ContextParam = string(provider[len(value)+2:])
			break
		}
	}
	if this.Type == CTUnkown {
		return fmt.Errorf("Unkown configuration provinder [%s]", provider)
	}
	return
}

// with provider parse config
func (this *Driver) LoadProvider() (Provider, error) {
	// only once
	if this.Provider != nil {
		return this.Provider, nil
	}
	var err error = nil
	if this.Type == CTFileConf {
		this.Provider = NewFileProvider(this.ContextParam)
	//} else if this.Type == CTEtcd {
	//	this.Provider = NewEtcdProvider(this.ContextParam)
	} else if this.Type == CTEnv {
		this.Provider = NewEnvProvider()
	} else {
		return nil, ErrUnkownProvider
	}
	return this.Provider, err
}

func (this *Driver) Buffer() *TreeBuffer {
	b, err := this.Provider.GetBuffer()
	if err != nil {
		panic(err)
	}
	return b
}

var driver *Driver
