// Package server wires the transport layer of the user service.
// It registers the LoginService handler on both the gRPC and HTTP transports
// provided by lynx-grpc and lynx-http respectively.
package server

import "github.com/google/wire"

// ProviderSet is the Wire provider set for the server layer.
var ProviderSet = wire.NewSet(
	NewGRPCServer,
	NewHTTPServer,
)
