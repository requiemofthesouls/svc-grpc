package def

import (
	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/requiemofthesouls/container"
	"google.golang.org/grpc"

	"github.com/requiemofthesouls/svc-grpc/server"
)

type diMonitoring struct {
	name string
}

func (m diMonitoring) GetDefinitions(serverInstances []server.Config) []container.Def {
	var containers = make([]container.Def, 0, len(serverInstances)*len(serverTypes))
	for _, serverInstance := range serverInstances {
		for _, serverType := range serverTypes {
			containers = append(containers, container.Def{
				Name:  GetDIMiddlewareName(serverInstance.Name, serverType, m.name),
				Build: m.getBuildFunc(serverType == UnaryServerType),
			})
		}
	}

	return containers
}

func (m diMonitoring) getBuildFunc(isUnary bool) func(container.Container) (interface{}, error) {
	if isUnary {
		return func(container.Container) (interface{}, error) {
			return []grpc.UnaryServerInterceptor{
				grpcPrometheus.UnaryServerInterceptor,
			}, nil
		}
	}

	return func(container.Container) (interface{}, error) {
		return []grpc.StreamServerInterceptor{
			grpcPrometheus.StreamServerInterceptor,
		}, nil
	}
}
