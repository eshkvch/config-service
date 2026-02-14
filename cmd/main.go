package main

import "config-service/internal/di"

func main() {
	app := di.NewApp()
	app.Run()
}
