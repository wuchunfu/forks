package cmd

import (
	"bufio"
	"compress/gzip"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
	"forks.com/m/assets"
	"forks.com/m/common"
	"forks.com/m/models"
	"forks.com/m/utils"
)

var (
	port           int
	address        string
	token          string
	customToken    string
	tempTokenStore   = make(map[string]string) // 临时token存储
	tempTokenMutex   sync.RWMutex

	// 批量克隆 ID 列表存储
	batchCloneIDsStore = make(map[string][]int)
	batchCloneIDsMutex sync.RWMutex

	// 代理配置
	httpProxy  string
	httpsProxy string
	noProxy    string

	// 快捷代理配置
	localProxy  string // 本地代理端口，例如：1080
	proxyType   string // 代理类型：http, socks5
	proxyPreset string // 代理预设：shadowsocks, v2ray, clash

	// 扫描任务管理
	scanTaskMutex sync.RWMutex
	scanTasks     = make(map[string]*ScanTask) // 扫描任务map

	// 运行时代理配置管理
	currentProxyConfig models.ProxyConfig
	proxyConfigMutex   sync.RWMutex
)

// ScanTask 扫描任务状态
type ScanTask struct {
	mu            sync.RWMutex
	ID            string
	Status        string // "running", "completed", "error"
	Progress      int
	TotalDirs     int
	ScannedDirs   int
	NewRepos      []gin.H
	MissingRepos  []gin.H
	Error         string
	StartTime     time.Time
}

// generateToken 生成随机token
func generateToken() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// maskToken 脱敏显示 token（前4位...后4位）
func maskToken(t string) string {
	if len(t) <= 8 {
		return "****"
	}
	return t[:4] + "****" + t[len(t)-4:]
}

// getAuthFile 获取认证配置文件路径
func getAuthFile() string {
	return utils.ConfigInstance.GetAppConfigDir() + "/auth.json"
}

// saveTokenToFile 保存 token 到认证配置文件
func saveTokenToFile(tok string) error {
	authFile := getAuthFile()

	authData := map[string]string{"token": tok}
	newData, err := json.MarshalIndent(authData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(authFile, newData, 0600)
}

// loadTokenFromFile 从认证配置文件加载 token
func loadTokenFromFile() string {
	authFile := getAuthFile()

	data, err := os.ReadFile(authFile)
	if err != nil {
		return ""
	}

	var authData struct {
		Token string `json:"token"`
	}

	if err := json.Unmarshal(data, &authData); err != nil {
		return ""
	}

	return authData.Token
}

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

// authMiddleware token认证中间件
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"code": 401, "message": "未提供认证信息"})
			c.Abort()
			return
		}

		if authHeader != "Bearer "+token {
			c.JSON(401, gin.H{"code": 401, "message": "认证失败"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// sseAuthMiddleware SSE端点认证中间件，支持临时token认证
func sseAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		path := c.Request.URL.Path
		log.Printf("🔐 [SSE认证] 请求路径: %s", path)

		// 首先尝试Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			if authHeader == "Bearer "+token {
				log.Printf("✅ [SSE认证] Authorization header 认证成功")
				c.Next()
				return
			}
		}

		// 检查临时token
		queryToken := c.Query("tempToken")
		if queryToken != "" {
			log.Printf("🔑 [SSE认证] 使用 tempToken 认证")

			// 根据路径确定操作类型
			var operation string
			var keyID string

			if strings.Contains(path, "scan-status") {
				// 扫描任务使用查询参数taskId
				taskID := c.Query("taskId")
				keyID = "scan_" + taskID
				log.Printf("🔑 [SSE认证] 扫描任务, keyID=%s", keyID)
			} else if strings.Contains(path, "batch-clone-status") {
				// 批量克隆任务
				keyID = "batch_clone"
				log.Printf("🔑 [SSE认证] 批量克隆任务, keyID=%s", keyID)
			} else {
				// 其他操作使用路径参数id
				if strings.Contains(path, "clone-status") {
					operation = "_clone"
				} else if strings.Contains(path, "pull-status") {
					operation = "_pull"
				} else if strings.Contains(path, "reset-status") {
					operation = "_reset"
				}
				keyID = id + operation
				log.Printf("🔑 [SSE认证] 其他操作, keyID=%s", keyID)
			}

			tempTokenMutex.RLock()
			expectedToken, exists := tempTokenStore[keyID]
			tempTokenMutex.RUnlock()
			log.Printf("🔑 [SSE认证] 期望token存在: %v, 匹配: %v", exists, expectedToken == queryToken)

			if exists && queryToken == expectedToken {
				// 对于扫描任务和批量克隆，不删除token（支持重连）
				if !strings.Contains(path, "scan-status") && !strings.Contains(path, "batch-clone-status") {
					tempTokenMutex.Lock()
					delete(tempTokenStore, keyID)
					tempTokenMutex.Unlock()
				}
				log.Printf("✅ [SSE认证] 认证成功")
				c.Next()
				return
			}
		}

		// 认证失败，对于SSE返回错误事件
		log.Printf("❌ [SSE认证] 认证失败")
		c.Header("Content-Type", "text/event-stream")
		c.SSEvent("error", gin.H{"message": "认证失败"})
		c.Abort()
	}
}

// corsMiddleware CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, Accept, Mcp-Protocol-Version")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
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

// getTokenInfo 获取当前 Token 信息（脱敏）
func getTokenInfo(c *gin.Context) {
	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"token": maskToken(token),
		},
	})
}

