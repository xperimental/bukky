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

type Router struct {
	log    logrus.FieldLogger
	store  *memory.Store
	router *mux.Router
}

func NewRouter(log logrus.FieldLogger, store *memory.Store) *Router {
	r := &Router{
		log:    log,
		store:  store,
		router: mux.NewRouter(),
	}

	objects := r.router.Path("/objects/{bucket}/{objectID}").Subrouter()
	objects.Methods(http.MethodGet).HandlerFunc(r.getHandler)
	objects.Methods(http.MethodPut).HandlerFunc(r.putHandler)
	objects.Methods(http.MethodDelete).HandlerFunc(r.deleteHandler)
	r.router.Path("/").HandlerFunc(r.healthHandler)

	return r
}

func (r *Router) Handler() http.Handler {
	return r.router
}

func (r *Router) healthHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, "Running.")
}

func (r *Router) getHandler(w http.ResponseWriter, req *http.Request) {
	bucket, objectID := reqVars(req)
	content, err := r.store.Get(bucket, objectID)
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
}

func (r *Router) putHandler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	bucket, objectID := reqVars(req)
	content, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("can not read body: %s", err), http.StatusInternalServerError)
		return
	}

	id, err := r.store.Put(bucket, objectID, string(content))
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
		r.log.Errorf("Error writing response: %s", err)
	}
}

func (r *Router) deleteHandler(w http.ResponseWriter, req *http.Request) {
	bucket, objectID := reqVars(req)
	err := r.store.Delete(bucket, objectID)
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
}
