package test

// import (
// 	"net"
// 	"testing"

// 	"github.com/amsalt/cluster/balancer"
// 	_ "github.com/amsalt/cluster/balancer/loader" // for import
// 	_ "github.com/amsalt/cluster/balancer/random"
// 	_ "github.com/amsalt/cluster/balancer/roundrobin"
// 	"github.com/amsalt/cluster/balancer/stickiness"
// 	"github.com/amsalt/cluster/resolver"
// 	"github.com/amsalt/cluster/resolver/static"
// 	"github.com/amsalt/log"
// 	"github.com/amsalt/nginet/bytes"
// 	"github.com/amsalt/nginet/core"
// 	"github.com/amsalt/nginet/gnetlog"
// )

// type rawconn struct{}

// func (r rawconn) Read(buf bytes.ReadOnlyBuffer) error { return nil }
// func (r rawconn) Write(data []byte)                   {}
// func (r rawconn) Close() error                        { return nil }
// func (r rawconn) LocalAddr() net.Addr                 { return nil }
// func (r rawconn) RemoteAddr() net.Addr                { return nil }

// var rsv resolver.Resolver

// func init() {
// 	gnetlog.Init()

// }
// func TestLoaderBalancer(t *testing.T) {
// 	rsv := static.NewConfigBasedResolver()
// 	rsv.RegisterSubChannel("game", core.NewDefaultSubChannel(&rawconn{}, 0, 0))
// 	rsv.RegisterSubChannel("game", core.NewDefaultSubChannel(&rawconn{}, 0, 0))

// 	b := balancer.GetBuilder("loader").Build(balancer.WithServName("game"), balancer.WithResolver(rsv))
// 	for i := 0; i < 10; i++ {
// 		c, err := b.Pick(nil)
// 		log.Infof("loader choose c: %+v, err: %+v", c, err)
// 	}
// }

// func TestRandomBalancer(t *testing.T) {
// 	rsv := static.NewConfigBasedResolver()
// 	rsv.RegisterSubChannel("game", core.NewDefaultSubChannel(&rawconn{}, 0, 0))
// 	rsv.RegisterSubChannel("game", core.NewDefaultSubChannel(&rawconn{}, 0, 0))

// 	b := balancer.GetBuilder("random").Build(balancer.WithServName("game"), balancer.WithResolver(rsv))
// 	for i := 0; i < 10; i++ {
// 		c, err := b.Pick(nil)
// 		log.Infof("random choose c: %+v, err: %+v", c, err)
// 	}
// }

// func TestRoundRobinBalancer(t *testing.T) {
// 	rsv := static.NewConfigBasedResolver()
// 	rsv.RegisterSubChannel("game", core.NewDefaultSubChannel(&rawconn{}, 0, 0))
// 	rsv.RegisterSubChannel("game", core.NewDefaultSubChannel(&rawconn{}, 0, 0))

// 	b := balancer.GetBuilder("roundrobin").Build(balancer.WithServName("game"), balancer.WithResolver(rsv))
// 	for i := 0; i < 10; i++ {
// 		c, err := b.Pick(nil)
// 		log.Infof("roundrobin choose c: %+v, err: %+v", c, err)
// 	}
// }

// func TestStickinessBalancer(t *testing.T) {
// 	rsv := static.NewConfigBasedResolver()
// 	rsv.RegisterSubChannel("game", core.NewDefaultSubChannel(&rawconn{}, 0, 0))
// 	rsv.RegisterSubChannel("game", core.NewDefaultSubChannel(&rawconn{}, 0, 0))

// 	b := balancer.GetBuilder("stickiness").Build(stickiness.WithServName("game"), stickiness.WithResolver(rsv))
// 	for i := 0; i < 10; i++ {
// 		if i < 5 {
// 			c, err := b.Pick("a")
// 			log.Infof("stickiness choose c: %+v, err: %+v", c, err)
// 		} else {
// 			c, err := b.Pick("b")
// 			log.Infof("stickiness choose c: %+v, err: %+v", c, err)
// 		}

// 	}
// }
