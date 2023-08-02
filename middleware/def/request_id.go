package def

import (
	"github.com/requiemofthesouls/container"
	logDef "github.com/requiemofthesouls/logger/def"
	"google.golang.org/grpc"

	"github.com/requiemofthesouls/svc-grpc/middleware"
	"github.com/requiemofthesouls/svc-grpc/server"
)

type diRequestID struct {
	name string
}

func (r diRequestID) GetDefinitions(serverInstances []server.Config) []container.Def {
	var containers = make([]container.Def, 0, len(serverInstances)*len(serverTypes))
	for _, serverInstance := range serverInstances {
		for _, serverType := range serverTypes {
			containers = append(containers, container.Def{
				Name:  GetDIMiddlewareName(serverInstance.Name, serverType, r.name),
				Build: r.getBuildFunc(serverType == UnaryServerType),
			})
		}
	}

	return containers
}

func (r diRequestID) getBuildFunc(isUnary bool) func(container.Container) (interface{}, error) {
	if isUnary {
		return func(cont container.Container) (interface{}, error) {
			var l logDef.Wrapper
			if err := cont.Fill(logDef.DIWrapper, &l); err != nil {
				return nil, err
			}

			return []grpc.UnaryServerInterceptor{
				middleware.UnaryRequestIDBuilder(map[string]struct{}{healthCheckRoute: {}}, l),
			}, nil
		}
	}

	return func(cont container.Container) (interface{}, error) {
		var l logDef.Wrapper
		if err := cont.Fill(logDef.DIWrapper, &l); err != nil {
			return nil, err
		}

		return []grpc.StreamServerInterceptor{
			middleware.StreamRequestIDBuilder(map[string]struct{}{healthCheckRoute: {}}, l),
		}, nil
	}
}
