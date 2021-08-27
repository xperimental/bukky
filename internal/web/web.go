package web

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/xperimental/bukky/internal/store"
)

type Router struct {
	log     logrus.FieldLogger
	backend store.Store
	router  *mux.Router
}

func NewRouter(log logrus.FieldLogger, backend store.Store) *Router {
	r := &Router{
		log:     log,
		backend: backend,
		router:  mux.NewRouter(),
	}

	objects := r.router.Path("/objects/{bucket}/{objectID}").Subrouter()
	objects.Methods(http.MethodGet).HandlerFunc(r.getHandler)
	objects.Methods(http.MethodPut).HandlerFunc(r.putHandler)
	objects.Methods(http.MethodDelete).HandlerFunc(r.deleteHandler)

	r.router.Path("/stats").Methods(http.MethodGet).HandlerFunc(r.statsHandler)
	r.router.Path("/health").HandlerFunc(r.healthHandler)

	return r
}

func (r *Router) Handler() http.Handler {
	return r.router
}

func (r *Router) healthHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, "Running.")
}

func (r *Router) statsHandler(w http.ResponseWriter, req *http.Request) {
	stats := r.backend.Stats()

	sendJSON(r.log, w, http.StatusOK, stats)
}

func (r *Router) getHandler(w http.ResponseWriter, req *http.Request) {
	bucket, objectID := reqVars(req)
	content, err := r.backend.Get(bucket, objectID)
	switch {
	case err == store.ErrNotFound:
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

	id, err := r.backend.Put(bucket, objectID, string(content))
	if err != nil {
		http.Error(w, fmt.Sprintf("can not save object: %s", err), http.StatusInternalServerError)
		return
	}

	response := struct {
		ID string `json:"id"`
	}{
		ID: id,
	}
	sendJSON(r.log, w, http.StatusCreated, response)
}

func (r *Router) deleteHandler(w http.ResponseWriter, req *http.Request) {
	bucket, objectID := reqVars(req)
	err := r.backend.Delete(bucket, objectID)
	switch {
	case err == store.ErrNotFound:
		http.Error(w, fmt.Sprintf("object not found: %s/%s", bucket, objectID), http.StatusNotFound)
		return
	case err != nil:
		http.Error(w, fmt.Sprintf("can not get object: %s", err), http.StatusInternalServerError)
		return
	default:
	}

	w.WriteHeader(http.StatusNoContent)
}
