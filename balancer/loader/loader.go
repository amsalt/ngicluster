package loader

import (
	"errors"
	"fmt"
	"sync"

	"github.com/amsalt/cluster/balancer"
	"github.com/amsalt/log"
	"github.com/amsalt/nginet/core"
)

func init() {
	balancer.Register(&builder{})
}

type builder struct{}

func (*builder) Name() string {
	return "loader"
}

func (b *builder) Build(opt ...balancer.BuildOption) balancer.Balancer {
	opts := balancer.Option{}
	for _, o := range opt {
		o(&opts)
	}

	return newLoader(&opts)
}

type loader struct {
	*balancer.Option

	sync.Mutex
	loads sync.Map
}

func newLoader(opt *balancer.Option) balancer.Balancer {
	l := &loader{}
	l.Option = opt

	return l
}

func (l *loader) Pick(ctx interface{}) (core.SubChannel, error) {
	conns, err := l.Resolver.ResolveSubChannel(l.Name)
	if err != nil {
		return nil, err
	}

	if len(conns) <= 0 {
		return nil, errors.New("no available connections")
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
	return l.Name
}
