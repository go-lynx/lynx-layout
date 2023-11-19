package main

import (
	"github.com/go-lynx/lynx/boot"
	"github.com/go-lynx/lynx/plugin/grpc"
	"github.com/go-lynx/lynx/plugin/http"
	"github.com/go-lynx/lynx/plugin/mysql"
	"github.com/go-lynx/lynx/plugin/redis"
	"github.com/go-lynx/lynx/plugin/token"
	"github.com/go-lynx/lynx/plugin/token/login"
	"github.com/go-lynx/lynx/plugin/tracer"
	_ "go.uber.org/automaxprocs"
)

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
