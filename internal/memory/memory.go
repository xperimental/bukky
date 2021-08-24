package memory

import (
	"errors"

	"github.com/sirupsen/logrus"
)

var (
	ErrNotExist = errors.New("object does not exist")
)

type Bucket map[string]string

type Store struct {
	log     logrus.FieldLogger
	buckets map[string]Bucket
}

func NewStore(log logrus.FieldLogger) *Store {
	return &Store{
		log:     log,
		buckets: make(map[string]Bucket),
	}
}

func (s *Store) Get(bucket, objectID string) (string, error) {
	b, ok := s.buckets[bucket]
	if !ok {
		return "", ErrNotExist
	}

	obj, ok := b[objectID]
	if !ok {
		return "", ErrNotExist
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
		return ErrNotExist
	}

	if _, ok := b[objectID]; !ok {
		return ErrNotExist
	}

	delete(b, objectID)
	return nil
}
