// internal/transport/http/handlers.go
package http

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Buricz9/notification-service/internal/model"
	"github.com/Buricz9/notification-service/internal/repo"
	"github.com/google/uuid"
)

type Handlers struct {
	Repo repo.NotificationRepository
}

func NewHandlers(r repo.NotificationRepository) *Handlers {
	return &Handlers{Repo: r}
}

// POST /notifications
func (h *Handlers) CreateNotification(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID   string          `json:"user_id"`
		Channel  string          `json:"channel"`
		Payload  json.RawMessage `json:"payload"`
		SendAt   time.Time       `json:"send_at"`  // musi przyjść w RFC3339
		Timezone string          `json:"timezone"` // np. "Europe/Warsaw"
		Priority int             `json:"priority"` // np. 0–10
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	uid, err := uuid.Parse(req.UserID)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	n := &model.Notification{
		UserID:   uid,
		Channel:  req.Channel,
		Payload:  req.Payload,
		SendAt:   req.SendAt,
		Timezone: req.Timezone,
		Priority: req.Priority,
		Status:   "pending",
	}
	id, err := h.Repo.Create(r.Context(), n)
	if err != nil {
		http.Error(w, "could not create notification", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id.String()})
}

// GET /notifications/{id}
func (h *Handlers) GetNotification(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/notifications/")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	n, err := h.Repo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(n)
}

// Health check
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// GET /notifications?status={pending|sent|failed}
func (h *Handlers) ListNotifications(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	list, err := h.Repo.ListByStatus(r.Context(), status)
	if err != nil {
		http.Error(w, "could not list notifications", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// GET /metrics?from=2025-05-01T00:00:00Z&to=2025-05-04T23:59:59Z
func (h *Handlers) Metrics(w http.ResponseWriter, r *http.Request) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		http.Error(w, "invalid from", http.StatusBadRequest)
		return
	}
	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		http.Error(w, "invalid to", http.StatusBadRequest)
		return
	}

	stats, err := h.Repo.Stats(r.Context(), from, to)
	if err != nil {
		http.Error(w, "could not fetch metrics", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// POST /notifications/{id}/send-now
func (h *Handlers) ForceSend(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/notifications/")
	idStr = strings.TrimSuffix(idStr, "/send-now")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	ok, err := h.Repo.ForceSend(r.Context(), id)
	if err != nil {
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "not pending or not found", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /notifications/{id}/cancel
func (h *Handlers) Cancel(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/notifications/")
	idStr = strings.TrimSuffix(idStr, "/cancel")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	ok, err := h.Repo.Cancel(r.Context(), id)
	if err != nil {
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "not pending or not found", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
