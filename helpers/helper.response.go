package helpers

import (
	"net/http"
)

type APIResponse struct {
	StatCode    int    `json:"stat_code,omitempty"`
	StatMessage string `json:"stat_message,omitempty"`
	ErrMessage  string `json:"err_message,omitempty"`
}

func ApiResponse(rw http.ResponseWriter, payload *APIResponse) {
	rw.Header().Set("Content-Type", "application/json")

	if payload.StatCode >= http.StatusBadRequest {
		rw.WriteHeader(payload.StatCode)
	}

	parser := NewParser()
	payloadByte, err := parser.Marshal(payload)

	if err != nil {
		rw.Write([]byte(err.Error()))
	} else {
		rw.Write(payloadByte)
	}
}
