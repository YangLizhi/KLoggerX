package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

var ctx = context.Background()

func SetCache(key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return RDB.Set(ctx, key, data, ttl).Err()
}

func GetCache(key string, dest interface{}) error {
	data, err := RDB.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func DeleteCache(key string) error {
	return RDB.Del(ctx, key).Err()
}

func DeleteCacheByPattern(pattern string) error {
	keys, err := RDB.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return RDB.Del(ctx, keys...).Err()
	}
	return nil
}

func DocTreeCacheKey(userID uint) string {
	return fmt.Sprintf("cache:doc_tree:%d", userID)
}

func UserCacheKey(userID uint) string {
	return fmt.Sprintf("cache:user:%d", userID)
}

func InvalidateDocTreeCache(userID uint) {
	DeleteCache(DocTreeCacheKey(userID))
}

func InvalidateAllDocTreeCache() {
	DeleteCacheByPattern("cache:doc_tree:*")
}