// updateTokenInfo 更新 Token
func updateTokenInfo(c *gin.Context) {
	var req struct {
		Token       string `json:"token"`
		Regenerate  bool   `json:"regenerate"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	var newToken string
	if req.Regenerate {
		newToken = generateToken()
	} else if req.Token != "" {
		if len(req.Token) < 8 {
			c.JSON(400, gin.H{"code": 400, "message": "Token 长度不能少于8位"})
			return
		}
		newToken = req.Token
	} else {
		c.JSON(400, gin.H{"code": 400, "message": "请提供 token 或 regenerate 参数"})
		return
	}

	// 更新内存中的 token（authMiddleware 和 sseAuthMiddleware 直接读该变量）
	token = newToken

	// 持久化到配置文件
	if err := saveTokenToFile(newToken); err != nil {
		log.Printf("⚠️ 保存 token 到文件失败: %v", err)
	}

	log.Printf("🔑 Token 已更新: %s", maskToken(newToken))

	c.JSON(200, gin.H{
		"code":    0,
		"message": "Token 已更新",
		"data": gin.H{
			"token": newToken,
		},
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

// setupRoutes 设置路由
func setupRoutes(r *gin.Engine) {
	// 添加CORS中间件
	r.Use(corsMiddleware())

	// API路由组
	api := r.Group("/api")
	api.Use(authMiddleware())
	{
		// 仓库管理接口
		api.GET("/repos", getRepos)
		api.POST("/repos", addRepo)
		api.POST("/repos/scan", scanRepos) // 扫描本地Git仓库
		api.DELETE("/repos/:id", deleteRepo)
		api.GET("/repos/:id", getRepo)
		api.PUT("/repos/:id/update", updateRepoInfo)

		// 作者管理接口
		api.GET("/authors", getAuthors)

		// 统计接口
		api.GET("/stats", getStats)

		// 活动记录接口
		api.GET("/activities", getActivities)
		api.POST("/activities", addActivity)
		api.DELETE("/activities", clearActivities)

		// 系统信息接口
		api.GET("/info", getSystemInfo)

		// 代码查看接口
		api.GET("/repos/:id/files", getRepoFiles)

		// 批量克隆
		api.POST("/repos/batch-clone", batchCloneRepos)
		api.GET("/repos/:id/file-content", getFileContent)

		// Git 操作接口
		api.POST("/repos/:id/clone", cloneRepo)
		api.POST("/repos/:id/pull", pullRepo)
		api.POST("/repos/:id/reset", resetRepo)
		api.GET("/repos/:id/status", getRepoStatus)
		api.GET("/repos/:id/diff", getRepoDiff)
		api.GET("/repos/:id/commits", getRepoCommits)
		api.GET("/repos/:id/branches", getRepoBranches)
		api.POST("/repos/:id/open-folder", openRepoFolder)
		api.DELETE("/repos/:id/local", deleteLocalRepo)

		// 代理配置接口
		api.GET("/proxy", getProxyConfig)
		api.POST("/proxy", updateProxyConfig)

		// Token 管理接口
		api.GET("/token", getTokenInfo)
		api.POST("/token", updateTokenInfo)

		// Git 镜像准备接口（供 fclone 调用）
		api.POST("/git/prepare", prepareGitMirror)
	}

	// SSE 路由组 - 使用特殊的SSE认证中间件
	sseApi := r.Group("/api")
	sseApi.Use(sseAuthMiddleware())
	{
		sseApi.GET("/repos/:id/clone-status", cloneRepoSSE)
		sseApi.GET("/repos/:id/pull-status", pullRepoSSE)
		sseApi.GET("/repos/:id/reset-status", resetRepoSSE)
		sseApi.GET("/repos/scan-status", scanReposSSE)
		sseApi.GET("/repos/batch-clone-status", batchCloneSSE)
	}

	// MCP 路由 — Streamable HTTP，复用 Bearer Token 认证
	mcpServer := utils.SetupMCPServer()
	mcpHandler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		return mcpServer
	}, nil)
	mcpGroup := r.Group("/mcp")
	mcpGroup.Use(func(c *gin.Context) {
		if token == "" {
			c.Next()
			return
		}
		authHeader := c.GetHeader("Authorization")
		if authHeader != "Bearer "+token {
			c.JSON(401, gin.H{"code": 401, "message": "认证失败"})
			c.Abort()
			return
		}
		c.Next()
	})
	mcpGroup.Any("", gin.WrapF(mcpHandler.ServeHTTP))

	// Git Smart HTTP 路由 — 允许从本服务克隆已缓存的仓库
	r.Any("/git/*path", gitHTTPHandler)

	// 使用嵌入的静态文件 - 使用自定义中间件避免覆盖API路由
	r.Use(func(c *gin.Context) {
		// 如果是API路径，跳过静态文件处理
		if strings.HasPrefix(c.Request.URL.Path, "/api") || strings.HasPrefix(c.Request.URL.Path, "/mcp") || strings.HasPrefix(c.Request.URL.Path, "/git") {
			c.Next()
			return
		}

		// 创建静态文件系统
		fs, err := static.EmbedFolder(assets.WebDist, "web/dist")
		if err != nil {
			c.String(500, "Failed to create static file system: "+err.Error())
			return
		}

		// 使用静态文件中间件
		staticHandler := static.Serve("/", fs)
		staticHandler(c)
	})

	// 处理SPA路由 - 所有非API路由返回index.html
	r.NoRoute(func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/api") && !strings.HasPrefix(c.Request.URL.Path, "/git") {
			c.Header("Content-Type", "text/html; charset=utf-8")
			indexFile, err := assets.WebDist.ReadFile("web/dist/index.html")
			if err != nil {
				c.String(500, "Failed to read index.html: "+err.Error())
				return
			}
			c.Data(200, "text/html; charset=utf-8", indexFile)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "API endpoint not found"})
		}
	})
}

// getRepos 获取仓库列表（分页）
func getRepos(c *gin.Context) {
	// 获取分页参数
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "10")
	search := c.Query("search")
	author := c.Query("author")
	status := c.Query("status") // cloned, not-cloned
	source := c.Query("source") // github, gitee

	pageNum, err := strconv.Atoi(page)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}

	pageSizeNum, err := strconv.Atoi(pageSize)
	if err != nil || pageSizeNum < 1 || pageSizeNum > 100 {
		pageSizeNum = 10
	}

	offset := (pageNum - 1) * pageSizeNum

	// 构建查询条件
	var whereClauses []string
	var args []interface{}
	var countArgs []interface{}

	// 搜索条件
	if search != "" {
		searchPattern := "%" + search + "%"
		whereClauses = append(whereClauses, "(author LIKE ? OR repo LIKE ? OR description LIKE ?)")
		args = append(args, searchPattern, searchPattern, searchPattern)
		countArgs = append(countArgs, searchPattern, searchPattern, searchPattern)
	}

	// 作者筛选
	if author != "" {
		whereClauses = append(whereClauses, "author = ?")
		args = append(args, author)
		countArgs = append(countArgs, author)
	}

	// 克隆状态筛选
	if status == "cloned" {
		whereClauses = append(whereClauses, "is_cloned = 1")
	} else if status == "not-cloned" {
		whereClauses = append(whereClauses, "(is_cloned != 1)")
	}

	// 平台筛选
	if source != "" {
		whereClauses = append(whereClauses, "source = ?")
		args = append(args, source)
		countArgs = append(countArgs, source)
	}

	// 构建 WHERE 子句
	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// 构建完整 SQL
	countSQL := "SELECT COUNT(*) FROM repos" + whereSQL
	querySQL := "SELECT id, author, repo, url, description, stars, forks, topics, license, created_at, COALESCE(updated_at, ''), COALESCE(is_cloned, 0), source FROM repos" + whereSQL + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, pageSizeNum, offset)

	log.Printf("🔍 [getRepos] SQL: %s, args: %v", querySQL, args)

	// 获取总数
	var total int
	err = common.Db.QueryRow(countSQL, countArgs...).Scan(&total)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "获取总数失败"})
		return
	}

	// 查询数据
	rows, err := common.Db.Query(querySQL, args...)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "数据库查询失败: " + err.Error()})
		return
	}
	defer rows.Close()

	var repos []gin.H
	for rows.Next() {
		var repo models.GitRepoInfo
		var id int
		var createdAt string
		var updatedAt string
		var isCloned int
		err := rows.Scan(&id, &repo.Author, &repo.Repo, &repo.URL, &repo.Description, &repo.Stars, &repo.Fork, &repo.Topics, &repo.License, &createdAt, &updatedAt, &isCloned, &repo.Source)
		if err != nil {
			continue
		}

		repoData := gin.H{
			"id":          id,
			"author":      repo.Author,
			"repo":        repo.Repo,
			"url":         repo.URL,
			"description": repo.Description,
			"stars":       repo.Stars,
			"forks":       repo.Fork,
			"topics":      repo.Topics,
			"license":     repo.License,
			"created_at":  createdAt,
			"updated_at":  updatedAt,
			"is_cloned":   isCloned,
			"source":      repo.Source,
		}
		repos = append(repos, repoData)
	}

	// 计算分页信息
	totalPages := (total + pageSizeNum - 1) / pageSizeNum

	log.Printf("✅ [getRepos] 返回 %d 条记录, 总数 %d, 筛选条件: status=%s, author=%s", len(repos), total, status, author)

	c.JSON(200, gin.H{
		"code": 0,
		"data": gin.H{
			"list":        repos,
			"total":       total,
			"page":        pageNum,
			"page_size":   pageSizeNum,
			"total_pages": totalPages,
		},
		"message": "success",
	})
}

// getAuthors 获取作者列表
func getAuthors(c *gin.Context) {
	search := c.Query("search")
	source := c.Query("source") // github, gitee
	sortBy := c.DefaultQuery("sort_by", "repo_count") // repo_count, name
	sortOrder := c.DefaultQuery("sort_order", "desc") // asc, desc

	// 构建查询 - 从 repos 表聚合作者信息
	querySQL := `
		SELECT
			author,
			source,
			COUNT(*) as repo_count,
			SUM(CASE WHEN is_cloned = 1 THEN 1 ELSE 0 END) as cloned_count,
			MAX(created_at) as last_updated
		FROM repos
		WHERE 1=1
	`

	var args []interface{}

	// 搜索条件
	if search != "" {
		querySQL += " AND author LIKE ?"
		args = append(args, "%"+search+"%")
	}

	// 平台筛选
	if source != "" {
		querySQL += " AND source = ?"
		args = append(args, source)
	}

	// 分组
	querySQL += " GROUP BY author, source"

	// 排序
	orderClause := " ORDER BY "
	if sortBy == "name" {
		orderClause += "author"
	} else {
		orderClause += "repo_count"
	}
	if sortOrder == "asc" {
		orderClause += " ASC"
	} else {
		orderClause += " DESC"
	}
	querySQL += orderClause

	log.Printf("🔍 [getAuthors] SQL: %s, args: %v", querySQL, args)

	// 查询数据
	rows, err := common.Db.Query(querySQL, args...)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "数据库查询失败: " + err.Error()})
		return
	}
	defer rows.Close()

	var authors []gin.H
	for rows.Next() {
		var author string
		var source string
		var repoCount int
		var clonedCount int
		var lastUpdated string

		err := rows.Scan(&author, &source, &repoCount, &clonedCount, &lastUpdated)
		if err != nil {
			continue
		}

		authors = append(authors, gin.H{
			"author":       author,
			"source":       source,
			"repo_count":   repoCount,
			"cloned_count": clonedCount,
			"last_updated": lastUpdated,
		})
	}

	log.Printf("✅ [getAuthors] 返回 %d 个作者", len(authors))

	c.JSON(200, gin.H{
		"code": 0,
		"data": gin.H{
			"list":  authors,
			"total": len(authors),
		},
		"message": "success",
	})
}

// addRepo 添加仓库
func addRepo(c *gin.Context) {
	var req struct {
		URL       string `json:"url" binding:"required"`
		AutoClone bool   `json:"autoClone"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 根据URL匹配平台
	platform, err := utils.GetPlatform(req.URL)
	if err != nil {
		c.JSON(400, gin.H{"code": 400, "message": err.Error()})
		return
	}

	author, repo, pageURL, err := platform.ParseURL(req.URL)
	if err != nil {
		c.JSON(400, gin.H{"code": 400, "message": err.Error()})
		return
	}

	platformName := platform.Name()
	proxyEnabled := IsProxyEnabledForPlatform(platformName)
	if proxyEnabled {
		proxyConfig := getCurrentProxyConfig()
		var pURL string
		if proxyConfig.Type == "socks5" {
			pURL = fmt.Sprintf("socks5://%s:%d", proxyConfig.Host, proxyConfig.Port)
		} else {
			pURL = fmt.Sprintf("http://%s:%d", proxyConfig.Host, proxyConfig.Port)
		}
		log.Printf("➕ [添加] %s/%s | 平台: %s | 代理: %s | URL: %s", author, repo, platformName, pURL, pageURL)
	} else {
		log.Printf("➕ [添加] %s/%s | 平台: %s | 直连 | URL: %s", author, repo, platformName, pageURL)
	}

	exists, err := platform.CheckExist(pageURL)
	if err != nil {
		log.Printf("❌ [添加] 检查仓库状态失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "message": "检查仓库状态失败，请稍后重试"})
		return
	}
	if exists {
		c.JSON(409, gin.H{"code": 409, "message": "该仓库已经添加过了"})
		return
	}

	// 获取仓库信息
	var repoInfo *models.GitRepoInfo
	if giteePlatform, ok := platform.(*utils.Gitee); ok {
		// Gitee 使用 API 获取信息
		repoInfo, err = giteePlatform.FetchRepoInfo(author, repo)
		if err != nil {
			log.Printf("❌ [添加] Gitee API 请求失败: %v", err)
			c.JSON(500, gin.H{"code": 500, "message": "无法获取仓库信息，请检查URL是否正确"})
			return
		}
	} else {
		// 其他平台使用 HTML 解析
		log.Printf("🔄 [添加] 正在抓取页面: %s", pageURL)
		doc, err := platform.FetchDocs(pageURL)
		if err != nil {
			c.JSON(500, gin.H{"code": 500, "message": "无法获取仓库信息，请检查URL是否正确"})
			return
		}
		repoInfo = platform.ParseDoc(doc)
	}

	repoInfo.URL = pageURL
	repoInfo.Author = author
	repoInfo.Repo = repo
	repoInfo.Source = platform.Name()

	// 保存到数据库
	err = platform.SaveRecords(repoInfo)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "保存仓库到数据库失败，请稍后重试"})
		return
	}

	// 获取刚插入的仓库ID
	var repoId int64
	err = common.Db.QueryRow("SELECT id FROM repos WHERE author = ? AND repo = ? AND source = ? ORDER BY created_at DESC LIMIT 1",
		author, repo, platform.Name()).Scan(&repoId)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "获取仓库ID失败，请稍后重试"})
		return
	}

	responseData := gin.H{
		"repoInfo":  repoInfo,
		"repoId":    repoId,
		"autoClone": req.AutoClone,
	}

	// 记录活动
	repoFullName := author + "/" + repo
	addActivityRecord("success", "添加仓库", fmt.Sprintf("成功添加仓库 %s", repoFullName), repoId, repoFullName)

	c.JSON(200, gin.H{"code": 0, "data": responseData, "message": "仓库添加成功"})
}

// deleteRepo 删除仓库
func deleteRepo(c *gin.Context) {
	id := c.Param("id")

	// 先获取仓库名称
	var repoName string
	common.Db.QueryRow("SELECT author || '/' || repo FROM repos WHERE id = ?", id).Scan(&repoName)

	_, err := common.Db.Exec("DELETE FROM repos WHERE id = ?", id)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "删除仓库失败"})
		return
	}

	// 记录活动（repo_id 为 0 因为已删除）
	addActivityRecord("warning", "删除仓库", fmt.Sprintf("已删除仓库 %s", repoName), 0, repoName)

	c.JSON(200, gin.H{"code": 0, "message": "删除成功"})
}

// getRepo 获取单个仓库信息
func getRepo(c *gin.Context) {
	id := c.Param("id")
	var repo models.GitRepoInfo
	var createdAt, updatedAt string
	var isCloned int
	var source string
	err := common.Db.QueryRow("SELECT author, repo, url, description, stars, forks, topics, license, created_at, COALESCE(updated_at, ''), COALESCE(is_cloned, 0), source FROM repos WHERE id = ?", id).
		Scan(&repo.Author, &repo.Repo, &repo.URL, &repo.Description, &repo.Stars, &repo.Fork, &repo.Topics, &repo.License, &createdAt, &updatedAt, &isCloned, &source)
	if err != nil {
		c.JSON(404, gin.H{"code": 404, "message": "仓库不存在"})
		return
	}
	c.JSON(200, gin.H{"code": 0, "data": gin.H{
		"id":          id,
		"author":      repo.Author,
		"repo":        repo.Repo,
		"url":         repo.URL,
		"description": repo.Description,
		"stars":       repo.Stars,
		"forks":       repo.Fork,
		"topics":      repo.Topics,
		"license":     repo.License,
		"created_at":  createdAt,
		"updated_at":  updatedAt,
		"is_cloned":   isCloned,
		"source":      source,
	}, "message": "success"})
}

// getStats 获取统计数据
func getStats(c *gin.Context) {
	// 仓库总数
	var totalRepos int
	err := common.Db.QueryRow("SELECT COUNT(*) FROM repos").Scan(&totalRepos)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "获取统计数据失败"})
		return
	}

	// 已克隆数量
	var clonedCount int
	err = common.Db.QueryRow("SELECT COUNT(*) FROM repos WHERE is_cloned = 1").Scan(&clonedCount)
	if err != nil {
		clonedCount = 0
	}

	// 未克隆数量（包括 0, NULL, -1 失效）
	var notClonedCount int
	err = common.Db.QueryRow("SELECT COUNT(*) FROM repos WHERE is_cloned != 1").Scan(&notClonedCount)
	if err != nil {
		notClonedCount = 0
	}

	// 作者数量
	var authorCount int
	err = common.Db.QueryRow("SELECT COUNT(DISTINCT author) FROM repos").Scan(&authorCount)
	if err != nil {
		authorCount = 0
	}

	c.JSON(200, gin.H{
		"code": 0,
		"data": gin.H{
			"total_repos":      totalRepos,
			"cloned_count":     clonedCount,
			"not_cloned_count": notClonedCount,
			"author_count":     authorCount,
		},
		"message": "success",
	})
}

// addActivityRecord 添加活动记录的辅助函数
func addActivityRecord(activityType, title, description string, repoId int64, repoName string) {
	_, err := common.Db.Exec(`
		INSERT INTO activities (type, title, description, repo_id, repo_name, created_at)
		VALUES (?, ?, ?, ?, ?, datetime('now', 'localtime'))
	`, activityType, title, description, repoId, repoName)
	if err != nil {
		log.Printf("添加活动记录失败: %v", err)
	}
}

