package resolver

import (
	"sync"

	"github.com/amsalt/nginet/core"
)

// package resolver privides the service discovery strategy.

// Resolver represents a naming service which used for service discovery.
// By this, can register new a service at address with name,and also resolve the name
// and return all services named `name`
// A resolver can be implemented by zookeeper, etcd, a self-implemented service or just a
// static config for test.
type Resolver interface {
	// Register regiters new service at address with name.
	Register(name, address string)

	// Resolve resolves the name and return a address list.
	Resolve(name string) (addrs []string, err error)

	// ResolveSubChannel resolves the name and return a SubChannel list.
	ResolveSubChannel(name string) (conns []core.SubChannel, err error)

	// RegisterSubChannel regiters new service SubChannel with name `name`.
	// The implementation must be thread-safe.
	RegisterSubChannel(name string, conn core.SubChannel)

	// UnregisterSubChannel unregisters the service.
	// The implementation must be thread-safe.
	UnregisterSubChannel(name string, conn core.SubChannel)
}

type NamingConnMap map[string][]core.SubChannel
type BaseResolver struct {
	sync.RWMutex

	namingConn NamingConnMap
}

func NewBaseResolver() *BaseResolver {
	br := new(BaseResolver)
	br.namingConn = make(NamingConnMap)
	return br
}

// RegisterSubChannel regiters new SubChannel.
func (br *BaseResolver) RegisterSubChannel(name string, conn core.SubChannel) {
	br.Lock()
	defer br.Unlock()
	if ok, _ := containsConn(br.namingConn[name], conn); !ok {
		br.namingConn[name] = append(br.namingConn[name], conn)
	}
}

// UnregisterSubChannel removes registered SubChannel.
func (br *BaseResolver) UnregisterSubChannel(name string, conn core.SubChannel) {
	br.Lock()
	defer br.Unlock()
	if ok, i := containsConn(br.namingConn[name], conn); ok {
		br.namingConn[name] = append(br.namingConn[name][:i], br.namingConn[name][i+1:]...)
	}
}

// ResolveSubChannel resolves the name and return a SubChannel list.
func (br *BaseResolver) ResolveSubChannel(name string) (conns []core.SubChannel, err error) {
	br.RLock()
	defer br.RUnlock()
	return br.namingConn[name], nil
}

func containsConn(arr []core.SubChannel, e core.SubChannel) (bool, int) {
	for i, item := range arr {
		if item == e {
			return true, i
		}
	}
	return false, 0
}
