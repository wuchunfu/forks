package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"forks.com/m/models"
	"forks.com/m/utils"
)

var (
	// 代理配置
	httpProxy  string
	httpsProxy string
	noProxy    string

	// 快捷代理配置
	localProxy  string
	proxyType   string
	proxyPreset string

	// 运行时代理配置管理
	currentProxyConfig models.ProxyConfig
	proxyConfigMutex   sync.RWMutex
)

// processShortcutProxy 处理快捷代理配置
func processShortcutProxy() {
	// 处理预设配置
	if proxyPreset != "" {
		switch proxyPreset {
		case "shadowsocks", "ss":
			localProxy = "1080"
			proxyType = "socks5"
			fmt.Printf("🔧 使用Shadowsocks预设 (SOCKS5:1080)\n")
		case "v2ray":
			localProxy = "1080"
			proxyType = "http"
			fmt.Printf("🔧 使用V2Ray预设 (HTTP:1080)\n")
		case "clash":
			localProxy = "7890"
			proxyType = "http"
			fmt.Printf("🔧 使用Clash预设 (HTTP:7890)\n")
		case "surge":
			localProxy = "6152"
			proxyType = "http"
			fmt.Printf("🔧 使用Surge预设 (HTTP:6152)\n")
		case "qv2ray":
			localProxy = "8080"
			proxyType = "http"
			fmt.Printf("🔧 使用Qv2ray预设 (HTTP:8080)\n")
		default:
			fmt.Printf("⚠️  未知的代理预设: %s\n", proxyPreset)
		}
	}

	// 处理本地代理端口配置
	if localProxy != "" {
		var proxyURL string

		// 构建代理URL
		if proxyType == "socks5" {
			proxyURL = fmt.Sprintf("socks5://127.0.0.1:%s", localProxy)
		} else {
			// 默认使用HTTP代理
			proxyURL = fmt.Sprintf("http://127.0.0.1:%s", localProxy)
		}

		// 如果没有明确指定http/https代理，则使用快捷配置
		if httpProxy == "" && httpsProxy == "" {
			httpProxy = proxyURL
			httpsProxy = proxyURL
			fmt.Printf("🚀 快捷代理配置: %s\n", proxyURL)
		}

		// 设置常见的no-proxy列表
		if noProxy == "" {
			noProxy = "localhost,127.0.0.1,0.0.0.0,::1,*.local"
		}
	}
}

// setupProxy 设置代理配置
func setupProxy() {
	// 首先处理快捷代理配置
	processShortcutProxy()

	// 如果命令行参数为空，尝试从环境变量获取
	if httpProxy == "" {
		httpProxy = os.Getenv("HTTP_PROXY")
		if httpProxy == "" {
			httpProxy = os.Getenv("http_proxy")
		}
	}

	if httpsProxy == "" {
		httpsProxy = os.Getenv("HTTPS_PROXY")
		if httpsProxy == "" {
			httpsProxy = os.Getenv("https_proxy")
		}
	}

	if noProxy == "" {
		noProxy = os.Getenv("NO_PROXY")
		if noProxy == "" {
			noProxy = os.Getenv("no_proxy")
		}
	}

	// 设置HTTP代理环境变量
	if httpProxy != "" {
		os.Setenv("HTTP_PROXY", httpProxy)
		os.Setenv("http_proxy", httpProxy)
		fmt.Printf("✓ 设置HTTP代理: %s\n", httpProxy)
	}

	if httpsProxy != "" {
		os.Setenv("HTTPS_PROXY", httpsProxy)
		os.Setenv("https_proxy", httpsProxy)
		fmt.Printf("✓ 设置HTTPS代理: %s\n", httpsProxy)
	}

	if noProxy != "" {
		os.Setenv("NO_PROXY", noProxy)
		os.Setenv("no_proxy", noProxy)
		fmt.Printf("✓ 设置代理排除列表: %s\n", noProxy)
	}

	// 同时设置Git的代理配置
	if httpProxy != "" {
		exec.Command("git", "config", "--global", "http.proxy", httpProxy).Run()
		fmt.Printf("✓ 设置Git HTTP代理: %s\n", httpProxy)
	}

	if httpsProxy != "" {
		exec.Command("git", "config", "--global", "https.proxy", httpsProxy).Run()
		fmt.Printf("✓ 设置Git HTTPS代理: %s\n", httpsProxy)
	}

	// 如果有代理设置，显示总结信息
	if httpProxy != "" || httpsProxy != "" {
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("🌐 代理配置已应用，Git操作将通过代理服务器进行")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	}
}

// getCurrentProxyConfig 获取当前代理配置
func getCurrentProxyConfig() models.ProxyConfig {
	proxyConfigMutex.RLock()
	defer proxyConfigMutex.RUnlock()
	return currentProxyConfig
}

// getProxyConfig 获取代理配置API
func getProxyConfig(c *gin.Context) {
	config := getCurrentProxyConfig()
	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data":    config,
	})
}