// getActivities 获取活动记录列表
func getActivities(c *gin.Context) {
	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if val, err := strconv.Atoi(ps); err == nil && val > 0 && val <= 100 {
			pageSize = val
		}
	}

	offset := (page - 1) * pageSize

	// 获取总数
	var total int
	err := common.Db.QueryRow("SELECT COUNT(*) FROM activities").Scan(&total)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "获取活动记录失败"})
		return
	}

	// 获取列表
	rows, err := common.Db.Query(`
		SELECT id, type, title, description, repo_id, repo_name, metadata, created_at
		FROM activities
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, pageSize, offset)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "查询活动记录失败"})
		return
	}
	defer rows.Close()

	activities := []map[string]interface{}{}
	for rows.Next() {
		var id int64
		var activityType, title, createdAt string
		var description, repoName, metadata sql.NullString
		var repoId sql.NullInt64

		err := rows.Scan(&id, &activityType, &title, &description, &repoId, &repoName, &metadata, &createdAt)
		if err != nil {
			log.Printf("扫描活动记录失败: %v", err)
			continue
		}

		activity := map[string]interface{}{
			"id":         id,
			"type":       activityType,
			"title":      title,
			"created_at": createdAt,
		}
		if description.Valid {
			activity["description"] = description.String
		} else {
			activity["description"] = ""
		}
		if repoId.Valid {
			activity["repo_id"] = repoId.Int64
		} else {
			activity["repo_id"] = 0
		}
		if repoName.Valid {
			activity["repo_name"] = repoName.String
		} else {
			activity["repo_name"] = ""
		}
		if metadata.Valid {
			activity["metadata"] = metadata.String
		}

		activities = append(activities, activity)
	}

	c.JSON(200, gin.H{
		"code": 0,
		"data": gin.H{
			"list":       activities,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
			"total_page": (total + pageSize - 1) / pageSize,
		},
		"message": "success",
	})
}

// addActivity 添加活动记录
func addActivity(c *gin.Context) {
	var req struct {
		Type        string `json:"type" binding:"required"`
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		RepoID      int    `json:"repo_id"`
		RepoName    string `json:"repo_name"`
		Metadata    string `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 验证类型
	validTypes := map[string]bool{
		"success": true,
		"info":    true,
		"warning": true,
		"error":   true,
	}
	if !validTypes[req.Type] {
		c.JSON(400, gin.H{"code": 400, "message": "无效的活动类型"})
		return
	}

	result, err := common.Db.Exec(`
		INSERT INTO activities (type, title, description, repo_id, repo_name, metadata, created_at)
		VALUES (?, ?, ?, ?, ?, ?, datetime('now', 'localtime'))
	`, req.Type, req.Title, req.Description, req.RepoID, req.RepoName, req.Metadata)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "添加活动记录失败"})
		return
	}

	id, _ := result.LastInsertId()

	c.JSON(200, gin.H{
		"code": 0,
		"data": gin.H{
			"id": id,
		},
		"message": "success",
	})
}

// clearActivities 清空所有活动记录
func clearActivities(c *gin.Context) {
	_, err := common.Db.Exec("DELETE FROM activities")
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "清空活动记录失败"})
		return
	}
	c.JSON(200, gin.H{"code": 0, "message": "活动记录已清空"})
}

// getSystemInfo 获取系统信息
func getSystemInfo(c *gin.Context) {
	config, err := utils.ConfigInstance.ReadConfig()
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "读取配置失败"})
		return
	}

	c.JSON(200, gin.H{
		"code": 0,
		"data": gin.H{
			"store_root_path": config.StoreRootPath,
			"version":         "1.0.0",
		},
		"message": "success",
	})
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

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start HTTP server with web frontend",
	Long:  `Start HTTP server to provide web interface and API services for GitPlus`,
	Run: func(cmd *cobra.Command, args []string) {
		// 环境变量覆盖端口和地址
		if envPort := os.Getenv("FORKS_PORT"); envPort != "" {
			if p, err := strconv.Atoi(envPort); err == nil {
				port = p
			}
		}
		if envAddr := os.Getenv("FORKS_ADDRESS"); envAddr != "" {
			address = envAddr
		}

		// 配置代理（先执行命令行配置）
		setupProxy()

		// 初始化运行时代理配置
		initRuntimeProxyConfig()

		// 使用自定义token或生成随机token
		// 优先级：命令行参数 > 配置文件 > 随机生成
		if customToken != "" {
			token = customToken
		} else {
			savedToken := loadTokenFromFile()
			if savedToken != "" {
				token = savedToken
			} else {
				token = generateToken()
			}
		}

		// 设置Gin模式
		gin.SetMode(gin.ReleaseMode)

		// 创建Gin引擎
		r := gin.Default()

		// 设置路由
		setupRoutes(r)

		// 打印访问信息
		url := fmt.Sprintf("http://%s:%d?token=%s", address, port, token)
		fmt.Printf("GitPlus Web服务已启动\n")
		fmt.Printf("访问地址: %s\n", url)
		if customToken != "" {
			fmt.Printf("API Token: %s (自定义)\n", token)
		} else if loadTokenFromFile() != "" {
			fmt.Printf("API Token: %s (配置文件)\n", token)
		} else {
			fmt.Printf("API Token: %s (随机生成)\n", token)
		}
		fmt.Printf("按 Ctrl+C 停止服务\n\n")

		// 启动服务器
		if err := r.Run(address + ":" + strconv.Itoa(port)); err != nil {
			fmt.Printf("启动服务器失败: %v\n", err)
		}
	},
}

// FileNode 文件树节点结构
type FileNode struct {
	ID          string      `json:"id"`
	Key         string      `json:"key"`
	FileName    string      `json:"fileName"`
	FilePath    string      `json:"filePath"`
	IsDirectory int         `json:"isDirectory"`
	Children    []*FileNode `json:"children,omitempty"`
}

// getRepoFiles 获取仓库文件结构
func getRepoFiles(c *gin.Context) {
	id := c.Param("id")

	// 获取仓库信息
	var repo models.GitRepoInfo
	err := common.Db.QueryRow("SELECT author, repo, source FROM repos WHERE id = ?", id).
		Scan(&repo.Author, &repo.Repo, &repo.Source)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "仓库不存在"})
		return
	}

	// 读取配置获取存储路径
	config, err := utils.ConfigInstance.ReadConfig()
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "读取配置失败"})
		return
	}

	// 构建仓库本地路径，包含平台标识
	repoPath := filepath.Join(config.StoreRootPath, repo.Source, repo.Author, repo.Repo)

	// 检查路径是否存在
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		c.JSON(200, gin.H{"code": 0, "data": gin.H{"tree": []*FileNode{}}, "message": fmt.Sprintf("仓库文件夹不存在: %s", repoPath)})
		return
	}

	// 构建文件树
	fileTree, err := buildFileTree(repoPath, "")
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "读取文件结构失败"})
		return
	}

	c.JSON(200, gin.H{
		"code": 0,
		"data": gin.H{
			"tree": fileTree,
		},
		"message": "success",
	})
}

// buildFileTree 递归构建文件树
func buildFileTree(basePath, relativePath string) ([]*FileNode, error) {
	var nodes []*FileNode

	fullPath := filepath.Join(basePath, relativePath)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		// 跳过隐藏文件和git目录
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		entryPath := filepath.Join(relativePath, entry.Name())
		nodeID := strings.ReplaceAll(entryPath, string(os.PathSeparator), "/")
		if nodeID == "" {
			nodeID = entry.Name()
		}

		node := &FileNode{
			ID:       nodeID,
			Key:      nodeID,
			FileName: entry.Name(),
			FilePath: entryPath,
		}

		if entry.IsDir() {
			node.IsDirectory = 1
			// 递归获取子目录
			children, err := buildFileTree(basePath, entryPath)
			if err == nil {
				node.Children = children
			}
		} else {
			node.IsDirectory = 0
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// getFileContent 获取文件内容
func getFileContent(c *gin.Context) {
	id := c.Param("id")
	filePath := c.Query("path")

	if filePath == "" {
		c.JSON(400, gin.H{"code": 400, "message": "文件路径参数缺失"})
		return
	}

	// 获取仓库信息
	var repo models.GitRepoInfo
	err := common.Db.QueryRow("SELECT author, repo, source FROM repos WHERE id = ?", id).
		Scan(&repo.Author, &repo.Repo, &repo.Source)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "仓库不存在"})
		return
	}

	// 读取配置获取存储路径
	config, err := utils.ConfigInstance.ReadConfig()
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "读取配置失败"})
		return
	}

	// 构建完整文件路径，包含平台标识
	fullFilePath := filepath.Join(config.StoreRootPath, repo.Source, repo.Author, repo.Repo, filePath)

	// 安全检查：确保文件在仓库目录内
	repoPath := filepath.Join(config.StoreRootPath, repo.Source, repo.Author, repo.Repo)
	absRepoPath, err := filepath.Abs(repoPath)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "路径解析失败"})
		return
	}
	absFilePath, err := filepath.Abs(fullFilePath)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "路径解析失败"})
		return
	}

	relPath, err := filepath.Rel(absRepoPath, absFilePath)
	if err != nil || strings.HasPrefix(relPath, "..") || filepath.IsAbs(relPath) {
		c.JSON(403, gin.H{"code": 403, "message": "访问被拒绝"})
		return
	}

	// 读取文件内容
	content, err := os.ReadFile(fullFilePath)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "文件不存在或读取失败"})
		return
	}

	// 检查是否请求二进制内容（用于图片等）
	contentType := c.Query("type")
	if contentType == "blob" {
		// 检测文件类型并设置适当的Content-Type
		ext := strings.ToLower(filepath.Ext(fullFilePath))
		var mimeType string
		switch ext {
		// 图片类型
		case ".jpg", ".jpeg":
			mimeType = "image/jpeg"
		case ".png":
			mimeType = "image/png"
		case ".gif":
			mimeType = "image/gif"
		case ".svg":
			mimeType = "image/svg+xml"
		case ".bmp":
			mimeType = "image/bmp"
		case ".webp":
			mimeType = "image/webp"
		case ".ico":
			mimeType = "image/x-icon"
		// 音频类型
		case ".mp3":
			mimeType = "audio/mpeg"
		case ".wav":
			mimeType = "audio/wav"
		case ".ogg":
			mimeType = "audio/ogg"
		case ".aac":
			mimeType = "audio/aac"
		case ".flac":
			mimeType = "audio/flac"
		case ".m4a":
			mimeType = "audio/mp4"
		case ".wma":
			mimeType = "audio/x-ms-wma"
		// 视频类型
		case ".mp4":
			mimeType = "video/mp4"
		case ".webm":
			mimeType = "video/webm"
		case ".avi":
			mimeType = "video/x-msvideo"
		case ".mov":
			mimeType = "video/quicktime"
		case ".mkv":
			mimeType = "video/x-matroska"
		case ".wmv":
			mimeType = "video/x-ms-wmv"
		case ".flv":
			mimeType = "video/x-flv"
		case ".m4v":
			mimeType = "video/mp4"
		default:
			mimeType = "application/octet-stream"
		}

		c.Header("Content-Type", mimeType)
		c.Header("Content-Length", fmt.Sprintf("%d", len(content)))
		c.Data(200, mimeType, content)
		return
	}

	c.JSON(200, gin.H{
		"code": 0,
		"data": gin.H{
			"content": string(content),
			"path":    filePath,
		},
		"message": "success",
	})
}

// cloneRepo 克隆仓库
func cloneRepo(c *gin.Context) {
	id := c.Param("id")

	// 获取仓库信息
	var repo models.GitRepoInfo
	err := common.Db.QueryRow("SELECT author, repo, url, source FROM repos WHERE id = ?", id).
		Scan(&repo.Author, &repo.Repo, &repo.URL, &repo.Source)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "仓库不存在"})
		return
	}

	// 读取配置
	config, err := utils.ConfigInstance.ReadConfig()
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "读取配置失败"})
		return
	}

	// 构建仓库路径
	repoPath := filepath.Join(config.StoreRootPath, repo.Source, repo.Author, repo.Repo)

	// 检查本地仓库是否已存在
	if _, err := os.Stat(repoPath); err == nil {
		// 检查是否是有效的git仓库
		if _, err := os.Stat(filepath.Join(repoPath, ".git")); err == nil {
			// 更新数据库中的克隆状态
			common.Db.Exec("UPDATE repos SET is_cloned = 1 WHERE id = ?", id)
			c.JSON(200, gin.H{
				"code":    0,
				"message": "本地仓库已存在",
				"data": gin.H{
					"useSSE":    false,
					"exists":    true,
				},
			})
			return
		}
	}

	// 生成临时token并存储在内存中（实际项目中可以使用Redis）
	tempToken := generateToken()
	tempTokenMutex.Lock()
	tempTokenStore[id+"_clone"] = tempToken
	tempTokenMutex.Unlock()

	c.JSON(200, gin.H{
		"code":    0,
		"message": "开始克隆仓库，请查看实时状态",
		"data": gin.H{
			"useSSE":    true,
			"tempToken": tempToken,
		},
	})
}

