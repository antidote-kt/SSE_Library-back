package utils

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/go-redis/redis/v8"
)

const aiSummaryCacheTTL = 7 * 24 * time.Hour

// AISummaryCacheValue Redis 中缓存的摘要（不写业务表，仅 KV）。
type AISummaryCacheValue struct {
	Summary   string `json:"summary"`
	SummaryID uint64 `json:"summaryId"`
	SrcHash   string `json:"srcHash"`
}

// HashSourceText 用于判断正文是否变化，避免旧缓存命中。
func HashSourceText(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func aiSummaryRedisKey(contentType string, contentID uint64) string {
	return fmt.Sprintf("ai_summary:%s:%d", contentType, contentID)
}

// GetAISummaryFromCache 若存在且 srcHash 一致则返回缓存；否则 nil。
func GetAISummaryFromCache(ctx context.Context, contentType string, contentID uint64, currentSrcHash string) (*AISummaryCacheValue, error) {
	rdb := config.GetRedisClient()
	key := aiSummaryRedisKey(contentType, contentID)
	raw, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var v AISummaryCacheValue
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return nil, err
	}
	if v.SrcHash != currentSrcHash {
		return nil, nil
	}
	return &v, nil
}

// SetAISummaryCache 写入摘要缓存。
func SetAISummaryCache(ctx context.Context, contentType string, contentID uint64, v *AISummaryCacheValue) error {
	rdb := config.GetRedisClient()
	key := aiSummaryRedisKey(contentType, contentID)
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return rdb.Set(ctx, key, string(b), aiSummaryCacheTTL).Err()
}