// updateProxyConfig 更新代理配置API
func updateProxyConfig(c *gin.Context) {
	var req models.ProxyConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	// 验证代理类型
	if req.Type != "none" && req.Type != "http" && req.Type != "socks5" {
		c.JSON(400, gin.H{"code": 400, "message": "无效的代理类型"})
		return
	}

	// 如果启用代理，验证端口
	if req.Type != "none" {
		if req.Port <= 0 || req.Port > 65535 {
			c.JSON(400, gin.H{"code": 400, "message": "无效的端口号"})
			return
		}
		if req.Host == "" {
			req.Host = "127.0.0.1"
		}
	}

	// 更新配置
	proxyConfigMutex.Lock()
	currentProxyConfig = req
	proxyConfigMutex.Unlock()

	// 应用新配置
	if err := applyProxyConfig(req); err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "应用代理配置失败: " + err.Error()})
		return
	}

	// 保存到配置文件
	if err := saveProxyConfigToFile(req); err != nil {
		log.Printf("⚠️ 保存代理配置到文件失败: %v", err)
	}

	c.JSON(200, gin.H{
		"code":    0,
		"message": "代理配置已更新",
		"data":    req,
	})
}

// applyProxyConfig 应用代理配置到系统和Git
func applyProxyConfig(config models.ProxyConfig) error {
	// 清除现有代理设置
	os.Unsetenv("HTTP_PROXY")
	os.Unsetenv("http_proxy")
	os.Unsetenv("HTTPS_PROXY")
	os.Unsetenv("https_proxy")
	os.Unsetenv("NO_PROXY")
	os.Unsetenv("no_proxy")

	// 清除Git代理设置
	exec.Command("git", "config", "--global", "--unset", "http.proxy").Run()
	exec.Command("git", "config", "--global", "--unset", "https.proxy").Run()

	// 如果禁用代理，直接返回
	if config.Type == "none" || !config.Enabled {
		log.Println("🌐 代理已禁用")
		return nil
	}

	// 构建代理URL
	var proxyURL string
	if config.Type == "socks5" {
		proxyURL = fmt.Sprintf("socks5://%s:%d", config.Host, config.Port)
	} else {
		proxyURL = fmt.Sprintf("http://%s:%d", config.Host, config.Port)
	}

	// 设置环境变量
	os.Setenv("HTTP_PROXY", proxyURL)
	os.Setenv("http_proxy", proxyURL)
	os.Setenv("HTTPS_PROXY", proxyURL)
	os.Setenv("https_proxy", proxyURL)

	// 设置no_proxy
	noProxyValue := config.NoProxy
	if noProxyValue == "" {
		noProxyValue = "localhost,127.0.0.1,0.0.0.0,::1,*.local"
	}
	os.Setenv("NO_PROXY", noProxyValue)
	os.Setenv("no_proxy", noProxyValue)

	// 设置Git代理
	exec.Command("git", "config", "--global", "http.proxy", proxyURL).Run()
	exec.Command("git", "config", "--global", "https.proxy", proxyURL).Run()

	log.Printf("🌐 代理配置已应用: %s", proxyURL)
	return nil
}

// IsProxyEnabledForPlatform 判断某个平台是否启用代理
// 如果 platforms 中有该平台的配置，返回其值
// 否则返回全局 enabled 状态
func IsProxyEnabledForPlatform(platformName string) bool {
	config := getCurrentProxyConfig()
	if !config.Enabled || config.Type == "none" {
		return false
	}
	if config.Platforms != nil {
		if enabled, ok := config.Platforms[platformName]; ok {
			return enabled
		}
	}
	return config.Enabled
}

// applyGitProxyForPlatform 根据平台是否启用代理，临时设置或取消 git proxy 环境变量
// 返回一个 restore 函数，调用方可恢复原状态
func applyGitProxyForPlatform(platformName string) (proxyURL string, restore func()) {
	config := getCurrentProxyConfig()
	enabled := IsProxyEnabledForPlatform(platformName)

	if enabled {
		log.Printf("🔧 [代理] 平台 %s 启用代理 (类型: %s)", platformName, config.Type)
	} else {
		log.Printf("🔧 [代理] 平台 %s 未启用代理, 直连", platformName)
	}

	// 保存当前环境变量
	savedHTTPProxy := os.Getenv("HTTP_PROXY")
	savedHTTPSProxy := os.Getenv("HTTPS_PROXY")

	if enabled {
		// 构建代理URL
		if config.Type == "socks5" {
			proxyURL = fmt.Sprintf("socks5://%s:%d", config.Host, config.Port)
		} else {
			proxyURL = fmt.Sprintf("http://%s:%d", config.Host, config.Port)
		}
		os.Setenv("HTTP_PROXY", proxyURL)
		os.Setenv("http_proxy", proxyURL)
		os.Setenv("HTTPS_PROXY", proxyURL)
		os.Setenv("https_proxy", proxyURL)
	} else {
		// 该平台不需要代理，临时清除
		os.Unsetenv("HTTP_PROXY")
		os.Unsetenv("http_proxy")
		os.Unsetenv("HTTPS_PROXY")
		os.Unsetenv("https_proxy")
	}

	restore = func() {
		if savedHTTPProxy != "" {
			os.Setenv("HTTP_PROXY", savedHTTPProxy)
			os.Setenv("http_proxy", savedHTTPProxy)
		} else {
			os.Unsetenv("HTTP_PROXY")
			os.Unsetenv("http_proxy")
		}
		if savedHTTPSProxy != "" {
			os.Setenv("HTTPS_PROXY", savedHTTPSProxy)
			os.Setenv("https_proxy", savedHTTPSProxy)
		} else {
			os.Unsetenv("HTTPS_PROXY")
			os.Unsetenv("https_proxy")
		}
	}
	if enabled && proxyURL != "" {
		log.Printf("🔧 [代理] 已设置代理环境变量: HTTP_PROXY=%s, HTTPS_PROXY=%s", proxyURL, proxyURL)
	}
	return proxyURL, restore
}

