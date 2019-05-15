package roundrobin

import (
	"errors"
	"sync"

	"github.com/amsalt/cluster/balancer"
	"github.com/amsalt/nginet/core"
)

func init() {
	balancer.Register(&builder{})
}

type builder struct{}

func (*builder) Name() string {
	return "roundrobin"
}

func (b *builder) Build(opt ...balancer.BuildOption) balancer.Balancer {
	opts := balancer.Option{}
	for _, o := range opt {
		o(&opts)
	}

	return newRoundRobin(&opts)
}

type roundrobin struct {
	*balancer.Option

	sync.Mutex
	next int
}

func newRoundRobin(opt *balancer.Option) balancer.Balancer {
	rr := &roundrobin{}
	rr.Option = opt

	return rr
}

func (s *roundrobin) Pick(ctx interface{}) (core.SubChannel, error) {
	conns, err := s.Resolver.ResolveSubChannel(s.Name)
	if err != nil {
		return nil, err
	}

	if len(conns) <= 0 {
		return nil, errors.New("no available connections")
	}

	sc := conns[s.next]
	s.Lock()
	s.next = (s.next + 1) % len(conns)
	s.Unlock()

	return sc, nil
}

func (s *roundrobin) ServName() string {
	return s.Name
}
