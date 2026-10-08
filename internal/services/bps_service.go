package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"statio/internal/dto"
	"strings"
	"time"
)

type BPSService struct {
	endpoint string
	client   *http.Client
}

func NewBPSService() *BPSService {
	return &BPSService{
		endpoint: "https://gerbang.web.bps.go.id/api/v1/auth/login",
		client:   &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *BPSService) GetUserInfo(code string, realm string, authType string) (*dto.UserInfoResponse, error) {
	payload := map[string]string{
		"code":  code,
		"type":  authType,
		"realm": realm,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", s.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "id-ID,id;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Origin", "https://gerbang.web.bps.go.id")
	req.Header.Set("Referer", "https://gerbang.web.bps.go.id/")

	client := s.client
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request to Gerbang BPS: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read Gerbang BPS response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		logGerbangResponse(resp, responseBody)
		return nil, fmt.Errorf("Gerbang BPS returned HTTP %d: %s", resp.StatusCode, gerbangErrorMessage(responseBody))
	}

	var wrapped dto.BPSAuthLoginResponse
	if err := json.Unmarshal(responseBody, &wrapped); err != nil {
		logGerbangResponse(resp, responseBody)
		return nil, fmt.Errorf("decode Gerbang BPS response: %w; body: %s", err, gerbangErrorMessage(responseBody))
	}
	if strings.TrimSpace(wrapped.Data.Email) == "" {
		logGerbangResponse(resp, responseBody)
		return nil, fmt.Errorf("Gerbang BPS response is missing data.email: %s", gerbangErrorMessage(responseBody))
	}
	return &wrapped.Data, nil
}

func logGerbangResponse(resp *http.Response, responseBody []byte) {
	const maxLogBody = 4096
	bodyPreview := responseBody
	if len(bodyPreview) > maxLogBody {
		bodyPreview = bodyPreview[:maxLogBody]
	}
	log.Printf(
		"Gerbang BPS login response: status=%d content_type=%q server=%q request_id=%q correlation_id=%q body=%q truncated=%t",
		resp.StatusCode,
		resp.Header.Get("Content-Type"),
		resp.Header.Get("Server"),
		resp.Header.Get("X-Request-ID"),
		resp.Header.Get("X-Correlation-ID"),
		string(bodyPreview),
		len(responseBody) > maxLogBody,
	)
}

func gerbangErrorMessage(responseBody []byte) string {
	var response struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal(responseBody, &response); err == nil {
		if message := strings.TrimSpace(response.Message); message != "" {
			return message
		}
		if message := strings.TrimSpace(response.Error); message != "" {
			return message
		}
	}
	if len(strings.TrimSpace(string(responseBody))) == 0 {
		return "empty response body"
	}
	const maxErrorBody = 500
	message := strings.Join(strings.Fields(string(responseBody)), " ")
	if len(message) > maxErrorBody {
		message = message[:maxErrorBody] + "..."
	}
	return message
}
