package detection

import (
	"bytes"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
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

	part, err := writer.CreateFormFile("file", "frame.jpg")
	if err != nil {
		return fmt.Errorf("create form file: %w", err)
	}
	part.Write(imageData)
	writer.Close()

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
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	log.Printf("Detection[%s]: success (%s)", scenarioID, resp.Status)
	return nil
}
