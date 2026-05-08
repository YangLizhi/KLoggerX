package ws

import (
	"context"
	"fmt"
	"sync"
	"time"

	"kloggerx-server/internal/pkg/logger"
	rds "kloggerx-server/internal/repository/redis"

	"github.com/redis/go-redis/v9"
)

// === varUint 编解码 ===
// y-protocols 使用变长整数编码（每字节 7 位数据 + 1 位延续标志）

// writeVarUint 编码 varUint 追加到 buf
func writeVarUint(buf []byte, num uint64) []byte {
	for {
		b := byte(num & 0x7f)
		num >>= 7
		if num > 0 {
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
	length, newOffset := readVarUint(data, offset)
	if newOffset+int(length) > len(data) {
		return nil, newOffset
	}
	result := data[newOffset : newOffset+int(length)]
	return result, newOffset + int(length)
}

// === 协议常量 ===
const (
	MessageSync      = 0
	MessageAwareness = 1

	SyncStep1  = 0
	SyncStep2  = 1
	SyncUpdate = 2
)

// === Redis 持久化 ===
const yjsDocKeyPrefix = "yjs:doc:"

// SaveYjsState 保存Y.Doc状态到Redis（无过期时间）
func SaveYjsState(docID uint, state []byte) error {
	if len(state) == 0 {
		return nil
	}
	key := fmt.Sprintf("%s%d", yjsDocKeyPrefix, docID)
	return rds.RDB.Set(context.Background(), key, state, 0).Err()
}

// LoadYjsState 从Redis加载Y.Doc状态
func LoadYjsState(docID uint) ([]byte, error) {
	key := fmt.Sprintf("%s%d", yjsDocKeyPrefix, docID)
	result, err := rds.RDB.Get(context.Background(), key).Bytes()
	if err == redis.Nil {
		return nil, nil // 无保存的状态
	}
	return result, err
}

// === YjsRoom - 每个文档一个房间 ===
type YjsRoom struct {
	documentID uint
	updates    [][]byte        // 累积的原始 Update 数据（不含协议头）
	awareness  map[uint][]byte // clientID -> 最新 awareness 状态
	clients    map[*Client]bool
	dirty      bool         // 是否有未保存到Redis的变更
	saveTicker *time.Ticker // 定期自动保存
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

// GetOrCreateRoom 获取或创建文档房间，isFirst 表示是否为房间第一个客户端
func (m *YjsManager) GetOrCreateRoom(docID uint) (room *YjsRoom, isFirst bool) {
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

		// 从 Redis 加载已有状态
		state, err := LoadYjsState(docID)
		if err != nil {
			logger.Errorf("Yjs load state from Redis failed: docID=%d, err=%v", docID, err)
		} else if len(state) > 0 {
			parsed := loadPersistedUpdates(state)
			if len(parsed) > 0 {
				room.updates = parsed
			} else {
				// 兼容旧格式：整个state作为单个update
				room.updates = append(room.updates, state)
			}
			logger.Infof("Yjs state loaded from Redis: docID=%d, size=%d bytes, updates=%d", docID, len(state), len(room.updates))
		}

		// 启动自动保存
		room.startAutoSave()
		isFirst = true
	}
	return room, isFirst
}

// GetRoom 获取房间（不创建）
func (m *YjsManager) GetRoom(docID uint) *YjsRoom {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.rooms[docID]
}

// CleanupIdleRooms 清理空闲超过 5 分钟的空房间
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
			room.stopAutoSave()
			delete(m.rooms, docID)
			logger.Infof("Yjs room cleaned up: docID=%d", docID)
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
	logger.Infof("Yjs client connected: docID=%d, userID=%d, clients=%d", r.documentID, c.UserID, len(r.clients))
}

// RemoveClient 客户端离开房间，返回房间是否已空
func (r *YjsRoom) RemoveClient(c *Client) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clients, c)
	r.lastActive = time.Now()
	isEmpty := len(r.clients) == 0

	logger.Infof("Yjs client disconnected: docID=%d, userID=%d, remaining=%d", r.documentID, c.UserID, len(r.clients))

	// 最后一个用户离开时，强制保存状态到Redis
	if isEmpty && r.dirty {
		state := r.mergeAllUpdates()
		if err := SaveYjsState(r.documentID, state); err != nil {
			logger.Errorf("Yjs save state on last leave failed: docID=%d, err=%v", r.documentID, err)
		} else {
			r.dirty = false
			logger.Infof("Yjs state saved on last leave: docID=%d, size=%d bytes", r.documentID, len(state))
		}
	}

	return isEmpty
}

// HandleMessage 处理收到的二进制消息
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
		// 客户端请求同步：发送服务端累积的状态作为 SyncStep2
		r.sendSyncStep2WithState(sender)
		// 转发 SyncStep1 给其他客户端（让他们也发送自身状态）
		r.broadcastExcept(sender, data)

	case SyncStep2:
		// 来自其他客户端的同步响应，提取 update 数据并存储
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

