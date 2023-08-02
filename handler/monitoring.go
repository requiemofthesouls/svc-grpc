package handler

import (
	"context"
	"sync/atomic"

	"github.com/requiemofthesouls/monitoring"
	"google.golang.org/grpc/stats"
)

type monitoringHandler struct {
	m monitoring.Wrapper
	c int32
}

func NewMonitoringHandler(m monitoring.Wrapper) stats.Handler {
	return &monitoringHandler{
		m: m,
	}
}

// HandleConn exists to satisfy gRPC stats.Handler.
func (s *monitoringHandler) HandleConn(_ context.Context, cs stats.ConnStats) {
	var delta int32
	switch cs.(type) {
	case *stats.ConnEnd:
		delta = -1
	case *stats.ConnBegin:
		delta = 1
	}

	atomic.AddInt32(&s.c, delta)

	var val uint64
	if s.c > 0 {
		val = uint64(s.c)
	}

	s.m.Val(&monitoring.Metric{
		Namespace: "grpc",
		Name:      "connection_count",
	}, val)
}

// TagConn exists to satisfy gRPC stats.Handler.
func (s *monitoringHandler) TagConn(ctx context.Context, _ *stats.ConnTagInfo) context.Context {
	// no-op
	return ctx
}

// HandleRPC implements per-RPC tracing and stats instrumentation.
func (s *monitoringHandler) HandleRPC(_ context.Context, _ stats.RPCStats) {
	// no-op
}

// TagRPC implements per-RPC context management.
func (s *monitoringHandler) TagRPC(ctx context.Context, _ *stats.RPCTagInfo) context.Context {
	// no-op
	return ctx
}
