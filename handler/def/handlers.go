package def

import (
	"github.com/requiemofthesouls/container"
	monDef "github.com/requiemofthesouls/monitoring/def"

	"github.com/requiemofthesouls/svc-grpc/handler"
)

const (
	DIHandlerMonitoring  = "grpc.handler.monitoring"
	DIHandlerHealthCheck = "grpc.handler.health_check"
)

func init() {
	container.Register(func(builder *container.Builder, _ map[string]interface{}) error {
		return builder.Add(
			container.Def{
				Name: DIHandlerMonitoring,
				Build: func(cont container.Container) (interface{}, error) {
					var m monDef.Wrapper
					if err := cont.Fill(monDef.DIWrapper, &m); err != nil {
						return nil, err
					}

					return handler.NewMonitoringHandler(m), nil
				},
			},
			container.Def{
				Name: DIHandlerHealthCheck,
				Build: func(cont container.Container) (interface{}, error) {
					return handler.NewHealthCheckServer(), nil
				},
			},
		)
	})
}
