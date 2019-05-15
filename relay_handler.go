package cluster

import (
	"errors"

	"github.com/amsalt/log"
	"github.com/amsalt/nginet/bytes"
	"github.com/amsalt/nginet/core"
	"github.com/amsalt/nginet/message/packet"
)

// RelayHandler relays message by definition in relay router.
type RelayHandler struct {
	*core.DefaultInboundHandler

	clus            *Cluster
	stickinessKey   string
	currentServType string
}

// NewRelayHandler creates a new RelayHandler.
func NewRelayHandler(currentServType string, c *Cluster, stickinessKey ...string) *RelayHandler {
	rh := &RelayHandler{DefaultInboundHandler: core.NewDefaultInboundHandler(), clus: c, currentServType: currentServType}

	if len(stickinessKey) > 0 {
		rh.stickinessKey = stickinessKey[0]
	}

	return rh
}

// OnRead called when reads new data.
func (rh *RelayHandler) OnRead(ctx *core.ChannelContext, msg interface{}) {
	if params, ok := msg.([]interface{}); ok && len(params) > 1 {
		id := params[0]
		msgBuf, ok := params[1].(bytes.ReadOnlyBuffer)
		if ok {
			servType := rh.clus.Route(id)
			if servType != "" && servType != rh.currentServType {

				buf := make([]byte, msgBuf.Len())
				copy(buf, msgBuf.Bytes()[0:])
				newPacket := packet.NewRawPacket(id, buf)

				var stickinessValue interface{}
				if rh.stickinessKey != "" {
					stickinessValue = ctx.Attr().Value(rh.stickinessKey)
				}

				rh.clus.Write(servType, newPacket, stickinessValue)
			} else {
				log.Warningf("no information found in router map for msg: %+v", msg)
				ctx.FireRead(msg)
			}
			return
		}
		ctx.FireError(errors.New("invalid msg type, a bytes.ReadOnlyBuffer required"))
		return
	}
	ctx.FireError(errors.New("invalid msg type, an array required"))
}
