package utils

import (
	"encoding/json"
	"net/http"
)

type BaseResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func RespondWithJSON(w http.ResponseWriter, status int, message string, data interface{}) {
	response := BaseResponse{
		Status:  status,
		Message: message,
		Data:    data,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

func RespondWithError(w http.ResponseWriter, status int, message string) {
	RespondWithJSON(w, status, message, nil)
}
