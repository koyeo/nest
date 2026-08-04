package storage

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

const (
	// multipartThreshold is the size above which an upload is split into parts.
	// Below it, the extra Initiate/Complete round trips cost more than a single
	// stream saves.
	multipartThreshold = 32 << 20

	// basePartSize is the part size used unless the file is big enough to exceed
	// maxParts, in which case it is doubled until the part count fits.
	basePartSize = 4 << 20

	// maxParts is the number of parts a single OSS multipart upload may have.
	maxParts = 10000

	// uploadRoutines is how many parts upload concurrently. Each routine gets its
	// own connection, which is where the speedup on high-latency links comes from:
	// a single TCP stream is capped by RTT and packet loss, N streams are not.
	uploadRoutines = 5
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

func (s *OSSStorage) Upload(ctx context.Context, key string, filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	if info.Size() < multipartThreshold {
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer func() { _ = file.Close() }()
		return s.bucket.PutObject(key, file, oss.WithContext(ctx))
	}

	// UploadFile aborts the multipart upload itself if any part fails, so a failed
	// run leaves no orphan parts behind.
	return s.bucket.UploadFile(key, filePath, ossPartSize(info.Size()),
		oss.Routines(uploadRoutines),
		oss.WithContext(ctx),
	)
}

// ossPartSize picks a part size that keeps the part count within maxParts.
func ossPartSize(size int64) int64 {
	part := int64(basePartSize)
	for size/part > maxParts {
		part *= 2
	}
	return part
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
