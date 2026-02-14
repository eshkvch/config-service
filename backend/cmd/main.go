package main

import "config-service/backend/internal/di"

func main() {
	app := di.NewApp()
	app.Run()
}
