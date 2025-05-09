package websocket

import (
	"context"
	"encoding/json"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v2/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/xhcdpg/crypto-trade/models"
	"log"
	"sync"
)

type WebsocketService struct {
	clients    map[*websocket.Conn]string // conn -> userID
	clientMu   sync.Mutex
	publisher  message.Publisher
	subscriber message.Subscriber
}

func NewWebsocketService(publisher message.Publisher, amqpURI string) *WebsocketService {
	logger := watermill.NewStdLogger(false, false)
	amqpConfig := amqp.NewDurablePubSubConfig(amqpURI, nil)
	subscriber, err := amqp.NewSubscriber(amqpConfig, logger)
	if err != nil {
		log.Fatal("failed to create amqp subscriber", err)
	}
	return &WebsocketService{
		clients:    make(map[*websocket.Conn]string),
		publisher:  publisher,
		subscriber: subscriber,
	}
}

func (ws *WebsocketService) Start(ctx context.Context) {
	messages, err := ws.subscriber.Subscribe(ctx, "position_updated")
	if err != nil {
		log.Fatal("failed to subscribe to position_updated", err)
	}
	go func() {
		for msg := range messages {
			var position models.Position
			if err := json.Unmarshal(msg.Payload, &position); err != nil {
				log.Println("failed to unmarshal position", err)
				continue
			}
			ws.clientMu.Lock()
			for conn, userID := range ws.clients {
				if userID == position.UserID {
					if err := conn.WriteJSON(position); err != nil {
						log.Println("failed to send position to client", userID, err)
						conn.Close()
						delete(ws.clients, conn)
					}
				}
			}
			ws.clientMu.Unlock()
			msg.Ack()
		}
	}()
}

func (ws *WebsocketService) HandleConnection(conn *websocket.Conn) {
	defer conn.Close()
	var msg struct {
		Type    string      `json:"type"`
		UserID  string      `json:"user_id"`
		Token   string      `json:"token"`
		Payload interface{} `json:"payload"`
	}
	for {
		if err := conn.ReadJSON(&msg); err != nil {
			log.Println("failed to read message", err)
			ws.clientMu.Lock()
			delete(ws.clients, conn)
			ws.clientMu.Unlock()
			return
		}
		switch msg.Type {
		case "auth":
			user, err := ws.userService.GetUser(msg.UserID)
			if err != nil || user.Token != msg.Token {
				conn.WriteJSON(gin.H{"error": "auth failed"})
				return
			}
			ws.clientMu.Lock()
			ws.clients[conn] = msg.UserID
			ws.clientMu.Unlock()
			conn.WriteJSON(gin.H{
				"message": "auth success",
			})
		case "subscribe":
			var payload struct {
				Topic string `json:"topic"`
			}
			if err := json.Unmarshal(json.RawMessage(msg.Payload.(string)), &payload); err != nil {
				log.Println("failed to unmarshal payload", err)
				continue
			}
			conn.WriteJSON(gin.H{
				"message": "subscribe success",
				"topic":   payload.Topic,
			})
		case "ping":
			conn.WriteJSON(gin.H{"message": "pong"})
		default:
			conn.WriteJSON(gin.H{"error": "unknown type: " + msg.Type})
		}
	}
}
