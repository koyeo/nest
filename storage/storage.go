package storage

import (
	"context"
	"io"
	"time"
)

// ObjectInfo holds metadata about a stored object.
type ObjectInfo struct {
	Key  string
	Size int64
}

// ObjectStorage defines the interface for cloud storage operations.
type ObjectStorage interface {
	// Upload sends data to the given object key.
	Upload(ctx context.Context, key string, reader io.Reader, size int64) error

	// Head returns the size of the object in bytes.
	// If the object does not exist, it returns -1 and nil error.
	Head(ctx context.Context, key string) (size int64, err error)

	// ListObjects returns all objects whose key starts with the given prefix.
	ListObjects(ctx context.Context, prefix string) ([]ObjectInfo, error)

	// DeleteObjects deletes all objects with the given keys.
	DeleteObjects(ctx context.Context, keys []string) error

	// PresignedURL generates a time-limited download URL for the given object key.
	PresignedURL(ctx context.Context, key string, expires time.Duration) (string, error)
}