// pullRepo 拉取更新
func pullRepo(c *gin.Context) {
	id := c.Param("id")
	repoPath, err := getRepoPath(id)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}

	// 检查仓库是否存在
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		c.JSON(500, gin.H{"code": 500, "message": "本地仓库不存在，请先克隆"})
		return
	}

	// 生成临时token并存储
	tempToken := generateToken()
	tempTokenMutex.Lock()
	tempTokenStore[id+"_pull"] = tempToken
	tempTokenMutex.Unlock()

	c.JSON(200, gin.H{
		"code":    0,
		"message": "开始拉取更新，请查看实时状态",
		"data": gin.H{
			"useSSE":    true,
			"tempToken": tempToken,
		},
	})
}

// getRepoStatus 获取仓库状态
func getRepoStatus(c *gin.Context) {
	id := c.Param("id")
	repoPath, err := getRepoPath(id)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}

	status := gin.H{
		"exists":   false,
		"status":   "未克隆",
		"repoPath": repoPath,
	}

	// 检查仓库是否存在
	if _, err := os.Stat(repoPath); err == nil {
		status["exists"] = true
		status["status"] = "已克隆"

		// 获取当前分支
		cmd := exec.Command("git", "branch", "--show-current")
		cmd.Dir = repoPath
		if output, err := cmd.Output(); err == nil {
			status["branch"] = strings.TrimSpace(string(output))
		}

		// 获取最后提交时间
		cmd = exec.Command("git", "log", "-1", "--format=%ci")
		cmd.Dir = repoPath
		if output, err := cmd.Output(); err == nil {
			status["lastUpdate"] = strings.TrimSpace(string(output))
		}

		// 获取最后提交哈希和消息
		cmd = exec.Command("git", "log", "-1", "--format=%H")
		cmd.Dir = repoPath
		if output, err := cmd.Output(); err == nil {
			hash := strings.TrimSpace(string(output))
			if len(hash) >= 8 {
				status["lastCommitHash"] = hash[:8] // 短哈希
			} else if len(hash) > 0 {
				status["lastCommitHash"] = hash
			}
		}

		cmd = exec.Command("git", "log", "-1", "--format=%s")
		cmd.Dir = repoPath
		if output, err := cmd.Output(); err == nil {
			status["lastCommitMessage"] = strings.TrimSpace(string(output))
		}

		// 获取远程URL
		cmd = exec.Command("git", "remote", "get-url", "origin")
		cmd.Dir = repoPath
		if output, err := cmd.Output(); err == nil {
			status["remoteUrl"] = strings.TrimSpace(string(output))
		}

		// 检查工作区状态
		cmd = exec.Command("git", "status", "--porcelain")
		cmd.Dir = repoPath
		if output, err := cmd.Output(); err == nil {
			changes := strings.TrimSpace(string(output))
			status["hasChanges"] = len(changes) > 0
			if len(changes) > 0 {
				status["changes"] = strings.Split(changes, "\n")
			}
		}

		// 检查是否有未推送的提交
		cmd = exec.Command("git", "log", "origin/HEAD..HEAD", "--oneline")
		cmd.Dir = repoPath
		if output, err := cmd.Output(); err == nil {
			unpushed := strings.TrimSpace(string(output))
			status["hasUnpushed"] = len(unpushed) > 0
			if len(unpushed) > 0 {
				status["unpushedCount"] = len(strings.Split(unpushed, "\n"))
			}
		}

		// 检查是否有未拉取的提交
		logGitCmd("fetch", "origin")
		exec.Command("git", "fetch", "origin").Run() // 先获取最新信息
		cmd = exec.Command("git", "log", "HEAD..origin/HEAD", "--oneline")
		cmd.Dir = repoPath
		if output, err := cmd.Output(); err == nil {
			behind := strings.TrimSpace(string(output))
			status["hasBehind"] = len(behind) > 0
			if len(behind) > 0 {
				status["behindCount"] = len(strings.Split(behind, "\n"))
			}
		}

		// 统计文件数量
		if fileCount, err := countFiles(repoPath); err == nil {
			status["fileCount"] = fileCount
		}

		// 获取仓库大小
		if size, err := getDirSize(repoPath); err == nil {
			status["repoSize"] = formatBytes(size)
		}

		// 获取所有分支
		cmd = exec.Command("git", "branch", "-a")
		cmd.Dir = repoPath
		if output, err := cmd.Output(); err == nil {
			branches := []string{}
			for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
				branch := strings.TrimSpace(line)
				if branch != "" {
					branches = append(branches, branch)
				}
			}
			status["allBranches"] = branches
		}
	}

	c.JSON(200, gin.H{"code": 0, "data": status, "message": "success"})
}

// getRepoDiff 获取仓库差异
func getRepoDiff(c *gin.Context) {
	id := c.Param("id")
	repoPath, err := getRepoPath(id)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}

	// 检查仓库是否存在
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		c.JSON(500, gin.H{"code": 500, "message": "本地仓库不存在"})
		return
	}

	// 获取远程更新
	logGitCmd("fetch")
	cmd := exec.Command("git", "fetch")
	cmd.Dir = repoPath
	cmd.Run()

	// 获取差异
	logGitCmd("log", "HEAD..origin/HEAD", "--oneline")
	cmd = exec.Command("git", "log", "HEAD..origin/HEAD", "--oneline")
	cmd.Dir = repoPath
	output, err := cmd.Output()

	diff := gin.H{
		"behind":     strings.TrimSpace(string(output)),
		"hasChanges": len(strings.TrimSpace(string(output))) > 0,
	}

	c.JSON(200, gin.H{"code": 0, "data": diff, "message": "success"})
}

// getRepoCommits 获取提交历史
func getRepoCommits(c *gin.Context) {
	id := c.Param("id")
	repoPath, err := getRepoPath(id)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}

	// 检查仓库是否存在
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		c.JSON(500, gin.H{"code": 500, "message": "本地仓库不存在"})
		return
	}

	// 获取提交历史
	cmd := exec.Command("git", "log", "--oneline", "-10")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "获取提交历史失败"})
		return
	}

	commits := strings.Split(strings.TrimSpace(string(output)), "\n")
	c.JSON(200, gin.H{"code": 0, "data": gin.H{"commits": commits}, "message": "success"})
}

// getRepoBranches 获取分支信息
func getRepoBranches(c *gin.Context) {
	id := c.Param("id")
	repoPath, err := getRepoPath(id)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}

	// 检查仓库是否存在
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		c.JSON(500, gin.H{"code": 500, "message": "本地仓库不存在"})
		return
	}

	// 获取分支信息
	cmd := exec.Command("git", "branch", "-a")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "获取分支信息失败"})
		return
	}

	branches := strings.Split(strings.TrimSpace(string(output)), "\n")
	c.JSON(200, gin.H{"code": 0, "data": gin.H{"branches": branches}, "message": "success"})
}

// openRepoFolder 打开仓库文件夹
func openRepoFolder(c *gin.Context) {
	id := c.Param("id")
	repoPath, err := getRepoPath(id)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}

	// 检查仓库是否存在
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		c.JSON(500, gin.H{"code": 500, "message": "本地仓库不存在"})
		return
	}

	// 根据操作系统打开文件夹
	var cmd *exec.Cmd
	switch {
	case strings.Contains(strings.ToLower(os.Getenv("OS")), "windows"):
		cmd = exec.Command("explorer", repoPath)
	case fileExists("/usr/bin/xdg-open"):
		cmd = exec.Command("xdg-open", repoPath)
	case fileExists("/usr/bin/open"):
		cmd = exec.Command("open", repoPath)
	default:
		c.JSON(500, gin.H{"code": 500, "message": "不支持的操作系统"})
		return
	}

	if err := cmd.Start(); err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "打开文件夹失败"})
		return
	}

	c.JSON(200, gin.H{"code": 0, "message": "文件夹已打开"})
}

// deleteLocalRepo 删除本地仓库
func deleteLocalRepo(c *gin.Context) {
	id := c.Param("id")
	repoPath, err := getRepoPath(id)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}

	// 检查仓库是否存在
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		c.JSON(200, gin.H{"code": 0, "message": "本地仓库不存在"})
		return
	}

	// 删除仓库文件夹
	if err := os.RemoveAll(repoPath); err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "删除本地仓库失败"})
		return
	}

	c.JSON(200, gin.H{"code": 0, "message": "本地仓库删除成功"})
}

// getRepoPath 获取仓库路径的辅助函数
func getRepoPath(id string) (string, error) {
	// 获取仓库信息
	var repo models.GitRepoInfo
	err := common.Db.QueryRow("SELECT author, repo, source FROM repos WHERE id = ?", id).
		Scan(&repo.Author, &repo.Repo, &repo.Source)
	if err != nil {
		return "", fmt.Errorf("仓库不存在")
	}

	// 读取配置
	config, err := utils.ConfigInstance.ReadConfig()
	if err != nil {
		return "", fmt.Errorf("读取配置失败")
	}

	// 构建路径
	return filepath.Join(config.StoreRootPath, repo.Source, repo.Author, repo.Repo), nil
}

// countFiles 统计文件数量的辅助函数
func countFiles(dirPath string) (int, error) {
	count := 0
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误，继续计数
		}
		// 跳过隐藏文件和.git目录
		if strings.HasPrefix(filepath.Base(path), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !info.IsDir() {
			count++
		}
		return nil
	})
	return count, err
}

// fileExists 检查文件是否存在的辅助函数
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// getDirSize 获取目录大小
func getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// formatBytes 格式化字节数
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// cloneRepoSSE 克隆仓库的SSE接口
func cloneRepoSSE(c *gin.Context) {
	id := c.Param("id")
	log.Printf("🔍 [cloneRepoSSE] 开始克隆, id=%s", id)

	// 设置SSE头部
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 获取仓库信息
	var repo models.GitRepoInfo
	err := common.Db.QueryRow("SELECT author, repo, url, source FROM repos WHERE id = ?", id).
		Scan(&repo.Author, &repo.Repo, &repo.URL, &repo.Source)
	if err != nil {
		log.Printf("❌ [cloneRepoSSE] 仓库不存在, id=%s", id)
		c.SSEvent("error", gin.H{"message": "仓库不存在"})
		c.Writer.Flush()
		return
	}

	// 读取配置
	config, err := utils.ConfigInstance.ReadConfig()
	if err != nil {
		c.SSEvent("error", gin.H{"message": "读取配置失败"})
		c.Writer.Flush()
		return
	}

	// 构建仓库路径
	repoPath := filepath.Join(config.StoreRootPath, repo.Source, repo.Author, repo.Repo)

	// 发送开始信息
	c.SSEvent("start", gin.H{"message": "开始克隆仓库..."})
	c.Writer.Flush()

	// 检查并创建父目录
	parentDir := filepath.Dir(repoPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		c.SSEvent("error", gin.H{"message": "创建目录失败"})
		c.Writer.Flush()
		return
	}

	// 如果目录已存在，先删除
	if _, err := os.Stat(repoPath); err == nil {
		c.SSEvent("progress", gin.H{"message": "清理已存在的目录..."})
		c.Writer.Flush()
		os.RemoveAll(repoPath)
	}

	// 执行git clone
	c.SSEvent("progress", gin.H{"message": "正在克隆仓库，请耐心等待..."})
	c.Writer.Flush()

	// 按平台动态设置代理
	_, restoreProxy := applyGitProxyForPlatform(repo.Source)
	defer restoreProxy()

	logGitCmd("clone", "--progress", repo.URL, repoPath)
	cmd := exec.Command("git", "clone", "--progress", repo.URL, repoPath)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		c.SSEvent("error", gin.H{"message": "创建命令失败"})
		c.Writer.Flush()
		return
	}

	if err := cmd.Start(); err != nil {
		c.SSEvent("error", gin.H{"message": "启动克隆失败: " + err.Error()})
		c.Writer.Flush()
		return
	}

	// 读取git输出
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			c.SSEvent("progress", gin.H{"message": line})
			c.Writer.Flush()
		}
	}

	if err := cmd.Wait(); err != nil {
		c.SSEvent("error", gin.H{"message": "克隆失败: " + err.Error()})
		c.Writer.Flush()
		return
	}

	// 更新数据库中的克隆状态
	_, err = common.Db.Exec("UPDATE repos SET is_cloned = 1 WHERE id = ?", id)
	if err != nil {
		log.Printf("⚠️ [cloneRepoSSE] 更新克隆状态失败: %v", err)
	}

	// 记录活动
	repoFullName := repo.Author + "/" + repo.Repo
	repoIdInt, _ := strconv.ParseInt(id, 10, 64)
	addActivityRecord("success", "克隆成功", fmt.Sprintf("成功克隆仓库 %s", repoFullName), repoIdInt, repoFullName)

	log.Printf("✅ [cloneRepoSSE] 克隆成功, id=%s", id)
	c.SSEvent("complete", gin.H{"message": "仓库克隆成功!"})
	c.Writer.Flush()
}

