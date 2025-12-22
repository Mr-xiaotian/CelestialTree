package tree

// Event 是 CelestialTree 的“最小历史原子”：不可变、可追加、可分岔、可合并。
type Event struct {
	ID       uint64         `json:"id"`
	Hash     string         `json:"hash"`
	TimeUnix int64          `json:"time_unix"`
	Type     string         `json:"type"`
	Parents  []uint64       `json:"parents"`
	Payload  map[string]any `json:"payload,omitempty"`
	Meta     map[string]any `json:"meta,omitempty"`
}

// EmitRequest 是客户端发来的“写入事件”的请求体。
type EmitRequest struct {
	Type    string         `json:"type"`
	Parents []uint64       `json:"parents"`
	Payload map[string]any `json:"payload"`
	Meta    map[string]any `json:"meta"`
}

// EmitResponse 是 /emit 返回的响应体。
type EmitResponse struct {
	Event Event `json:"event"`
}
