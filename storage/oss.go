package storage

import (
	"context"
	"io"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// OSSStorage implements ObjectStorage for Alibaba Cloud OSS.
type OSSStorage struct {
	bucket *oss.Bucket
}

// NewOSSStorage creates a new OSSStorage client.
func NewOSSStorage(endpoint, accessKeyID, accessKeySecret, bucketName string) (*OSSStorage, error) {
	client, err := oss.New(endpoint, accessKeyID, accessKeySecret)
	if err != nil {
		return nil, err
	}
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return nil, err
	}
	return &OSSStorage{bucket: bucket}, nil
}

func (s *OSSStorage) Upload(ctx context.Context, key string, reader io.Reader, size int64) error {
	return s.bucket.PutObject(key, reader)
}

func (s *OSSStorage) PresignedURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	return s.bucket.SignURL(key, oss.HTTPGet, int64(expires.Seconds()))
}
