package server

import (
	"context"
	"net"

	grpcprometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/requiemofthesouls/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"

	"github.com/requiemofthesouls/svc-grpc/middleware"
)

type (
	ListenerRegistrant func(srv *grpc.Server)
	GatewayRegistrant  func(context.Context, *runtime.ServeMux, *grpc.ClientConn) error

	Config struct {
		Name        string
		Address     string
		HTTP        *HTTPConfig
		Middlewares middleware.Config
	}

	Server interface {
		Start() error
		Stop()
		IsStarted() bool
		GetHTTPGateway() HTTPGatewayServer
	}
	server struct {
		cfg           Config
		server        *grpc.Server
		gatewayServer HTTPGatewayServer
		isStarted     bool
		l             logger.Wrapper
	}
)

func New(
	cfg Config,
	httpGatewayServer HTTPGatewayServer,
	unaryServerOptions []grpc.UnaryServerInterceptor,
	streamServerOptions []grpc.StreamServerInterceptor,
	listenerRegistrants []ListenerRegistrant,
	monitoringHandler stats.Handler,
	l logger.Wrapper,
) (Server, error) {
	return &server{
		cfg:           cfg,
		server:        createGrpcServer(unaryServerOptions, streamServerOptions, listenerRegistrants, monitoringHandler),
		gatewayServer: httpGatewayServer,
		l:             l,
	}, nil
}

func createGrpcServer(
	unaryServerOptions []grpc.UnaryServerInterceptor,
	streamServerOptions []grpc.StreamServerInterceptor,
	listenerRegistrants []ListenerRegistrant,
	monitoringHandler stats.Handler,
) *grpc.Server {
	var grpcServer = grpc.NewServer(
		[]grpc.ServerOption{
			grpc.StatsHandler(monitoringHandler),
			grpc.ChainUnaryInterceptor(unaryServerOptions...),
			grpc.ChainStreamInterceptor(streamServerOptions...),
		}...,
	)

	for _, registrant := range listenerRegistrants {
		registrant(grpcServer)
	}

	grpcprometheus.EnableHandlingTimeHistogram()

	return grpcServer
}

func (s *server) Start() error {
	grpcprometheus.Register(s.server)

	var (
		listener net.Listener
		err      error
	)
	if listener, err = net.Listen("tcp", s.cfg.Address); err != nil {
		return err
	}

	s.isStarted = true
	defer func() { s.isStarted = false }()

	if err = s.server.Serve(listener); err != nil {
		return err
	}

	return nil
}

func (s *server) Stop() {
	s.server.Stop()
}

func (s *server) IsStarted() bool {
	return s.isStarted
}

func (s *server) GetHTTPGateway() HTTPGatewayServer {
	return s.gatewayServer
}
