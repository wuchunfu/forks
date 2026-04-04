package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const Version = "1.3.0"

// 平台域名映射
var PlatformDomains = map[string]string{
	"github":    "github.com",
	"gitee":     "gitee.com",
	"gitlab":    "gitlab.com",
	"bitbucket": "bitbucket.org",
}

// Config fclone 本地配置
type Config struct {
	Server string `json:"server,omitempty"` // 镜像服务地址，如 http://192.168.1.100:8080
	Token  string `json:"token,omitempty"`
}

// GetConfigPath 返回配置文件路径 (~/.fclone.json)
func GetConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".fclone.json")
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
