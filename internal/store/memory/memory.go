package memory

import (
	"github.com/sirupsen/logrus"
	"github.com/xperimental/bukky/internal/store"
)

type Bucket map[string]string

type Store struct {
	log     logrus.FieldLogger
	buckets map[string]Bucket
}

func NewStore(log logrus.FieldLogger) store.Store {
	return &Store{
		log:     log,
		buckets: make(map[string]Bucket),
	}
}

func (s *Store) Get(bucket, objectID string) (string, error) {
	b, ok := s.buckets[bucket]
	if !ok {
		return "", store.ErrNotFound
	}

	obj, ok := b[objectID]
	if !ok {
		return "", store.ErrNotFound
	}

	return obj, nil
}

func (s *Store) Put(bucket string, objectID string, content string) (string, error) {
	b, ok := s.buckets[bucket]
	if !ok {
		b = make(Bucket)
		s.buckets[bucket] = b
	}

	b[objectID] = content
	return objectID, nil
}

func (s *Store) Delete(bucket, objectID string) error {
	b, ok := s.buckets[bucket]
	if !ok {
		return store.ErrNotFound
	}

	if _, ok := b[objectID]; !ok {
		return store.ErrNotFound
	}

	delete(b, objectID)
	return nil
}
