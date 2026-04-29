package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"

	"github.com/cicbyte/forks/models"
)

var ConfigInstance = Config{}

type Config struct {
	HomeDir       string
	AppSeriesDir  string
	AppDir        string
	AppConfigDir  string
	AppConfigFile string
	LogDir        string
	DbDir         string
	DbPath        string
}

func (c *Config) ReadConfig() (*models.GitPlusConfig, error) {
	// 读取配置文件
	file, err := os.Open(c.GetAppConfigFile())
	if err != nil {
		fmt.Println("无法打开文件")
		return nil, err
	}
	defer file.Close()

	// 读取文件内容
	byteValue, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("无法读取文件")
		return nil, err
	}
	var model models.GitPlusConfig
	err = json.Unmarshal(byteValue, &model)
	if err != nil {
		fmt.Println("配置文件格式错误")
		return nil, err
	}
	// 环境变量覆盖仓库存储路径
	if dir := os.Getenv("FORKS_REPO_PATH"); dir != "" {
		model.StoreRootPath = dir
	}
	return &model, nil
}

func (c *Config) GetHomeDir() string {
	if c.HomeDir != "" {
		return c.HomeDir
	}
	usr, err := user.Current()
	if err != nil {
		panic(fmt.Sprintf("Failed to get current user: %v", err))
	}
	c.HomeDir = usr.HomeDir
	return c.HomeDir
}

func (c *Config) GetAppSeriesDir() string {
	if c.AppSeriesDir != "" {
		return c.AppSeriesDir
	}
	c.AppSeriesDir = c.GetHomeDir() + "/.cicbyte"
	return c.AppSeriesDir
}

func (c *Config) GetAppDir() string {
	if c.AppDir != "" {
		return c.AppDir
	}
	// 环境变量覆盖，用于 Docker 等容器化部署
	if dir := os.Getenv("FORKS_HOME"); dir != "" {
		c.AppDir = dir
		return c.AppDir
	}
	c.AppDir = c.GetAppSeriesDir() + "/forks"
	return c.AppDir
}

func (c *Config) GetAppConfigDir() string {
	if c.AppConfigDir != "" {
		return c.AppConfigDir
	}
	c.AppConfigDir = c.GetAppDir() + "/config"
	return c.AppConfigDir
}

func (c *Config) GetAppConfigFile() string {
	if c.AppConfigFile != "" {
		return c.AppConfigFile
	}
	c.AppConfigFile = c.GetAppConfigDir() + "/config.json"
	return c.AppConfigFile
}

func (c *Config) GetLogDir() string {
	if c.LogDir != "" {
		return c.LogDir
	}
	homeDir := filepath.Join(c.GetAppDir(), "logs")
	c.LogDir = homeDir
	return c.LogDir
}

func (c *Config) GetDbDir() string {
	if c.DbDir != "" {
		return c.DbDir
	}
	dbDir := filepath.Join(c.GetAppDir(), "db")
	c.DbDir = dbDir
	return c.DbDir
}

func (c *Config) GetDbPath() string {
	if c.DbPath != "" {
		return c.DbPath
	}
	c.DbPath = filepath.Join(c.GetDbDir(), "forks.db")
	return c.DbPath
}
