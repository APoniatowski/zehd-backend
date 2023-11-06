package helper

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse Boilerplate error response
func ErrorResponse(w http.ResponseWriter, message string, httpStatusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode)
	response := make(map[string]string)
	response["message"] = message
	jsonResp, _ := json.Marshal(response)
	_, err := w.Write(jsonResp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
