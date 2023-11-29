package main

import (
	"github.com/go-lynx/lynx/boot"
	_ "go.uber.org/automaxprocs"
)

func main() {
	boot.LynxApplication(wireApp).Run()
}
