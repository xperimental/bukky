package web

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	"github.com/xperimental/bukky/internal/store"
)

var (
	log = logrus.New()
)

type fakeStore struct {
	t            *testing.T
	wantBucket   string
	wantObjectID string
	wantContent  string
	getContent   string
	putID        string
	err          error
}

func (f fakeStore) checkBucketObject(bucket, objectID string) {
	if bucket != f.wantBucket {
		f.t.Errorf("got bucket %q, want %q", bucket, f.wantBucket)
	}

	if objectID != f.wantObjectID {
		f.t.Errorf("got object ID %q, want %q", objectID, f.wantObjectID)
	}
}

func (f fakeStore) Get(bucket, objectID string) (content string, err error) {
	f.checkBucketObject(bucket, objectID)
	return f.getContent, f.err
}

func (f fakeStore) Put(bucket, objectID, content string) (id string, err error) {
	f.checkBucketObject(bucket, objectID)
	if content != f.wantContent {
		f.t.Errorf("got content %q, want %q", content, f.wantContent)
	}
	return f.putID, f.err
}

func (f fakeStore) Delete(bucket, objectID string) error {
	f.checkBucketObject(bucket, objectID)
	return f.err
}

func (f fakeStore) Stats() store.StoreStats {
	return store.StoreStats{
		Buckets: map[string]store.BucketStats{},
	}
}

type errorReader struct{}

func (e errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestGet(t *testing.T) {

	tt := []struct {
		desc       string
		store      store.Store
		wantStatus int
		wantBody   string
	}{
		{
			desc: "success",
			store: &fakeStore{
				t:            t,
				wantBucket:   "test-bucket",
				wantObjectID: "test-object",
				getContent:   "test-content",
				err:          nil,
			},
			wantStatus: http.StatusOK,
			wantBody:   "test-content",
		},
		{
			desc: "not found",
			store: &fakeStore{
				t:            t,
				wantBucket:   "test-bucket",
				wantObjectID: "test-object",
				getContent:   "",
				err:          store.ErrNotFound,
			},
			wantStatus: http.StatusNotFound,
			wantBody:   "object not found: test-bucket/test-object\n",
		},
		{
			desc: "backend error",
			store: &fakeStore{
				t:            t,
				wantBucket:   "test-bucket",
				wantObjectID: "test-object",
				getContent:   "",
				err:          errors.New("test-error"),
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "can not get object: test-error\n",
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			r := NewRouter(log, tc.store)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/objects/test-bucket/test-object", nil)

			r.Handler().ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("got status %v, want %v", rec.Code, tc.wantStatus)
			}

			body := rec.Body.String()
			if diff := cmp.Diff(body, tc.wantBody); diff != "" {
				t.Errorf("body differs: -got+want\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tt := []struct {
		desc       string
		store      store.Store
		wantStatus int
		wantBody   string
	}{
		{
			desc: "success",
			store: &fakeStore{
				t:            t,
				wantBucket:   "test-bucket",
				wantObjectID: "test-object",
				err:          nil,
			},
			wantStatus: http.StatusNoContent,
			wantBody:   "",
		},
		{
			desc: "not found",
			store: &fakeStore{
				t:            t,
				wantBucket:   "test-bucket",
				wantObjectID: "test-object",
				err:          store.ErrNotFound,
			},
			wantStatus: http.StatusNotFound,
			wantBody:   "object not found: test-bucket/test-object\n",
		},
		{
			desc: "backend error",
			store: &fakeStore{
				t:            t,
				wantBucket:   "test-bucket",
				wantObjectID: "test-object",
				err:          errors.New("test-error"),
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "can not get object: test-error\n",
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			r := NewRouter(log, tc.store)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/objects/test-bucket/test-object", nil)

			r.Handler().ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("got status %v, want %v", rec.Code, tc.wantStatus)
			}

			body := rec.Body.String()
			if diff := cmp.Diff(body, tc.wantBody); diff != "" {
				t.Errorf("body differs: -got+want\n%s", diff)
			}
		})
	}
}

func TestPut(t *testing.T) {
	tt := []struct {
		desc       string
		store      store.Store
		body       io.Reader
		wantStatus int
		wantBody   string
	}{
		{
			desc: "success",
			store: &fakeStore{
				t:            t,
				wantBucket:   "test-bucket",
				wantObjectID: "test-object",
				wantContent:  "test-content",
				putID:        "test-object-id",
				err:          nil,
			},
			body:       strings.NewReader("test-content"),
			wantStatus: http.StatusCreated,
			wantBody: `{"id":"test-object-id"}
`,
		},
		{
			desc: "read error",
			store: &fakeStore{
				t: t,
			},
			body:       &errorReader{},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "can not read body: read error\n",
		},
		{
			desc: "backend error",
			store: &fakeStore{
				t:            t,
				wantBucket:   "test-bucket",
				wantObjectID: "test-object",
				wantContent:  "test-content",
				err:          errors.New("backend error"),
			},
			body:       strings.NewReader("test-content"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "can not save object: backend error\n",
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			r := NewRouter(log, tc.store)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, "/objects/test-bucket/test-object", tc.body)

			r.Handler().ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("got status %v, want %v", rec.Code, tc.wantStatus)
			}

			body := rec.Body.String()
			if diff := cmp.Diff(body, tc.wantBody); diff != "" {
				t.Errorf("body differs: -got+want\n%s", diff)
			}
		})
	}
}

func TestSimpleHandlers(t *testing.T) {
	tt := []struct {
		desc     string
		path     string
		wantBody string
	}{
		{
			desc:     "healthcheck",
			path:     "/health",
			wantBody: "Running.\n",
		},
		{
			desc:     "empty stats",
			path:     "/stats",
			wantBody: `{"buckets":{}}
`,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			r := NewRouter(log, fakeStore{})
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)

			r.Handler().ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("got non-ok status: %d", rec.Code)
			}

			body := rec.Body.String()
			if diff := cmp.Diff(body, tc.wantBody); diff != "" {
				t.Errorf("body differs: -got+want\n%s", diff)
			}
		})
	}
}
