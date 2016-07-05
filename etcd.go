package configuration

import (
	"fmt"
	"github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/coreos/etcd/client"
	"strings"
	"time"
)

type EtcdProvider struct {
	host   string
	dir    string
	buffer *TreeBuffer
}

func urlParse(s string) (proto, host, path string) {
	kvs := strings.Split(s, "://")
	if len(kvs) < 2 {
		return kvs[0], "", "/"
	} else {
		proto = kvs[0]
		kkvs := strings.SplitN(kvs[1], "/", 2)
		host = kkvs[0]
		if len(kkvs) > 1 {
			path = "/" + kkvs[1]
		} else {
			path = "/"
		}
		return
	}
}

func NewEtcdProvider(cp string) *EtcdProvider {
	_, host, path := urlParse(cp)
	return &EtcdProvider{
		host: host,
		dir:  path,
	}
}

func (e *EtcdProvider) loadEtcd() (*TreeBuffer, error) {
	buffer := NewTreeBuffer()
	cfg := client.Config{
		Endpoints:               []string{e.host},
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second * 3,
	}
	c, err := client.New(cfg)
	if err != nil {
		return nil, err
	}
	api := client.NewKeysAPI(c)
	opt := client.GetOptions{Recursive: true}
	resp, err := api.Get(context.Background(), e.dir, &opt)
	if err != nil {
		if !client.IsKeyNotFound(err) {
			return nil, err
		}
	} else {
		if !resp.Node.Dir {
			return nil, fmt.Errorf("etcd provider isn't a directory")
		} else {
			e.setRecursiveDir(buffer, resp.Node)
		}
	}
	envBuffer, _ := NewEnvProvider().GetBuffer()
	buffer.MergeFrom(envBuffer, false)
	e.buffer = buffer
	return e.buffer, nil
}

func (e *EtcdProvider) setRecursiveDir(buffer *TreeBuffer, node *client.Node) {
	for _, n := range node.Nodes {
		if n == nil {
			continue
		}
		if n.Dir {
			e.setRecursiveDir(buffer, n)
		} else {
			k := strings.Replace(n.Key, "/", ".", -1)
			buffer.Set(k, n.Value)
		}
	}
}

func (f *EtcdProvider) GetBuffer() (*TreeBuffer, error) {
	if f.buffer == nil {
		return f.loadEtcd()
	} else {
		return f.buffer, nil
	}
}
