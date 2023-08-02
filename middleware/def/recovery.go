package def

import (
	"github.com/requiemofthesouls/container"
	logDef "github.com/requiemofthesouls/logger/def"
	monDef "github.com/requiemofthesouls/monitoring/def"

	"github.com/requiemofthesouls/svc-grpc/middleware"
	"github.com/requiemofthesouls/svc-grpc/server"
)

type diRecovery struct {
	name string
}

func (r diRecovery) GetDefinitions(serverInstances []server.Config) []container.Def {
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

func (r diRecovery) getBuildFunc(isUnary bool) func(container.Container) (interface{}, error) {
	return func(cont container.Container) (interface{}, error) {
		var log logDef.Wrapper
		if err := cont.Fill(logDef.DIWrapper, &log); err != nil {
			return nil, err
		}

		var mon monDef.Wrapper
		if err := cont.Fill(monDef.DIWrapper, &mon); err != nil {
			return nil, err
		}

		if isUnary {
			return middleware.UnaryRecoveryBuilder(mon, log), nil
		}

		return middleware.StreamRecoveryBuilder(mon, log), nil
	}
}
