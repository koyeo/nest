package storage

import (
	"fmt"

	"github.com/koyeo/nest/config"
)

// NewFromBucket creates an ObjectStorage from a decrypted BucketCredential.
func NewFromBucket(cred *config.BucketCredential) (ObjectStorage, error) {
	switch cred.Provider {
	case "oss":
		if cred.Endpoint == "" {
			return nil, fmt.Errorf("OSS requires 'endpoint'")
		}
		return NewOSSStorage(cred.Endpoint, cred.AccessKeyID, cred.AccessKeySecret, cred.BucketName)
	case "s3":
		if cred.Region == "" {
			return nil, fmt.Errorf("S3 requires 'region'")
		}
		return NewS3Storage(cred.Region, cred.AccessKeyID, cred.AccessKeySecret, cred.BucketName, cred.Endpoint)
	default:
		return nil, fmt.Errorf("unsupported provider: '%s' (use 'oss' or 's3')", cred.Provider)
	}
}
