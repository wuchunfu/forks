package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type TrendingSyncTask struct {
	Language          string `json:"language"`
	SpokenLanguageCode string `json:"spoken_language_code"`
	Since             string `json:"since"`
}

type TrendingSyncConfig struct {
	Enabled      bool               `json:"enabled"`
	SyncTime     string             `json:"sync_time"`
	LastSyncDate string             `json:"last_sync_date"`
	Tasks        []TrendingSyncTask `json:"tasks"`
}

var (
	syncConfigMu     sync.RWMutex
	syncConfigPath   string
	syncRunning      bool
	syncRunningMu    sync.Mutex
)

func syncConfigFilePath() string {
	if syncConfigPath == "" {
		syncConfigPath = filepath.Join(ConfigInstance.GetAppConfigDir(), "sync_config.json")
	}
	return syncConfigPath
}

func LoadSyncConfig() (*TrendingSyncConfig, error) {
	syncConfigMu.RLock()
	defer syncConfigMu.RUnlock()

	raw, err := os.ReadFile(syncConfigFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return defaultSyncConfig(), nil
		}
		return nil, err
	}
	var cfg TrendingSyncConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func SaveSyncConfig(cfg *TrendingSyncConfig) error {
	syncConfigMu.Lock()
	defer syncConfigMu.Unlock()

	dir := filepath.Dir(syncConfigFilePath())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	raw, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(syncConfigFilePath(), raw, 0644)
}

func defaultSyncConfig() *TrendingSyncConfig {
	return &TrendingSyncConfig{
		Enabled:  false,
		SyncTime: "18:00",
		Tasks:    []TrendingSyncTask{},
	}
}

// RunSyncTasks 执行所有同步任务
func RunSyncTasks() error {
	syncRunningMu.Lock()
	if syncRunning {
		syncRunningMu.Unlock()
		return fmt.Errorf("同步正在执行中")
	}
	syncRunning = true
	syncRunningMu.Unlock()

	defer func() {
		syncRunningMu.Lock()
		syncRunning = false
		syncRunningMu.Unlock()
	}()

	cfg, err := LoadSyncConfig()
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	if len(cfg.Tasks) == 0 {
		return fmt.Errorf("没有配置同步任务")
	}

	today := time.Now().Format("2006-01-02")
	successCount := 0

	for i, task := range cfg.Tasks {
		since := task.Since
		if since == "" {
			since = "daily"
		}

		log.Printf("[Trending 同步] %d/%d: language=%s spoken=%s since=%s",
			i+1, len(cfg.Tasks), task.Language, task.SpokenLanguageCode, since)

		repos, err := FetchTrendingData(task.Language, since, task.SpokenLanguageCode)
		if err != nil {
			log.Printf("[Trending 同步] 失败: %v", err)
			continue
		}

		if err := SaveTrendingData(today, since, task.Language, task.SpokenLanguageCode, repos); err != nil {
			log.Printf("[Trending 同步] 保存失败: %v", err)
			continue
		}
		successCount++
	}

	cfg.LastSyncDate = today
	_ = SaveSyncConfig(cfg)

	log.Printf("[Trending 同步] 完成: %d/%d 成功", successCount, len(cfg.Tasks))
	return nil
}

// IsSyncRunning 返回同步是否正在执行
func IsSyncRunning() bool {
	syncRunningMu.Lock()
	defer syncRunningMu.Unlock()
	return syncRunning
}

// StartSyncScheduler 启动定时同步调度器
func StartSyncScheduler() {
	go func() {
		// 启动时检查：如果今天还没同步且已过 sync_time，立即执行
		if err := checkAndRunOnStartup(); err != nil {
			log.Printf("[Trending 调度] 启动检查: %v", err)
		}

		// 每分钟检查一次
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			cfg, err := LoadSyncConfig()
			if err != nil || !cfg.Enabled || len(cfg.Tasks) == 0 {
				continue
			}

			now := time.Now()
			today := now.Format("2006-01-02")
			if cfg.LastSyncDate == today {
				continue
			}

			syncHour, syncMin := parseSyncTime(cfg.SyncTime)
			if now.Hour() == syncHour && now.Minute() == syncMin {
				log.Printf("[Trending 调度] 到达同步时间 %s，开始执行", cfg.SyncTime)
				if err := RunSyncTasks(); err != nil {
					log.Printf("[Trending 调度] 执行失败: %v", err)
				}
			}
		}
	}()
}

func checkAndRunOnStartup() error {
	cfg, err := LoadSyncConfig()
	if err != nil || !cfg.Enabled || len(cfg.Tasks) == 0 {
		return nil
	}

	today := time.Now().Format("2006-01-02")
	if cfg.LastSyncDate == today {
		return nil // 今天已同步
	}

	now := time.Now()
	syncHour, syncMin := parseSyncTime(cfg.SyncTime)
	currentMinutes := now.Hour()*60 + now.Minute()
	syncMinutes := syncHour*60 + syncMin

	if currentMinutes >= syncMinutes {
		log.Printf("[Trending 调度] 启动时补执行同步 (已过 %s)", cfg.SyncTime)
		return RunSyncTasks()
	}
	return nil
}

func parseSyncTime(t string) (hour, minute int) {
	hour, minute = 18, 0 // default 18:00 (UTC+8 ≈ GitHub 10:00 UTC)
	parts := parseTimeParts(t)
	if len(parts) == 2 {
		if h, err := atoi(parts[0]); err == nil && h >= 0 && h < 24 {
			hour = h
		}
		if m, err := atoi(parts[1]); err == nil && m >= 0 && m < 60 {
			minute = m
		}
	}
	return
}

func parseTimeParts(t string) []string {
	for _, sep := range []string{":", "："} {
		if parts := splitN(t, sep, 2); len(parts) == 2 {
			return parts
		}
	}
	return nil
}

func splitN(s, sep string, n int) []string {
	result := make([]string, 0, n)
	for i := 0; i < len(s) && len(result) < n-1; i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[:i])
			result = append(result, s[i+len(sep):])
			return result
		}
	}
	return nil
}

func atoi(s string) (int, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid")
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}
