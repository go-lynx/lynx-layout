<p align="center"><a href="https://go-lynx.cn/" target="_blank"><img width="120" src="https://avatars.githubusercontent.com/u/150900434?s=250&u=8f8e9a5d1fab6f321b4aa350283197fc1d100efa&v=4" alt="logo"></a></p>

<p align="center">
<a href="https://pkg.go.dev/github.com/go-lynx/lynx"><img src="https://pkg.go.dev/badge/github.com/go-lynx/lynx/v2" alt="GoDoc"></a>
<a href="https://codecov.io/gh/go-lynx/lynx"><img src="https://codecov.io/gh/go-lynx/lynx/master/graph/badge.svg" alt="codeCov"></a>
<a href="https://goreportcard.com/report/github.com/go-lynx/lynx"><img src="https://goreportcard.com/badge/github.com/go-lynx/lynx" alt="Go Report Card"></a>
<a href="https://github.com/go-lynx/lynx/blob/main/LICENSE"><img src="https://img.shields.io/github/license/go-lynx/lynx" alt="License"></a>
<a href="https://discord.gg/2vq2Zsqq"><img src="https://img.shields.io/discord/1174545542689337497?label=chat&logo=discord" alt="Discord"></a>
</p>


# Lynx-layout ：Lynx Microservice Template 

Welcome aboard! 🎉 This Microservice Template Project serves as a scaffold for quickly initiating microservice logic development. It's designed to facilitate the creation of microservices by providing a well-structured, easy-to-navigate project layout.

The project has a strong association with Polaris, utilizing its service discovery, rate limiting, and degradation functions.

Here's a tree-like representation of the project structure:

```
📦Microservice Template Project
 ┣ 📂api
 ┃ ┗📚 Protobuf files for API declaration and management
 ┣ 📂biz
 ┃ ┗🏢 Business logic, focusing on the overall process
 ┣ 📂bo
 ┃ ┗🔄 Data flow management between `biz` and `data` layers
 ┣ 📂code
 ┃ ┗🏷 Application status codes for quick issue identification
 ┣ 📂conf
 ┃ ┗⚙️ Configuration files with customizable mappings
 ┣ 📂data
 ┃ ┗💽 Data processing, including database and remote calls
 ┣ 📂service
 ┃ ┗🛎️ Service declarations, parameter validation, and data conversion
 ┗ 📂server
   ┗🖥️ Service configuration for API interfaces
```

## Quick Install

> If you want to use this microservice code template, all you need to do is execute the following command to install the Lynx CLI command-line tool, and then run the new command to automatically initialize a runnable project (the new command can support multiple project names).

```shell
go install github.com/go-lynx/lynx/cmd/lynx@latest
```

```shell
lynx new demo1 demo2 demo3
```

## Quick Start Code

To get your microservice up and running in no time, use the following code (Some functionalities can be plugged in or out based on your choice.):

```go
func main() {
    boot.LynxApplication(wireApp).Run()
}
```

This code initializes and runs the application with essential components like HTTP, gRPC with TLS, MySQL, Redis, Tracer, and Token.

仓库当前通过 `make wire` 自动发现 `cmd/**/wire.go` 所在目录，并逐个执行 Wire 生成依赖注入代码。本地开发如果修改了 Wire provider / injector，或首次需要补齐依赖注入生成物，可以先执行 `make wire`，然后再运行应用或执行构建命令；`make build` 也会隐式先执行这一步。

We hope this guide helps you navigate our Microservice Template Project. Happy coding! 🎉

## 本地开发（无需 Polaris）

如果只想在本地调试服务且不依赖 Polaris，可按照下面的步骤操作：

1. **使用 Go 1.25.3**
   ```bash
   go env -w GOTOOLCHAIN=go1.25.3
   ```
   或者确认 `go version` 输出为 `go1.25.3`。
2. **按需安装 Makefile 依赖工具**
   ```bash
   make init
   ```
   `make init` 会安装 `lynx`、`protoc-gen-go`、`protoc-gen-go-grpc` 和 `protoc-gen-openapi`。如果只直接执行 `go run`，且本地已经有可用的生成物，可以跳过这一步。
