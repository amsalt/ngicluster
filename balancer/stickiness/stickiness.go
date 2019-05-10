package stickiness

import (
	"sync"

	"github.com/amsalt/cluster/balancer"
	_ "github.com/amsalt/cluster/balancer/roundrobin" // for import
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
		opts.dependency = balancer.GetBuilder("roundrobin").Build(balancer.WithServName(opts.name), balancer.WithResolver(opts.resolver))
	}
	return newStickiness(&opts)
}

type Storage interface {
	// Get returns the SubChannel by key.
	// The implementation must be thread-safe.
	Get(key interface{}) core.SubChannel

	// Set binds key with SubChannel.
	// The implementation must be thread-safe.
	Set(key interface{}, channel core.SubChannel)
}

type DefaultStorage struct {
	sync.RWMutex
	storage map[interface{}]core.SubChannel
}

func NewDefaultStorage() Storage {
	s := new(DefaultStorage)
	s.storage = make(map[interface{}]core.SubChannel)
	return s
}

func (s DefaultStorage) Get(key interface{}) core.SubChannel {
	s.RLock()
	defer s.RUnlock()
	return s.storage[key]
}

func (s DefaultStorage) Set(key interface{}, channel core.SubChannel) {
	s.Lock()
	defer s.Unlock()
	s.storage[key] = channel
}

type option struct {
	name          string            // the service name to resolve
	resolver      resolver.Resolver // the naming resolver
	stickinessKey string
	dependency    balancer.Balancer
	storage       Storage
}

type stickiness struct {
	sync.RWMutex

	*option
}

func newStickiness(opt *option) balancer.Balancer {
	s := &stickiness{}
	s.option = opt
	if s.storage == nil {
		s.storage = NewDefaultStorage()
	}
	return s
}

func (s *stickiness) Pick(ctx interface{}) (conn core.SubChannel, err error) {
	if ctx == nil {
		conn, err = s.selector(ctx)
		return
	}
	stickinessKey := ctx
	if stickinessKey != nil {
		stored := s.storage.Get(stickinessKey)
		if stored != nil {
			return stored, nil
		}
	}

	conn, err = s.selector(ctx)
	if err != nil {
		return
	}

	if stickinessKey != nil {
		s.storage.Set(stickinessKey, conn)
	}

	return
}

func (s *stickiness) selector(ctx interface{}) (conn core.SubChannel, err error) {
	return s.dependency.Pick(ctx)
}

func (s *stickiness) ServName() string {
	return s.name
}
