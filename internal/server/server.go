package server

import (
	"github.com/go-lynx/lynx"
	_ "github.com/go-lynx/lynx/plugins/polaris"
	"github.com/google/wire"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(
	NewGRPCServer,
	NewHTTPServer,
	lynx.GetServiceRegistry,
	lynx.GetServiceDiscovery,
)
