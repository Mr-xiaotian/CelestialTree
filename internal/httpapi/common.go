package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Mr-xiaotian/CelestialTree/internal/tree"
)

func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		writeJSON(w, 405, tree.ResponseError{Error: "method not allowed"})
		return false
	}
	return true
}

func parsePathUint64(w http.ResponseWriter, path, prefix string) (uint64, bool) {
	idStr := strings.TrimPrefix(path, prefix)
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		writeJSON(w, 400, tree.ResponseError{Error: "bad id"})
		return 0, false
	}
	return id, true
}

func normalizeView(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func readJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
