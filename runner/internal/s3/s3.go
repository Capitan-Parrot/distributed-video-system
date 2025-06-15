package s3

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/Capitan-Parrot/distributed-video-system/runner/internal/models"
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

func (c *Client) DownloadFilesFromURL(ctx context.Context, fileURL string) ([][]byte, error) {
	u, err := url.Parse(fileURL)
	if err != nil {
		return nil, err
	}

	parts := strings.SplitN(strings.TrimPrefix(u.Path, "/"), "/", 2)
	if len(parts) != 2 {
		return nil, err
	}
	bucket, folder := parts[0], parts[1]

	objectCh := c.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Prefix:    folder,
		Recursive: true,
	})

	var files [][]byte

	// Обрабатываем каждый объект
	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}

		// Пропускаем саму папку (если она есть в списке)
		if strings.HasSuffix(object.Key, "/") {
			continue
		}

		// Получаем объект
		obj, err := c.client.GetObject(ctx, bucket, object.Key, minio.GetObjectOptions{})
		if err != nil {
			return nil, err
		}

		// Читаем содержимое файла
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, obj)
		obj.Close()
		if err != nil {
			return nil, err
		}

		files = append(files, buf.Bytes())
	}

	return files, nil
}

// SaveDetectionResults сохраняет результаты детекции в бакет predictions
// в папку с именем сценария под именем файла - индексом файла
func (c *Client) SaveDetectionResults(ctx context.Context, scenarioID string, fileIndex int, detections []models.Detection) error {
	// Конвертируем детекции в JSON
	jsonData, err := json.Marshal(detections)
	if err != nil {
		return fmt.Errorf("failed to marshal detections: %w", err)
	}

	// Формируем путь для сохранения
	objectPath := fmt.Sprintf("%s/%d.json", scenarioID, fileIndex)

	// Загружаем данные в MinIO
	_, err = c.client.PutObject(
		ctx,
		"predictions",             // бакет
		objectPath,                // путь к файлу
		bytes.NewReader(jsonData), // данные
		int64(len(jsonData)),      // размер данных
		minio.PutObjectOptions{
			ContentType: "application/json",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to save detections to S3: %w", err)
	}

	return nil
}

// CountFilesInFolder возвращает количество файлов в указанной папке бакета predictions
func (c *Client) CountFilesInFolder(ctx context.Context, folderPath string) (int, error) {
	count := 0
	objectCh := c.client.ListObjects(ctx, "predictions", minio.ListObjectsOptions{
		Prefix:    folderPath,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return 0, fmt.Errorf("error listing objects: %w", object.Err)
		}

		if strings.HasSuffix(object.Key, "/") {
			continue
		}

		count++
	}

	return count, nil
}
