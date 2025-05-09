package main

import "database/sql"

type Config struct {
	DatabaseURL string
	RedisURL    string
	AmqpURL     string
	HttpPort    string
}

type App struct {
	db        *sql.DB
	redis     *redis.Client
	pubSub    message.PubSub
	matching  *matching.MatchingEngine
	risk      *risk.RiskManager
	position  *position.PositionManager
	user      *user.Service
	websocket *websocket.WebsocketService
	router    *gin.Engine
}
