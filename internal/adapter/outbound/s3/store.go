package s3

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/alexburley/ask-howard/internal/port/outbound"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Config holds the settings needed to connect to an S3-compatible store.
type Config struct {
	Endpoint        string
	PresignEndpoint string // public-facing endpoint used in presigned URLs (e.g. localhost vs internal hostname)
	Bucket          string
	Region          string
	AccessKey       string
	SecretKey       string
	UsePathStyle    bool
}

// Store is an S3-compatible ObjectStore backed by aws-sdk-go-v2. It works
// against AWS S3, Scaleway Object Storage, and MinIO (path-style).
type Store struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucket        string
}

var _ outbound.ObjectStore = (*Store)(nil)

func NewStore(ctx context.Context, cfg *Config) (*Store, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		o.UsePathStyle = cfg.UsePathStyle
	})

	// Use a separate client for presigning when a public endpoint differs from the internal one.
	presignBase := client
	if cfg.PresignEndpoint != "" && cfg.PresignEndpoint != cfg.Endpoint {
		presignBase = s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.PresignEndpoint)
			o.UsePathStyle = cfg.UsePathStyle
		})
	}

	return &Store{
		client:        client,
		presignClient: s3.NewPresignClient(presignBase),
		bucket:        cfg.Bucket,
	}, nil
}

func (s *Store) PresignPut(ctx context.Context, key, contentType string, expiry time.Duration) (string, error) {
	req, err := s.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("presign put %q: %w", key, err)
	}
	return req.URL, nil
}

func (s *Store) PresignGet(ctx context.Context, key string, expiry time.Duration) (string, error) {
	req, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("presign get %q: %w", key, err)
	}
	return req.URL, nil
}

func (s *Store) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("get object %q: %w", key, err)
	}
	return out.Body, nil
}

func (s *Store) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          r,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("put object %q: %w", key, err)
	}
	return nil
}

func (s *Store) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete object %q: %w", key, err)
	}
	return nil
}
