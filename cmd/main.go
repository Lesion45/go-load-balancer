package main

import (
	"github.com/Lesion45/go-load-balancer/internal/app"
)

func main() {
	application := app.NewApp()
	application.Run()
}
