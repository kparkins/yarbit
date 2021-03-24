package node

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func readRequestJson(request *http.Request, result interface{}) error {
	content, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return errors.Wrap(err, "invalid request body")
	}
	if err := json.Unmarshal(content, &result); err != nil {
		return errors.Wrap(err, "failed to deserialize request body")
	}
	return nil
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