// handleAwarenessMessage 处理 awareness 消息（光标位置等）
func (r *YjsRoom) handleAwarenessMessage(sender *Client, data []byte) {
	// 直接广播给房间内其他客户端
	r.broadcastExcept(sender, data)
}

// sendSyncStep2WithState 发送包含服务端累积状态的 SyncStep2
func (r *YjsRoom) sendSyncStep2WithState(c *Client) {
	r.mu.RLock()
	state := r.mergeAllUpdatesLocked()
	r.mu.RUnlock()

	// 构造 SyncStep2 消息：[0(sync)] [1(step2)] [varByteArray(state)]
	msg := make([]byte, 0, len(state)+16)
	msg = writeVarUint(msg, MessageSync)
	msg = writeVarUint(msg, SyncStep2)
	msg = writeVarByteArray(msg, state)
	c.SendBinary(msg)
}

// sendAccumulatedUpdates 发送累积的更新给客户端（逐条发送为 SyncUpdate）
func (r *YjsRoom) sendAccumulatedUpdates(c *Client) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, update := range r.updates {
		msg := make([]byte, 0, len(update)+16)
		msg = writeVarUint(msg, MessageSync)
		msg = writeVarUint(msg, SyncUpdate)
		msg = writeVarByteArray(msg, update)
		c.SendBinary(msg)
	}
}

// storeUpdate 存储更新并标记dirty
func (r *YjsRoom) storeUpdate(update []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	// 存储副本
	stored := make([]byte, len(update))
	copy(stored, update)
	r.updates = append(r.updates, stored)
	r.lastActive = time.Now()
	r.dirty = true

	// 限制累积更新数量，超过 1000 条时合并压缩
	if len(r.updates) > 1000 {
		merged := r.mergeAllUpdatesLocked()
		r.updates = [][]byte{merged}
	}
}

// mergeAllUpdates 合并所有update为一个（需外部不持锁或可重入）
func (r *YjsRoom) mergeAllUpdates() []byte {
	// Yjs update 是可拼接的二进制，简单拼接即可由客户端 applyUpdate 依次应用
	// 这里将所有 updates 按顺序拼接为一个大的 byte slice
	totalLen := 0
	for _, u := range r.updates {
		totalLen += len(u)
	}
	if totalLen == 0 {
		return nil
	}
	// 存储格式：[varUint(count)] + [varByteArray(update)]...
	// 为简单起见，我们存储为: 所有updates依次写入（每个前面加长度前缀）
	buf := make([]byte, 0, totalLen+len(r.updates)*4+8)
	buf = writeVarUint(buf, uint64(len(r.updates)))
	for _, u := range r.updates {
		buf = writeVarByteArray(buf, u)
	}
	return buf
}

// mergeAllUpdatesLocked 合并所有update（调用者已持有读锁）
func (r *YjsRoom) mergeAllUpdatesLocked() []byte {
	totalLen := 0
	for _, u := range r.updates {
		totalLen += len(u)
	}
	if totalLen == 0 {
		return nil
	}
	buf := make([]byte, 0, totalLen+len(r.updates)*4+8)
	buf = writeVarUint(buf, uint64(len(r.updates)))
	for _, u := range r.updates {
		buf = writeVarByteArray(buf, u)
	}
	return buf
}

// startAutoSave 启动定期自动保存（每30秒）
func (r *YjsRoom) startAutoSave() {
	r.saveTicker = time.NewTicker(30 * time.Second)
	go func() {
		for range r.saveTicker.C {
			r.mu.Lock()
			if r.dirty && len(r.updates) > 0 {
				state := r.mergeAllUpdatesLocked()
				r.mu.Unlock()
				if err := SaveYjsState(r.documentID, state); err != nil {
					logger.Errorf("Yjs auto-save failed: docID=%d, err=%v", r.documentID, err)
				} else {
					r.mu.Lock()
					r.dirty = false
					r.mu.Unlock()
					logger.Infof("Yjs auto-saved: docID=%d, size=%d bytes", r.documentID, len(state))
				}
			} else {
				r.mu.Unlock()
			}
		}
	}()
}

// stopAutoSave 停止自动保存
func (r *YjsRoom) stopAutoSave() {
	if r.saveTicker != nil {
		r.saveTicker.Stop()
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

// === 加载持久化状态的辅助函数 ===

// loadPersistedUpdates 从Redis加载的持久化格式中恢复updates列表
func loadPersistedUpdates(data []byte) [][]byte {
	if len(data) == 0 {
		return nil
	}
	count, offset := readVarUint(data, 0)
	updates := make([][]byte, 0, int(count))
	for i := uint64(0); i < count; i++ {
		update, newOffset := readVarByteArray(data, offset)
		if update == nil {
			break
		}
		stored := make([]byte, len(update))
		copy(stored, update)
		updates = append(updates, stored)
		offset = newOffset
	}
	return updates
}
