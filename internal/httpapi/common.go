package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Mr-xiaotian/CelestialTree/internal/tree"
)

// requireMethod 校验 HTTP 方法，不匹配则返回 405 并返回 false。
func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		writeJSON(w, 405, tree.ResponseError{Error: "method not allowed"})
		return false
	}
	return true
}

// parsePathUint64 从 URL 路径中解析 uint64 类型的 ID，失败则返回 400。
func parsePathUint64(w http.ResponseWriter, path, prefix string) (uint64, bool) {
	idStr := strings.TrimPrefix(path, prefix)
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		writeJSON(w, 400, tree.ResponseError{Error: "bad id"})
		return 0, false
	}
	return id, true
}

// normalizeView 将 view 参数统一为小写并去除首尾空白。
func normalizeView(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// writeJSON 将 v 序列化为 JSON 写入响应，设置 Content-Type 和状态码。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// readJSON 从请求体解码 JSON，拒绝未知字段。
func readJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
