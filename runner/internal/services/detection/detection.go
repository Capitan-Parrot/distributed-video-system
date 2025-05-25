package detection

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
)

type Client struct {
	URL string
}

func NewClient(baseURL string) *Client {
	return &Client{URL: baseURL}
}

// SendFrame отправляет изображение JPEG байтами на /predict
func (c *Client) SendFrame(imageData []byte, scenarioID string) error {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Создаем form field с правильным Content-Type
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="frame.jpg"`)
	h.Set("Content-Type", "image/jpeg")

	part, err := writer.CreatePart(h)
	if err != nil {
		return fmt.Errorf("create form part: %w", err)
	}

	if _, err := part.Write(imageData); err != nil {
		return fmt.Errorf("write image data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("close writer: %w", err)
	}

	req, err := http.NewRequest("POST", c.URL+"/predict", &buf)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bad status: %s, error: %s", resp.Status, bodyBytes)
	}

	log.Printf("Detection[%s]: success (%s)", scenarioID, resp.Status)
	return nil
}
