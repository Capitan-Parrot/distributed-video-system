package detection

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"

	"github.com/Capitan-Parrot/distributed-video-system/runner/internal/models"
)

type Client struct {
	URL string
}

func NewClient(baseURL string) *Client {
	return &Client{URL: baseURL}
}

// SendFrame отправляет изображение JPEG байтами на /predict и возвращает детекции
func (c *Client) SendFrame(imageData []byte, scenarioID string) ([]models.Detection, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Создаем form field с правильным Content-Type
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="frame.jpg"`)
	h.Set("Content-Type", "image/jpeg")

	part, err := writer.CreatePart(h)
	if err != nil {
		return nil, fmt.Errorf("create form part: %w", err)
	}

	if _, err := part.Write(imageData); err != nil {
		return nil, fmt.Errorf("write image data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close writer: %w", err)
	}

	req, err := http.NewRequest("POST", c.URL+"/predict", &buf)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bad status: %s, error: %s", resp.Status, bodyBytes)
	}

	// Обрабатываем JSON-ответ
	var response struct {
		Detections []models.Detection `json:"detections"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	//log.Printf("Detection[%s]: success (%s), found %d objects", scenarioID, resp.Status, len(response.Detections))
	return response.Detections, nil
}
