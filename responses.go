package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

var (
	ErrNotFound   = errors.New("Post not found")
	ErrUserExists = errors.New("Name or email already taken")
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Println(err)
	}
}

func writeError(w http.ResponseWriter, status int, msg, details string) {
	writeJSON(w, status, ErrorResponse{Error: msg, Details: details})
}
