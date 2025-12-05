package server

import (
	"github.com/go-lynx/lynx"
	"github.com/google/wire"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(
	NewGRPCServer,
	NewHTTPServer,
	lynx.GetServiceRegistry,
	lynx.GetServiceDiscovery,
)
