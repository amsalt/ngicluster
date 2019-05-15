package static

import (
	"sync"

	"github.com/amsalt/cluster/resolver"
)

// NamingMap represents a map, which key is name, value is address array.
type NamingMap map[string][]string

// ConfigBasedResolver represents a static configuration based resolver.
// Generally used for test.
type ConfigBasedResolver struct {
	mutex  sync.Mutex
	naming NamingMap
	*resolver.BaseResolver
}

// NewConfigBasedResolver creates and returns a *ConfigBasedResolver instance.
func NewConfigBasedResolver() resolver.Resolver {
	cbr := &ConfigBasedResolver{}
	cbr.naming = make(NamingMap)
	cbr.BaseResolver = resolver.NewBaseResolver()
	return cbr
}

// Register regiters new service at address with name.
func (cbr *ConfigBasedResolver) Register(name, address string) {
	cbr.mutex.Lock()
	defer cbr.mutex.Unlock()
	if ok, _ := contains(cbr.naming[name], address); !ok {
		cbr.naming[name] = append(cbr.naming[name], address)
	}
}

// Resolve resolves the name and return a address list.
func (cbr *ConfigBasedResolver) Resolve(name string) (addrs []string, err error) {
	return cbr.naming[name], nil
}

func (cbr *ConfigBasedResolver) ResolveMulti(names []string) ([]*resolver.ResolveEntry, error) {
	cbr.mutex.Lock()
	defer cbr.mutex.Unlock()
	var result []*resolver.ResolveEntry
	for _, n := range names {
		result = append(result, &resolver.ResolveEntry{Name: n, Addrs: cbr.naming[n]})
	}
	return result, nil
}

func contains(arr []string, e string) (bool, int) {
	for i, item := range arr {
		if item == e {
			return true, i
		}
	}
	return false, 0
}
