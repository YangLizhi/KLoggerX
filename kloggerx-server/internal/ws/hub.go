package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"

	"kloggerx-server/internal/pkg/jwt"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type ClientMessage struct {
	MsgType int
	Data    []byte
}

type Client struct {
	Hub        *Hub
	Conn       *websocket.Conn
	Send       chan ClientMessage
	UserID     uint
	UserName   string
	DocumentID uint
	Color      string
}

type Hub struct {
	rooms      map[uint]map[*Client]bool
	broadcast  chan *RoomMessage
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

type RoomMessage struct {
	DocumentID uint
	MsgType    int
	Data       []byte
	Sender     *Client
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[uint]map[*Client]bool),
		broadcast:  make(chan *RoomMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.rooms[client.DocumentID] == nil {
				h.rooms[client.DocumentID] = make(map[*Client]bool)
			}
			h.rooms[client.DocumentID][client] = true
			h.mu.Unlock()
			h.broadcastCollaborators(client.DocumentID)

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.rooms[client.DocumentID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.rooms, client.DocumentID)
					}
				}
			}
			h.mu.Unlock()
			h.broadcastCollaborators(client.DocumentID)
			h.broadcastToRoom(client.DocumentID, Message{
				Type: "user_leave",
				Data: map[string]interface{}{"userId": client.UserID},
			}, nil)

		case msg := <-h.broadcast:
			h.mu.RLock()
			clients := h.rooms[msg.DocumentID]
			for c := range clients {
				if c != msg.Sender {
					select {
					case c.Send <- ClientMessage{MsgType: msg.MsgType, Data: msg.Data}:
					default:
						close(c.Send)
						delete(clients, c)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

var colors = []string{"#3370ff", "#f54a45", "#36b37e", "#ff7d00", "#9254de", "#00b8d9", "#f5a623", "#eb2f96"}

func (h *Hub) broadcastCollaborators(docID uint) {
	h.mu.RLock()
	clients := h.rooms[docID]
	var collaborators []map[string]interface{}
	i := 0
	for c := range clients {
		collaborators = append(collaborators, map[string]interface{}{
			"userId":   c.UserID,
			"userName": c.UserName,
			"color":    colors[i%len(colors)],
			"isOnline": true,
		})
		i++
	}
	h.mu.RUnlock()

	h.broadcastToRoom(docID, Message{Type: "collaborators", Data: collaborators}, nil)
}

func (h *Hub) broadcastToRoom(docID uint, msg Message, exclude *Client) {
	data, _ := json.Marshal(msg)
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.rooms[docID] {
		if c != exclude {
			select {
			case c.Send <- ClientMessage{MsgType: websocket.TextMessage, Data: data}:
			default:
			}
		}
	}
}

func HandleWebSocket(hub *Hub, c *gin.Context) {
	token := c.Query("token")
	claims, err := jwt.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	docIDStr := c.Param("id")
	docID, _ := strconv.ParseUint(docIDStr, 10, 64)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		Hub:        hub,
		Conn:       conn,
		Send:       make(chan ClientMessage, 256),
		UserID:     claims.UserID,
		UserName:   claims.Username,
		DocumentID: uint(docID),
	}

	hub.register <- client

	go client.writePump()
	go client.readPump()
}

// HandleYjsWebSocket handles Yjs binary sync messages for real-time collaboration
func HandleYjsWebSocket(hub *Hub, c *gin.Context) {
	token := c.Query("token")
	claims, err := jwt.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	docIDStr := c.Param("id")
	docID, _ := strconv.ParseUint(docIDStr, 10, 64)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Yjs WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		Hub:        hub,
		Conn:       conn,
		Send:       make(chan ClientMessage, 256),
		UserID:     claims.UserID,
		UserName:   claims.Username,
		DocumentID: uint(docID),
	}

	// For Yjs, we use a separate hub namespace by offsetting document IDs
	// to avoid conflicts with the text message hub
	yjsDocID := uint(docID) + 1000000
	client.DocumentID = yjsDocID

	hub.register <- client

	go client.writePump()
	go client.readYjsPump()
}

func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()
	for {
		msgType, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		if msgType == websocket.BinaryMessage {
			// Forward binary messages (Yjs sync data) directly
			c.Hub.broadcast <- &RoomMessage{
				DocumentID: c.DocumentID,
				MsgType:    websocket.BinaryMessage,
				Data:       message,
				Sender:     c,
			}
			continue
		}

		// Text messages: JSON protocol
		var msg Message
		if json.Unmarshal(message, &msg) != nil {
			continue
		}
		if msg.Type == "ping" {
			c.Send <- ClientMessage{
				MsgType: websocket.TextMessage,
				Data:    mustJSON(Message{Type: "pong", Data: nil}),
			}
			continue
		}
		c.Hub.broadcast <- &RoomMessage{
			DocumentID: c.DocumentID,
			MsgType:    websocket.TextMessage,
			Data:       message,
			Sender:     c,
		}
	}
}

func (c *Client) readYjsPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()
	for {
		msgType, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		// Forward all messages (binary Yjs data) to room
		c.Hub.broadcast <- &RoomMessage{
			DocumentID: c.DocumentID,
			MsgType:    msgType,
			Data:       message,
			Sender:     c,
		}
	}
}

func (c *Client) writePump() {
	defer c.Conn.Close()
	for msg := range c.Send {
		if err := c.Conn.WriteMessage(msg.MsgType, msg.Data); err != nil {
			break
		}
	}
}

func mustJSON(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}
