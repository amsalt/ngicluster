package etcd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/amsalt/cluster/resolver"
	"go.etcd.io/etcd/client"
)

// Etcd represents a etcd wrapper for Resolver service.
type Etcd struct {
	*resolver.BaseResolver
	timeout   time.Duration
	apiClient client.KeysAPI
	root      string
}

func NewEtcdResolver(root string, hosts []string, timeout time.Duration) resolver.Resolver {
	cfg := client.Config{
		Endpoints:               hosts,
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	kapi := client.NewKeysAPI(c)

	etcd := new(Etcd)
	etcd.BaseResolver = resolver.NewBaseResolver()
	etcd.timeout = timeout
	etcd.apiClient = kapi
	if etcd.root == "" {
		etcd.root = root
	}

	return etcd
}

func (etcd *Etcd) Register(serverType, host string) {
	go etcd.heartbeat(serverType, host)
}

func (etcd *Etcd) Resolve(serverType string) (list []string, err error) {
	resp, err := etcd.apiClient.Get(context.Background(), etcd.root+"/"+serverType, &client.GetOptions{
		Recursive: true,
	})

	if err == nil && resp.Node != nil {
		for _, node := range resp.Node.Nodes {
			list = append(list, node.Value)
		}
		return list, nil
	}
	return nil, err
}

func (etcd *Etcd) heartbeat(serverType, host string) {
	path := etcd.root + "/" + serverType + "/" + host
	for {
		_, err := etcd.apiClient.Set(context.Background(), path, host, &client.SetOptions{
			TTL: time.Second * 10,
		})
		if err != nil {
			fmt.Printf("EtcdPlugin heartbeat error: %v", err)
		}
		time.Sleep(time.Second * 4)
	}
}
