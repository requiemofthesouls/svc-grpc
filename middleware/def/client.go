package def

import (
	cfgCfg "github.com/requiemofthesouls/config/def"
	"github.com/requiemofthesouls/container"
	logDef "github.com/requiemofthesouls/logger/def"
	"google.golang.org/grpc"

	"github.com/requiemofthesouls/svc-grpc/middleware"
	"github.com/requiemofthesouls/svc-grpc/server"
)

type diClient struct {
	name string
}

func (c diClient) GetDefinitions(serverInstances []server.Config) []container.Def {
	var containers = make([]container.Def, 0, len(serverInstances)*len(serverTypes))
	for _, serverInstance := range serverInstances {
		for _, serverType := range serverTypes {
			containers = append(containers, container.Def{
				Name:  GetDIMiddlewareName(serverInstance.Name, serverType, c.name),
				Build: c.getBuildFunc(serverType == UnaryServerType),
			})
		}
	}

	return containers
}

func (c diClient) getBuildFunc(isUnary bool) func(container.Container) (interface{}, error) {
	if isUnary {
		return func(cont container.Container) (interface{}, error) {
			var l logDef.Wrapper
			if err := cont.Fill(logDef.DIWrapper, &l); err != nil {
				return nil, err
			}

			var cfg cfgCfg.Wrapper
			if err := cont.Fill(cfgCfg.DIWrapper, &cfg); err != nil {
				return nil, err
			}

			return []grpc.UnaryServerInterceptor{
				middleware.UnaryClientBuilder(map[string]struct{}{healthCheckRoute: {}}, l, cfg.GetBool("isStage")),
			}, nil
		}
	}

	return func(cont container.Container) (interface{}, error) {
		var l logDef.Wrapper
		if err := cont.Fill(logDef.DIWrapper, &l); err != nil {
			return nil, err
		}

		var cfg cfgCfg.Wrapper
		if err := cont.Fill(cfgCfg.DIWrapper, &cfg); err != nil {
			return nil, err
		}

		return []grpc.StreamServerInterceptor{
			middleware.StreamClientBuilder(map[string]struct{}{healthCheckRoute: {}}, l, cfg.GetBool("isStage")),
		}, nil
	}
}
