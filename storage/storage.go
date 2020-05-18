package storage

import "io"

// Storage defines an interface for storing and retreiving files
// from local or remote providers.
type Storage interface {
	Put(name string, body io.Reader) error
}