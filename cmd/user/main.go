package main

import (
	"github.com/go-lynx/lynx/boot"
	_ "go.uber.org/automaxprocs"
)

// start the user service
func main() {
	err := boot.NewLynxApplication(wireApp).Run()
	if err != nil {
		panic(err)
	}
}
