package config

import (
	"os"
	"strings"
)

const (
	IntegrationWebsiteBaseURLEnvKey = "INTEGRATION_WEBSITE_BASE_URL"

	DefaultIntegrationWebsiteBaseURL = "https://bontangkota.archive.bps.go.id"
	DefaultIntegrationWebsiteReferer = "/backend/index.php/dataDynamic/menu/id/2"
)

type IntegrationConfig struct {
	WebsiteBaseURL string
	WebsiteReferer string
}

func LoadIntegrationConfig() (*IntegrationConfig, error) {
	baseURL := strings.TrimSpace(os.Getenv(IntegrationWebsiteBaseURLEnvKey))
	if baseURL == "" {
		baseURL = DefaultIntegrationWebsiteBaseURL
	}

	referer := strings.TrimRight(baseURL, "/") + DefaultIntegrationWebsiteReferer

	return &IntegrationConfig{
		WebsiteBaseURL: baseURL,
		WebsiteReferer: referer,
	}, nil
}
