package server

import (
	"github.com/go-lynx/lynx/app"
	_ "github.com/go-lynx/plugins/polaris/v2"
	"github.com/google/wire"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(
	NewGRPCServer,
	NewHTTPServer,
	app.GetServiceRegistry,
	app.GetServiceDiscovery,
)
