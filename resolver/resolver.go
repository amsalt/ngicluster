package resolver

import "github.com/amsalt/nginet/core"

// package resolver privides the service discovery strategy.

// Resolver represents a naming service which used for service discovery.
// By this, can register new a service at address with name,and also resolve the name
// and return all services named `name`
// A resolver can be implemented by zookeeper, etcd, a self-implemented service or just a
// static config for test.
type Resolver interface {
	// Register regiters new service at address with name.
	Register(name, address string)

	// Ungregister unregisters the service.
	Unregister(name, address string)

	// RegisterSubChannel regiters new service SubChannel with name `name`.
	// The implementation must be thread-safe.
	RegisterSubChannel(name string, conn core.SubChannel)

	// UnregisterSubChannel unregisters the service.
	// The implementation must be thread-safe.
	UnregisterSubChannel(name string, conn core.SubChannel)

	// Resolve resolves the name and return a address list.
	Resolve(name string) (addrs []string, err error)

	// ResolveSubChannel resolves the name and return a SubChannel list.
	ResolveSubChannel(name string) (conns []core.SubChannel, err error)
}
