package server

import (
	"github.com/go-lynx/lynx/app"
	"github.com/google/wire"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(
	NewGRPCServer,
	NewHTTPServer,
	app.GetServiceRegistry,
	app.GetServiceDiscovery,
)
