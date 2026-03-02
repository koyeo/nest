package storage

import (
	"context"
	"io"
	"net/http"
	"strconv"
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

func (s *OSSStorage) Head(ctx context.Context, key string) (int64, error) {
	resp, err := s.bucket.GetObjectMeta(key)
	if err != nil {
		if serviceErr, ok := err.(oss.ServiceError); ok && serviceErr.StatusCode == http.StatusNotFound {
			return -1, nil
		}
		return -1, err
	}
	sizeStr := resp.Get("Content-Length")
	if sizeStr == "" {
		return -1, nil
	}
	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return -1, nil
	}
	return size, nil
}

func (s *OSSStorage) ListObjects(ctx context.Context, prefix string) ([]ObjectInfo, error) {
	var objects []ObjectInfo
	marker := ""
	for {
		result, err := s.bucket.ListObjects(oss.Prefix(prefix), oss.Marker(marker), oss.MaxKeys(1000))
		if err != nil {
			return nil, err
		}
		for _, obj := range result.Objects {
			objects = append(objects, ObjectInfo{Key: obj.Key, Size: obj.Size})
		}
		if !result.IsTruncated {
			break
		}
		marker = result.NextMarker
	}
	return objects, nil
}

func (s *OSSStorage) DeleteObjects(ctx context.Context, keys []string) error {
	// OSS supports batch delete up to 1000 keys at a time
	for i := 0; i < len(keys); i += 1000 {
		end := i + 1000
		if end > len(keys) {
			end = len(keys)
		}
		_, err := s.bucket.DeleteObjects(keys[i:end])
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *OSSStorage) PresignedURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	return s.bucket.SignURL(key, oss.HTTPGet, int64(expires.Seconds()))
}
