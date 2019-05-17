package balancer

import (
	"github.com/amsalt/nginet/core"
)

// package balancer provides the load balancer strategy.

// Balancer defines the balance strategy for load balancer between the services with same name.
// Balancer picks one SubChannel by apply the load balance strategy,
type Balancer interface {
	// Pick returns the SubChannel to be used.
	Pick(ctx interface{}) (conn core.SubChannel, err error)

	// ServName returns the service name of the Balancer concerned.
	ServName() string
}

// Storage represents a storage which stores service name-> key->SubChannel mapping used by stickiness balander.
// Add servName params to support sharing storage in different Stickiness Balancers.
type Storage interface {
	// Get returns the SubChannel by key.
	// The implementation must be thread-safe.
	Get(servName string, key interface{}) core.SubChannel

	// Set binds key with SubChannel.
	// The implementation must be thread-safe.
	Set(servName string, key interface{}, channel core.SubChannel)

	// Del removes the key from storage.
	Del(servName string, key interface{})
}
