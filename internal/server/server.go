package server

import (
	"github.com/go-lynx/lynx/boot"
	"github.com/google/wire"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(
	NewGRPCServer,
	NewHTTPServer,
	boot.NewServiceRegistry,
	boot.NewServiceDiscovery,
)
