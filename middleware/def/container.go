package def

import (
	"fmt"

	cfgDef "github.com/requiemofthesouls/config/def"
	"github.com/requiemofthesouls/container"

	"github.com/requiemofthesouls/svc-grpc/server"
)

const (
	Client     = "client"
	Logging    = "logging"
	Monitoring = "monitoring"
	Recovery   = "recovery"
	RequestID  = "requestID"

	healthCheckRoute = "grpc.health.v1.Health/Check"
)

type DIMiddleware interface {
	GetDefinitions(serverInstances []server.Config) []container.Def
}

const (
	UnaryServerType  = "unary"
	StreamServerType = "stream"
)

var (
	serverTypes = []string{UnaryServerType, StreamServerType}

	diMiddlewares = []DIMiddleware{
		diClient{name: Client},
		diLogging{name: Logging},
		diMonitoring{name: Monitoring},
		diRecovery{name: Recovery},
		diRequestID{name: RequestID},
	}
)

func init() {
	container.Register(func(builder *container.Builder, _ map[string]interface{}) error {
		cont := builder.Build()

		var cfg cfgDef.Wrapper
		if err := cont.Fill(cfgDef.DIWrapper, &cfg); err != nil {
			return err
		}

		var serverInstances []server.Config
		if err := cfg.UnmarshalKey("grpcServers", &serverInstances); err != nil {
			return err
		}

		return builder.Add(getContainers(serverInstances)...)
	})
}

func GetDIMiddlewareName(serverInstance string, serverType string, middlewareName string) string {
	return fmt.Sprintf("grpc.%s.%s.middleware.%s", serverInstance, serverType, middlewareName)
}

func getContainers(serverInstances []server.Config) []container.Def {
	var containers = make([]container.Def, 0, len(diMiddlewares))
	for _, md := range diMiddlewares {
		containers = append(containers, md.GetDefinitions(serverInstances)...)
	}

	return containers
}
