package zookeeper

import (
	"log"
	"time"

	"github.com/amsalt/cluster/resolver"
	"github.com/samuel/go-zookeeper/zk"
)

type zookeeper struct {
	*resolver.BaseResolver

	hosts   []string
	timeout time.Duration
	conn    *zk.Conn
	root    string
}

// NewZookeeper creates zookeeper client as service-discovery component.
func NewZookeeper(hosts []string, timeout time.Duration, root string) resolver.Resolver {
	z := new(zookeeper)
	z.hosts = hosts
	z.timeout = timeout
	z.root = root
	z.BaseResolver = resolver.NewBaseResolver()
	z.init()
	return z
}

// Register register service.
func (z *zookeeper) Register(servType, host string) {
	err := z.ensurePath(z.root + "/" + servType)
	if err == nil {
		_, err = z.conn.Create(z.root+"/"+servType+"/"+host, nil, zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	}
	return
}

// Resolve return service list.
func (z *zookeeper) Resolve(servType string) (list []string, err error) {
	list, _, err = z.conn.Children(z.root + "/" + servType)
	return
}

func (z *zookeeper) init() {
	conn, _, err := zk.Connect(z.hosts, z.timeout)
	if err != nil {
		log.Printf("zkPlugin init error: %v", err)
		return
	}

	z.conn = conn
	z.ensurePath(z.root)
}

func (z *zookeeper) ensurePath(path string) error {
	exists, _, err := z.conn.Exists(path)
	if err != nil {
		return err
	}
	if !exists {
		_, err := z.conn.Create(path, []byte(""), 0, zk.WorldACL(zk.PermAll)) // todo:
		if err != nil && err != zk.ErrNodeExists {
			return err
		}
	}
	return nil
}
