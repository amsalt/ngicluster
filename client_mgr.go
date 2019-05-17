package ngicluster

import (
	"fmt"
	"sync"
	"time"

	"github.com/amsalt/log"
	"github.com/amsalt/ngicluster/balancer"
	"github.com/amsalt/ngicluster/resolver"
	"github.com/amsalt/nginet/core"
)

type connetInfo struct {
	servType string
	addr     string
}

type SubChannelMap map[string][]core.SubChannel
type ChannelInfo map[interface{}]*connetInfo
type ServiceClients map[string]*Client
type ConnectedAddrs map[string][]string

type ConnectingAddr map[string][]string

type clientMgr struct {
	resolver resolver.Resolver

	caresServTypes []string // record the cares service type.

	clients ServiceClients // record the configuration.

	subChannels    SubChannelMap  // record the conntected SubChannel
	channelInfo    ChannelInfo    // record the connected information.
	connectedAddrs ConnectedAddrs // record the connected address, for fast lookup.
	connecting     ConnectingAddr // record the address connecting.

	balancers Balancers // record balancer for each service type.

	rwMutex sync.RWMutex
	close   chan byte
}

func newClientMgr(resolver resolver.Resolver) *clientMgr {
	cm := &clientMgr{resolver: resolver}
	cm.clients = make(ServiceClients)
	cm.channelInfo = make(ChannelInfo)
	cm.subChannels = make(SubChannelMap)
	cm.connectedAddrs = make(ConnectedAddrs)
	cm.balancers = make(Balancers)
	cm.connecting = make(ConnectingAddr)
	cm.close = make(chan byte)

	return cm
}

func (cm *clientMgr) Start() {
	go cm.syncService()
}

func (cm *clientMgr) Stop() {
	close(cm.close)
}

func (cm *clientMgr) RegisterClient(servType string, sc *Client, balancer balancer.Balancer) {
	cm.rwMutex.Lock()
	defer cm.rwMutex.Unlock()
	log.Debugf("clientmgr register new client with balancer: %+v", balancer)
	if cm.clients[servType] != nil {
		panic(fmt.Errorf("duplicate register service client for type: %+v", servType))
	}
	cm.caresServTypes = append(cm.caresServTypes, servType)
	sc.OnDisconnect(cm.onDisconnected)
	cm.clients[servType] = sc
	cm.balancers[servType] = balancer
}

// syncService will updates the service list and connect to new server address.
func (cm *clientMgr) syncService() {
	for {
		select {
		case <-cm.close:
			goto closed
		default:
		}

		for _, t := range cm.caresServTypes {
			addr, err := cm.resolver.Resolve(t)

			if err == nil {
				for _, address := range addr {
					if !cm.isConnectedOrConnecting(t, address) {
						cm.connect(t, address)
					}
				}
			} else {
				log.Errorf("clientMgr.syncService resolve service list of type: %+v failed for %+v", t, err)
			}
		}

		time.Sleep(time.Second)
	}

closed:
}

func (cm *clientMgr) connect(t, addr string) {
	if cm.clients[t] == nil {
		log.Errorf("clientMgr.connect service of type: %+v client not found", t)
		return
	}

	cm.rwMutex.Lock()
	defer cm.rwMutex.Unlock()
	cm.connecting[t] = append(cm.connecting[t], addr)
	subChannel, err := cm.clients[t].Connect(addr)
	if err == nil {
		cm.newChannel(subChannel, addr, t)
		cm.removeConnecting(t, addr)
		cm.resolver.RegisterSubChannel(t, subChannel)
		log.Debugf("connect server %+v success", addr)
	} else {
		cm.removeConnecting(t, addr)
		log.Errorf("connect server %+v failed %+v", addr, err)
	}
}

func (cm *clientMgr) newChannel(subChannel core.SubChannel, addr string, servType string) {
	cm.channelInfo[subChannel] = &connetInfo{addr: addr, servType: servType}
	cm.subChannels[servType] = append(cm.subChannels[servType], subChannel)
	cm.connectedAddrs[servType] = append(cm.connectedAddrs[servType], addr)
}

func (cm *clientMgr) removeConnecting(servType, addr string) {
	for i, a := range cm.connecting[servType] {
		if a == addr {
			cm.connecting[servType] = append(cm.connecting[servType][:i], cm.connecting[servType][i+1:]...)
		}
	}
}

func (cm *clientMgr) isConnectedOrConnecting(servType, addr string) bool {
	cm.rwMutex.RLock()
	defer cm.rwMutex.RUnlock()

	for _, a := range cm.connecting[servType] {
		log.Debugf("connecting list: %+v", a)
		if a == addr {
			return true
		}
	}

	for _, address := range cm.connectedAddrs[servType] {
		if address == addr {
			return true
		}
	}
	return false
}

func (cm *clientMgr) onDisconnected(ctx *core.ChannelContext) {
	cm.rwMutex.Lock()
	defer cm.rwMutex.Unlock()

	channel := ctx.Channel()
	info := cm.channelInfo[channel]
	addr := info.addr
	servType := info.servType
	cm.removeAddr(servType, addr)
	cm.removeChannel(servType, channel.(core.SubChannel))
}

func (cm *clientMgr) removeAddr(servType, addr string) {
	for i, address := range cm.connectedAddrs[servType] {
		if address == addr {
			cm.connectedAddrs[servType] = append(cm.connectedAddrs[servType][:i], cm.connectedAddrs[servType][i+1:]...)
		}
	}
}

func (cm *clientMgr) removeChannel(servType string, channel core.SubChannel) {
	for i, c := range cm.subChannels[servType] {
		if c == channel {
			cm.subChannels[servType] = append(cm.subChannels[servType][:i], cm.subChannels[servType][i+1:]...)
		}
	}
	cm.resolver.UnregisterSubChannel(servType, channel)
	delete(cm.channelInfo, channel)
}

func (cm *clientMgr) Channels(servType string) []core.SubChannel {
	return cm.subChannels[servType]
}

func (cm *clientMgr) Write(servType string, msg interface{}, ctx interface{}) error {
	b := cm.balancers[servType]

	if b == nil {
		return fmt.Errorf("no balancer found for service: %+v", servType)
	}

	channel, err := b.Pick(ctx)
	if err == nil {
		if ctx != nil {
			channel.Write(msg, &ExtraMsg{Params: ctx})
		} else {
			channel.Write(msg)
		}

		return nil
	}
	return err
}
