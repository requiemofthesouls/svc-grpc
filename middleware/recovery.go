package middleware

import (
	"context"
	"encoding/json"
	"path"

	clienterrors "github.com/requiemofthesouls/client-errors"
	"github.com/requiemofthesouls/logger"
	"github.com/requiemofthesouls/monitoring"
	"google.golang.org/grpc"
)

func UnaryRecoveryBuilder(mon monitoring.Wrapper, log logger.Wrapper) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		func(
			ctx context.Context,
			req interface{},
			info *grpc.UnaryServerInfo,
			handler grpc.UnaryHandler,
		) (_ interface{}, err error) {
			defer func() {
				if rec := recover(); rec != nil {
					handlePanic(mon, log, info.FullMethod, req, rec)
					err = clienterrors.ErrInternalServer
				}
			}()

			return handler(ctx, req)
		},
	}
}

func StreamRecoveryBuilder(mon monitoring.Wrapper, log logger.Wrapper) []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{
		func(
			srv interface{},
			stream grpc.ServerStream,
			info *grpc.StreamServerInfo,
			handler grpc.StreamHandler,
		) (err error) {
			defer func() {
				if rec := recover(); rec != nil {
					handlePanic(mon, log, info.FullMethod, nil, rec)
					err = clienterrors.ErrInternalServer
				}
			}()

			return handler(srv, stream)
		},
	}
}

func handlePanic(mon monitoring.Wrapper, log logger.Wrapper, fullMethod string, req interface{}, rec interface{}) {
	var service, method = path.Dir(fullMethod)[1:], path.Base(fullMethod)

	mon.Inc(&monitoring.Metric{
		Namespace: "grpc",
		Name:      "panic",
		ConstLabels: map[string]string{
			"grpc_service": service,
			"grpc_method":  method,
		},
	})

	reqBytes, _ := json.Marshal(req)
	recBytes, _ := json.Marshal(rec)

	log.Error("grpc panic",
		logger.String(logger.KeyGRPCService, service),
		logger.String(logger.KeyGRPCMethod, method),
		logger.ByteString(logger.KeyGRPCRequestBody, reqBytes),
		logger.ByteString(logger.KeyGRPCRequestResponse, recBytes),
	)
}
