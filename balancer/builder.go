package balancer

import (
	"fmt"

	"github.com/amsalt/ngicluster/resolver"
)

// Builder creates a Balancer.
type Builder interface {
	// Build builds a new Balancer.
	// name is the name of the resovler service.
	// resolver is a resolver.
	Build(opts ...BuildOption) Balancer

	// Name returns the name of builder.
	Name() string
}

// BuildOption represents a build option method.
type BuildOption func(interface{})

// BuilderMap is a map which contains all the registered naming builder.
type BuilderMap map[string]Builder

var builders BuilderMap

func init() {
	builders = make(BuilderMap)
}

// Register registers a new builder.
func Register(builder Builder) {
	builders[builder.Name()] = builder
}

// GetBuilder return a builder by name.
// If the builder with name `name` not found, a panic occurred
func GetBuilder(name string) Builder {
	b, exist := builders[name]
	if !exist {
		panic(fmt.Errorf("builder named %+v not exist", name))
	}
	return b
}

// WithServName creates stickiness balancer with Service name `n`
func WithServName(n string) BuildOption {
	return func(o interface{}) {
		o.(*Option).Name = n
	}
}

// WithResolver creates stickiness balancer with resolver `r`
func WithResolver(r resolver.Resolver) BuildOption {
	return func(o interface{}) {
		o.(*Option).Resolver = r
	}
}

type Option struct {
	Name     string // the service name to resolve
	Resolver resolver.Resolver
}
