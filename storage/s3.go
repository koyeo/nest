package storage

import (
	"context"
	"io"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Storage implements ObjectStorage for AWS S3 (and S3-compatible services).
type S3Storage struct {
	client     *s3.Client
	presigner  *s3.PresignClient
	bucketName string
}

// NewS3Storage creates a new S3Storage client.
func NewS3Storage(region, accessKeyID, accessKeySecret, bucketName, endpoint string) (*S3Storage, error) {
	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKeyID, accessKeySecret, ""),
		),
	)
	if err != nil {
		return nil, err
	}

	var client *s3.Client
	if endpoint != "" {
		client = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = &endpoint
		})
	} else {
		client = s3.NewFromConfig(cfg)
	}

	return &S3Storage{
		client:     client,
		presigner:  s3.NewPresignClient(client),
		bucketName: bucketName,
	}, nil
}

func (s *S3Storage) Upload(ctx context.Context, key string, reader io.Reader, size int64) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        &s.bucketName,
		Key:           &key,
		Body:          reader,
		ContentLength: &size,
	})
	return err
}

func (s *S3Storage) PresignedURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	req, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.bucketName,
		Key:    &key,
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}
