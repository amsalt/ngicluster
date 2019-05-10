package cluster

import (
	"github.com/amsalt/log"
	"github.com/amsalt/nginet/core"
)

// package cluster implementats the communication between internal servers.

const (
	// DefaultReadBufSize sets the default ReadBufSize as 100k.
	DefaultReadBufSize = 1024 * 100

	// DefaultWriteBufSize sets the default WriteBufSize as 100k.
	DefaultWriteBufSize = 1024 * 100
)

type handlerWrapper struct {
	*core.DefaultInboundHandler

	onClose   func(ctx *core.ChannelContext)
	onConnect func(ctx *core.ChannelContext, channel core.Channel)
}

func newHandlerWrapper() *handlerWrapper {
	h := &handlerWrapper{}
	h.DefaultInboundHandler = core.NewDefaultInboundHandler()
	return h
}

func (h *handlerWrapper) OnDisconnect(ctx *core.ChannelContext) {
	log.Debugf("handlerWrapper.OnDisconnect: channel %+v closed", ctx.Channel().ID())
	if h.onClose != nil {
		h.onClose(ctx)
	}
	ctx.FireDisconnect()
}

func (h *handlerWrapper) OnConnect(ctx *core.ChannelContext, channel core.Channel) {
	log.Debugf("handlerWrapper.OnConnect:channel %+v Connect", channel.ID())
	if h.onConnect != nil {
		h.onConnect(ctx, channel)
	}
	ctx.FireConnect(channel)
}
