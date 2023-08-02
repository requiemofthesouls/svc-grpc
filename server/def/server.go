package def

import (
	"github.com/requiemofthesouls/container"
	logDef "github.com/requiemofthesouls/logger/def"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"

	"github.com/requiemofthesouls/svc-grpc/handler"
	handlerDef "github.com/requiemofthesouls/svc-grpc/handler/def"
	grpcHealthcheck "github.com/requiemofthesouls/svc-grpc/pb/grpc-healthcheck"
	"github.com/requiemofthesouls/svc-grpc/server"
)

const (
	DIServerFactory = "grpc.server_factory"

	DIMiddlewaresUnaryListPrefix  = "grpc.middleware.unary_list."
	DIMiddlewaresStreamListPrefix = "grpc.middleware.stream_list."

	DIGatewayHTTPStatusMapPrefix  = "grpc.gateway.http_status_map."
	DIGatewayRegistrantListPrefix = "grpc.gateway.registrant_list."

	DIListenerListPrefix = "grpc.listener_list."
)

type ServerFactory func(cfg server.Config) (server.Server, error)

func init() {
	container.Register(func(builder *container.Builder, _ map[string]interface{}) error {
		return builder.Add(container.Def{
			Name: DIServerFactory,
			Build: func(cont container.Container) (interface{}, error) {
				var l logDef.Wrapper
				if err := cont.Fill(logDef.DIWrapper, &l); err != nil {
					return nil, err
				}

				return func(cfg server.Config) (server.Server, error) {
					var listenerRegistrants []server.ListenerRegistrant
					if err := cont.Fill(DIListenerListPrefix+cfg.Name, &listenerRegistrants); err != nil {
						return nil, err
					}

					// добавляем во все grpc-сервера дополнительный метод healthcheck
					listenerRegistrants = append(
						listenerRegistrants,
						func(srv *grpc.Server) {
							grpcHealthcheck.RegisterHealthCheckServer(srv, handler.NewHealthCheckServer())
						},
					)

					var unaryMiddlewares []string
					if err := cont.Fill(DIMiddlewaresUnaryListPrefix+cfg.Name, &unaryMiddlewares); err != nil {
						return nil, err
					}

					unarySrvOpts := make([]grpc.UnaryServerInterceptor, 0, len(unaryMiddlewares))
					for _, def := range unaryMiddlewares {
						var opt []grpc.UnaryServerInterceptor
						if err := cont.UnscopedFill(def, &opt); err != nil {
							return nil, err
						}
						unarySrvOpts = append(unarySrvOpts, opt...)
					}

					var streamMiddlewares []string
					if err := cont.Fill(DIMiddlewaresStreamListPrefix+cfg.Name, &streamMiddlewares); err != nil {
						return nil, err
					}

					streamSrvOps := make([]grpc.StreamServerInterceptor, 0, len(streamMiddlewares))
					for _, def := range streamMiddlewares {
						var opt []grpc.StreamServerInterceptor
						if err := cont.UnscopedFill(def, &opt); err != nil {
							return nil, err
						}
						streamSrvOps = append(streamSrvOps, opt...)
					}

					var h stats.Handler
					if err := cont.Fill(handlerDef.DIHandlerMonitoring, &h); err != nil {
						return nil, err
					}

					var httpGwServer server.HTTPGatewayServer
					if cfg.HTTP != nil {
						var gwRegistrants []server.GatewayRegistrant
						if err := cont.Fill(DIGatewayRegistrantListPrefix+cfg.Name, &gwRegistrants); err != nil {
							return nil, err
						}

						var httpStatusMap server.HTTPStatusMap
						if err := cont.Fill(DIGatewayHTTPStatusMapPrefix+cfg.Name, &httpStatusMap); err != nil {
							return nil, err
						}

						var err error
						if httpGwServer, err = server.NewHTTPGatewayServer(
							*cfg.HTTP,
							cfg.Name,
							cfg.Address,
							gwRegistrants,
							server.NewErrorHandler(l, httpStatusMap),
							l,
						); err != nil {
							return nil, err
						}
					}

					return server.New(cfg, httpGwServer, unarySrvOpts, streamSrvOps, listenerRegistrants, h, l)
				}, nil
			},
		})
	})
}
