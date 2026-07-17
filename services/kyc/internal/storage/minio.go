package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

type MinIOStorage struct {
	client *minio.Client
	bucket string
	expiry time.Duration
	logger *zap.Logger
}

func NewMinIOStorage(endpoint, accessKey, secretKey, bucket string, useSSL bool, logger *zap.Logger) (*MinIOStorage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio init: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("minio bucket check: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("minio bucket create: %w", err)
		}
	}

	return &MinIOStorage{
		client: client,
		bucket: bucket,
		expiry: 15 * time.Minute,
		logger: logger,
	}, nil
}

func (s *MinIOStorage) GenerateUploadURL(ctx context.Context, verificationID, docType string) (string, error) {
	objectName := buildObjectName(verificationID, docType)

	policy := minio.NewPostPolicy()
	_ = policy.SetBucket(s.bucket)
	_ = policy.SetKey(objectName)
	_ = policy.SetExpires(time.Now().Add(s.expiry))
	_ = policy.SetContentLengthRange(1024, 10*1024*1024)
	_ = policy.SetContentType("image/jpeg")
	_ = policy.SetContentTypeStartsWith("image/")

	url, formData, err := s.client.PresignedPostPolicy(ctx, policy)
	if err != nil {
		return "", fmt.Errorf("presigned URL: %w", err)
	}

	result := url.String()
	for k, v := range formData {
		result += "&" + k + "=" + v
	}
	return result, nil
}

func (s *MinIOStorage) GetDocument(ctx context.Context, verificationID string) ([]byte, error) {
	objectCh := s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix:    verificationID + "/",
		Recursive: false,
	})

	var objectName string
	for obj := range objectCh {
		if obj.Err != nil {
			return nil, obj.Err
		}
		objectName = obj.Key
		break
	}
	if objectName == "" {
		return nil, fmt.Errorf("document non trouvé: %s", verificationID)
	}

	obj, err := s.client.GetObject(ctx, s.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("read object: %w", err)
	}
	return data, nil
}

func (s *MinIOStorage) DeleteDocument(ctx context.Context, verificationID string) error {
	objectCh := s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix:    verificationID + "/",
		Recursive: true,
	})

	var deleteErrors []error
	for obj := range objectCh {
		if obj.Err != nil {
			deleteErrors = append(deleteErrors, obj.Err)
			continue
		}
		err := s.client.RemoveObject(ctx, s.bucket, obj.Key, minio.RemoveObjectOptions{})
		if err != nil {
			deleteErrors = append(deleteErrors, err)
		} else {
			s.logger.Info("document deleted (BCEAO compliance)",
				zap.String("verification_id", verificationID),
				zap.String("object", obj.Key),
			)
		}
	}

	if len(deleteErrors) > 0 {
		return fmt.Errorf("delete errors: %v", deleteErrors)
	}
	return nil
}

func buildObjectName(verificationID, docType string) string {
	return fmt.Sprintf("%s/%s_%d", verificationID, docType, time.Now().UnixMilli())
}
