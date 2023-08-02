package def

import (
	"github.com/requiemofthesouls/container"
	logDef "github.com/requiemofthesouls/logger/def"
	"github.com/requiemofthesouls/logger/shortener"
	"github.com/requiemofthesouls/logger/trimmer"
	"google.golang.org/grpc"

	"github.com/requiemofthesouls/svc-grpc/middleware"
	"github.com/requiemofthesouls/svc-grpc/server"
)

type diLogging struct {
	name string
}

func (l diLogging) GetDefinitions(serverInstances []server.Config) []container.Def {
	var containers = make([]container.Def, 0, len(serverInstances)*len(serverTypes))
	for _, serverInstance := range serverInstances {
		for _, serverType := range serverTypes {
			containers = append(containers, container.Def{
				Name:  GetDIMiddlewareName(serverInstance.Name, serverType, l.name),
				Build: l.getBuildFunc(serverType == UnaryServerType, serverInstance.Middlewares.Logging),
			})
		}
	}

	return containers
}

func (l diLogging) getBuildFunc(isUnary bool, conf middleware.LoggingConfig) func(container.Container) (interface{}, error) {
	return func(cont container.Container) (interface{}, error) {
		var l logDef.Wrapper
		if err := cont.Fill(logDef.DIWrapper, &l); err != nil {
			return nil, err
		}

		if isUnary {
			return []grpc.UnaryServerInterceptor{
				middleware.UnaryLoggerBuilder(l, trimmer.New(conf.TrimmedFields), shortener.New(conf.LoggedFields)),
			}, nil
		}

		return []grpc.StreamServerInterceptor{
			middleware.StreamLoggerBuilder(l),
		}, nil
	}
}
