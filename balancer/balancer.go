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