// logGitCmd 打印 git 命令及当前代理环境
func logGitCmd(args ...string) {
	cmdStr := "git " + strings.Join(args, " ")
	httpProxy := os.Getenv("HTTP_PROXY")
	if httpProxy != "" {
		log.Printf("⚙️ [Git] %s  | 代理: %s", cmdStr, httpProxy)
	} else {
		log.Printf("⚙️ [Git] %s  | 直连", cmdStr)
	}
}

// saveProxyConfigToFile 保存代理配置到配置文件
func saveProxyConfigToFile(config models.ProxyConfig) error {
	configFile := utils.ConfigInstance.GetAppConfigFile()

	// 读取现有配置
	var fullConfig map[string]interface{}
	data, err := os.ReadFile(configFile)
	if err != nil {
		// 文件不存在，创建新配置
		fullConfig = make(map[string]interface{})
	} else {
		if err := json.Unmarshal(data, &fullConfig); err != nil {
			fullConfig = make(map[string]interface{})
		}
	}

	// 更新代理配置
	fullConfig["proxy"] = config

	// 保存到文件
	newData, err := json.MarshalIndent(fullConfig, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, newData, 0644)
}

// loadProxyConfigFromFile 从配置文件加载代理配置
func loadProxyConfigFromFile() models.ProxyConfig {
	configFile := utils.ConfigInstance.GetAppConfigFile()

	data, err := os.ReadFile(configFile)
	if err != nil {
		return models.ProxyConfig{Type: "none"}
	}

	var fullConfig struct {
		Proxy models.ProxyConfig `json:"proxy"`
	}

	if err := json.Unmarshal(data, &fullConfig); err != nil {
		return models.ProxyConfig{Type: "none"}
	}

	return fullConfig.Proxy
}

// initRuntimeProxyConfig 初始化运行时代理配置
func initRuntimeProxyConfig() {
	// 注册代理配置回调，供 utils 包获取当前代理配置
	utils.SetProxyConfigCallback(func() models.ProxyConfig {
		return getCurrentProxyConfig()
	})

	// 尝试从配置文件加载
	savedConfig := loadProxyConfigFromFile()
	if savedConfig.Type != "" {
		proxyConfigMutex.Lock()
		currentProxyConfig = savedConfig
		proxyConfigMutex.Unlock()

		// 应用保存的配置
		if savedConfig.Enabled && savedConfig.Type != "none" {
			applyProxyConfig(savedConfig)
			return
		}
	}

	// 如果没有保存的配置，使用命令行参数
	proxyConfigMutex.Lock()
	if httpProxy != "" || httpsProxy != "" || localProxy != "" {
		// 从命令行参数构建配置
		currentProxyConfig.Enabled = true
		currentProxyConfig.NoProxy = noProxy

		// 处理快捷代理配置
		if localProxy != "" {
			port, _ := strconv.Atoi(localProxy)
			currentProxyConfig.Port = port
			currentProxyConfig.Host = "127.0.0.1"
			if proxyType == "socks5" {
				currentProxyConfig.Type = "socks5"
			} else {
				currentProxyConfig.Type = "http"
			}
		} else if httpProxy != "" {
			currentProxyConfig.Type = "http"
			// 尝试解析URL
			if strings.HasPrefix(httpProxy, "http://") {
				parts := strings.TrimPrefix(httpProxy, "http://")
				if idx := strings.LastIndex(parts, ":"); idx > 0 {
					currentProxyConfig.Host = parts[:idx]
					port, _ := strconv.Atoi(parts[idx+1:])
					currentProxyConfig.Port = port
				}
			}
		}
	} else {
		currentProxyConfig.Type = "none"
		currentProxyConfig.Enabled = false
	}
	proxyConfigMutex.Unlock()
}
