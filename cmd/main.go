package main

import "github.com/Woodfyn/chat-api-backend-go/internal/app"

// @title Social Network Backend
// @description API Server

// @host 51.44.24.230:8080
// @BasePath /api

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	app.Run()
}
