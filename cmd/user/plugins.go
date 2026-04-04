package main

import (
	// Explicit plugin manifest for the user-service bootstrap path.
	_ "github.com/go-lynx/lynx-grpc"
	_ "github.com/go-lynx/lynx-http"
	_ "github.com/go-lynx/lynx-mysql"
	_ "github.com/go-lynx/lynx-redis"
	_ "github.com/go-lynx/lynx-redis-lock"
	_ "github.com/go-lynx/lynx-tracer"
)
