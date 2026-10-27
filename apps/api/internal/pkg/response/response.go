package response

import (
	"encoding/json"
	"net/http"
)

// ResponseEnvelope represents the standardized top-level API structure.
type ResponseEnvelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
	Error   *ErrorData  `json:"error,omitempty"`
}

// ErrorData contains granular error information following API conventions.
type ErrorData struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// MetaCursor provides pagination cursor metadata.
type MetaCursor struct {
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
	Total      *int64 `json:"total,omitempty"`
}

// JSON writes any value as an indented JSON HTTP response with proper headers.
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// Success writes a standardized success response envelope.
func Success(w http.ResponseWriter, status int, data interface{}, meta interface{}) {
	JSON(w, status, ResponseEnvelope{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// Error writes a standardized error response envelope.
func Error(w http.ResponseWriter, status int, code string, message string, details interface{}) {
	JSON(w, status, ResponseEnvelope{
		Success: false,
		Error: &ErrorData{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}
