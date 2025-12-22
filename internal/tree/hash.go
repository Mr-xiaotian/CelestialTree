package tree

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// hashEvent 计算事件内容的 hash（不含 Hash 字段本身）
// 用于提供“不可篡改感”和调试一致性。
func hashEvent(ev Event) string {
	type H struct {
		ID       uint64         `json:"id"`
		TimeUnix int64          `json:"time_unix_nano"`
		Type     string         `json:"type"`
		Parents  []uint64       `json:"parents"`
		Payload  map[string]any `json:"payload,omitempty"`
		Meta     map[string]any `json:"meta,omitempty"`
	}

	h := H{
		ID:       ev.ID,
		TimeUnix: ev.TimeUnix,
		Type:     ev.Type,
		Parents:  ev.Parents,
		Payload:  ev.Payload,
		Meta:     ev.Meta,
	}

	b, _ := json.Marshal(h)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
