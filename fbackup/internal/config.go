package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const Version = "0.1.0"

// Config fbackup 本地配置
type Config struct {
	Server string `json:"server,omitempty"` // Forks 服务地址，如 http://192.168.1.100:8080
	Token  string `json:"token,omitempty"`
	Dir    string `json:"dir,omitempty"` // 本地备份目录
}

// GetConfigPath 返回配置文件路径 (~/.fbackup.json)
func GetConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".fbackup.json")
}

// LoadConfig 从文件加载配置
func LoadConfig() Config {
	path := GetConfigPath()
	if path == "" {
		return Config{}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}
	}
	var cfg Config
	json.Unmarshal(data, &cfg)
	return cfg
}

// SaveConfig 保存配置到文件
func SaveConfig(cfg Config) error {
	path := GetConfigPath()
	if path == "" {
		return fmt.Errorf("无法确定配置文件路径")
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// MaskToken 脱敏显示 token
func MaskToken(t string) string {
	if len(t) <= 8 {
		return "****"
	}
	return t[:4] + "****" + t[len(t)-4:]
}
