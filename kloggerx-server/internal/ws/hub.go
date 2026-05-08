package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"kloggerx-server/internal/pkg/jwt"
	"kloggerx-server/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin:  func(r *http.Request) bool { return true },
	Subprotocols: []string{"access_token"},
}

const (
	MaxConnectionsPerUser = 10 // 单用户最多10个WS连接
	MaxConnectionsPerRoom = 50 // 单房间最多50个连接
	pingInterval          = 30 * time.Second
	pongTimeout           = 90 * time.Second
	writeWait             = 10 * time.Second
)

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
	UserAvatar string
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

// countUserConnections 统计某用户当前所有连接数（需持有读锁或写锁）
func (h *Hub) countUserConnections(userID uint) int {
	count := 0
	for _, clients := range h.rooms {
		for c := range clients {
			if c.UserID == userID {
				count++
			}
		}
	}
	return count
}

func (h *Hub) Run() {
	// 定期清理空闲的 Yjs 房间（每分钟一次）
	cleanupTicker := time.NewTicker(1 * time.Minute)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-cleanupTicker.C:
			GlobalYjsManager.CleanupIdleRooms()

		case client := <-h.register:
			h.mu.Lock()
			// 检查用户连接数
			userConns := h.countUserConnections(client.UserID)
			if userConns >= MaxConnectionsPerUser {
				h.mu.Unlock()
				client.Conn.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(4001, "连接数超限"))
				client.Conn.Close()
				continue
			}
			// 检查房间连接数
			if len(h.rooms[client.DocumentID]) >= MaxConnectionsPerRoom {
				h.mu.Unlock()
				client.Conn.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(4002, "房间已满"))
				client.Conn.Close()
				continue
			}
			if h.rooms[client.DocumentID] == nil {
				h.rooms[client.DocumentID] = make(map[*Client]bool)
			}
			h.rooms[client.DocumentID][client] = true
			h.mu.Unlock()
			h.broadcastCollaborators(client.DocumentID)
			h.broadcastToRoom(client.DocumentID, Message{
				Type: "user_join",
				Data: map[string]interface{}{
					"userId":    client.UserID,
					"userName":  client.UserName,
					"userAvatar": client.UserAvatar,
					"color":     client.Color,
					"isOnline":  true,
				},
			}, nil)
			// Record join event asynchronously (non-blocking)
			// Note: For Yjs connections, documentID is offset by 1000000
			actualDocID := client.DocumentID
			if actualDocID > 1000000 {
				actualDocID = actualDocID - 1000000
			}
			service.RecordCollaborateEventAsync("join", actualDocID, client.UserID, "加入了文档协作")

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
			// Record leave event asynchronously (non-blocking)
			// Note: For Yjs connections, documentID is offset by 1000000
			actualDocID := client.DocumentID
			if actualDocID > 1000000 {
				actualDocID = actualDocID - 1000000
			}
			service.RecordCollaborateEventAsync("leave", actualDocID, client.UserID, "离开了文档协作")

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

// 12 distinct colors for collaboration
var colors = []string{
	"#3370ff", // 蓝色
	"#f54a45", // 红色
	"#36b37e", // 绿色
	"#ff7d00", // 橙色
	"#9254de", // 紫色
	"#00b8d9", // 青色
	"#f5a623", // 金色
	"#eb2f96", // 粉色
	"#722ed1", // 深紫
	"#13c2c2", // 青绿
	"#fa8c16", // 橙黄
	"#52c41a", // 草绿
}

// getUserColor returns a fixed color for a user based on their userId hash
func getUserColor(userID uint) string {
	// Simple hash: use userID modulo number of colors
	index := int(userID) % len(colors)
	return colors[index]
}

func (h *Hub) broadcastCollaborators(docID uint) {
	h.mu.RLock()
	clients := h.rooms[docID]
	var collaborators []map[string]interface{}
	for c := range clients {
		collaborators = append(collaborators, map[string]interface{}{
			"userId":    c.UserID,
			"userName":  c.UserName,
			"userAvatar": c.UserAvatar,
			"color":     c.Color,
			"isOnline":  true,
		})
	}
	h.mu.RUnlock()

	h.broadcastToRoom(docID, Message{Type: "collaborators_update", Data: collaborators}, nil)
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

// getTokenFromRequest 从请求中获取token（按优先级）
// 1. Sec-WebSocket-Protocol header（格式: "access_token, <actual-token>"）
// 2. Query parameter（deprecated，兼容旧客户端）
func getTokenFromRequest(c *gin.Context) string {
	// 优先从 Sec-WebSocket-Protocol 获取
	protocols := c.GetHeader("Sec-WebSocket-Protocol")
	if protocols != "" {
		parts := strings.Split(protocols, ", ")
		if len(parts) >= 2 && parts[0] == "access_token" {
			return parts[1]
		}
	}
	// fallback to query param (deprecated)
	return c.Query("token")
}

func HandleWebSocket(hub *Hub, c *gin.Context) {
	token := getTokenFromRequest(c)
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

	// Get user avatar from database
	user, err := service.GetUserByID(claims.UserID)
	userAvatar := ""
	if err == nil && user != nil {
		userAvatar = user.Avatar
	}

	client := &Client{
		Hub:        hub,
		Conn:       conn,
		Send:       make(chan ClientMessage, 256),
		UserID:     claims.UserID,
		UserName:   claims.Username,
		UserAvatar: userAvatar,
		DocumentID: uint(docID),
		Color:      getUserColor(claims.UserID),
	}

	hub.register <- client

	go client.writePump()
	go client.readPump()
}

// HandleYjsWebSocket handles Yjs binary sync messages for real-time collaboration
func HandleYjsWebSocket(hub *Hub, c *gin.Context) {
	token := getTokenFromRequest(c)
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

	// Get user avatar from database
	user, err := service.GetUserByID(claims.UserID)
	userAvatar := ""
	if err == nil && user != nil {
		userAvatar = user.Avatar
	}

	client := &Client{
		Hub:        hub,
		Conn:       conn,
		Send:       make(chan ClientMessage, 256),
		UserID:     claims.UserID,
		UserName:   claims.Username,
		UserAvatar: userAvatar,
		DocumentID: uint(docID),
		Color:      getUserColor(claims.UserID),
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
	c.Conn.SetReadDeadline(time.Now().Add(pongTimeout))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongTimeout))
		return nil
	})
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
	// 获取真实的文档 ID（去除偏移量）
	actualDocID := c.DocumentID
	if actualDocID > 1000000 {
		actualDocID = actualDocID - 1000000
	}

	// 获取或创建 Yjs 房间
	yjsRoom, _ := GlobalYjsManager.GetOrCreateRoom(actualDocID)
	yjsRoom.AddClient(c)

	defer func() {
		c.Hub.unregister <- c
		yjsRoom.RemoveClient(c)
		c.Conn.Close()
	}()

	for {
		msgType, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		// 只处理二进制消息（Yjs 协议）
		if msgType == websocket.BinaryMessage {
			yjsRoom.HandleMessage(c, message)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.Send:
			if !ok {
				// channel已关闭
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(msg.MsgType, msg.Data); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SendBinary 发送二进制消息（用于 Yjs 协议）
func (c *Client) SendBinary(data []byte) {
	select {
	case c.Send <- ClientMessage{MsgType: websocket.BinaryMessage, Data: data}:
	default:
		// channel 满了，客户端可能已断开
	}
}

func mustJSON(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}
