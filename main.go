package main

import (
	"database/sql"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/xhcdpg/crypto-trade/matching"
	"github.com/xhcdpg/crypto-trade/position"
	"github.com/xhcdpg/crypto-trade/risk"
	u "github.com/xhcdpg/crypto-trade/user"
	"github.com/xhcdpg/crypto-trade/websocket"
)

type Config struct {
	DatabaseURL string
	RedisURL    string
	AmqpURL     string
	HttpPort    string
}

type App struct {
	db        *sql.DB
	redis     *redis.Client
	pubSub    message.Publisher
	matching  *matching.MatchingEngine
	risk      *risk.RiskManager
	position  *position.PositionManager
	user      *u.UserService
	websocket *websocket.WebsocketService
	router    *gin.Engine
}