// batchCloneRepos 批量克隆仓库（一键克隆）
func batchCloneRepos(c *gin.Context) {
	log.Printf("🚀 [batchCloneRepos] 开始批量克隆")

	// 解析可选的 ID 列表
	var reqBody struct {
		IDs []int `json:"ids"`
	}
	c.ShouldBindJSON(&reqBody)

	var query string
	var args []interface{}

	if len(reqBody.IDs) > 0 {
		// 只克隆指定 ID 的未克隆仓库
		placeholders := ""
		for i, id := range reqBody.IDs {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args = append(args, id)
		}
		query = fmt.Sprintf("SELECT id, author, repo, url, source FROM repos WHERE id IN (%s) AND (is_cloned = 0 OR is_cloned IS NULL) ORDER BY id", placeholders)
	} else {
		query = "SELECT id, author, repo, url, source FROM repos WHERE is_cloned = 0 OR is_cloned IS NULL ORDER BY id"
	}

	// 查询未克隆的仓库
	rows, err := common.Db.Query(query, args...)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "查询数据库失败"})
		return
	}
	defer rows.Close()

	type RepoInfo struct {
		ID     int
		Author string
		Repo   string
		URL    string
		Source string
	}

	var repos []RepoInfo
	var repoIDs []int
	for rows.Next() {
		var r RepoInfo
		if err := rows.Scan(&r.ID, &r.Author, &r.Repo, &r.URL, &r.Source); err == nil {
			repos = append(repos, r)
			repoIDs = append(repoIDs, r.ID)
		}
	}

	if len(repos) == 0 {
		c.JSON(200, gin.H{"code": 0, "message": "没有需要克隆的仓库", "data": gin.H{"total": 0, "cloned": 0, "failed": 0, "invalid": 0}})
		return
	}

	log.Printf("📋 [batchCloneRepos] 找到 %d 个未克隆仓库", len(repos))

	// 生成临时token
	tempToken := generateToken()
	tempTokenMutex.Lock()
	tempTokenStore["batch_clone"] = tempToken
	tempTokenMutex.Unlock()

	// 将选中的 ID 列表存入内存，供 batchCloneSSE 读取
	batchCloneIDsMutex.Lock()
	batchCloneIDsStore[tempToken] = repoIDs
	batchCloneIDsMutex.Unlock()

	c.JSON(200, gin.H{
		"code":    0,
		"message": "开始批量克隆",
		"data": gin.H{
			"useSSE":    true,
			"tempToken": tempToken,
			"total":     len(repos),
		},
	})
}

// batchCloneSSE 批量克隆进度推送
func batchCloneSSE(c *gin.Context) {
	log.Printf("🚀 [batchCloneSSE] 客户端连接")

	// 设置SSE头部
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 读取配置
	config, err := utils.ConfigInstance.ReadConfig()
	if err != nil {
		c.SSEvent("error", gin.H{"message": "读取配置失败"})
		c.Writer.Flush()
		return
	}

	// 从内存中读取 ID 列表
	tempToken := c.Query("tempToken")
	batchCloneIDsMutex.RLock()
	selectedIDs, hasIDs := batchCloneIDsStore[tempToken]
	batchCloneIDsMutex.RUnlock()

	var query string
	var args []interface{}

	if hasIDs && len(selectedIDs) > 0 {
		// 使用指定的 ID 列表
		placeholders := ""
		for i, id := range selectedIDs {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args = append(args, id)
		}
		query = fmt.Sprintf("SELECT id, author, repo, url, source FROM repos WHERE id IN (%s) AND (is_cloned = 0 OR is_cloned IS NULL) ORDER BY id", placeholders)

		// 清理内存
		batchCloneIDsMutex.Lock()
		delete(batchCloneIDsStore, tempToken)
		batchCloneIDsMutex.Unlock()
	} else {
		query = "SELECT id, author, repo, url, source FROM repos WHERE is_cloned = 0 OR is_cloned IS NULL ORDER BY id"
	}

	// 查询未克隆的仓库
	rows, err := common.Db.Query(query, args...)
	if err != nil {
		c.SSEvent("error", gin.H{"message": "查询数据库失败"})
		c.Writer.Flush()
		return
	}
	defer rows.Close()

	type RepoInfo struct {
		ID     int
		Author string
		Repo   string
		URL    string
		Source string
	}

	var repos []RepoInfo
	for rows.Next() {
		var r RepoInfo
		if err := rows.Scan(&r.ID, &r.Author, &r.Repo, &r.URL, &r.Source); err == nil {
			repos = append(repos, r)
		}
	}

	total := len(repos)
	if total == 0 {
		c.SSEvent("complete", gin.H{"message": "没有需要克隆的仓库", "total": 0, "cloned": 0, "failed": 0, "invalid": 0})
		c.Writer.Flush()
		return
	}

	// 发送开始事件
	c.SSEvent("start", gin.H{"message": "开始批量克隆", "total": total})
	c.Writer.Flush()

	clonedCount := 0
	failedCount := 0
	invalidCount := 0

	for i, repo := range repos {
		// 检查客户端是否断开
		select {
		case <-c.Request.Context().Done():
			log.Printf("🔌 [batchCloneSSE] 客户端断开连接")
			return
		default:
		}

		repoPath := filepath.Join(config.StoreRootPath, repo.Source, repo.Author, repo.Repo)
		repoName := fmt.Sprintf("%s/%s", repo.Author, repo.Repo)

		// 发送进度
		c.SSEvent("progress", gin.H{
			"current": i + 1,
			"total":   total,
			"repo":    repoName,
			"status":  "checking",
			"message": fmt.Sprintf("[%d/%d] 检查 %s...", i+1, total, repoName),
		})
		c.Writer.Flush()

		// 检查URL是否有效（通过git ls-remote）
		_, restoreCheckProxy := applyGitProxyForPlatform(repo.Source)
		logGitCmd("ls-remote", "--exit-code", repo.URL)
		checkCmd := exec.Command("git", "ls-remote", "--exit-code", repo.URL)
		checkErr := checkCmd.Run()
		restoreCheckProxy()

		if checkErr != nil {
			// URL无效，标记为失效
			invalidCount++
			common.Db.Exec("UPDATE repos SET is_cloned = -1 WHERE id = ?", repo.ID) // -1 表示失效

			c.SSEvent("progress", gin.H{
				"current": i + 1,
				"total":   total,
				"repo":    repoName,
				"status":  "invalid",
				"message": fmt.Sprintf("[%d/%d] %s - 仓库不存在或无法访问", i+1, total, repoName),
			})
			c.Writer.Flush()
			continue
		}

		// 检查是否已存在
		if _, err := os.Stat(repoPath); err == nil {
			// 目录已存在，更新状态
			common.Db.Exec("UPDATE repos SET is_cloned = 1 WHERE id = ?", repo.ID)
			clonedCount++

			c.SSEvent("progress", gin.H{
				"current": i + 1,
				"total":   total,
				"repo":    repoName,
				"status":  "exists",
				"message": fmt.Sprintf("[%d/%d] %s - 已存在", i+1, total, repoName),
			})
			c.Writer.Flush()
			continue
		}

		// 创建目录
		parentDir := filepath.Dir(repoPath)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			failedCount++
			c.SSEvent("progress", gin.H{
				"current": i + 1,
				"total":   total,
				"repo":    repoName,
				"status":  "error",
				"message": fmt.Sprintf("[%d/%d] %s - 创建目录失败", i+1, total, repoName),
			})
			c.Writer.Flush()
			continue
		}

		// 发送克隆开始
		c.SSEvent("progress", gin.H{
			"current": i + 1,
			"total":   total,
			"repo":    repoName,
			"status":  "cloning",
			"message": fmt.Sprintf("[%d/%d] 正在克隆 %s...", i+1, total, repoName),
		})
		c.Writer.Flush()

		// 执行克隆
		_, restoreCloneProxy := applyGitProxyForPlatform(repo.Source)
		logGitCmd("clone", "--depth", "1", repo.URL, repoPath)
		cloneCmd := exec.Command("git", "clone", "--depth", "1", repo.URL, repoPath)
		cloneErr := cloneCmd.Run()
		restoreCloneProxy()

		if cloneErr != nil {
			failedCount++
			c.SSEvent("progress", gin.H{
				"current": i + 1,
				"total":   total,
				"repo":    repoName,
				"status":  "failed",
				"message": fmt.Sprintf("[%d/%d] %s - 克隆失败: %v", i+1, total, repoName, cloneErr),
			})
			c.Writer.Flush()
			continue
		}

		// 更新状态
		common.Db.Exec("UPDATE repos SET is_cloned = 1 WHERE id = ?", repo.ID)
		clonedCount++

		c.SSEvent("progress", gin.H{
			"current": i + 1,
			"total":   total,
			"repo":    repoName,
			"status":  "success",
			"message": fmt.Sprintf("[%d/%d] %s - 克隆成功 ✓", i+1, total, repoName),
		})
		c.Writer.Flush()
	}

	// 发送完成事件
	c.SSEvent("complete", gin.H{
		"message": "批量克隆完成",
		"total":   total,
		"cloned":  clonedCount,
		"failed":  failedCount,
		"invalid": invalidCount,
	})
	c.Writer.Flush()

	// 记录活动
	addActivityRecord("success", "批量克隆完成", fmt.Sprintf("成功 %d 个, 失败 %d 个, 失效 %d 个", clonedCount, failedCount, invalidCount), 0, "")

	log.Printf("✅ [batchCloneSSE] 批量克隆完成: 总数=%d, 成功=%d, 失败=%d, 失效=%d", total, clonedCount, failedCount, invalidCount)
}

// pullRepoSSE 拉取更新的SSE接口
func pullRepoSSE(c *gin.Context) {
	id := c.Param("id")

	// 设置SSE头部
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	repoPath, err := getRepoPath(id)
	if err != nil {
		c.SSEvent("error", gin.H{"message": err.Error()})
		return
	}

	// 获取仓库 source 用于代理判断
	var repoSource string
	common.Db.QueryRow("SELECT source FROM repos WHERE id = ?", id).Scan(&repoSource)

	// 检查仓库是否存在
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		c.SSEvent("error", gin.H{"message": "本地仓库不存在，请先克隆"})
		return
	}

	// 发送开始信息
	c.SSEvent("start", gin.H{"message": "开始拉取更新..."})
	c.Writer.Flush()

	// 按平台动态设置代理
	_, restorePullProxy := applyGitProxyForPlatform(repoSource)
	defer restorePullProxy()

	// 执行git pull
	logGitCmd("pull", "--progress")
	cmd := exec.Command("git", "pull", "--progress")
	cmd.Dir = repoPath

	// 获取stdout和stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		c.SSEvent("error", gin.H{"message": "创建命令失败"})
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		c.SSEvent("error", gin.H{"message": "创建命令失败"})
		return
	}

	if err := cmd.Start(); err != nil {
		c.SSEvent("error", gin.H{"message": "启动拉取失败: " + err.Error()})
		return
	}

	// 创建goroutine读取输出
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				c.SSEvent("progress", gin.H{"message": line})
				c.Writer.Flush()
			}
		}
	}()

	// 读取stderr
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			c.SSEvent("progress", gin.H{"message": line})
			c.Writer.Flush()
		}
	}

	if err := cmd.Wait(); err != nil {
		c.SSEvent("error", gin.H{"message": "拉取失败: " + err.Error()})
		return
	}

	c.SSEvent("complete", gin.H{"message": "拉取更新成功!"})
	c.Writer.Flush()
}

