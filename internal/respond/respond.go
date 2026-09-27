package respond

import (
	"encoding/json"
	"log"
	"net/http"
)

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{"error": message})
}

func ServerError(w http.ResponseWriter, op string, err error) {
	log.Printf("%s: %v", op, err)
	Error(w, http.StatusInternalServerError, "ошибка сервера")
}
