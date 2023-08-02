package def

import (
	"github.com/requiemofthesouls/container"

	middlewareDef "github.com/requiemofthesouls/svc-grpc/middleware/def"
	"github.com/requiemofthesouls/svc-grpc/server"
)

type (
	DefinitionBuilder struct {
		Listener    Listener
		Gateways    Gateways
		Middlewares Middlewares
	}

	Listener func(cont container.Container) (interface{}, error)

	Gateways struct {
		RegistrantList []server.GatewayRegistrant
		HttpStatusMap  server.HTTPStatusMap
	}

	Middlewares struct {
		Unary  []string
		Stream []string
	}
)

var (
	DefaultMiddlewares = []string{
		middlewareDef.Logging,
		middlewareDef.Monitoring,
		middlewareDef.Recovery,
	}

	PublicClientMiddlewares = append(
		DefaultMiddlewares,
		middlewareDef.RequestID,
		middlewareDef.Client,
	)
)

func AddDefinitions(builder *container.Builder, definitionsBuilders map[string]DefinitionBuilder) error {
	var diDefs = make([]container.Def, 0, 10)
	for serverInstance, definitionBuilder := range definitionsBuilders {
		diDefs = append(diDefs, definitionBuilder.Listener.getDefinition(serverInstance))
		diDefs = append(diDefs, definitionBuilder.Gateways.getDefinitions(serverInstance)...)
		diDefs = append(diDefs, definitionBuilder.Middlewares.getDefinitions(serverInstance)...)
	}

	return builder.Add(diDefs...)
}

func (l Listener) getDefinition(serverInstance string) container.Def {
	return container.Def{
		Name:  DIListenerListPrefix + serverInstance,
		Build: l,
	}
}

func (m Middlewares) getDefinitions(serverInstance string) []container.Def {
	return []container.Def{
		{
			Name: DIMiddlewaresUnaryListPrefix + serverInstance,
			Build: func(_ container.Container) (interface{}, error) {
				return getMiddlewaresDINames(serverInstance, middlewareDef.UnaryServerType, m.Unary), nil
			},
		},
		{
			Name: DIMiddlewaresStreamListPrefix + serverInstance,
			Build: func(_ container.Container) (interface{}, error) {
				return getMiddlewaresDINames(serverInstance, middlewareDef.StreamServerType, m.Stream), nil
			},
		},
	}
}

func getMiddlewaresDINames(serverInstance string, serverType string, middlewaresNames []string) []string {
	var middlewaresDINames = make([]string, 0, len(middlewaresNames))
	for _, unaryMiddleware := range middlewaresNames {
		middlewaresDINames = append(
			middlewaresDINames,
			middlewareDef.GetDIMiddlewareName(serverInstance, serverType, unaryMiddleware),
		)
	}

	return middlewaresDINames
}

func (g Gateways) getDefinitions(serverInstance string) []container.Def {
	return []container.Def{
		{
			Name: DIGatewayRegistrantListPrefix + serverInstance,
			Build: func(_ container.Container) (interface{}, error) {
				return g.RegistrantList, nil
			},
		},
		{
			Name: DIGatewayHTTPStatusMapPrefix + serverInstance,
			Build: func(_ container.Container) (interface{}, error) {
				return g.HttpStatusMap, nil
			},
		},
	}
}
