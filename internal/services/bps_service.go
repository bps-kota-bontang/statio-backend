package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"statio/internal/dto"
)

type BPSService struct {
}

func NewBPSService() *BPSService {
	return &BPSService{}
}

func (s *BPSService) GetUserInfo(code string, realm string, authType string) (*dto.UserInfoResponse, error) {
	endpoint := "https://gerbang.web.bps.go.id/api/v1/auth/login"

	payload := map[string]string{
		"code":  code,
		"type":  authType,
		"realm": realm,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "id-ID,id;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Origin", "https://gerbang.web.bps.go.id")
	req.Header.Set("Referer", "https://gerbang.web.bps.go.id/")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch user info from gerbang BPS: status %d", resp.StatusCode)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo dto.UserInfoResponse
	if err := json.Unmarshal(responseBody, &userInfo); err == nil && userInfo.Email != "" {
		return &userInfo, nil
	}

	var wrapped dto.BPSAuthLoginResponse
	if err := json.Unmarshal(responseBody, &wrapped); err == nil && wrapped.Data.Email != "" {
		return &dto.UserInfoResponse{
			Sub:   wrapped.Data.Sub,
			Name:  wrapped.Data.Name,
			Email: wrapped.Data.Email,
		}, nil
	}

	return nil, fmt.Errorf("failed to fetch user info from gerbang BPS")
}
