package roundrobin

import (
	"errors"
	"sync"

	"github.com/amsalt/cluster/balancer"
	"github.com/amsalt/cluster/resolver"
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
	return "roundrobin"
}

func (b *builder) Build(opt ...balancer.BuildOption) balancer.Balancer {
	opts := option{}
	for _, o := range opt {
		o(&opts)
	}

	return newRoundRobin(&opts)
}

type roundrobin struct {
	*option

	sync.Mutex
	next int
}

func newRoundRobin(opt *option) balancer.Balancer {
	rr := &roundrobin{}
	rr.option = opt

	return rr
}

func (s *roundrobin) Pick(ctx *core.ChannelContext) (core.SubChannel, error) {
	conns, err := s.resolver.ResolveSubChannel(s.name)
	if err != nil {
		return nil, err
	}

	if len(conns) <= 0 {
		return nil, errors.New("no available balancer")
	}

	sc := conns[s.next]
	s.Lock()
	s.next = (s.next + 1) % len(conns)
	s.Unlock()

	return sc, nil
}

func (s *roundrobin) ServName() string {
	return s.name
}