3. **启动本地依赖（PostgreSQL & Redis）**
   ```bash
   docker compose -f deployments/docker-compose.local.yml up -d
   ```
   该 compose 文件会启动 `postgres://lynx:lynx@127.0.0.1:5432/lynx` 与 `redis://127.0.0.1:6379`，并自动暴露到本机端口。
4. **按需重新生成 Wire 依赖注入代码**
   ```bash
   make wire
   ```
   这个目标会自动发现 `cmd/**/wire.go` 所在目录，并在每个目录下执行 `go run -mod=mod github.com/google/wire/cmd/wire`，直接复用仓库当前 `go.mod` 中锁定的 Wire 版本，不要求本机单独安装 `wire` 二进制。
5. **使用本地配置启动应用**（该配置不会加载 Polaris）
   ```bash
   go run ./cmd/user -conf ./configs/bootstrap.local.yaml
   ```
   如果你有自己的数据库/Redis，可以修改 `configs/bootstrap.local.yaml` 中的 `lynx.mysql` 与 `lynx.redis` 配置。注意：仓库当前 `deployments/docker-compose.local.yml` 启动的是 PostgreSQL 与 Redis，而 `bootstrap.local.yaml` 默认示例使用的是 MySQL；本地运行前请按你的实际依赖把二者对齐。
6. **调试完成后关闭依赖**
   ```bash
   docker compose -f deployments/docker-compose.local.yml down
   ```

默认的 `configs/bootstrap.yaml` 仍然保留对 Polaris 的配置，方便需要接入 Polaris 的环境使用。

## 可选：接入外部登录鉴权 gRPC 服务

模板默认不会强制依赖额外的鉴权微服务，`configs/bootstrap.local.yaml` 也没有预置这段配置。只有当你希望把登录后的 token 签发委托给其他服务时，才需要补充 `lynx.layout.auth.grpc.*` 或对应环境变量。

```yaml
lynx:
  layout:
    auth:
      grpc:
        service: auth-service
        method: /layout.auth.v1.AuthService/IssueLoginToken
        timeout: 5s
```

可选环境变量：

- `LYNX_LAYOUT_AUTH_GRPC_SERVICE`
- `LYNX_LAYOUT_AUTH_GRPC_METHOD`
- `LYNX_LAYOUT_AUTH_GRPC_TIMEOUT`

说明：

- 配置文件与环境变量可以同时存在，环境变量优先级更高。
- `timeout` 默认值为 `5s`。
- 如果当前服务不需要接入外部登录鉴权，可以保持这些配置为空，不会影响模板本身的本地启动说明；只有实际执行 `LoginAuth` 远程鉴权链路时，才需要提供有效配置。

## 构建命令

| 命令 | 说明 |
|------|------|
| `make wire` | 自动发现 `cmd/**/wire.go` 所在目录，并在每个目录下执行 `go run -mod=mod github.com/google/wire/cmd/wire`，更新 Wire 依赖注入生成代码 |
| `make build` | 目标依赖 `wire`，执行 `make build` 时会先跑 `make wire`；`wire` 成功后再执行 `mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/ ./...`，把构建产物输出到 `bin/` 目录 |
| `make test` | 直接执行 `go test ./...`，不会自动运行 `make wire` |
| `make all` | 顺序执行 `make api`、`make config`、`make generate`、`make ent`、`make validate`、`make build`；由于 `make build` 依赖 `wire`，所以 `make all` 也会隐式完成 Wire 生成后再构建 |

> **前置条件**：
> - `make wire` 需要当前 Go 环境可用，并能根据 `go.mod` 拉起仓库锁定版本的 `github.com/google/wire/cmd/wire`。
> - `make build` 隐含依赖 `make wire`，因此也有同样的 Go 环境与模块依赖前置条件。
> - `make test` 只依赖当前 Go 环境与仓库依赖，不会主动生成 Wire 代码。
> - `make all` 会串起生成链路与构建链路，因此也隐含依赖 `make build` / `make wire` 对应的 Go 环境与模块依赖。
