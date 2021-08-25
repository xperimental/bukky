package store

import "errors"

var (
	// ErrNotFound is returned when an operation is done on a not existing bucket or object.
	ErrNotFound = errors.New("object not found")
)

type StoreStats struct {
	Buckets []BucketStats `json:"buckets"`
}

type BucketStats struct {
	NumObjects  uint `json:"objects"`
	NumContents uint `json:"contents"`
}

// Store provides the interface to the storage backend
type Store interface {
	Get(bucket, objectID string) (content string, err error)
	Put(bucket, objectID, content string) (id string, err error)
	Delete(bucket, objectID string) error
	Stats() StoreStats
}
