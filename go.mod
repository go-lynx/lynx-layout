module github.com/go-lynx/lynx-layout

go 1.24.3

require (
	entgo.io/ent v0.14.4
	github.com/go-kratos/kratos/v2 v2.8.4
	github.com/go-lynx/lynx v0.1.4-beta
	github.com/go-lynx/lynx-layout/api v0.0.0-20231226093010-b62a1b27588b
	github.com/go-lynx/plugins/db/mysql/v2 v2.0.0
	github.com/go-lynx/plugins/db/redis/v2 v2.0.0
	github.com/go-lynx/plugins/service/grpc/v2 v2.0.0
	github.com/go-lynx/plugins/service/http/v2 v2.0.0
	github.com/go-lynx/plugins/polaris/v2 v2.0.0
	github.com/go-sql-driver/mysql v1.7.1
	github.com/google/wire v0.6.0
	github.com/redis/go-redis/v9 v9.8.0
	go.uber.org/automaxprocs v1.5.1
	google.golang.org/protobuf v1.35.2
)

require (
	ariga.io/atlas v0.31.1-0.20250212144724-069be8033e83 // indirect
	dario.cat/mergo v1.0.0 // indirect
	github.com/agext/levenshtein v1.2.1 // indirect
	github.com/apparentlymart/go-textseg/v13 v13.0.0 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/bmatcuk/doublestar v1.3.4 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/envoyproxy/protoc-gen-validate v1.1.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-kratos/aegis v0.2.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/inflect v0.19.0 // indirect
	github.com/go-playground/form/v4 v4.2.1 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.2 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/hashicorp/hcl/v2 v2.13.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mitchellh/go-wordwrap v0.0.0-20150314170334-ad45545899c7 // indirect
	github.com/rs/zerolog v1.34.0 // indirect
	github.com/zclconf/go-cty v1.14.4 // indirect
	github.com/zclconf/go-cty-yaml v1.1.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel v1.33.0 // indirect
	go.opentelemetry.io/otel/metric v1.33.0 // indirect
	go.opentelemetry.io/otel/trace v1.33.0 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/mod v0.23.0 // indirect
	golang.org/x/net v0.39.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20241209162323-e6fa225c2576 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241209162323-e6fa225c2576 // indirect
	google.golang.org/grpc v1.68.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/go-lynx/lynx-layout/api => ./api

replace github.com/go-lynx/lynx => ../lynx

replace github.com/go-lynx/plugins/service/grpc/v2 => ../lynx/plugins/service/grpc

replace github.com/go-lynx/plugins/service/http/v2 => ../lynx/plugins/service/http

replace github.com/go-lynx/plugins/db/redis/v2 => ../lynx/plugins/db/redis

replace github.com/go-lynx/plugins/db/mysql/v2 => ../lynx/plugins/db/mysql

replace github.com/go-lynx/plugins/polaris/v2 => ../lynx/plugins/polaris
