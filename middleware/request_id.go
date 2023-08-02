package middleware

import (
	"context"

	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	clienterrors "github.com/requiemofthesouls/client-errors"
	"github.com/requiemofthesouls/logger"
	client "github.com/requiemofthesouls/user-client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func UnaryRequestIDBuilder(skipFor map[string]struct{}, l logger.Wrapper) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if _, ok := skipFor[info.FullMethod[1:]]; ok {
			return handler(ctx, req)
		}

		var err error
		if ctx, err = extractRequestIDFromMD(ctx, l); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

func StreamRequestIDBuilder(skipFor map[string]struct{}, l logger.Wrapper) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if _, ok := skipFor[info.FullMethod[1:]]; ok {
			return handler(srv, stream)
		}

		var (
			wrapped = grpcMiddleware.WrapServerStream(stream)
			err     error
		)
		if wrapped.WrappedContext, err = extractRequestIDFromMD(stream.Context(), l); err != nil {
			return err
		}

		return handler(srv, wrapped)
	}
}

func extractRequestIDFromMD(ctx context.Context, l logger.Wrapper) (context.Context, error) {
	var (
		md metadata.MD
		ok bool
	)
	if md, ok = metadata.FromIncomingContext(ctx); !ok {
		l.Warn(clienterrors.ErrMetadataNotSent.Error())
		return ctx, clienterrors.ErrMetadataNotSent
	}

	requestID := header(md, client.HeaderRequestID)
	if requestID == "" {
		errMsg := "requestID is not sent or empty"
		l.Warn(errMsg)
		return ctx, status.Error(codes.InvalidArgument, errMsg)
	}

	return client.RequestIDToContext(ctx, requestID), nil
}
