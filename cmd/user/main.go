package main

import (
	"fmt"
	"os"

	"github.com/go-lynx/lynx/boot"
)

// start the user service
func main() {
	app := boot.NewApplication(wireApp)
	if app == nil {
		_, _ = fmt.Fprintln(os.Stderr, "failed to create user service application")
		os.Exit(1)
	}

	if err := app.Run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to run user service: %v\n", err)
		os.Exit(1)
	}
}
