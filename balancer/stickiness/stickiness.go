package stickiness

import (
	"sync"

	"github.com/amsalt/cluster/balancer"
	_ "github.com/amsalt/cluster/balancer/roundrobin" // for import
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

// WithStorage sets storage for balancer.
func WithStorage(s balancer.Storage) balancer.BuildOption {
	return func(o interface{}) {
		o.(*option).storage = s
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

type DefaultStorage struct {
	sync.RWMutex
	storage map[string]map[interface{}]core.SubChannel
}

func NewDefaultStorage() balancer.Storage {
	s := new(DefaultStorage)
	s.storage = make(map[string]map[interface{}]core.SubChannel)
	return s
}

func (s DefaultStorage) Get(servName string, key interface{}) core.SubChannel {
	s.RLock()
	defer s.RUnlock()
	return s.storage[servName][key]
}

func (s DefaultStorage) Set(servName string, key interface{}, channel core.SubChannel) {
	s.Lock()
	defer s.Unlock()
	if s.storage[servName] == nil {
		s.storage[servName] = make(map[interface{}]core.SubChannel)
	}
	s.storage[servName][key] = channel
}

func (s DefaultStorage) Del(servName string, key interface{}) {
	s.Lock()
	defer s.Unlock()
	toDelete := s.storage[servName]
	delete(toDelete, key)
}

type option struct {
	name          string            // the service name to resolve
	resolver      resolver.Resolver // the naming resolver
	stickinessKey string
	dependency    balancer.Balancer
	storage       balancer.Storage
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

// Pick support key or map as params.
func (s *stickiness) Pick(params interface{}) (conn core.SubChannel, err error) {
	if params == nil {
		conn, err = s.selector(params)
		return
	}

	if data, ok := params.(map[string]interface{}); ok {
		if s.stickinessKey != "" && data[s.stickinessKey] != nil {
			stored := s.storage.Get(s.name, data[s.stickinessKey])
			if stored != nil {
				return stored, nil
			}
		}
	} else if key, ok := params.(string); ok {
		stored := s.storage.Get(s.name, key)
		if stored != nil {
			return stored, nil
		}
	} else {
		log.Errorf("unsupported params type: %T", params)
		return
	}

	conn, err = s.selector(params)
	if err != nil {
		return
	}

	if params != nil {
		if data, ok := params.(map[string]interface{}); ok {
			if s.stickinessKey != "" && data[s.stickinessKey] != nil {
				s.storage.Set(s.name, data[s.stickinessKey], conn)
			}
		} else if key, ok := params.(string); ok {
			s.storage.Set(s.name, key, conn)
		}
	}

	return
}

func (s *stickiness) selector(ctx interface{}) (conn core.SubChannel, err error) {
	return s.dependency.Pick(ctx)
}

func (s *stickiness) ServName() string {
	return s.name
}
