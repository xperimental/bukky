package memory

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/xperimental/bukky/internal/digest"
	"github.com/xperimental/bukky/internal/store"
)

type bucket struct {
	objects  map[string]digest.Digest
	contents map[digest.Digest]string
}

type Store struct {
	log         logrus.FieldLogger
	buckets     map[string]*bucket
	bucketMutex *sync.RWMutex
	digester    digest.Digester
}

func NewStore(log logrus.FieldLogger) *Store {
	return &Store{
		log:         log,
		buckets:     make(map[string]*bucket),
		bucketMutex: &sync.RWMutex{},
		digester:    digest.SHA256,
	}
}

func (s *Store) Stats() store.StoreStats {
	s.bucketMutex.RLock()
	defer s.bucketMutex.RUnlock()

	buckets := map[string]store.BucketStats{}
	for k, b := range s.buckets {
		buckets[k] = store.BucketStats{
			NumObjects:  uint(len(b.objects)),
			NumContents: uint(len(b.contents)),
		}
	}

	return store.StoreStats{
		Buckets: buckets,
	}
}

func (s *Store) Get(bucketName, objectID string) (string, error) {
	s.bucketMutex.RLock()
	defer s.bucketMutex.RUnlock()

	b, ok := s.buckets[bucketName]
	if !ok {
		return "", store.ErrNotFound
	}

	obj, ok := b.objects[objectID]
	if !ok {
		return "", store.ErrNotFound
	}

	content, ok := b.contents[obj]
	if !ok {
		return "", fmt.Errorf("can not find content with digest %q", obj)
	}

	return content, nil
}

func (s *Store) Put(bucketName string, objectID string, content string) (string, error) {
	s.bucketMutex.Lock()
	defer s.bucketMutex.Unlock()

	contentDigest, err := s.digester(content)
	if err != nil {
		return "", fmt.Errorf("can not create digest: %w", err)
	}

	b, ok := s.buckets[bucketName]
	if !ok {
		b = &bucket{
			objects:  make(map[string]digest.Digest),
			contents: make(map[digest.Digest]string),
		}
		s.buckets[bucketName] = b
	}

	if _, ok := b.contents[contentDigest]; !ok {
		b.contents[contentDigest] = content
	}
	b.objects[objectID] = contentDigest

	return objectID, nil
}

func (s *Store) Delete(bucketName, objectID string) error {
	s.bucketMutex.Lock()
	defer s.bucketMutex.Unlock()

	b, ok := s.buckets[bucketName]
	if !ok {
		return store.ErrNotFound
	}

	contentDigest, ok := b.objects[objectID]
	if !ok {
		return store.ErrNotFound
	}

	delete(b.objects, objectID)

	found := false
loop:
	for _, d := range b.objects {
		if d == contentDigest {
			found = true
			break loop
		}
	}

	if !found {
		delete(b.contents, contentDigest)
	}

	return nil
}
