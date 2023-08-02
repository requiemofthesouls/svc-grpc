package middleware

import (
	"context"
	"encoding/json"
	"path"
	"strings"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/requiemofthesouls/logger"
	"github.com/requiemofthesouls/logger/shortener"
	"github.com/requiemofthesouls/logger/trimmer"
	"google.golang.org/grpc"
)

type LoggingConfig struct {
	TrimmedFields trimmer.TrimmedFields
	LoggedFields  shortener.LoggedFields
}

func UnaryLoggerBuilder(l logger.Wrapper, tr trimmer.Trimmer, sh shortener.Shortener) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		var fields = serverCallFields(info.FullMethod)

		if requestJSON, err := json.Marshal(req); err == nil {
			var (
				methodNameParts = strings.Split(info.FullMethod, "/")
				methodName      string
			)
			if len(methodNameParts) > 2 {
				methodName = methodNameParts[2]
			}

			fields = append(fields, logger.ByteString(
				logger.KeyGRPCRequestBody,
				tr.Trim(methodName, sh.Shorten(methodName, requestJSON)),
			))
		}

		return handler(newLoggerForCall(ctx, l, time.Now(), fields), req)
	}
}

func StreamLoggerBuilder(l logger.Wrapper) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		var (
			newCtx context.Context
			fields = serverCallFields(info.FullMethod)
		)
		newCtx = newLoggerForCall(stream.Context(), l, time.Now(), fields)
		var wrappedStream = grpc_middleware.WrapServerStream(stream)
		wrappedStream.WrappedContext = newCtx

		return handler(srv, wrappedStream)
	}
}

func serverCallFields(fullMethodString string) []logger.Field {
	return []logger.Field{
		logger.String(logger.KeyGRPCService, path.Dir(fullMethodString)[1:]),
		logger.String(logger.KeyGRPCMethod, path.Base(fullMethodString)),
	}
}

func newLoggerForCall(ctx context.Context, l logger.Wrapper, start time.Time, fields []logger.Field) context.Context {
	fields = append(fields, logger.String(logger.KeyGRPCRequestStartTime, start.Format(time.RFC3339)))

	if deadline, ok := ctx.Deadline(); ok {
		fields = append(fields, logger.String(logger.KeyGRPCRequestDeadline, deadline.Format(time.RFC3339)))
	}

	return ctxzap.ToContext(ctx, l.With(append(fields, ctxzap.TagsToFields(ctx)...)...).GetLogger())
}
