package main

import "github.com/Woodfyn/chat-api-backend-go/internal/app"

// @title Social Network Backend
// @description API Server

// @host 35.180.225.96:8080
// @BasePath /api

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	app.Run()
}
