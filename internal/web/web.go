package web

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/xperimental/bukky/internal/memory"
)

func MainHandler(log logrus.FieldLogger, store *memory.Store) http.Handler {
	r := mux.NewRouter()
	objects := r.Path("/objects/{bucket}/{objectID}").Subrouter()
	objects.Methods(http.MethodGet).Handler(getHandler(store))
	objects.Methods(http.MethodPut).Handler(putHandler(log, store))
	objects.Methods(http.MethodDelete).Handler(deleteHandler(store))
	r.Path("/").Handler(healthHandler())

	return r
}

func healthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Running.")
	})
}

func reqVars(r *http.Request) (bucket, objectID string) {
	vars := mux.Vars(r)
	bucket = vars["bucket"]
	objectID = vars["objectID"]
	return bucket, objectID
}

func getHandler(objectStore *memory.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bucket, objectID := reqVars(r)
		content, err := objectStore.Get(bucket, objectID)
		switch {
		case err == memory.ErrNotExist:
			http.Error(w, fmt.Sprintf("object not found: %s/%s", bucket, objectID), http.StatusNotFound)
			return
		case err != nil:
			http.Error(w, fmt.Sprintf("can not get object: %s", err), http.StatusInternalServerError)
			return
		default:
		}

		w.Write([]byte(content))
	})
}

func putHandler(log logrus.FieldLogger, objectStore *memory.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		bucket, objectID := reqVars(r)
		content, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("can not read body: %s", err), http.StatusInternalServerError)
			return
		}

		id, err := objectStore.Put(bucket, objectID, string(content))
		if err != nil {
			http.Error(w, fmt.Sprintf("can not save object: %s", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		response := struct {
			ID string `json:"id"`
		}{
			ID: id,
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Errorf("Error writing response: %s", err)
		}
	})
}

func deleteHandler(objectStore *memory.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bucket, objectID := reqVars(r)
		err := objectStore.Delete(bucket, objectID)
		switch {
		case err == memory.ErrNotExist:
			http.Error(w, fmt.Sprintf("object not found: %s/%s", bucket, objectID), http.StatusNotFound)
			return
		case err != nil:
			http.Error(w, fmt.Sprintf("can not get object: %s", err), http.StatusInternalServerError)
			return
		default:
		}

		w.WriteHeader(http.StatusNoContent)
	})
}
