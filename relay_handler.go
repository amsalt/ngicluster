package ngicluster

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
	log.Debugf("RelayHandler onread: %+v", msg)
	if params, ok := msg.([]interface{}); ok && len(params) > 1 {
		id := params[0]
		msgBuf, ok := params[1].(bytes.ReadOnlyBuffer)
		if ok {
			log.Debugf("RelayHandler.OnRead params check passed")
			servType := rh.clus.Route(id)
			log.Debugf("RelayHandler.OnRead server type: %+v", servType)
			if servType != "" && servType != rh.currentServType {
				log.Debugf("found router for msg: %+v", id)
				buf := make([]byte, msgBuf.Len())
				copy(buf, msgBuf.Bytes()[0:])
				newPacket := packet.NewRawPacket(id, buf)

				var stickinessValue interface{}
				if rh.stickinessKey != "" {
					stickinessValue = ctx.Attr().Value(rh.stickinessKey)

					// test code
					// rh.stickinessKey = "UserID"
					// stickinessValue = "test"
				}
				if stickinessValue != nil {
					params := make(map[string]interface{})
					params[rh.stickinessKey] = stickinessValue
					rh.clus.Write(servType, newPacket, params)
				} else {
					rh.clus.Write(servType, newPacket, stickinessValue)
				}

			} else {
				ctx.FireRead(msg)
			}
			return
		}
		ctx.FireError(errors.New("invalid msg type, a bytes.ReadOnlyBuffer required"))
		return
	}
	ctx.FireError(errors.New("invalid msg type, an array required"))
}
