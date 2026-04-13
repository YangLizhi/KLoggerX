package ws

import (
	"sync"
	"time"
)

// === varUint 编解码 ===
// y-protocols 使用变长整数编码（每字节 7 位数据 + 1 位延续标志）

// writeVarUint 编码 varUint 追加到 buf
func writeVarUint(buf []byte, num uint64) []byte {
	for {
		// 取低 7 位
		b := byte(num & 0x7f)
		num >>= 7
		if num > 0 {
			// 还有更多位，设置延续标志
			b |= 0x80
		}
		buf = append(buf, b)
		if num == 0 {
			break
		}
	}
	return buf
}

// readVarUint 从 offset 读取 varUint，返回值和新 offset
func readVarUint(data []byte, offset int) (uint64, int) {
	var num uint64
	var shift uint
	for {
		if offset >= len(data) {
			return num, offset
		}
		b := data[offset]
		offset++
		num |= uint64(b&0x7f) << shift
		if b&0x80 == 0 {
			break
		}
		shift += 7
	}
	return num, offset
}

// writeVarByteArray 写入 varUint(len) + raw bytes
func writeVarByteArray(buf []byte, data []byte) []byte {
	buf = writeVarUint(buf, uint64(len(data)))
	buf = append(buf, data...)
	return buf
}

// readVarByteArray 从 offset 读取 varByteArray
func readVarByteArray(data []byte, offset int) ([]byte, int) {
	length, offset := readVarUint(data, offset)
	if offset+int(length) > len(data) {
		return nil, offset
	}
	result := data[offset : offset+int(length)]
	return result, offset + int(length)
}

// === 协议常量 ===
const (
	MessageSync      = 0
	MessageAwareness = 1

	SyncStep1  = 0
	SyncStep2  = 1
	SyncUpdate = 2
)

// === YjsRoom - 每个文档一个房间 ===
type YjsRoom struct {
	documentID uint
	updates    [][]byte          // 累积的原始 Update 数据（不含协议头）
	awareness  map[uint][]byte   // clientID -> 最新 awareness 状态
	clients    map[*Client]bool
	mu         sync.RWMutex
	lastActive time.Time
}

// === YjsManager - 全局管理器 ===
type YjsManager struct {
	rooms map[uint]*YjsRoom // documentID -> room
	mu    sync.RWMutex
}

// GlobalYjsManager 全局 Yjs 房间管理器
var GlobalYjsManager = &YjsManager{rooms: make(map[uint]*YjsRoom)}

// GetOrCreateRoom 获取或创建文档房间
func (m *YjsManager) GetOrCreateRoom(docID uint) *YjsRoom {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, exists := m.rooms[docID]
	if !exists {
		room = &YjsRoom{
			documentID: docID,
			updates:    make([][]byte, 0),
			awareness:  make(map[uint][]byte),
			clients:    make(map[*Client]bool),
			lastActive: time.Now(),
		}
		m.rooms[docID] = room
	}
	return room
}

// CleanupIdleRooms 清理空闲超过 5 分钟的房间
func (m *YjsManager) CleanupIdleRooms() {
	m.mu.Lock()
	defer m.mu.Unlock()

	timeout := 5 * time.Minute
	for docID, room := range m.rooms {
		room.mu.RLock()
		isEmpty := len(room.clients) == 0
		idle := time.Since(room.lastActive) > timeout
		room.mu.RUnlock()

		if isEmpty && idle {
			delete(m.rooms, docID)
		}
	}
}

// === 房间方法 ===

// AddClient 客户端加入房间
func (r *YjsRoom) AddClient(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[c] = true
	r.lastActive = time.Now()
}

// RemoveClient 客户端离开房间
func (r *YjsRoom) RemoveClient(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clients, c)
	r.lastActive = time.Now()
}

// HandleMessage 处理收到的二进制消息
// 解析协议类型和消息子类型，执行对应逻辑
func (r *YjsRoom) HandleMessage(sender *Client, data []byte) {
	if len(data) < 1 {
		return
	}

	protoType, offset := readVarUint(data, 0)

	switch protoType {
	case MessageSync:
		r.handleSyncMessage(sender, data, offset)
	case MessageAwareness:
		r.handleAwarenessMessage(sender, data)
	default:
		// 未知协议，直接广播
		r.broadcastExcept(sender, data)
	}
}

// handleSyncMessage 处理同步协议消息
func (r *YjsRoom) handleSyncMessage(sender *Client, data []byte, offset int) {
	msgType, offset := readVarUint(data, offset)

	switch msgType {
	case SyncStep1:
		// 客户端请求同步
		// 1. 回复空 SyncStep2（服务端无独立 CRDT 状态）
		r.sendEmptySyncStep2(sender)
		// 2. 发送所有累积的更新给新客户端
		r.sendAccumulatedUpdates(sender)
		// 3. 转发 SyncStep1 给其他客户端（让他们响应）
		r.broadcastExcept(sender, data)

	case SyncStep2:
		// 来自其他客户端的同步响应
		// 提取 update 数据并存储
		updateData, _ := readVarByteArray(data, offset)
		if len(updateData) > 0 {
			r.storeUpdate(updateData)
		}
		// 转发给其他客户端
		r.broadcastExcept(sender, data)

	case SyncUpdate:
		// 增量更新
		updateData, _ := readVarByteArray(data, offset)
		if len(updateData) > 0 {
			r.storeUpdate(updateData)
		}
		// 广播给其他客户端
		r.broadcastExcept(sender, data)
	}
}

// handleAwarenessMessage 处理 awareness 消息（直接广播）
func (r *YjsRoom) handleAwarenessMessage(sender *Client, data []byte) {
	// 直接广播给房间内其他客户端
	r.broadcastExcept(sender, data)
}

// sendEmptySyncStep2 发送空的 SyncStep2 响应
func (r *YjsRoom) sendEmptySyncStep2(c *Client) {
	// 构造消息：[0(sync)] [1(step2)] [0(empty update length)]
	msg := make([]byte, 0, 8)
	msg = writeVarUint(msg, MessageSync)
	msg = writeVarUint(msg, SyncStep2)
	msg = writeVarByteArray(msg, []byte{})
	c.SendBinary(msg)
}

// sendAccumulatedUpdates 发送累积的更新给客户端
func (r *YjsRoom) sendAccumulatedUpdates(c *Client) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, update := range r.updates {
		// 构造 Update 消息：[0(sync)] [2(update)] [varByteArray(update)]
		msg := make([]byte, 0, len(update)+16)
		msg = writeVarUint(msg, MessageSync)
		msg = writeVarUint(msg, SyncUpdate)
		msg = writeVarByteArray(msg, update)
		c.SendBinary(msg)
	}
}

// storeUpdate 存储更新
func (r *YjsRoom) storeUpdate(update []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	// 存储副本
	stored := make([]byte, len(update))
	copy(stored, update)
	r.updates = append(r.updates, stored)
	r.lastActive = time.Now()

	// 限制累积更新数量，超过 1000 条时保留后 500 条
	if len(r.updates) > 1000 {
		r.updates = r.updates[len(r.updates)-500:]
	}
}

// broadcastExcept 广播给房间内除 sender 外的所有客户端
func (r *YjsRoom) broadcastExcept(sender *Client, data []byte) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for c := range r.clients {
		if c != sender {
			c.SendBinary(data)
		}
	}
}
