// Package api exposes the analytics service over HTTP.
package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"loginpulse/internal/analytics"
)

type Handler struct {
	svc *analytics.Service
}

func NewHandler(svc *analytics.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /logins", h.postLogin)
	mux.HandleFunc("GET /analytics/daily", h.getDaily)
	mux.HandleFunc("GET /analytics/monthly", h.getMonthly)
}

type loginRequest struct {
	UserID  string    `json:"user_id"`
	LoginAt time.Time `json:"login_at"`
}

func (h *Handler) postLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.LoginAt.IsZero() {
		req.LoginAt = time.Now()
	}
	if err := h.svc.RecordLogin(r.Context(), userID, req.LoginAt); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) getDaily(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	count, err := h.svc.GetDailyUniqueUsers(r.Context(), date)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, map[string]any{"date": date, "unique_users": count})
}

func (h *Handler) getMonthly(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	count, err := h.svc.GetMonthlyUniqueUsers(r.Context(), month)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, map[string]any{"month": month, "unique_users": count})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}
