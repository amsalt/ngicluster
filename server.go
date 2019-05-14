package cluster

import (
	"errors"
	"net"

	"github.com/amsalt/cluster/resolver"
	"github.com/amsalt/nginet/core"
	"github.com/amsalt/nginet/core/tcp"
	"github.com/amsalt/nginet/encoding"
	"github.com/amsalt/nginet/encoding/json"
	"github.com/amsalt/nginet/handler"
	"github.com/amsalt/nginet/message"
	"github.com/amsalt/nginet/message/idparser"
)

type TimeOutOpts struct {
	TimeoutSec int
	PeriodSec  int
}

// Server represents a server-side server.
type Server struct {
	handler *handlerWrapper

	servType string // the type of service.
	addr     string // the address of service.
	resolver resolver.Resolver

	readBuf, writeBuf, maxConn int
	executor                   core.Executor

	acceptor core.AcceptorChannel

	timeoutSec       int
	timeoutPeriodSec int
}

// NewServer creates an empty Server instance.
func NewServer(servType string, resolver resolver.Resolver, timeoutSec ...*TimeOutOpts) *Server {
	s := new(Server)
	s.resolver = resolver
	s.servType = servType
	s.handler = newHandlerWrapper()
	if len(timeoutSec) > 0 {
		s.timeoutSec = timeoutSec[0].TimeoutSec
		s.timeoutPeriodSec = timeoutSec[0].PeriodSec
	}
	return s
}

// NewServerWithBufSize creates an Server instance with readBufSize and writeBufSize.
func NewServerWithBufSize(servType string, resolver resolver.Resolver, readBuf, writeBuf, maxConn int, timeoutSec ...*TimeOutOpts) *Server {
	s := NewServer(servType, resolver, timeoutSec...)
	s.readBuf = readBuf
	s.writeBuf = writeBuf
	s.maxConn = maxConn

	return s
}

// SetAcceptor enables set a full-control AcceptorChannel.
func (s *Server) SetAcceptor(acceptor core.AcceptorChannel) {
	s.acceptor = acceptor
	s.acceptor.InitSubChannel(func(channel core.SubChannel) {
		s.acceptor.SubChannelInitializer()(channel)
		channel.Pipeline().AddLast(nil, "OnOpenOrCloseHandler", s.handler)
	})
}

// InitAcceptor inits a customized AcceptorChannel.
func (s *Server) InitAcceptor(executor core.Executor, register message.Register, processorMgr message.ProcessorMgr) {
	s.readBuf = DefaultReadBufSize
	s.writeBuf = DefaultWriteBufSize

	s.acceptor = core.GetAcceptorBuilder(core.TCPServBuilder).Build(
		tcp.WithReadBufSize(s.readBuf),
		tcp.WithWriteBufSize(s.writeBuf),
		tcp.WithMaxConnNum(s.maxConn),
	)

	parser := idparser.NewUint16ID()
	codec := encoding.MustGetCodec(json.CodecJSON)
	idParser := handler.NewIDParser(register, parser)
	serializer := handler.NewMessageSerializer(register, codec)
	deserializer := handler.NewMessageDeserializer(register, codec)

	s.acceptor.InitSubChannel(func(channel core.SubChannel) {
		if s.timeoutSec > 0 {
			channel.Pipeline().AddLast(nil, "IdleStateHandler", handler.NewIdleStateHandler(s.timeoutSec, s.timeoutSec, true))
			channel.Pipeline().AddLast(nil, "HeartbeatHandler", NewHeartbeatHandler(s.timeoutSec, s.timeoutPeriodSec))
		}
		channel.Pipeline().AddLast(nil, "PacketLengthDecoder", handler.NewPacketLengthDecoder(2))
		channel.Pipeline().AddLast(nil, "PacketLengthPrepender", handler.NewPacketLengthPrepender(2))
		channel.Pipeline().AddLast(nil, "MessageSerializer", serializer)
		channel.Pipeline().AddLast(nil, "IDParser", idParser)
		channel.Pipeline().AddLast(nil, "MessageDeserializer", deserializer)
		channel.Pipeline().AddLast(executor, "processor", handler.NewDefaultMessageHandler(processorMgr))
		channel.Pipeline().AddLast(nil, "OnOpenOrCloseHandler", s.handler)
	})
}

func (s *Server) Listen(addr string) error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)

	if err != nil {
		return errors.New("bad net addr")
	}
	// TODO: support external IP
	s.addr = tcpAddr.String()
	s.acceptor.Listen(tcpAddr)
	return nil
}

func (s *Server) Accept() {
	if s.resolver != nil {
		s.resolver.Register(s.servType, s.addr)
	}

	s.acceptor.Accept()
}

func (s *Server) OnDisconnect(f func(ctx *core.ChannelContext)) {
	s.handler.onClose = f
}

func (s *Server) OnConnect(f func(ctx *core.ChannelContext, channel core.Channel)) {
	s.handler.onConnect = f
}

func (s *Server) SubChannelInitializer() func(channel core.SubChannel) {
	return s.acceptor.SubChannelInitializer()
}

func (s *Server) AddAfterHandler(afterName string, executor core.Executor, name string, h interface{}) {
	initialize := s.acceptor.SubChannelInitializer()
	s.acceptor.InitSubChannel(func(channel core.SubChannel) {
		initialize(channel)
		channel.Pipeline().AddAfter(afterName, executor, name, h)
	})
}

func (s *Server) AddLastHandler(executor core.Executor, name string, h interface{}) {
	initialize := s.acceptor.SubChannelInitializer()
	s.acceptor.InitSubChannel(func(channel core.SubChannel) {
		initialize(channel)
		channel.Pipeline().AddLast(executor, name, h)
	})
}
