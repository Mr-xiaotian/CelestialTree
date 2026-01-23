package tree

import "encoding/json"

// Event 是 CelestialTree 的“最小历史原子”。
type Event struct {
	ID           uint64          `json:"id"`
	TimeUnixNano int64           `json:"time_unix_nano"`
	Type         string          `json:"type"`
	Message      string          `json:"message,omitempty"`
	Payload      json.RawMessage `json:"payload,omitempty"`
	Parents      []uint64        `json:"parents"`
}

// ===============================
// 			请求体结构
// ===============================

// EmitRequest 是客户端发来的“写入事件”的请求体。
type EmitRequest struct {
	Type    string          `json:"type"`
	Message string          `json:"message,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
	Parents []uint64        `json:"parents"`
}

// TreeBatchRequest 用于批量查询 descendants/provenance。
type TreeBatchRequest struct {
	IDs  []uint64 `json:"ids"`
	View string   `json:"view,omitempty"`
}

// ===============================
// 			响应体结构
// ===============================

// EmitResponse 是 /emit 返回的响应体。
type EmitResponse struct {
	ID uint64 `json:"id"`
}

// DescendantsTree 用于表示某个事件及其所有后代（树形结构）
type DescendantsTree struct {
	ID       uint64            `json:"id"`
	IsRef    bool              `json:"is_ref"`
	Children []DescendantsTree `json:"children"`
}

// DescendantsTreeMeta 用于表示某个事件及其所有后代（树形结构），并且包含时间戳
type DescendantsTreeMeta struct {
	ID           uint64                `json:"id"`
	TimeUnixNano int64                 `json:"time_unix_nano"`
	Type         string                `json:"type"`
	IsRef        bool                  `json:"is_ref"`
	Message      string                `json:"message,omitempty"`
	Payload      json.RawMessage       `json:"payload,omitempty"`
	Children     []DescendantsTreeMeta `json:"children"`
}

// ProvenanceTree 用于表示某个事件及其所有祖先（树形结构，向上追溯）
type ProvenanceTree struct {
	ID      uint64           `json:"id"`
	IsRef   bool             `json:"is_ref"`
	Parents []ProvenanceTree `json:"parents"`
}

// ProvenanceTreeMeta 用于表示某个事件及其所有祖先（树形结构），并且包含时间戳和类型
type ProvenanceTreeMeta struct {
	ID           uint64               `json:"id"`
	TimeUnixNano int64                `json:"time_unix_nano"`
	Type         string               `json:"type"`
	IsRef        bool                 `json:"is_ref"`
	Message      string               `json:"message,omitempty"`
	Payload      json.RawMessage      `json:"payload,omitempty"`
	Parents      []ProvenanceTreeMeta `json:"parents"`
}

// ResponseError 是错误响应的响应体。
type ResponseError struct {
	Error  string `json:"error"`
	Detail string `json:"detail,omitempty"`
}
