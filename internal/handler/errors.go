package handler

import (
	"context"
	"delivery-api/internal/apperror"
	"errors"
	"log"
	"net/http"
)

func handleError(w http.ResponseWriter, err error) {
	if errors.Is(err, context.Canceled) {
		return
	}
	if errors.Is(err, context.DeadlineExceeded) {
		writeError(w, http.StatusServiceUnavailable, "service temporarily unavailable, try again later")
		return
	}
	if errors.Is(err, apperror.ErrOrderNotFound) {
		writeError(w, http.StatusNotFound, "order not found")
		return
	}
	log.Printf("Unexpected error: %v", err)
	writeError(w, http.StatusInternalServerError, "internal server error")
}
