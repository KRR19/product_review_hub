package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"product_review_hub/internal/api"
)

// responseJSON writes a JSON response with the given status code.
func responseJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// responseError writes an error response with the given status code and message.
func responseError(w http.ResponseWriter, statusCode int, message string) {
	errorResponse := api.ErrorResponse{
		Error: message,
	}
	responseJSON(w, statusCode, errorResponse)
}

// errValidation creates a validation error with field and message.
func errValidation(field, message string) error {
	return fmt.Errorf("validation error: %s - %s", field, message)
}

// getStringValue safely returns string value from pointer, or empty string if nil.
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
