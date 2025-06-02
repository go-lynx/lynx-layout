package main

import (
	"github.com/go-lynx/lynx/boot"
	_ "go.uber.org/automaxprocs"
)

func main() {
	err := boot.NewLynxApplication(wireApp).Run()
	if err != nil {
		panic(err)
	}
}
