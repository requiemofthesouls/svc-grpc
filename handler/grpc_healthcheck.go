package handler

import (
	"context"

	grpcHealthcheck "github.com/requiemofthesouls/svc-grpc/pb/grpc-healthcheck"
)

type grpcHealthCheck struct {
	grpcHealthcheck.UnsafeHealthCheckServer
}

func NewHealthCheckServer() grpcHealthcheck.HealthCheckServer {
	return &grpcHealthCheck{}
}

func (*grpcHealthCheck) Check(
	context.Context,
	*grpcHealthcheck.HealthCheckRequest,
) (*grpcHealthcheck.HealthCheckResponse, error) {
	return &grpcHealthcheck.HealthCheckResponse{
		Status: grpcHealthcheck.HealthCheckResponse_SERVING,
	}, nil
}
