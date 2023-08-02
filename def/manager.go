package def

import (
	cfgDef "github.com/requiemofthesouls/config/def"
	"github.com/requiemofthesouls/container"
	logDef "github.com/requiemofthesouls/logger/def"

	grpcService "github.com/requiemofthesouls/svc-grpc"
	"github.com/requiemofthesouls/svc-grpc/server"
	serverDef "github.com/requiemofthesouls/svc-grpc/server/def"
)

const DIServerManager = "grpc.server_manager"

func init() {
	container.Register(func(builder *container.Builder, _ map[string]interface{}) error {
		return builder.Add(container.Def{
			Name: DIServerManager,
			Build: func(cont container.Container) (interface{}, error) {
				var serverFactory serverDef.ServerFactory
				if err := cont.Fill(serverDef.DIServerFactory, &serverFactory); err != nil {
					return nil, err
				}

				var cfg cfgDef.Wrapper
				if err := cont.Fill(cfgDef.DIWrapper, &cfg); err != nil {
					return nil, err
				}

				var serversCfg []server.Config
				if err := cfg.UnmarshalKey("grpcServers", &serversCfg); err != nil {
					return nil, err
				}

				var servers = make(map[string]server.Server)
				for _, sCfg := range serversCfg {
					var (
						s        server.Server
						err      error
						sCfgCopy = sCfg
					)
					if s, err = serverFactory(sCfgCopy); err != nil {
						return nil, err
					}

					servers[sCfg.Name] = s
				}

				var l logDef.Wrapper
				if err := cont.Fill(logDef.DIWrapper, &l); err != nil {
					return nil, err
				}

				return grpcService.NewManager(servers, l)
			},
		})
	})
}
