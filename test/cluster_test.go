package test

import (
	"testing"
	"time"

	"github.com/amsalt/cluster"
	"github.com/amsalt/cluster/balancer"
	"github.com/amsalt/cluster/balancer/roundrobin"
	"github.com/amsalt/cluster/resolver/static"
	"github.com/amsalt/log"
	"github.com/amsalt/nginet/core"
	"github.com/amsalt/nginet/message"
)

type tcpChannel struct {
	Msg string
}

func TestCluster(t *testing.T) {
	resolver := static.NewConfigBasedResolver()

	register := message.NewRegister()
	register.RegisterMsgByID(1, &tcpChannel{})

	processMgr := message.NewProcessorMgr(register)
	processMgr.RegisterProcessorByID(1, func(ctx *core.ChannelContext, msg interface{}, args ...interface{}) {
		if m, ok := msg.([]byte); ok {
			log.Infof("tcpChannel handler: %+v", string(m))
		} else {
			log.Infof("tcpChannel handler: %+v", msg)
		}
	})

	b := balancer.GetBuilder("roundrobin").Build(roundrobin.WithServName("game"), roundrobin.WithResolver(resolver))
	clus := cluster.NewCluster(resolver, b)

	s := clus.NewServer("game")
	s.InitAcceptor(nil, register, processMgr)
	s.Listen(":7878")
	go s.Accept()

	c := cluster.NewClient()
	c.InitConnector(nil, register, processMgr)
	clus.AddClient("game", c)

	time.Sleep(15 * time.Second)

}