// resetRepo 重置仓库到远程最新状态
func resetRepo(c *gin.Context) {
	id := c.Param("id")
	repoPath, err := getRepoPath(id)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": err.Error()})
		return
	}

	// 检查仓库是否存在
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		c.JSON(500, gin.H{"code": 500, "message": "本地仓库不存在，请先克隆"})
		return
	}

	// 生成临时token并存储
	tempToken := generateToken()
	tempTokenMutex.Lock()
	tempTokenStore[id+"_reset"] = tempToken
	tempTokenMutex.Unlock()

	c.JSON(200, gin.H{
		"code":    0,
		"message": "开始重置仓库，请查看实时状态",
		"data": gin.H{
			"useSSE":    true,
			"tempToken": tempToken,
		},
	})
}

// resetRepoSSE 重置仓库的SSE实现
func resetRepoSSE(c *gin.Context) {
	id := c.Param("id")

	// 设置SSE头部
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	repoPath, err := getRepoPath(id)
	if err != nil {
		c.SSEvent("error", gin.H{"message": err.Error()})
		return
	}

	// 获取仓库 source 用于代理判断
	var repoSource string
	common.Db.QueryRow("SELECT source FROM repos WHERE id = ?", id).Scan(&repoSource)

	// 检查仓库是否存在
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		c.SSEvent("error", gin.H{"message": "本地仓库不存在，请先克隆"})
		return
	}

	// 发送开始信息
	c.SSEvent("start", gin.H{"message": "开始重置仓库到远程最新状态..."})
	c.Writer.Flush()

	// 按平台动态设置代理
	_, restoreResetProxy := applyGitProxyForPlatform(repoSource)
	defer restoreResetProxy()

	// 1. 获取当前分支名
	logGitCmd("branch", "--show-current")
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = repoPath
	branchOutput, err := cmd.Output()
	if err != nil {
		c.SSEvent("error", gin.H{"message": "获取当前分支失败: " + err.Error()})
		return
	}
	currentBranch := strings.TrimSpace(string(branchOutput))

	c.SSEvent("progress", gin.H{"message": fmt.Sprintf("当前分支: %s", currentBranch)})
	c.Writer.Flush()

	// 2. 获取最新的远程信息
	c.SSEvent("progress", gin.H{"message": "正在获取远程更新..."})
	c.Writer.Flush()

	cmd = exec.Command("git", "fetch", "origin")
	cmd.Dir = repoPath
	logGitCmd("fetch", "origin")
	if output, err := cmd.CombinedOutput(); err != nil {
		errMsg := strings.TrimSpace(string(output))
		if errMsg == "" {
			errMsg = err.Error()
		}
		c.SSEvent("error", gin.H{"message": "获取远程更新失败: " + errMsg})
		return
	} else {
		if len(output) > 0 {
			c.SSEvent("progress", gin.H{"message": string(output)})
			c.Writer.Flush()
		}
	}

	// 3. 清理未跟踪的文件
	c.SSEvent("progress", gin.H{"message": "清理未跟踪的文件和目录..."})
	c.Writer.Flush()

	cmd = exec.Command("git", "clean", "-fd")
	cmd.Dir = repoPath
	logGitCmd("clean", "-fd")
	if output, err := cmd.CombinedOutput(); err != nil {
		c.SSEvent("progress", gin.H{"message": "清理文件警告: " + err.Error()})
		c.Writer.Flush()
	} else {
		if len(output) > 0 {
			c.SSEvent("progress", gin.H{"message": string(output)})
		} else {
			c.SSEvent("progress", gin.H{"message": "没有需要清理的文件"})
		}
		c.Writer.Flush()
	}

	// 4. 重置到远程分支的最新状态
	c.SSEvent("progress", gin.H{"message": fmt.Sprintf("重置到 origin/%s 最新状态...", currentBranch)})
	c.Writer.Flush()

	cmd = exec.Command("git", "reset", "--hard", fmt.Sprintf("origin/%s", currentBranch))
	cmd.Dir = repoPath
	logGitCmd("reset", "--hard", fmt.Sprintf("origin/%s", currentBranch))
	if output, err := cmd.CombinedOutput(); err != nil {
		c.SSEvent("error", gin.H{"message": "重置失败: " + err.Error()})
		return
	} else {
		c.SSEvent("progress", gin.H{"message": string(output)})
		c.Writer.Flush()
	}

	// 5. 检查最终状态
	c.SSEvent("progress", gin.H{"message": "验证重置结果..."})
	c.Writer.Flush()

	cmd = exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoPath
	statusOutput, err := cmd.Output()
	if err != nil {
		c.SSEvent("progress", gin.H{"message": "检查状态时出现问题，但重置可能已成功"})
	} else {
		if len(strings.TrimSpace(string(statusOutput))) == 0 {
			c.SSEvent("progress", gin.H{"message": "✅ 工作区干净，重置成功"})
		} else {
			c.SSEvent("progress", gin.H{"message": "⚠️ 仍有未提交的更改"})
		}
	}
	c.Writer.Flush()

	c.SSEvent("complete", gin.H{"message": "🔄 仓库已重置到远程最新状态!"})
	c.Writer.Flush()
}

// scanRepos 扫描本地文件系统中的Git仓库（异步）
func scanRepos(c *gin.Context) {
	log.Println("🔍 [扫描] 收到扫描请求")

	// 生成扫描任务ID
	taskID := generateToken()
	log.Printf("🔍 [扫描] 生成任务ID: %s", taskID)

	// 创建扫描任务
	task := &ScanTask{
		ID:        taskID,
		Status:    "running",
		StartTime: time.Now(),
		NewRepos:  []gin.H{},
		MissingRepos: []gin.H{},
	}

	scanTaskMutex.Lock()
	scanTasks[taskID] = task
	scanTaskMutex.Unlock()

	// 生成临时token用于SSE
	tempToken := generateToken()
	tempTokenMutex.Lock()
	tempTokenStore["scan_"+taskID] = tempToken
	tempTokenMutex.Unlock()

	log.Printf("🔍 [扫描] 任务创建成功, taskId=%s, tempToken=%s", taskID, tempToken)

	// 立即返回，后台启动扫描
	c.JSON(200, gin.H{
		"code": 0,
		"data": gin.H{
			"taskId":    taskID,
			"useSSE":    true,
			"tempToken": tempToken,
		},
		"message": "扫描任务已启动",
	})

	log.Println("🔍 [扫描] 已返回响应，启动后台扫描")

	// 后台异步执行扫描
	go performScan(taskID)
}

// performScan 后台执行扫描任务
func performScan(taskID string) {
	log.Printf("🔍 [扫描] 后台任务开始, taskID=%s", taskID)

	scanTaskMutex.RLock()
	task := scanTasks[taskID]
	scanTaskMutex.RUnlock()

	if task == nil {
		log.Printf("❌ [扫描] 任务不存在: %s", taskID)
		return
	}

	// 读取配置
	config, err := utils.ConfigInstance.ReadConfig()
	if err != nil {
		log.Printf("❌ [扫描] 读取配置失败: %v", err)
		task.mu.Lock()
		task.Status = "error"
		task.Error = "读取配置失败"
		task.mu.Unlock()
		return
	}

	storeRootPath := config.StoreRootPath
	log.Printf("🔍 [扫描] 存储路径: %s", storeRootPath)

	// 检查路径是否存在
	if _, err := os.Stat(storeRootPath); os.IsNotExist(err) {
		log.Printf("⚠️ [扫描] 路径不存在: %s", storeRootPath)
		task.mu.Lock()
		task.Status = "completed"
		task.mu.Unlock()
		return
	}

	log.Printf("🔍 [扫描] 开始扫描目录...")

	// 查询数据库中所有现有仓库，包含克隆状态
	type DBRepo struct {
		ID       int
		Author   string
		Repo     string
		Source   string
		URL      string
		IsCloned int
	}
	dbReposMap := make(map[string]DBRepo)
	dbReposByID := make(map[int]DBRepo)

	rows, err := common.Db.Query("SELECT id, author, repo, source, url, COALESCE(is_cloned, 0) FROM repos")
	if err == nil {
		for rows.Next() {
			var r DBRepo
			if err := rows.Scan(&r.ID, &r.Author, &r.Repo, &r.Source, &r.URL, &r.IsCloned); err == nil {
				key := fmt.Sprintf("%s/%s/%s", r.Source, r.Author, r.Repo)
				dbReposMap[key] = r
				dbReposByID[r.ID] = r
			}
		}
		rows.Close()
	}
	log.Printf("🔍 [扫描] 数据库中已有 %d 个仓库记录", len(dbReposMap))

	// 扫描结果
	var newRepos []gin.H
	var statusUpdated []gin.H // 状态已同步的仓库
	scannedCount := 0
	existingCount := 0

	// 记录本地存在的仓库路径（用于状态同步）
	localExistsMap := make(map[string]bool)

	// 第一层：平台目录（如 github, gitee）
	platforms, err := os.ReadDir(storeRootPath)
	if err != nil {
		log.Printf("❌ [扫描] 读取存储目录失败: %v", err)
		task.mu.Lock()
		task.Status = "error"
		task.Error = "读取存储目录失败"
		task.mu.Unlock()
		return
	}

	for _, platform := range platforms {
		if !platform.IsDir() {
			continue
		}

		source := platform.Name()
		platformPath := filepath.Join(storeRootPath, source)

		// 第二层：作者目录
		authors, err := os.ReadDir(platformPath)
		if err != nil {
			continue
		}

		for _, author := range authors {
			if !author.IsDir() {
				continue
			}

			authorName := author.Name()
			authorPath := filepath.Join(platformPath, authorName)

			// 第三层：仓库目录
			repos, err := os.ReadDir(authorPath)
			if err != nil {
				continue
			}

			for _, repo := range repos {
				if !repo.IsDir() {
					continue
				}

				repoName := repo.Name()
				scannedCount++

				// 更新进度
				task.mu.Lock()
				task.TotalDirs = scannedCount
				task.ScannedDirs = scannedCount
				task.mu.Unlock()

				// 构建数据库查找key
				dbKey := fmt.Sprintf("%s/%s/%s", source, authorName, repoName)
				localExistsMap[dbKey] = true

				// 检查数据库中是否已存在
				if dbRepo, exists := dbReposMap[dbKey]; exists {
					existingCount++
					delete(dbReposByID, dbRepo.ID)

					// 检查克隆状态是否需要更新（本地存在但数据库标记为未克隆）
					if dbRepo.IsCloned == 0 {
						// 更新数据库状态
						_, err := common.Db.Exec("UPDATE repos SET is_cloned = 1 WHERE id = ?", dbRepo.ID)
						if err == nil {
							statusUpdated = append(statusUpdated, gin.H{
								"id":        dbRepo.ID,
								"author":    dbRepo.Author,
								"repo":      dbRepo.Repo,
								"oldStatus": "未克隆",
								"newStatus": "已克隆",
							})
							log.Printf("🔄 [扫描] 状态同步: %s/%s 未克隆 → 已克隆", dbRepo.Author, dbRepo.Repo)
						}
					}
					continue
				}

				// 新仓库 - 直接拼接URL
				var repoURL string
				switch source {
				case "github":
					repoURL = fmt.Sprintf("https://github.com/%s/%s", authorName, repoName)
				case "gitee":
					repoURL = fmt.Sprintf("https://gitee.com/%s/%s", authorName, repoName)
				case "gitlab":
					repoURL = fmt.Sprintf("https://gitlab.com/%s/%s", authorName, repoName)
				default:
					repoURL = fmt.Sprintf("https://%s.com/%s/%s", source, authorName, repoName)
				}

				localPath := filepath.Join(authorPath, repoName)

				newRepos = append(newRepos, gin.H{
					"author":    authorName,
					"repo":      repoName,
					"url":       repoURL,
					"source":    source,
					"localPath": localPath,
				})

				log.Printf("✨ [扫描] 发现新仓库: %s/%s (%s)", authorName, repoName, source)
			}
		}
	}

	// 检查数据库中标记为已克隆但本地不存在的仓库
	var missingRepos []gin.H
	for _, repo := range dbReposByID {
		// 更新数据库状态为未克隆
		if repo.IsCloned == 1 {
			_, err := common.Db.Exec("UPDATE repos SET is_cloned = 0 WHERE id = ?", repo.ID)
			if err == nil {
				statusUpdated = append(statusUpdated, gin.H{
					"id":        repo.ID,
					"author":    repo.Author,
					"repo":      repo.Repo,
					"oldStatus": "已克隆",
					"newStatus": "未克隆",
				})
				log.Printf("🔄 [扫描] 状态同步: %s/%s 已克隆 → 未克隆", repo.Author, repo.Repo)
			}
		}

		missingRepos = append(missingRepos, gin.H{
			"id":       repo.ID,
			"author":   repo.Author,
			"repo":     repo.Repo,
			"source":   repo.Source,
			"url":      repo.URL,
			"isCloned": repo.IsCloned,
		})
	}

	task.mu.Lock()
	task.MissingRepos = missingRepos
	task.NewRepos = newRepos
	task.Status = "completed"
	task.ScannedDirs = scannedCount
	task.mu.Unlock()

	log.Printf("✅ [扫描] 扫描完成! 扫描=%d, 新仓库=%d, 未克隆=%d, 已存在=%d, 状态同步=%d",
		scannedCount, len(newRepos), len(missingRepos), existingCount, len(statusUpdated))

	// 如果有状态更新，添加到结果中
	if len(statusUpdated) > 0 {
		log.Printf("📋 [扫描] 已同步 %d 个仓库的克隆状态", len(statusUpdated))
	}
}

