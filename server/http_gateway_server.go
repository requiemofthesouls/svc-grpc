package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/requiemofthesouls/logger"
	httpServer "github.com/requiemofthesouls/svc-http/server"
	userclient "github.com/requiemofthesouls/user-client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

type (
	HTTPConfig struct {
		Address string
	}

	HTTPGatewayServer interface {
		httpServer.Server
	}

	httpGatewayServer struct {
		httpServer.Server
		cfg HTTPConfig
		l   logger.Wrapper
	}
)

type eventSourceMarshaler struct {
	runtime.JSONPb
}

func (m *eventSourceMarshaler) ContentType(_ interface{}) string {
	return "text/event-stream"
}

func NewHTTPGatewayServer(
	cfg HTTPConfig,
	name, grpcAddress string,
	registrants []GatewayRegistrant,
	errorHandler ErrorHandler,
	l logger.Wrapper,
) (HTTPGatewayServer, error) {
	var (
		h   httpServer.Handler
		err error
	)
	if h, err = createHttpHandler(grpcAddress, registrants, errorHandler); err != nil {
		return nil, err
	}

	return &httpGatewayServer{
		Server: httpServer.New(httpServer.Config{Name: name, Address: cfg.Address}, h),
		cfg:    cfg,
		l:      l,
	}, nil
}

func createHttpHandler(addr string, registrants []GatewayRegistrant, errorHandler ErrorHandler) (httpServer.Handler, error) {
	var gwMux = runtime.NewServeMux(
		runtime.WithMarshalerOption(
			runtime.MIMEWildcard,
			&runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					UseProtoNames:   true,
					EmitUnpopulated: true,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					DiscardUnknown: true,
				}},
		),
		runtime.WithMarshalerOption(
			"text/event-stream",
			&eventSourceMarshaler{
				runtime.JSONPb{
					MarshalOptions: protojson.MarshalOptions{
						UseProtoNames:   true,
						EmitUnpopulated: true,
					},
					UnmarshalOptions: protojson.UnmarshalOptions{
						DiscardUnknown: true,
					},
				},
			},
		),
		runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
			switch {
			case strings.HasPrefix(key, "X-"):
				return key, true
			case key == "User-Agent":
				return userclient.HeaderUserAgent, true
			default:
				return runtime.DefaultHeaderMatcher(key)
			}
		}),
		runtime.WithErrorHandler(errorHandler.HTTPError),
	)

	for _, registrant := range registrants {
		var (
			ctx      = context.Background()
			grpcConn *grpc.ClientConn
			err      error
		)
		if grpcConn, err = grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
			return nil, err
		}

		if err = registrant(ctx, gwMux, grpcConn); err != nil {
			return nil, err
		}
	}

	var mux = http.NewServeMux()
	mux.HandleFunc("/healthcheck", func(resp http.ResponseWriter, req *http.Request) {
		defer func() { _ = req.Body.Close() }()
		_, _ = fmt.Fprint(resp, "ok")
	})
	mux.Handle("/", gwMux)

	return mux, nil
}
