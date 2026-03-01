package storage

import (
	"context"
	"io"
	"time"
)

// ObjectStorage defines the interface for cloud storage operations.
type ObjectStorage interface {
	// Upload sends data to the given object key.
	Upload(ctx context.Context, key string, reader io.Reader, size int64) error

	// PresignedURL generates a time-limited download URL for the given object key.
	PresignedURL(ctx context.Context, key string, expires time.Duration) (string, error)
}