// scanReposSSE 扫描进度推送
func scanReposSSE(c *gin.Context) {
	taskID := c.Query("taskId")

	// 验证参数
	if taskID == "" {
		log.Printf("❌ [SSE] 缺少 taskId 参数")
		c.Header("Content-Type", "text/event-stream")
		c.SSEvent("error", gin.H{"message": "缺少 taskId 参数"})
		c.Writer.Flush()
		return
	}

	log.Printf("🔍 [SSE] 客户端连接, taskID=%s", taskID)

	// 设置SSE头部
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 发送开始事件
	c.SSEvent("start", gin.H{"taskId": taskID, "message": "开始扫描..."})
	c.Writer.Flush()

	// 轮询任务状态
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			scanTaskMutex.RLock()
			task := scanTasks[taskID]
			scanTaskMutex.RUnlock()

			if task == nil {
				log.Printf("❌ [SSE] 任务不存在: %s", taskID)
				c.SSEvent("error", gin.H{"message": "任务不存在"})
				c.Writer.Flush()
				return
			}

			// 发送进度更新
			task.mu.RLock()
			scannedDirs := task.ScannedDirs
			totalDirs := task.TotalDirs
			newCount := len(task.NewRepos)
			taskStatus := task.Status
			taskNewRepos := task.NewRepos
			taskMissingRepos := task.MissingRepos
			taskError := task.Error
			task.mu.RUnlock()

			c.SSEvent("progress", gin.H{
				"scannedDirs": scannedDirs,
				"totalDirs":   totalDirs,
				"newCount":    newCount,
			})
			c.Writer.Flush()

			// 检查是否完成
			if taskStatus == "completed" {
				log.Printf("✅ [SSE] 扫描完成, 发送结果, taskID=%s", taskID)

				// 记录活动
				if newCount > 0 {
					addActivityRecord("info", "扫描完成", fmt.Sprintf("发现 %d 个新仓库", newCount), 0, "")
				}

				c.SSEvent("complete", gin.H{
					"new_repos":     taskNewRepos,
					"missing_repos": taskMissingRepos,
					"summary": gin.H{
						"scanned_count":  totalDirs,
						"new_count":      newCount,
						"missing_count":  len(taskMissingRepos),
						"existing_count": scannedDirs - newCount,
					},
				})
				c.Writer.Flush()

				// 清理任务（延迟5秒后清理）
				go func() {
					time.Sleep(5 * time.Second)
					scanTaskMutex.Lock()
					delete(scanTasks, taskID)
					scanTaskMutex.Unlock()
					tempTokenMutex.Lock()
					delete(tempTokenStore, "scan_"+taskID)
					tempTokenMutex.Unlock()
					log.Printf("🗑️ [SSE] 任务已清理: %s", taskID)
				}()
				return
			}

			if taskStatus == "error" {
				log.Printf("❌ [SSE] 扫描错误: %s", taskError)
				c.SSEvent("error", gin.H{"message": taskError})
				c.Writer.Flush()
				return
			}

		case <-c.Request.Context().Done():
			log.Printf("🔌 [SSE] 客户端断开连接: %s", taskID)
			return
		}
	}
}

// updateRepoInfo 更新仓库信息
func updateRepoInfo(c *gin.Context) {
	id := c.Param("id")

	// 获取现有仓库信息
	var repo models.GitRepoInfo
	err := common.Db.QueryRow("SELECT author, repo, url, source FROM repos WHERE id = ?", id).
		Scan(&repo.Author, &repo.Repo, &repo.URL, &repo.Source)
	if err != nil {
		c.JSON(404, gin.H{"code": 404, "message": "仓库不存在"})
		return
	}

	// 根据平台获取仓库信息
	platform, err := utils.GetPlatform(repo.URL)
	if err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "暂不支持该类型仓库的信息更新"})
		return
	}

	// 打印代理状态
	proxyEnabled := IsProxyEnabledForPlatform(repo.Source)
	if proxyEnabled {
		proxyConfig := getCurrentProxyConfig()
		var pURL string
		if proxyConfig.Type == "socks5" {
			pURL = fmt.Sprintf("socks5://%s:%d", proxyConfig.Host, proxyConfig.Port)
		} else {
			pURL = fmt.Sprintf("http://%s:%d", proxyConfig.Host, proxyConfig.Port)
		}
		log.Printf("🔄 [更新] 更新仓库 %s/%s 信息 | 平台: %s | 代理: %s | URL: %s", repo.Author, repo.Repo, repo.Source, pURL, repo.URL)
	} else {
		log.Printf("🔄 [更新] 更新仓库 %s/%s 信息 | 平台: %s | 直连 | URL: %s", repo.Author, repo.Repo, repo.Source, repo.URL)
	}

	// 获取最新仓库信息
	var updatedRepo *models.GitRepoInfo
	if giteePlatform, ok := platform.(*utils.Gitee); ok {
		updatedRepo, err = giteePlatform.FetchRepoInfo(repo.Author, repo.Repo)
		if err != nil {
			log.Printf("❌ [更新] Gitee API 请求失败: %v", err)
			c.JSON(500, gin.H{"code": 500, "message": "无法获取最新仓库信息: " + err.Error()})
			return
		}
	} else {
		log.Printf("🔄 [更新] 正在抓取页面: %s", repo.URL)
		doc, err := platform.FetchDocs(repo.URL)
		if err != nil {
			log.Printf("❌ [更新] 抓取页面失败: %v", err)
			c.JSON(500, gin.H{"code": 500, "message": "无法获取最新仓库信息: " + err.Error()})
			return
		}
		updatedRepo = platform.ParseDoc(doc)
		log.Printf("✅ [更新] 页面解析完成: stars=%d, forks=%d", updatedRepo.Stars, updatedRepo.Fork)
	}

	updatedRepo.URL = repo.URL
	updatedRepo.Author = repo.Author
	updatedRepo.Repo = repo.Repo
	updatedRepo.Source = repo.Source

	// 更新数据库
	_, err = common.Db.Exec(`UPDATE repos SET
		description = ?,
		stars = ?,
		forks = ?,
		topics = ?,
		license = ?,
		languages = ?,
		updated_at = strftime('%Y-%m-%d %H:%M:%S', 'now','localtime')
		WHERE id = ?`,
		updatedRepo.Description,
		updatedRepo.Stars,
		updatedRepo.Fork,
		updatedRepo.Topics,
		updatedRepo.License,
		updatedRepo.Languages,
		id)

	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "更新数据库失败: " + err.Error()})
		return
	}

	// 读取更新后的时间
	var updatedAt string
	common.Db.QueryRow("SELECT COALESCE(updated_at, '') FROM repos WHERE id = ?", id).Scan(&updatedAt)

	// 返回更新后的信息
	c.JSON(200, gin.H{
		"code":    0,
		"message": "仓库信息更新成功",
		"data": gin.H{
			"description": updatedRepo.Description,
			"stars":       updatedRepo.Stars,
			"forks":       updatedRepo.Fork,
			"topics":      updatedRepo.Topics,
			"license":     updatedRepo.License,
			"languages":   updatedRepo.Languages,
			"updated_at":  updatedAt,
		},
	})
}

// ====================== Git 镜像准备接口 ======================

