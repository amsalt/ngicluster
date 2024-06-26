package ngicluster

import (
	"fmt"
	"sync"
)

type MsgID2ServType map[interface{}]string
type relayRouter struct {
	sync.Mutex
	msgID2ServType MsgID2ServType
}

func newRelayRouter() *relayRouter {
	rr := new(relayRouter)
	rr.msgID2ServType = make(MsgID2ServType)

	return rr
}

func (rr *relayRouter) Register(msgID interface{}, servType string) {
	rr.Lock()
	defer rr.Unlock()
	rr.msgID2ServType[fmt.Sprintf("%v", msgID)] = servType
}

func (rr *relayRouter) Route(msgID interface{}) string {
	rr.Lock()
	defer rr.Unlock()

	return rr.msgID2ServType[fmt.Sprintf("%v", msgID)]
}
