package web

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func reqVars(r *http.Request) (bucket, objectID string) {
	vars := mux.Vars(r)
	bucket = vars["bucket"]
	objectID = vars["objectID"]
	return bucket, objectID
}

func sendJSON(log logrus.FieldLogger, w http.ResponseWriter, statusCode int, content interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(content); err != nil {
		log.Errorf("Error encoding stats JSON: %s", err)
	}
}
