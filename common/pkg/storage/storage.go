package storage

import (
	"context"
	"io"
	"log"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type StorageConfig struct {
	AccessKey       string
	Endpoint        string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
}

type Storage struct {
	client     *minio.Client
	bucketName string
	endpoint   string
	useSSL     bool
}

func NewStorage(cfg StorageConfig) (*Storage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})

	if err != nil {
		return nil, err
	}

	s := &Storage{
		client:     client,
		bucketName: cfg.BucketName,
		endpoint:   cfg.Endpoint,
		useSSL:     cfg.UseSSL,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = s.ensureBucket(ctx); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Storage) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil {
		log.Println("Failed to check if bucket exists: ", err)
		return err
	}

	if exists {
		log.Printf("Bucket %s already exists\n", s.bucketName)
		return nil
	}

	if nweBucketErr := s.client.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{}); nweBucketErr != nil {
		log.Println("Failed to create bucket: ", nweBucketErr)
		return nweBucketErr
	}

	return nil
}

func (s *Storage) UploadImage(ctx context.Context, file io.Reader, key string, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucketName, key, file, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Printf("Failed to upload image to bucket %s with key %s: %v\n", s.bucketName, key, err)
		return err
	}
	return nil
}

func (s *Storage) GetImageURL(key string) string {
	scheme := "http"
	if s.useSSL {
		scheme = "https"
	}
	return scheme + "://" + s.endpoint + "/" + s.bucketName + "/" + key
}

func (s *Storage) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	reqParams := make(url.Values)
	presignedURL, err := s.client.PresignedGetObject(ctx, s.bucketName, key, expiry, reqParams)
	if err != nil {
		log.Printf("Failed to generate presigned URL for bucket %s with key %s: %v\n", s.bucketName, key, err)
		return "", err
	}

	return presignedURL.String(), nil
}

func (s *Storage) Delete(ctx context.Context, key string) error {
	err := s.client.RemoveObject(ctx, s.bucketName, key, minio.RemoveObjectOptions{})
	if err != nil {
		log.Printf("Failed to delete object from bucket %s with key %s: %v\n", s.bucketName, key, err)
		return err
	}
	return nil
}
