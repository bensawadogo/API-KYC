package handler

import "time"

type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *string     `json:"error,omitempty"`
	Timestamp string      `json:"timestamp"`
}

func SuccessResponse(data interface{}) APIResponse {
	return APIResponse{
		Success:   true,
		Data:      data,
		Error:     nil,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

func ErrorResponse(errMsg string) APIResponse {
	err := errMsg
	return APIResponse{
		Success:   false,
		Data:      nil,
		Error:     &err,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}
