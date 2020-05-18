package storage

import (
	"io"
	"io/ioutil"
	"sync"
)

// MemoryStorage holds settings for the memory storage provider.
type MemoryStorage struct {
	values map[string]string
	mutex *sync.Mutex
}

// NewS3Storage creates a new config for uploading and retrieving files.
func NewMemoryStorage() *MemoryStorage {
	values := make(map[string]string)
	mutex := &sync.Mutex{}
	return &MemoryStorage{values: values, mutex: mutex}
}

// Upload stores the 
func (c *MemoryStorage) Put(name string, body io.Reader) error {
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}

	c.mutex.Lock()
	c.values[name] = string(b)
	c.mutex.Unlock()
	return nil
}