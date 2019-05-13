package cluster

import (
	"errors"

	"github.com/amsalt/log"
	"github.com/amsalt/netkit/message/packet"
	"github.com/amsalt/nginet/bytes"
	"github.com/amsalt/nginet/core"
)

type RelayHandler struct {
	*core.DefaultInboundHandler
	clus            *Cluster
	stickinessKey   string
	currentServType string
}

func NewRelayHandler(currentServType string, c *Cluster, stickinessKey ...string) *RelayHandler {
	rh := new(RelayHandler)
	rh.DefaultInboundHandler = core.NewDefaultInboundHandler()
	rh.clus = c
	rh.currentServType = currentServType
	if len(stickinessKey) > 0 {
		rh.stickinessKey = stickinessKey[0]
	}

	return rh
}

// OnRead called when reads new data.
func (rh *RelayHandler) OnRead(ctx *core.ChannelContext, msg interface{}) {
	log.Debugf("RelayHandler read: %+v", msg)
	if params, ok := msg.([]interface{}); ok && len(params) > 1 {
		id := params[0]
		msgBuf, ok := params[1].(bytes.ReadOnlyBuffer)
		if ok {
			servType := rh.clus.Route(id)
			if servType != "" && servType != rh.currentServType {
				// copy buffer for safe.
				buf := make([]byte, msgBuf.Len())
				copy(buf, msgBuf.Bytes()[0:])
				newPacket := packet.NewRawPacket(id, buf)
				// relay message
				var stickinessValue interface{}
				if rh.stickinessKey != "" {
					stickinessValue = ctx.Attr().Value(rh.stickinessKey)
				}
				rh.clus.Write(servType, newPacket, stickinessValue)
			} else {
				log.Errorf("no router map for msg: %+v", msg)
				ctx.FireRead(msg)
			}

		} else {
			ctx.FireError(errors.New("MessageDeserializer.OnRead invalid msg type, a bytes.ReadOnlyBuffer required."))
		}

	} else {
		ctx.FireError(errors.New("MessageDeserializer.OnRead invalid msg type, an array required."))
	}
}
