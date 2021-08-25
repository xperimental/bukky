package digest

import (
	"crypto/sha256"
	"fmt"
)

// A Digest is the hash of an object's contents.
type Digest string

// Digester returns the Digest of an object.
type Digester func(content string) (Digest, error)

// OneToOne is a digest implementation that maps the content onto itself.
// Not for real-world usage, but no conflicts.
func OneToOne(content string) (Digest, error) {
	return Digest(content), nil
}

// SHA256 hashes the content with SHA-256 and returns the hash as a hex-string.
func SHA256(content string) (Digest, error) {
	sum := sha256.Sum256([]byte(content))
	str := fmt.Sprintf("%x", sum)
	return Digest(str), nil
}
