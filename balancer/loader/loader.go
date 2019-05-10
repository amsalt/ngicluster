package loader

import (
	"errors"
	"fmt"
	"sync"

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
	return "loader"
}

func (b *builder) Build(opt ...balancer.BuildOption) balancer.Balancer {
	opts := option{}
	for _, o := range opt {
		o(&opts)
	}

	return newLoader(&opts)
}

type loader struct {
	*option

	sync.Mutex
	loads sync.Map
}

func newLoader(opt *option) balancer.Balancer {
	l := &loader{}
	l.option = opt

	return l
}

func (l *loader) Pick(ctx *core.ChannelContext) (core.SubChannel, error) {
	conns, err := l.resolver.ResolveSubChannel(l.name)
	if err != nil {
		return nil, err
	}

	if len(conns) <= 0 {
		return nil, errors.New("no available balancer")
	}

	return l.selectLeastLoader(conns)
}

func (l *loader) selectLeastLoader(conns []core.SubChannel) (core.SubChannel, error) {
	var leastCounter int64
	var selected core.SubChannel
	for _, c := range conns {
		counter, ok := l.loads.Load(c)
		if !ok {
			l.loads.Store(c, int64(1))
			return c, nil
		}
		vcounter, ok := counter.(int64)
		if !ok {
			return nil, fmt.Errorf("loader.selectLeastLoader invalid counter type: %T", counter)
		}

		if leastCounter == 0 || vcounter < leastCounter {
			leastCounter = vcounter
			selected = c
		}
	}

	l.Lock()
	c, _ := l.loads.Load(selected)
	vc, ok := c.(int64)
	if ok {
		l.loads.Store(selected, vc+1)
	} else {
		log.Errorf("loader.selectLeastLoader invalid counter type not int64, but %T", c)
	}
	l.Unlock()

	return selected, nil
}

func (l *loader) ServName() string {
	return l.name
}
