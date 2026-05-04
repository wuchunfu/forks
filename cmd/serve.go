package cmd

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
	"github.com/cicbyte/forks/assets"
	"github.com/cicbyte/forks/common"
	"github.com/cicbyte/forks/models"
	"github.com/cicbyte/forks/utils"
)

var (
	port           int
	address        string
	token          string
	customToken    string
	tempTokenStore   = make(map[string]string) // 临时token存储
	tempTokenMutex   sync.RWMutex

	// 任务取消信号：key 为任务类型（如 "batch_pull"），value 为 close channel
	taskCancelChans sync.Map // map[string]chan struct{}
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

// RepoInfo 仓库基本信息（批量操作用）
type RepoInfo struct {
	ID     int
	Author string
	Repo   string
	Source string
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
			} else if strings.Contains(path, "batch-pull-status") {
				// 批量拉取任务
				keyID = "batch_pull"
				log.Printf("🔑 [SSE认证] 批量拉取任务, keyID=%s", keyID)
			} else if strings.Contains(path, "batch-update-info-status") {
				// 批量更新信息任务
				keyID = "batch_update_info"
				log.Printf("🔑 [SSE认证] 批量更新信息任务, keyID=%s", keyID)
			} else {
				// 其他操作使用路径参数id
				if strings.Contains(path, "clone-status") {
					operation = "_clone"
				} else if strings.Contains(path, "pull-status") {
					operation = "_pull"
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
				if !strings.Contains(path, "scan-status") && !strings.Contains(path, "batch-clone-status") && !strings.Contains(path, "batch-pull-status") && !strings.Contains(path, "batch-update-info-status") {
					tempTokenMutex.Lock()
					delete(tempTokenStore, keyID)
					tempTokenMutex.Unlock()
				}
				log.Printf("✅ [SSE认证] 认证成功")
				c.Next()
				return
			}
		}

		// tasks-stream 路径支持通过 query 参数 token 认证
		if strings.Contains(path, "tasks-stream") {
			queryTok := c.Query("token")
			if token == "" || queryTok == token {
				log.Printf("✅ [SSE认证] tasks-stream token 认证成功")
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

// tasksStreamSSE 任务列表 SSE 推送
func tasksStreamSSE(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 立即发送连接成功事件
	c.SSEvent("connected", gin.H{"message": "已连接"})
	c.Writer.Flush()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// 检查客户端是否断开
		select {
		case <-c.Request.Context().Done():
			return
		default:
		}

		// 查询所有任务概要（只查运行中和暂停的，用于实时更新）
		rows, err := common.Db.Query(`SELECT id, type, status, total, success_count, fail_count, updated_at FROM tasks ORDER BY id DESC LIMIT 50`)
		if err != nil {
			continue
		}

		var tasks []gin.H
		for rows.Next() {
			var id int64
			var taskType, status string
			var total, successCount, failCount int
			var updatedAt sql.NullString
			if err := rows.Scan(&id, &taskType, &status, &total, &successCount, &failCount, &updatedAt); err == nil {
				tasks = append(tasks, gin.H{
					"id":             id,
					"type":           taskType,
					"status":         status,
					"total":          total,
					"success_count":  successCount,
					"fail_count":     failCount,
					"updated_at":     updatedAt.String,
				})
			}
		}
		rows.Close()

		c.SSEvent("tasks", gin.H{"tasks": tasks})
		c.Writer.Flush()
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
		api.POST("/repos/:id/toggle-valid", toggleValid)

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
		api.GET("/version", getVersion)

		// MCP 工具列表
		api.GET("/mcp/tools", getMCPTools)

		// Trending 接口
		api.GET("/trending/dates", getTrendingDates)
		api.GET("/trending", getTrending)
		api.GET("/trending/languages", getTrendingLanguages)
		api.GET("/trending/sync-config", getTrendingSyncConfig)
		api.POST("/trending/sync-config", updateTrendingSyncConfig)
		api.POST("/trending/sync-now", syncTrendingNow)

		// 代码查看接口
		api.GET("/repos/:id/files", getRepoFiles)

		// 批量克隆
		api.POST("/repos/batch-clone", batchCloneRepos)
		api.GET("/repos/:id/file-content", getFileContent)

		// Git 操作接口
		api.POST("/repos/:id/clone", cloneRepo)
		api.POST("/repos/:id/pull", pullRepo)
		api.POST("/repos/batch-pull", batchPullRepos)
		api.POST("/repos/batch-update-info", batchUpdateInfoRepos)
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

		// 任务管理接口
		api.GET("/tasks", getTaskList)
		api.GET("/tasks/:id", getTaskDetail)
		api.DELETE("/tasks/:id", deleteTask)
		api.DELETE("/tasks", clearCompletedTasks)
		api.POST("/tasks/:id/pause", pauseTask)
		api.POST("/tasks/:id/resume", resumeTask)
		api.POST("/tasks/:id/cancel", cancelTask)
		api.POST("/tasks/:id/retry", retryTask)

		// Git 镜像准备接口（供 fclone 调用）
		api.POST("/git/prepare", prepareGitMirror)
	}

	// SSE 路由组 - 使用特殊的SSE认证中间件
	sseApi := r.Group("/api")
	sseApi.Use(sseAuthMiddleware())
	{
		sseApi.GET("/repos/:id/clone-status", cloneRepoSSE)
		sseApi.GET("/repos/:id/pull-status", pullRepoSSE)
		sseApi.GET("/repos/batch-pull-status", batchPullSSE)
		sseApi.GET("/repos/batch-update-info-status", batchUpdateInfoSSE)
		sseApi.GET("/repos/scan-status", scanReposSSE)
		sseApi.GET("/repos/batch-clone-status", batchCloneSSE)
		sseApi.GET("/tasks-stream", tasksStreamSSE)
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

		// 恢复启动时未完成的任务
		// 恢复未完成的任务
		resumeRunningTasks()

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

		// 启动 Trending 同步调度器
		utils.StartSyncScheduler()

		// 启动服务器
		if err := r.Run(address + ":" + strconv.Itoa(port)); err != nil {
			fmt.Printf("启动服务器失败: %v\n", err)
		}
	},
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

// getTrending 获取 GitHub Trending 数据（带存储）
func getTrending(c *gin.Context) {
	language := c.Query("language")
	since := c.DefaultQuery("since", "daily")
	spokenLanguageCode := c.Query("spoken_language_code")
	date := c.Query("date")
	refresh := c.Query("refresh") == "true"

	if since != "daily" && since != "weekly" && since != "monthly" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的 since 参数"})
		return
	}

	repos, err := utils.GetTrending(language, since, spokenLanguageCode, date, refresh)
	if err != nil {
		log.Printf("获取 GitHub Trending 失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	// 批量查询已存在的仓库 URL
	existingURLs := make(map[string]bool)
	if len(repos) > 0 {
		urls := make([]string, 0, len(repos))
		for _, r := range repos {
			urls = append(urls, r.URL)
		}
		placeholders := make([]string, 0, len(urls))
		args := make([]interface{}, 0, len(urls))
		for _, u := range urls {
			placeholders = append(placeholders, "?")
			args = append(args, u)
		}
		query := "SELECT url FROM repos WHERE url IN (" + strings.Join(placeholders, ",") + ")"
		rows, err := common.Db.Query(query, args...)
		if err == nil {
			for rows.Next() {
				var u string
				if rows.Scan(&u) == nil {
					existingURLs[u] = true
				}
			}
			rows.Close()
		}
	}

	type TrendingRepoWithExists struct {
		utils.TrendingRepo
		Exists bool `json:"_exists"`
	}

	items := make([]TrendingRepoWithExists, 0, len(repos))
	for _, r := range repos {
		items = append(items, TrendingRepoWithExists{
			TrendingRepo: r,
			Exists:       existingURLs[r.URL],
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"count": len(repos),
			"items": items,
		},
	})
}

// getTrendingLanguages 获取语言映射
func getTrendingLanguages(c *gin.Context) {
	mappings, err := utils.GetLanguageMappings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    mappings,
	})
}

// getTrendingDates 查询某月有数据的日期列表
func getTrendingDates(c *gin.Context) {
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(time.Now().Year())))
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(time.Now().Month()))))

	if year < 2020 || year > 2100 || month < 1 || month > 12 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的年月参数"})
		return
	}

	dates, err := utils.ListTrendingDates(year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    dates,
	})
}

// getMCPTools 返回 MCP 工具列表
func getMCPTools(c *gin.Context) {
	c.JSON(200, gin.H{
		"code": 0,
		"data": utils.GetMCPToolInfos(),
	})
}

// getTrendingSyncConfig 获取同步配置
func getTrendingSyncConfig(c *gin.Context) {
	cfg, err := utils.LoadSyncConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    cfg,
	})
}

// updateTrendingSyncConfig 更新同步配置
func updateTrendingSyncConfig(c *gin.Context) {
	var cfg utils.TrendingSyncConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	if err := utils.SaveSyncConfig(&cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// syncTrendingNow 立即执行同步
func syncTrendingNow(c *gin.Context) {
	if utils.IsSyncRunning() {
		c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "同步正在执行中"})
		return
	}
	go func() {
		if err := utils.RunSyncTasks(); err != nil {
			log.Printf("[Trending] 手动同步失败: %v", err)
		}
	}()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "同步已启动",
	})
}
