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
    boot.NewApplication(wireApp).Run()
}
```

This code initializes and runs the application with essential components like HTTP, gRPC with TLS, MySQL, Redis, Tracer, and Token.

仓库当前通过 `make wire` 自动发现 `cmd/**/wire.go` 所在目录，并逐个执行 Wire 生成依赖注入代码。本地开发如果修改了 Wire provider / injector，或首次需要补齐依赖注入生成物，可以先执行 `make wire`，然后再运行应用或执行构建命令；`make build` 也会隐式先执行这一步。

## Bootstrap 装配约定

当前模板把运行时装配边界固定在 [`cmd/user/plugins.go`](./cmd/user/plugins.go) 与 [`cmd/user/providers.go`](./cmd/user/providers.go)：

- `plugins.go` 是唯一的 side-effect 插件注册清单，集中声明 HTTP、gRPC、MySQL、Redis、Redis Lock、Tracer 等运行时插件。
- `providers.go` 负责把 Lynx runtime 中已经加载好的 app、config、service registrar、HTTP/gRPC server、MySQL provider、Redis provider facade 显式导出给 Wire。
- `internal/data`、`internal/server`、`internal/service` 不再直接读取 `lynx.Lynx()` 或插件 helper，模板内部层只消费显式注入的依赖。

这意味着模板的装配链路已经固定为：

1. `boot.NewApplication(wireApp)` 创建 boot shell；
2. shell 完成 Lynx app 初始化并加载插件；
3. `wireApp` 从 `cmd/user` provider 层显式提取 runtime 依赖；
4. `internal/*` 只接收已经准备好的 provider / transport / config。

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
3. **启动本地依赖（MySQL & Redis）**
   ```bash
   docker compose -f deployments/docker-compose.local.yml up -d
   ```
   该 compose 文件会启动 `mysql://lynx:lynx123456@tcp(127.0.0.1:3306)/lynx_test` 与无密码的 `redis://127.0.0.1:6379`，并自动暴露到本机端口。
4. **按需重新生成 Wire 依赖注入代码**
   ```bash
   make wire
   ```
   这个目标会自动发现 `cmd/**/wire.go` 所在目录，并在每个目录下执行 `go run -mod=mod github.com/google/wire/cmd/wire`，直接复用仓库当前 `go.mod` 中锁定的 Wire 版本，不要求本机单独安装 `wire` 二进制。
   如果你修改了 [`cmd/user/plugins.go`](./cmd/user/plugins.go) 或 [`cmd/user/providers.go`](./cmd/user/providers.go)，也需要重新执行这一步，确保显式装配代码与 `wire_gen.go` 保持一致。
5. **使用本地配置启动应用**（该配置不会加载 Polaris）
   ```bash
   go run ./cmd/user -conf ./configs/bootstrap.local.yaml
   ```
   如果你有自己的数据库/Redis，可以修改 `configs/bootstrap.local.yaml` 中的 `lynx.mysql` 与 `lynx.redis` 配置。仓库当前的 `deployments/docker-compose.local.yml` 已与默认示例对齐，直接使用即可跑通本地 MySQL + Redis 依赖；默认 Redis 示例不带密码，如果你切换到带密码实例，需要同时更新 `lynx.redis.password`。模板运行期会通过 `cmd/user/providers.go` 显式拿到 MySQL provider、Redis provider facade、HTTP/gRPC server 和 service registrar。
6. **调试完成后关闭依赖**
   ```bash
   docker compose -f deployments/docker-compose.local.yml down
   ```

默认的 `configs/bootstrap.yaml` 仍然保留对 Polaris 的配置，方便需要接入 Polaris 的环境使用。

## 生产使用说明

### 1. Local bootstrap 合同

模板当前把本地无 control-plane 的运行合同固定在 4 个文件上：

- `configs/bootstrap.local.yaml`
- `tests/integration/testdata/bootstrap.local.yaml`
- `deployments/docker-compose.local.yml`
- `cmd/user/providers.go`

这 4 处需要保持同一口径：

- 本地链路默认不接 Polaris / Apollo / Nacos / Etcd 等 control plane。
- Redis 默认使用无密码的 `127.0.0.1:6379`。
- `provideServiceRegistrar()` 在 no-control-plane 路径下返回 `nil, nil` 是合法结果，不应该被模板层当成启动失败。
- 模板业务层只消费显式注入的 provider / transport，不回退到 `internal/*` 里直接读取全局 app。

如果你修改了本地 MySQL / Redis / gRPC 地址或鉴权配置，请同时同步 bootstrap、integration fixture 与 README，避免文档、测试夹具和本地 compose 口径漂移。

### 2. 本地冒烟前置条件

本地冒烟链路仍然是：

```bash
docker compose -f deployments/docker-compose.local.yml up -d
go run ./cmd/user -conf ./configs/bootstrap.local.yaml
```

但这条链路依赖本机 Docker daemon 先可用。如果执行 `docker version` 或 `docker compose` 时出现 `Cannot connect to the Docker daemon`，这是环境阻塞，不是模板代码失败；需要先把 Docker Desktop / daemon 拉起，再继续做本地启动冒烟。

### 3. no-control-plane 合法路径

模板已经验证过 local no-control-plane 路径：

- `boot.NewApplication(wireApp)` 负责发布默认 app，并把 runtime shell 挂到 Lynx core。
- `cmd/user/providers.go` 从已发布 app 中显式提取 config、MySQL provider、Redis provider facade、HTTP/gRPC server。
- 没有 control plane 时，service registrar 可以为 `nil`；Kratos 应用仍可正常完成本地启动，不要求强制注册中心。

这意味着 `bootstrap.local.yaml` 本身就是生产前的最小本地回归入口，不需要为了本地调试额外伪造注册中心。

### 4. 与 transport readiness / resource alias 的边界

任务12已经把 transport/message/transaction/observability 插件统一到了 `<plugin>.plugin`、`<plugin>.readiness`、`<plugin>.health` 等 runtime contract。模板侧本轮只读对照后的结论是：

- `lynx-layout` 当前不直接消费这些 readiness / health / alias 资源。
- 模板侧唯一需要的 transport 接口仍然是 `cmd/user/providers.go` 提供的 HTTP / gRPC base server，然后在 `internal/server` 完成业务路由注册。
- 因此模板侧不需要为任务12再新增消费代码；真正的运行态健康判断应由 transport 插件自身的 runtime health / endpoint / readiness contract 提供。

### 5. 最终交付口径

当前模板侧可以直接向 owner 交付的说明口径如下：

就 production 交付判断而言，当前模板已经收敛到一条稳定口径：运行时依赖通过显式 provider 装配进入模板，本地允许 no-control-plane 启动；如果 `docker compose` 无法拉起，则应先归因为 Docker daemon 环境阻塞，而不是把它判定为当前模板改造的代码 blocker。

- 显式装配边界固定在 `cmd/user/plugins.go` 与 `cmd/user/providers.go`，其中 Redis 已统一为 provider facade 注入，不再把 raw client 作为模板单例依赖向下传递。
- `configs/bootstrap.local.yaml` 代表标准 local no-control-plane 路径；`provideServiceRegistrar()` 返回 `nil, nil` 属于合法启动分支，不构成启动失败。
- 本机无法执行 `docker compose ... && go run ...` 时，需要先排除 Docker daemon 环境阻塞；这属于环境问题，不属于当前模板代码 blocker。
- transport readiness / health / alias 资源继续由插件运行时自身负责，模板只消费 HTTP / gRPC base server，不额外承担 transport control-plane/resource contract 的解释责任。

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
