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

We hope this guide helps you navigate our Microservice Template Project. Happy coding! 🎉

## 本地开发（无需 Polaris）

如果只想在本地调试服务且不依赖 Polaris，可按照下面的步骤操作：

1. **使用 Go 1.25.3**  
   ```bash
   go env -w GOTOOLCHAIN=go1.25.3
   ```
   或者确认 `go version` 输出为 `go1.25.3`。
2. **启动本地依赖（PostgreSQL & Redis）**  
   ```bash
   docker compose -f deployments/docker-compose.local.yml up -d
   ```
   该 compose 文件会启动 `postgres://lynx:lynx@127.0.0.1:5432/lynx` 与 `redis://127.0.0.1:6379`，并自动暴露到本机端口。
3. **使用本地配置启动应用**（该配置不会加载 Polaris）  
   ```bash
   go run ./cmd/user -conf ./configs/bootstrap.local.yaml
   ```
   如果你有自己的数据库/Redis，可以修改 `configs/bootstrap.local.yaml` 中的 `lynx.pgsql` 与 `lynx.redis` 配置。
4. **调试完成后关闭依赖**  
   ```bash
   docker compose -f deployments/docker-compose.local.yml down
   ```

默认的 `configs/bootstrap.yaml` 仍然保留对 Polaris 的配置，方便需要接入 Polaris 的环境使用。
