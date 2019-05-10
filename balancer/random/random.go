package random

import (
	"errors"
	"math/rand"

	"github.com/amsalt/cluster/balancer"
	"github.com/amsalt/cluster/resolver"
	"github.com/amsalt/log"
	"github.com/amsalt/nginet/core"
)

func init() {
	balancer.Register(&builder{})
}

// WithServName creates stickiness balancer with Service name `n`
func WithServName(n string) balancer.BuildOption {
	return func(o interface{}) {
		o.(*option).name = n
	}
}

// WithResolver creates stickiness balancer with resolver `r`
func WithResolver(r resolver.Resolver) balancer.BuildOption {
	return func(o interface{}) {
		o.(*option).resolver = r
	}
}

type option struct {
	name     string // the service name to resolve
	resolver resolver.Resolver
}

type builder struct{}

func (*builder) Name() string {
	return "random"
}

func (b *builder) Build(opt ...balancer.BuildOption) balancer.Balancer {
	opts := option{}
	for _, o := range opt {
		o(&opts)
	}

	return newRandom(&opts)
}

type random struct {
	*option
}

func newRandom(opt *option) balancer.Balancer {
	rr := &random{}
	rr.option = opt

	return rr
}

func (r *random) Pick(ctx *core.ChannelContext) (core.SubChannel, error) {
	conns, err := r.resolver.ResolveSubChannel(r.name)
	if err != nil {
		return nil, err
	}

	if len(conns) <= 0 {
		return nil, errors.New("no available balancer")
	}

	log.Debugf("conn len: %+v", len(conns))

	v := rand.Intn(len(conns))
	log.Debugf("random choose: %+v", v)
	sc := conns[v]
	return sc, nil
}

func (r *random) ServName() string {
	return r.name
}
