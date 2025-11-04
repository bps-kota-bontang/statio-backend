package config

import (
	"os"
)

// Cloudflare menyimpan konfigurasi Cloudflare R2
type SchedulerConfig struct {
	CronExpressionSyncMainAvatar string
	CronExpressionSyncNewAvatar  string
}

func LoadSchedulerConfig() (*SchedulerConfig, error) {
	return &SchedulerConfig{
		CronExpressionSyncMainAvatar: os.Getenv("SCHEDULER_CRON_EXPRESSION_SYNC_MAIN_AVATAR"),
		CronExpressionSyncNewAvatar:  os.Getenv("SCHEDULER_CRON_EXPRESSION_SYNC_NEW_AVATAR"),
	}, nil
}
