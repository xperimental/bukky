package memory

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	"github.com/xperimental/bukky/internal/digest"
	"github.com/xperimental/bukky/internal/store"
	"github.com/xperimental/bukky/internal/testutil"
)

var (
	log = logrus.New()
)

func TestGet(t *testing.T) {
	tt := []struct {
		desc        string
		bucket      string
		objectID    string
		buckets     map[string]*bucket
		wantContent string
		wantErr     error
	}{
		{
			desc:        "empty",
			bucket:      "test-bucket",
			objectID:    "test-object",
			buckets:     map[string]*bucket{},
			wantContent: "",
			wantErr:     store.ErrNotFound,
		},
		{
			desc:     "empty bucket",
			bucket:   "test-bucket",
			objectID: "test-object",
			buckets: map[string]*bucket{
				"test-bucket": {},
			},
			wantContent: "",
			wantErr:     store.ErrNotFound,
		},
		{
			desc:     "content not found",
			bucket:   "test-bucket",
			objectID: "test-object",
			buckets: map[string]*bucket{
				"test-bucket": {
					objects: map[string]digest.Digest{
						"test-object": "digest",
					},
					contents: map[digest.Digest]string{},
				},
			},
			wantContent: "",
			wantErr:     errors.New(`can not find content with digest "digest"`),
		},
		{
			desc:     "success",
			bucket:   "test-bucket",
			objectID: "test-object",
			buckets: map[string]*bucket{
				"test-bucket": {
					objects: map[string]digest.Digest{
						"test-object": "digest",
					},
					contents: map[digest.Digest]string{
						"digest": "content",
					},
				},
			},
			wantContent: "content",
			wantErr:     nil,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			s := NewStore(log)
			s.buckets = tc.buckets

			content, err := s.Get(tc.bucket, tc.objectID)
			if !testutil.EqualErrorMessage(err, tc.wantErr) {
				t.Errorf("got error %q, want %q", err, tc.wantErr)
			}

			if err != nil {
				return
			}

			if content != tc.wantContent {
				t.Errorf("got content %q, want %q", content, tc.wantContent)
			}
		})
	}
}

func TestPut(t *testing.T) {
	digestError := errors.New("test-digest-error")
	tt := []struct {
		desc          string
		bucket        string
		objectID      string
		content       string
		digester      digest.Digester
		bucketsBefore map[string]*bucket
		wantBuckets   map[string]*bucket
		wantID        string
		wantErr       error
	}{
		{
			desc:     "success",
			bucket:   "test-bucket",
			objectID: "test-object",
			content:  "test-content",
			digester: func(content string) (digest.Digest, error) {
				if content != "test-content" {
					t.Errorf("got content to digest %q, want %q", content, "test-content")
				}

				return "test-digest", nil
			},
			bucketsBefore: map[string]*bucket{},
			wantBuckets: map[string]*bucket{
				"test-bucket": {
					objects: map[string]digest.Digest{
						"test-object": "test-digest",
					},
					contents: map[digest.Digest]string{
						"test-digest": "test-content",
					},
				},
			},
			wantID:  "test-object",
			wantErr: nil,
		},
		{
			desc:     "duplicate content",
			bucket:   "test-bucket",
			objectID: "test-object-two",
			content:  "test-content",
			digester: func(content string) (digest.Digest, error) {
				return digest.Digest(fmt.Sprintf("%s-digest", content)), nil
			},
			bucketsBefore: map[string]*bucket{
				"test-bucket": {
					objects: map[string]digest.Digest{
						"test-object": "test-content-digest",
					},
					contents: map[digest.Digest]string{
						"test-content-digest": "test-content",
					},
				},
			},
			wantBuckets: map[string]*bucket{
				"test-bucket": {
					objects: map[string]digest.Digest{
						"test-object":     "test-content-digest",
						"test-object-two": "test-content-digest",
					},
					contents: map[digest.Digest]string{
						"test-content-digest": "test-content",
					},
				},
			},
			wantID:  "test-object-two",
			wantErr: nil,
		},
		{
			desc:     "error in digest",
			bucket:   "test-bucket",
			objectID: "test-object",
			content:  "test-content",
			digester: func(content string) (digest.Digest, error) {
				return "", digestError
			},
			bucketsBefore: map[string]*bucket{},
			wantBuckets:   map[string]*bucket{},
			wantID:        "",
			wantErr:       errors.New("can not create digest: test-digest-error"),
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			s := NewStore(log)
			s.buckets = tc.bucketsBefore
			s.digester = tc.digester

			id, err := s.Put(tc.bucket, tc.objectID, tc.content)
			if !testutil.EqualErrorMessage(err, tc.wantErr) {
				t.Errorf("got error %q, want %q", err, tc.wantErr)
			}

			if diff := cmp.Diff(s.buckets, tc.wantBuckets, cmp.AllowUnexported(bucket{})); diff != "" {
				t.Errorf("resulting buckets differ: -got+want\n%s", diff)
			}

			if err != nil {
				return
			}

			if id != tc.wantID {
				t.Errorf("got id %q, want %q", id, tc.wantID)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tt := []struct {
		desc          string
		bucket        string
		objectID      string
		bucketsBefore map[string]*bucket
		wantBuckets   map[string]*bucket
		wantErr       error
	}{
		{
			desc:          "bucket not found",
			bucket:        "test-bucket",
			objectID:      "test-object",
			bucketsBefore: map[string]*bucket{},
			wantBuckets:   map[string]*bucket{},
			wantErr:       store.ErrNotFound,
		},
		{
			desc:     "object not found",
			bucket:   "test-bucket",
			objectID: "test-object",
			bucketsBefore: map[string]*bucket{
				"test-bucket": {},
			},
			wantBuckets: map[string]*bucket{
				"test-bucket": {},
			},
			wantErr: store.ErrNotFound,
		},
		{
			desc:     "delete object",
			bucket:   "test-bucket",
			objectID: "test-object",
			bucketsBefore: map[string]*bucket{
				"test-bucket": {
					objects: map[string]digest.Digest{
						"test-object": "test-digest",
					},
					contents: map[digest.Digest]string{
						"test-digest": "test-content",
					},
				},
			},
			wantBuckets: map[string]*bucket{
				"test-bucket": {
					objects:  map[string]digest.Digest{},
					contents: map[digest.Digest]string{},
				},
			},
			wantErr: nil,
		},
		{
			desc:     "used content remaining",
			bucket:   "test-bucket",
			objectID: "test-object",
			bucketsBefore: map[string]*bucket{
				"test-bucket": {
					objects: map[string]digest.Digest{
						"test-object":  "test-digest",
						"test-object2": "test-digest",
					},
					contents: map[digest.Digest]string{
						"test-digest": "test-content",
					},
				},
			},
			wantBuckets: map[string]*bucket{
				"test-bucket": {
					objects: map[string]digest.Digest{
						"test-object2": "test-digest",
					},
					contents: map[digest.Digest]string{
						"test-digest": "test-content",
					},
				},
			},
			wantErr: nil,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			s := NewStore(log)
			s.buckets = tc.bucketsBefore

			err := s.Delete(tc.bucket, tc.objectID)
			if !testutil.EqualErrorMessage(err, tc.wantErr) {
				t.Errorf("got error %q, want %q", err, tc.wantErr)
			}

			if diff := cmp.Diff(s.buckets, tc.wantBuckets, cmp.AllowUnexported(bucket{})); diff != "" {
				t.Errorf("resulting buckets differ: -got+want\n%s", diff)
			}

			if err != nil {
				return
			}
		})
	}
}
