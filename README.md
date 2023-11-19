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

## Quick Start Code

To get your microservice up and running in no time, use the following code:

```go
func main() {
    boot.NewApp(
    wireApp,
    http.Http(),
    grpc.Grpc(grpc.EnableTls()),
    mysql.Mysql(),
    redis.Redis(),
    tracer.Tracer(),
    token.Token(login.NewLogin()),
    ).Run()
}
```

This code initializes and runs the application with essential components like HTTP, gRPC with TLS, MySQL, Redis, Tracer, and Token.

We hope this guide helps you navigate our Microservice Template Project. Happy coding! 🎉