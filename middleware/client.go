package middleware

import (
	"context"

	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	clienterrors "github.com/requiemofthesouls/client-errors"
	"github.com/requiemofthesouls/logger"
	userclient "github.com/requiemofthesouls/user-client"
	userclientpb "github.com/requiemofthesouls/user-client/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func UnaryClientBuilder(skipFor map[string]struct{}, l logger.Wrapper, isStage bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if _, ok := skipFor[info.FullMethod[1:]]; ok {
			return handler(ctx, req)
		}

		var err error
		if ctx, err = clientMetadataToContext(ctx, l, isStage); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

func StreamClientBuilder(skipFor map[string]struct{}, l logger.Wrapper, isStage bool) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if _, ok := skipFor[info.FullMethod[1:]]; ok {
			return handler(srv, stream)
		}

		var (
			wrappedStream = grpcMiddleware.WrapServerStream(stream)
			err           error
		)
		if wrappedStream.WrappedContext, err = clientMetadataToContext(stream.Context(), l, isStage); err != nil {
			return err
		}

		return handler(srv, wrappedStream)
	}
}

func clientMetadataToContext(ctx context.Context, l logger.Wrapper, isStage bool) (context.Context, error) {
	var (
		md metadata.MD
		ok bool
	)
	if md, ok = metadata.FromIncomingContext(ctx); !ok {
		l.Warn(clienterrors.ErrMetadataNotSent.Error())
		return ctx, clienterrors.ErrMetadataNotSent
	}

	var c *userclientpb.Client
	if c = userclient.ExtractContextToClient(ctx); c == nil {
		c = &userclientpb.Client{}
	}

	c.Ip = header(md, userclient.HeaderUserIP)
	c.Host = header(md, userclient.HeaderUserHost)
	c.UserAgent = header(md, userclient.HeaderUserAgent)
	c.Language = header(md, userclient.HeaderUserLanguage)
	c.Location = header(md, userclient.HeaderUserLocation)
	c.Platform = header(md, userclient.HeaderApplicationPlatform)
	// Ручная передача страны пользователя в debug режиме
	if location := header(md, userclient.HeaderDebugCountry); isStage && location != "" {
		c.Location = location
	}

	return userclient.ClientToContext(ctx, c), nil
}

func header(md metadata.MD, key string) string {
	if value := md.Get(key); len(value) != 0 {
		return value[0]
	}

	return ""
}
