package test

import (
	"testing"
	"time"

	"github.com/amsalt/cluster"
	"github.com/amsalt/cluster/balancer"
	"github.com/amsalt/cluster/balancer/stickiness"
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

	b := balancer.GetBuilder("stickiness").Build(stickiness.WithServName("game"), stickiness.WithResolver(resolver))
	clus := cluster.NewCluster(resolver)

	relayHandler := cluster.NewRelayHandler("game", clus, "userID")
	s := clus.NewServer("game", &cluster.TimeOutOpts{TimeoutSec: 3, PeriodSec: 300})
	s.InitAcceptor(nil, register, processMgr)
	s.AddAfterHandler("IDParser", nil, "RelayHandler", relayHandler)
	s.Listen(":7878")
	go s.Accept()

	s1 := clus.NewServer("game")
	s1.InitAcceptor(nil, register, processMgr)
	s1.Listen(":7979")
	go s1.Accept()

	c := cluster.NewClient()
	c.InitConnector(nil, register, processMgr)

	clus.AddClient("game", c, b)
	time.Sleep(time.Second * 2)

	clus.Write("game", &tcpChannel{Msg: "cluster send message1"})
	clus.Write("game", &tcpChannel{Msg: "cluster send message2"})
	clus.Write("game", &tcpChannel{Msg: "cluster send message3"})
	time.Sleep(15 * time.Second)

}
