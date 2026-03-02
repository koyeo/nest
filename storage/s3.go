package storage

import (
	"context"
	"errors"
	"io"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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

func (s *S3Storage) Head(ctx context.Context, key string) (int64, error) {
	out, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &s.bucketName,
		Key:    &key,
	})
	if err != nil {
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			return -1, nil
		}
		return -1, err
	}
	if out.ContentLength != nil {
		return *out.ContentLength, nil
	}
	return -1, nil
}

func (s *S3Storage) ListObjects(ctx context.Context, prefix string) ([]ObjectInfo, error) {
	var objects []ObjectInfo
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: &s.bucketName,
		Prefix: &prefix,
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, obj := range page.Contents {
			size := int64(0)
			if obj.Size != nil {
				size = *obj.Size
			}
			objects = append(objects, ObjectInfo{Key: *obj.Key, Size: size})
		}
	}
	return objects, nil
}

func (s *S3Storage) DeleteObjects(ctx context.Context, keys []string) error {
	for i := 0; i < len(keys); i += 1000 {
		end := i + 1000
		if end > len(keys) {
			end = len(keys)
		}
		batch := make([]types.ObjectIdentifier, 0, end-i)
		for _, k := range keys[i:end] {
			key := k
			batch = append(batch, types.ObjectIdentifier{Key: &key})
		}
		_, err := s.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: &s.bucketName,
			Delete: &types.Delete{Objects: batch},
		})
		if err != nil {
			return err
		}
	}
	return nil
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
