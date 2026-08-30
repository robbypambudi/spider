package handler

import (
	"encoding/json"
	"net/http"

	"github.com/spider/spider/internal/spidererrors"
)

func WriteJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteError(w http.ResponseWriter, err error) {
	if se, ok := err.(*spidererrors.SpiderError); ok {
		WriteJSON(w, se.StatusCode, map[string]string{"error": se.Code, "message": se.Message})
		return
	}
	WriteJSON(w, http.StatusInternalServerError, map[string]string{
		"error": "spider_error", "message": err.Error(),
	})
}
