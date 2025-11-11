package services

import (
	"encoding/json"
	"errors"
	"net/http"
	"statio/internal/dto"
)

type BPSService struct {
}

func NewBPSService() *BPSService {
	return &BPSService{}
}

func (s *BPSService) GetUserInfo(token string) (*dto.UserInfoResponse, error) {
	endpoint := "https://sso.bps.go.id/auth/realms/pegawai-bps/protocol/openid-connect/userinfo"

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch user info")
	}

	var userInfo dto.UserInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}
