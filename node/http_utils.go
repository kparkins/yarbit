package node

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func writeResponse(w http.ResponseWriter, content interface{}) {
	contentJson, err := json.Marshal(content)
	if err != nil {
		writeErrorResponse(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(contentJson)
}

func writeErrorResponse(w http.ResponseWriter, e error, statusCode int) {
	w.WriteHeader(statusCode)
	if e != nil {
		contentJson, _ := json.Marshal(ErrorResponse{Error: e.Error()})
		w.Header().Set("Content-Type", "application/json")
		w.Write(contentJson)
	}
}