// prepareGitMirror 为 fclone 提供镜像准备服务
// POST /api/git/prepare
// 请求体: { "source": "github", "author": "torvalds", "repo": "linux" }
// 如果仓库已在数据库中且本地已缓存 → pull 更新后返回 ready
// 如果仓库已在数据库中但未缓存 → clone 到本地后返回 ready
// 如果仓库不在数据库中 → 自动注册并克隆，返回 ready
func prepareGitMirror(c *gin.Context) {
	var req struct {
		Source string `json:"source" binding:"required"`
		Author string `json:"author" binding:"required"`
		Repo   string `json:"repo" binding:"required"`
		Force  bool   `json:"force"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "参数错误: 需要 source, author, repo"})
		return
	}

	log.Printf("📦 [Prepare] 收到准备请求: %s/%s/%s", req.Source, req.Author, req.Repo)

	// 读取配置
	config, err := utils.ConfigInstance.ReadConfig()
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "读取配置失败"})
		return
	}

	repoPath := filepath.Join(config.StoreRootPath, req.Source, req.Author, req.Repo)
	gitDir := filepath.Join(repoPath, ".git")

	// 查询数据库
	var repoID int
	var isCloned int
	var repoURL string
	var lastPulledAt sql.NullString
	err = common.Db.QueryRow(
		"SELECT id, COALESCE(is_cloned, 0), url, COALESCE(last_pulled_at, '') FROM repos WHERE source = ? AND author = ? AND repo = ?",
		req.Source, req.Author, req.Repo,
	).Scan(&repoID, &isCloned, &repoURL, &lastPulledAt)

	if err != nil {
		// 仓库不在数据库中 → 自动注册
		log.Printf("📦 [Prepare] 仓库 %s/%s/%s 不在数据库中，自动注册并克隆", req.Source, req.Author, req.Repo)

		// 构建原始仓库 URL
		platformDomains := map[string]string{
			"github":    "github.com",
			"gitee":     "gitee.com",
			"gitlab":    "gitlab.com",
			"bitbucket": "bitbucket.org",
		}
		domain, ok := platformDomains[req.Source]
		if !ok {
			domain = req.Source + ".com"
		}
		repoURL = fmt.Sprintf("https://%s/%s/%s", domain, req.Author, req.Repo)

		// 通过平台接口获取仓库信息并保存
		platform, platformErr := utils.GetPlatform(repoURL)
		if platformErr != nil {
			log.Printf("⚠️ [Prepare] 不支持的平台: %s", req.Source)
			c.JSON(400, gin.H{"code": 400, "message": "不支持的平台: " + req.Source})
			return
		}

		var repoInfo *models.GitRepoInfo
		if giteePlatform, ok := platform.(*utils.Gitee); ok {
			repoInfo, err = giteePlatform.FetchRepoInfo(req.Author, req.Repo)
			if err != nil {
				log.Printf("⚠️ [Prepare] 获取 Gitee 仓库信息失败: %v", err)
				c.JSON(500, gin.H{"code": 500, "message": "获取仓库信息失败: " + err.Error()})
				return
			}
		} else {
			doc, fetchErr := platform.FetchDocs(repoURL)
			if fetchErr != nil {
				log.Printf("⚠️ [Prepare] 抓取页面失败: %v", fetchErr)
				c.JSON(500, gin.H{"code": 500, "message": "获取仓库信息失败: " + fetchErr.Error()})
				return
			}
			repoInfo = platform.ParseDoc(doc)
		}

		repoInfo.URL = repoURL
		repoInfo.Author = req.Author
		repoInfo.Repo = req.Repo
		repoInfo.Source = platform.Name()

		if saveErr := platform.SaveRecords(repoInfo); saveErr != nil {
			log.Printf("⚠️ [Prepare] 保存仓库记录失败: %v", saveErr)
			c.JSON(500, gin.H{"code": 500, "message": "保存仓库记录失败"})
			return
		}

		// 获取新插入的 ID
		common.Db.QueryRow(
			"SELECT id FROM repos WHERE source = ? AND author = ? AND repo = ? ORDER BY id DESC LIMIT 1",
			req.Source, req.Author, req.Repo,
		).Scan(&repoID)

		log.Printf("📦 [Prepare] 已注册仓库: %s (ID: %d)", repoURL, repoID)

		// 继续执行克隆（跳到下面的克隆逻辑）
		isCloned = 0
	}

	if isCloned == 1 && fileExists(gitDir) {
		// 检查今日是否已 pull，如果是则跳过（除非 force）
		if !req.Force && lastPulledAt.Valid {
			pulledDate := strings.SplitN(lastPulledAt.String, " ", 2)[0]
			today := time.Now().Format("2006-01-02")
			if pulledDate == today {
				log.Printf("📦 [Prepare] 仓库今日已更新，跳过 pull: %s/%s/%s", req.Source, req.Author, req.Repo)
				c.JSON(200, gin.H{
					"code":    0,
					"status":  "ready",
					"message": "仓库已就绪（今日已更新）",
				})
				return
			}
		}

		// 已缓存 → pull 更新
		log.Printf("📦 [Prepare] 仓库已缓存，执行 pull: %s", repoPath)

		_, restoreProxy := applyGitProxyForPlatform(req.Source)

		// 先 fetch
		logGitCmd("fetch", "origin")
		fetchCmd := exec.Command("git", "fetch", "origin")
		fetchCmd.Dir = repoPath
		if fetchOutput, err := fetchCmd.CombinedOutput(); err != nil {
			log.Printf("⚠️ [Prepare] fetch 失败: %v, %s", err, string(fetchOutput))
			restoreProxy()
			c.JSON(200, gin.H{
				"code":    0,
				"status":  "ready",
				"message": "仓库已缓存（fetch 更新失败，使用本地版本）",
			})
			return
		}

		// 再 pull
		logGitCmd("pull", "--ff-only")
		pullCmd := exec.Command("git", "pull", "--ff-only")
		pullCmd.Dir = repoPath
		pullOutput, _ := pullCmd.CombinedOutput()

		restoreProxy()

		log.Printf("✅ [Prepare] pull 完成: %s", strings.TrimSpace(string(pullOutput)))
		common.Db.Exec("UPDATE repos SET last_pulled_at = datetime('now', 'localtime') WHERE id = ?", repoID)
		addActivityRecord("info", "fclone 更新", fmt.Sprintf("通过 fclone 更新仓库 %s/%s/%s", req.Source, req.Author, req.Repo), int64(repoID), req.Author+"/"+req.Repo)
		c.JSON(200, gin.H{
			"code":    0,
			"status":  "ready",
			"message": "仓库已更新",
		})
		return
	}

	// 未缓存 → clone
	log.Printf("📦 [Prepare] 仓库未缓存，开始克隆: %s → %s", repoURL, repoPath)

	// 创建父目录
	parentDir := filepath.Dir(repoPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "创建目录失败"})
		return
	}

	// 如果目录已存在但不是有效 git 仓库，先删除
	if _, err := os.Stat(repoPath); err == nil {
		os.RemoveAll(repoPath)
	}

	_, restoreProxy := applyGitProxyForPlatform(req.Source)
	defer restoreProxy()

	logGitCmd("clone", "--progress", repoURL, repoPath)
	cloneCmd := exec.Command("git", "clone", "--progress", repoURL, repoPath)
	cloneOutput, cloneErr := cloneCmd.CombinedOutput()
	if cloneErr != nil {
		log.Printf("❌ [Prepare] 克隆失败: %v, %s", cloneErr, string(cloneOutput))
		addActivityRecord("error", "fclone 克隆失败", fmt.Sprintf("克隆 %s/%s/%s 失败", req.Source, req.Author, req.Repo), 0, req.Author+"/"+req.Repo)
		c.JSON(500, gin.H{
			"code":    500,
			"message": "克隆失败: " + strings.TrimSpace(string(cloneOutput)),
		})
		return
	}

	// 更新数据库状态
	common.Db.Exec("UPDATE repos SET is_cloned = 1 WHERE id = ?", repoID)
	common.Db.Exec("UPDATE repos SET last_pulled_at = datetime('now', 'localtime') WHERE id = ?", repoID)

	addActivityRecord("success", "fclone 克隆", fmt.Sprintf("通过 fclone 自动注册并克隆仓库 %s/%s/%s", req.Source, req.Author, req.Repo), int64(repoID), req.Author+"/"+req.Repo)
	log.Printf("✅ [Prepare] 克隆完成: %s", repoPath)
	c.JSON(200, gin.H{
		"code":    0,
		"status":  "ready",
		"message": "仓库已克隆到服务端",
	})
}

// ====================== Git Smart HTTP 协议 ======================

// gitHTTPHandler 处理 Git Smart HTTP 协议请求
// 支持从已克隆的仓库提供 git clone 服务
// 路径格式: /git/{source}/{author}/{repo}.git/{action}
// 例如: git clone http://localhost:8080/git/github/author/repo.git
func gitHTTPHandler(c *gin.Context) {
	// 解析路径: /git/source/author/repo.git/info/refs 或 /git/source/author/repo.git/git-upload-pack
	rawPath := c.Param("path")
	// 去掉前导斜杠
	rawPath = strings.TrimPrefix(rawPath, "/")

	// 解析路径: source/author/repo.git/...
	// 找到 .git/ 的位置
	gitSuffix := ".git/"
	idx := strings.Index(rawPath, gitSuffix)
	if idx < 0 {
		// 可能没有 .git 后缀，尝试另一种格式
		// 也许是 source/author/repo/info/refs
		parts := strings.SplitN(rawPath, "/", 4)
		if len(parts) < 4 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的 Git 路径格式，应为: /git/{source}/{author}/{repo}.git/{action}"})
			return
		}
		// 用 parts 重新构造
		rawPath = parts[0] + "/" + parts[1] + "/" + parts[2] + ".git/" + parts[3]
		idx = strings.Index(rawPath, gitSuffix)
		if idx < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的 Git 路径格式"})
			return
		}
	}

	repoPart := rawPath[:idx]   // source/author/repo
	action := rawPath[idx+5:]   // info/refs 或 git-upload-pack

	parts := strings.SplitN(repoPart, "/", 3)
	if len(parts) != 3 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的仓库路径，应为: {source}/{author}/{repo}"})
		return
	}

	source, author, repo := parts[0], parts[1], parts[2]

	// 构建本地仓库路径
	config, err := utils.ConfigInstance.ReadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "读取配置失败"})
		return
	}
	repoPath := filepath.Join(config.StoreRootPath, source, author, repo)

	// 检查仓库是否存在
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": fmt.Sprintf("仓库不存在或未克隆: %s/%s/%s", source, author, repo)})
		return
	}

	log.Printf("📦 [GitHTTP] %s %s → %s", c.Request.Method, rawPath, repoPath)

	switch action {
	case "info/refs":
		handleGitInfoRefs(c, repoPath)
	case "git-upload-pack":
		handleGitUploadPack(c, repoPath)
	case "git-receive-pack":
		// 只读服务，不支持推送
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "此服务为只读镜像，不支持 git push"})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "不支持的操作: " + action})
	}
}

// handleGitInfoRefs 处理 GET /git/.../info/refs?service=git-upload-pack
// 返回仓库引用列表（分支、标签等）
func handleGitInfoRefs(c *gin.Context, repoPath string) {
	service := c.Query("service")
	if service != "git-upload-pack" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "不支持的服务: " + service})
		return
	}

	// 执行 git-upload-pack --stateless-rpc --advertise-refs
	cmd := exec.Command("git-upload-pack", "--stateless-rpc", "--advertise-refs", repoPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("❌ [GitHTTP] upload-pack --advertise-refs 失败: %v, output: %s", err, strings.TrimSpace(string(output)))
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取仓库引用失败"})
		return
	}

	// 构造 Smart HTTP 响应
	// 格式: pkt-line("# service=git-upload-pack\n") + "0000" + 仓库引用数据
	pktLine := pktLineEncode("# service=" + service + "\n")

	c.Header("Content-Type", "application/x-"+service+"-advertisement")
	c.Header("Cache-Control", "no-cache")
	c.Writer.Write(pktLine)
	c.Writer.Write([]byte("0000"))
	c.Writer.Write(output)
}

// handleGitUploadPack 处理 POST /git/.../git-upload-pack
// 客户端通过此接口获取实际的仓库数据包
func handleGitUploadPack(c *gin.Context, repoPath string) {
	// 验证 Content-Type
	contentType := c.ContentType()
	if contentType != "application/x-git-upload-pack-request" {
		log.Printf("⚠️ [GitHTTP] 非标准 Content-Type: %s", contentType)
	}

	// 获取请求体 reader，处理 gzip 压缩
	bodyReader := io.Reader(c.Request.Body)
	if c.Request.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(c.Request.Body)
		if err != nil {
			log.Printf("❌ [GitHTTP] gzip 解压失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "解压请求体失败"})
			return
		}
		defer gzReader.Close()
		bodyReader = gzReader
	}

	// 执行 git-upload-pack --stateless-rpc
	cmd := exec.Command("git-upload-pack", "--stateless-rpc", repoPath)

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		log.Printf("❌ [GitHTTP] 创建 stdin 管道失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "内部错误"})
		return
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("❌ [GitHTTP] 创建 stdout 管道失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "内部错误"})
		return
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("❌ [GitHTTP] 创建 stderr 管道失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "内部错误"})
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("❌ [GitHTTP] 启动 upload-pack 失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "内部错误"})
		return
	}

	// 将（解压后的）请求体写入 stdin
	go func() {
		defer stdinPipe.Close()
		io.Copy(stdinPipe, bodyReader)
	}()

	// 读取 stderr 用于日志
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			log.Printf("⚙️ [GitHTTP] upload-pack stderr: %s", scanner.Text())
		}
	}()

	// 设置响应头
	c.Header("Content-Type", "application/x-git-upload-pack-result")
	c.Header("Cache-Control", "no-cache")

	// 流式传输 stdout 到响应
	written, err := io.Copy(c.Writer, stdoutPipe)
	if err != nil {
		log.Printf("⚠️ [GitHTTP] 传输数据中断: %v", err)
		return
	}

	if err := cmd.Wait(); err != nil {
		log.Printf("⚠️ [GitHTTP] upload-pack 退出异常: %v (已传输 %d 字节)", err, written)
		return
	}

	log.Printf("✅ [GitHTTP] upload-pack 完成, 传输 %d 字节", written)
}

// pktLineEncode 编码一个 pkt-line 格式的数据
// 格式: 4位十六进制长度 + 数据
func pktLineEncode(data string) []byte {
	length := len(data) + 4
	return []byte(fmt.Sprintf("%04x%s", length, data))
}

// ====================== Git Smart HTTP 结束 ======================

func init() {
	rootCmd.AddCommand(serveCmd)

	// 添加命令行参数
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "服务器端口")
	serveCmd.Flags().StringVarP(&address, "address", "a", "0.0.0.0", "服务器地址")
	serveCmd.Flags().StringVarP(&customToken, "token", "t", "", "自定义访问令牌（留空则随机生成）")

	// 代理配置参数
	serveCmd.Flags().StringVar(&httpProxy, "http-proxy", "", "HTTP代理服务器 (例如: http://proxy:8080)")
	serveCmd.Flags().StringVar(&httpsProxy, "https-proxy", "", "HTTPS代理服务器 (例如: https://proxy:8080)")
	serveCmd.Flags().StringVar(&noProxy, "no-proxy", "", "不使用代理的主机列表 (例如: localhost,127.0.0.1,*.local)")

	// 快捷代理配置参数
	serveCmd.Flags().StringVar(&localProxy, "local-proxy", "", "本地代理端口 (例如: 1080, 7890, 8080)")
	serveCmd.Flags().StringVar(&proxyType, "proxy-type", "http", "代理类型 (http, socks5) 默认: http")
	serveCmd.Flags().StringVar(&proxyPreset, "proxy-preset", "", "代理预设 (shadowsocks, v2ray, clash, surge)")
}
