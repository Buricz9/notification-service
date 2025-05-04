// internal/transport/http/router.go
package http

import (
	"net/http"
	"strings"
)

func NewRouter(h *Handlers) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/notifications", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.CreateNotification(w, r)
		case http.MethodGet:
			h.ListNotifications(w, r)
		default:
			http.NotFound(w, r)
		}
	})
	mux.HandleFunc("/notifications/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/notifications/")
		switch {
		case r.Method == http.MethodGet && !strings.Contains(path, "/"):
			h.GetNotification(w, r)
		case r.Method == http.MethodPost && strings.HasSuffix(path, "/send-now"):
			h.ForceSend(w, r)
		case r.Method == http.MethodPost && strings.HasSuffix(path, "/cancel"):
			h.Cancel(w, r)
		default:
			http.NotFound(w, r)
		}
	})
	mux.HandleFunc("/metrics", h.Metrics)
	return mux
}
