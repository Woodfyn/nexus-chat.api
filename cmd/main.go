package main

import "github.com/Woodfyn/chat-api-backend-go/internal/app"

// @title Social Network Backend
// @description API Server

// @host 51.44.7.199:8080
// @BasePath /api

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	app.Run()
}
