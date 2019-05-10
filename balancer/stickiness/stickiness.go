package stickiness

import (
	"sync"

	"github.com/amsalt/cluster/balancer"
	"github.com/amsalt/cluster/balancer/roundrobin"
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

// WithStickinessKey creates stickiness balancer with stickinessKey `key`
func WithStickinessKey(key string) balancer.BuildOption {
	return func(o interface{}) {
		o.(*option).stickinessKey = key
	}
}

// WithDependentBalancer creates stickiness balancer with dependent balancer `b`
// If not set, default dependent is roundrobin Balancer.
func WithDependentBalancer(b balancer.Balancer) balancer.BuildOption {
	return func(o interface{}) {
		o.(*option).dependency = b
	}
}

type builder struct {
}

func (b *builder) Name() string {
	return "stickiness"
}

func (b *builder) Build(opt ...balancer.BuildOption) balancer.Balancer {
	opts := option{}
	for _, o := range opt {
		o(&opts)
	}

	if opts.dependency == nil {
		opts.dependency = balancer.GetBuilder("roundrobin").Build(roundrobin.WithServName(opts.name), roundrobin.WithResolver(opts.resolver))
	}
	return newStickiness(&opts)
}

type option struct {
	name          string // the service name to resolve
	resolver      resolver.Resolver
	stickinessKey string
	dependency    balancer.Balancer
}

type stickiness struct {
	sync.RWMutex

	*option
	storage map[interface{}]core.SubChannel
}

func newStickiness(opt *option) balancer.Balancer {
	s := &stickiness{}
	s.option = opt
	s.storage = make(map[interface{}]core.SubChannel)
	return s
}

func (s *stickiness) Pick(ctx *core.ChannelContext) (conn core.SubChannel, err error) {
	if ctx == nil {
		conn, err = s.selector(ctx)
		return
	}
	stickinessKey := ctx.Attr().Value(s.stickinessKey)
	if stickinessKey != nil {
		s.RLock()
		stored := s.storage[stickinessKey]
		s.RUnlock()
		if stored != nil {
			return stored, nil
		}
	}

	conn, err = s.selector(ctx)
	if err != nil {
		return
	}

	if stickinessKey != nil {
		s.Lock()
		s.storage[stickinessKey] = conn
		s.Unlock()
	}

	return
}

func (s *stickiness) selector(ctx *core.ChannelContext) (conn core.SubChannel, err error) {
	return s.dependency.Pick(ctx)
}

func (s *stickiness) ServName() string {
	return s.name
}
