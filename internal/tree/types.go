package tree

// Event 是 CelestialTree 的“最小历史原子”。
type Event struct {
	ID           uint64   `json:"id"`
	TimeUnixNano int64    `json:"time_unix_nano"`
	Type         string   `json:"type"`
	Parents      []uint64 `json:"parents"`
	Message      string   `json:"message,omitempty"`
	Payload      []byte   `json:"payload,omitempty"`
}

// EventTreeNode 用于表示某个事件及其所有后代（树形结构）
type EventTreeNode struct {
	ID       uint64          `json:"id"`
	Children []EventTreeNode `json:"children"`
	IsRef    bool            `json:"is_ref"`
}

// EmitRequest 是客户端发来的“写入事件”的请求体。
type EmitRequest struct {
	Type    string   `json:"type"`
	Parents []uint64 `json:"parents"`
	Message string   `json:"message,omitempty"`
	Payload []byte   `json:"payload,omitempty"`
}

// EmitResponse 是 /emit 返回的响应体。
type EmitResponse struct {
	ID uint64 `json:"id"`
}
