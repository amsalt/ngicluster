package cluster

import (
	"github.com/amsalt/cluster/balancer"
	_ "github.com/amsalt/cluster/balancer/loader" // for import
	_ "github.com/amsalt/cluster/balancer/random"
	_ "github.com/amsalt/cluster/balancer/roundrobin"
	"github.com/amsalt/cluster/resolver"
	"github.com/amsalt/nginet/core"
)

type Cluster struct {
	resolver     resolver.Resolver
	clientMgr    *clientMgr
	balancerName string
	router       *relayRouter
}

func NewCluster(rsv resolver.Resolver, b ...string) *Cluster {
	cluster := new(Cluster)
	cluster.resolver = rsv

	cluster.clientMgr = newClientMgr(rsv)
	cluster.clientMgr.Start()
	cluster.router = newRelayRouter()

	if len(b) > 0 {
		cluster.balancerName = b[0]
	} else {
		cluster.balancerName = "roundrobin"
	}

	if cluster.balancerName == "stickiness" {
		panic("can't use `stickiness` as the default name, parameters insufficient")
	}

	return cluster
}

func (cluster *Cluster) NewServer(servType string) *Server {
	s := NewServer(servType, cluster.resolver)
	return s
}

func (cluster *Cluster) NewServerWithConfig(servType string, readBuf, writeBuf, maxConn int) *Server {
	s := NewServerWithConfig(servType, cluster.resolver, readBuf, writeBuf, maxConn)
	return s
}

func (cluster *Cluster) AddClient(servType string, c *Client, b ...balancer.Balancer) {
	if len(b) > 0 {
		cluster.clientMgr.RegisterClient(servType, c, b[0])
	} else {
		defaultBalancer := balancer.GetBuilder(cluster.balancerName).Build(balancer.WithServName(servType), balancer.WithResolver(cluster.resolver))
		cluster.clientMgr.RegisterClient(servType, c, defaultBalancer)
	}
}

func (cluster *Cluster) Clients(servType string) []core.SubChannel {
	return cluster.clientMgr.Channels(servType)
}

func (cluster *Cluster) Write(servType string, msg interface{}, ctx ...interface{}) error {
	if len(ctx) > 0 {
		return cluster.clientMgr.Write(servType, msg, ctx[0])
	}
	return cluster.clientMgr.Write(servType, msg, nil)
}

func (cluster *Cluster) Register(msgID interface{}, servType string) {
	cluster.router.Register(msgID, servType)
}

func (cluster *Cluster) Route(msgID interface{}) string {
	return cluster.router.Route(msgID)
}
