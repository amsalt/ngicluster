package static

import (
	"github.com/amsalt/cluster/resolver"
	"github.com/amsalt/nginet/core"
)

// NamingMap represents a map, which key is name, value is address array.
type NamingMap map[string][]string
type NamingConnMap map[string][]core.SubChannel

// ConfigBasedResolver represents a static configuration based resolver.
// Generally used for test.
type ConfigBasedResolver struct {
	naming     NamingMap
	namingConn NamingConnMap
}

// NewConfigBasedResolver creates and returns a *ConfigBasedResolver instance.
func NewConfigBasedResolver() resolver.Resolver {
	cbr := &ConfigBasedResolver{}
	cbr.naming = make(NamingMap)
	cbr.namingConn = make(NamingConnMap)
	return cbr
}

// Register regiters new service at address with name.
func (cbr *ConfigBasedResolver) Register(name, address string) {
	if ok, _ := contains(cbr.naming[name], address); !ok {
		cbr.naming[name] = append(cbr.naming[name], address)
	}
}

func (cbr *ConfigBasedResolver) Unregister(name, address string) {
	if ok, i := contains(cbr.naming[name], address); ok {
		cbr.naming[name] = append(cbr.naming[name][:i], cbr.naming[name][i+1:]...)
	}
}

func (cbr *ConfigBasedResolver) UnregisterSubChannel(name string, conn core.SubChannel) {
	if ok, i := containsConn(cbr.namingConn[name], conn); ok {
		cbr.namingConn[name] = append(cbr.namingConn[name][:i], cbr.namingConn[name][i+1:]...)
	}
}

// Register regiters new service at address with name.
func (cbr *ConfigBasedResolver) RegisterSubChannel(name string, conn core.SubChannel) {
	if ok, _ := containsConn(cbr.namingConn[name], conn); !ok {
		cbr.namingConn[name] = append(cbr.namingConn[name], conn)
	}
}

// Resolve resolves the name and return a address list.
func (cbr *ConfigBasedResolver) Resolve(name string) (addrs []string, err error) {
	return cbr.naming[name], nil
}

// ResolveSubChannel resolves the name and return a SubChannel list.
func (cbr *ConfigBasedResolver) ResolveSubChannel(name string) (conns []core.SubChannel, err error) {
	return cbr.namingConn[name], nil
}

func contains(arr []string, e string) (bool, int) {
	for i, item := range arr {
		if item == e {
			return true, i
		}
	}
	return false, 0
}

func containsConn(arr []core.SubChannel, e core.SubChannel) (bool, int) {
	for i, item := range arr {
		if item == e {
			return true, i
		}
	}
	return false, 0
}
