package web

import (
	"net/http"

	"github.com/gorilla/mux"
)

func reqVars(r *http.Request) (bucket, objectID string) {
	vars := mux.Vars(r)
	bucket = vars["bucket"]
	objectID = vars["objectID"]
	return bucket, objectID
}
