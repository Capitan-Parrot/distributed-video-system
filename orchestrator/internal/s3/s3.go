package s3

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	client *minio.Client
}

func NewMinioClient(endpoint, accessKey, secretKey string) (*Client, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	return &Client{client: client}, nil
}

func (c *Client) EnsureBucketExists(ctx context.Context, bucketName string) error {
	exists, err := c.client.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}
	if !exists {
		return c.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	}
	return nil
}

func (c *Client) UploadFileStream(ctx context.Context, bucketName, objectName string, reader io.Reader, size int64) (string, error) {
	if err := c.EnsureBucketExists(ctx, bucketName); err != nil {
		return "", fmt.Errorf("bucket error: %w", err)
	}

	_, err := c.client.PutObject(
		ctx,
		bucketName,
		objectName,
		reader,
		size,
		minio.PutObjectOptions{
			ContentType: "video/mp4",
		},
	)
	if err != nil {
		return "", fmt.Errorf("upload error: %w", err)
	}

	url := fmt.Sprintf("http://%s/%s/%s", c.client.EndpointURL().Host, bucketName, objectName)
	return url, nil
}
