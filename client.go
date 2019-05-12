package cluster

import (
	"errors"
	"net"

	"github.com/amsalt/log"
	"github.com/amsalt/nginet/core"
	"github.com/amsalt/nginet/core/tcp"
	"github.com/amsalt/nginet/encoding"
	"github.com/amsalt/nginet/encoding/json"
	"github.com/amsalt/nginet/handler"
	"github.com/amsalt/nginet/message"
	"github.com/amsalt/nginet/message/idparser"
)

// Client represents a client-side server.
type Client struct {
	handler *handlerWrapper

	readBuf, writeBuf int
	executor          core.Executor

	connector core.ConnectorChannel

	onClose   func(ctx *core.ChannelContext)
	onConnect func(ctx *core.ChannelContext, channel core.Channel)
}

// NewClient creates an empty Client instance.
func NewClient() *Client {
	c := new(Client)
	c.handler = newHandlerWrapper()
	return c
}

// NewClientWithBufSize creates an Client instance with readBufSize and writeBufSize.
func NewClientWithBufSize(readBuf, writeBuf int) *Client {
	c := NewClient()

	c.readBuf = readBuf
	c.writeBuf = writeBuf

	return c
}

// SetConnector enables set a full-control ConnectorChannel.
func (c *Client) SetConnector(connector core.ConnectorChannel) {
	c.connector = connector
	c.connector.InitSubChannel(func(channel core.SubChannel) {
		c.connector.SubChannelInitializer()(channel)
		channel.Pipeline().AddLast(nil, "OnOpenOrCloseHandler", c.handler)
	})
}

// InitConnector inits a customized ConnectorChannel.
func (c *Client) InitConnector(executor core.Executor, register message.Register, processorMgr message.ProcessorMgr) {
	c.readBuf = DefaultReadBufSize
	c.writeBuf = DefaultWriteBufSize

	c.connector = tcp.NewClientChannel(&tcp.Options{WriteBufSize: c.writeBuf, ReadBufSize: c.readBuf})
	parser := idparser.NewUint16ID()
	codec := encoding.MustGetCodec(json.CodecJSON)
	idParser := handler.NewIDParser(register, parser)
	serializer := handler.NewMessageSerializer(register, codec)
	deserializer := handler.NewMessageDeserializer(register, codec)

	c.connector.InitSubChannel(func(channel core.SubChannel) {
		channel.Pipeline().AddLast(nil, "PacketLengthDecoder", handler.NewPacketLengthDecoder(2))
		channel.Pipeline().AddLast(nil, "PacketLengthPrepender", handler.NewPacketLengthPrepender(2))
		channel.Pipeline().AddLast(nil, "MessageSerializer", serializer)
		channel.Pipeline().AddLast(nil, "IDParser", idParser)
		channel.Pipeline().AddLast(nil, "MessageDeserializer", deserializer)
		channel.Pipeline().AddLast(nil, "OnOpenOrCloseHandler", c.handler)
		channel.Pipeline().AddLast(c.executor, "MessageHandler", handler.NewDefaultMessageHandler(processorMgr))
	})
}

// Connect connects the address at 'addr'
func (c *Client) Connect(addr string) (core.SubChannel, error) {
	log.Debugf("Client connecting %+v ...", addr)
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, errors.New("bad net addr")
	}
	return c.connector.Connect(tcpAddr)
}

func (c *Client) OnDisconnect(f func(ctx *core.ChannelContext)) {
	c.handler.onClose = f
}

func (c *Client) OnConnect(f func(ctx *core.ChannelContext, channel core.Channel)) {
	c.handler.onConnect = f
}

func (c *Client) SubChannelInitializer() func(channel core.SubChannel) {
	return c.connector.SubChannelInitializer()
}

func (c *Client) AddAfterHandler(afterName string, executor core.Executor, name string, h interface{}) {
	initialize := c.connector.SubChannelInitializer()
	c.connector.InitSubChannel(func(channel core.SubChannel) {
		initialize(channel)
		channel.Pipeline().AddAfter(afterName, executor, name, h)
	})
}
