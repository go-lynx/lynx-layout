package main

import (
	"github.com/go-lynx/lynx/boot"
)

// start the user service
func main() {
	err := boot.NewApplication(wireApp).Run()
	if err != nil {
		panic(err)
	}
}
