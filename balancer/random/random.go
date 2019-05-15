package random

import (
	"errors"
	"math/rand"

	"github.com/amsalt/cluster/balancer"
	"github.com/amsalt/log"
	"github.com/amsalt/nginet/core"
)

func init() {
	balancer.Register(&builder{})
}

type builder struct{}

func (*builder) Name() string {
	return "random"
}

func (b *builder) Build(opt ...balancer.BuildOption) balancer.Balancer {
	opts := balancer.Option{}
	for _, o := range opt {
		o(&opts)
	}

	return newRandom(&opts)
}

type random struct {
	*balancer.Option
}

func newRandom(opt *balancer.Option) balancer.Balancer {
	rr := &random{}
	rr.Option = opt

	return rr
}

func (r *random) Pick(ctx interface{}) (core.SubChannel, error) {
	conns, err := r.Resolver.ResolveSubChannel(r.Name)
	if err != nil {
		return nil, err
	}

	if len(conns) <= 0 {
		return nil, errors.New("no available connections")
	}

	log.Debugf("conn len: %+v", len(conns))

	v := rand.Intn(len(conns))
	log.Debugf("random choose: %+v", v)
	sc := conns[v]
	return sc, nil
}

func (r *random) ServName() string {
	return r.Name
}
