package cluster

import (
	"github.com/amsalt/cluster/balancer"
	"github.com/amsalt/cluster/resolver"
)

type Cluster struct {
	resolver        resolver.Resolver
	clientMgr       *clientMgr
	defaultBalancer balancer.Balancer
}

func NewCluster(rsv resolver.Resolver, b ...balancer.Balancer) *Cluster {
	cluster := new(Cluster)
	cluster.resolver = rsv

	if len(b) > 0 {
		cluster.defaultBalancer = b[0]
	}

	cluster.clientMgr = newClientMgr(rsv)
	cluster.clientMgr.Start()

	return cluster
}

func (cluster *Cluster) NewServer(servType string) *Server {
	s := NewServer(servType, cluster.resolver)
	return s
}

func (cluster *Cluster) AddClient(servType string, c *Client, b ...balancer.Balancer) {
	if len(b) > 0 {
		cluster.clientMgr.RegisterClient(servType, c, b[0])
	} else {
		cluster.clientMgr.RegisterClient(servType, c, cluster.defaultBalancer)
	}
}
