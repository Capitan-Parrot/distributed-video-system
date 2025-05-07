package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

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

func (c *Client) DownloadFileFromURL(fileURL string) ([]byte, error) {
	u, err := url.Parse(fileURL)
	if err != nil {
		return nil, err
	}

	parts := strings.SplitN(strings.TrimPrefix(u.Path, "/"), "/", 2)
	if len(parts) != 2 {
		return nil, err
	}
	bucket, object := parts[0], parts[1]

	obj, err := c.client.GetObject(context.Background(), bucket, object, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, obj)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
