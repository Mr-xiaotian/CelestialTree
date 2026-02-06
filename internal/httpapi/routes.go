package httpapi

import (
	"net/http"

	"github.com/Mr-xiaotian/CelestialTree/internal/memory"
)

func RegisterRoutes(mux *http.ServeMux, store *memory.Store) {
	mux.HandleFunc("/emit", handleEmit(store))
	mux.HandleFunc("/event/", handleGetEvent(store))
	mux.HandleFunc("/children/", handleChildren(store))
	mux.HandleFunc("/ancestors/", handleAncestors(store))

	mux.HandleFunc("/heads", handleHeads(store))
	mux.HandleFunc("/healthz", handleHealthz())
	mux.HandleFunc("/version", handleVersion())

	// descendants: GET /descendants/{id}  &  POST /descendants {ids:[...]}
	mux.HandleFunc("/descendants/", handleDescendants(store))
	mux.HandleFunc("/descendants", handleDescendantsBatch(store))

	// provenance:  GET /provenance/{id}   &  POST /provenance  {ids:[...]}
	mux.HandleFunc("/provenance/", handleProvenance(store))
	mux.HandleFunc("/provenance", handleProvenanceBatch(store))

	// subscribe: GET /subscribe {id:...}
	mux.HandleFunc("/subscribe", handleSubscribe(store))
}
