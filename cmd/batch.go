package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/cicbyte/forks/common"
	"github.com/cicbyte/forks/models"
	"github.com/cicbyte/forks/utils"
)

var (
	// 批量克隆 ID 列表存储
	batchCloneIDsStore = make(map[string][]int)
	batchCloneIDsMutex sync.RWMutex
)

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

func batchPullRepos(c *gin.Context) {
	log.Printf("🔄 [batchPullRepos] 开始批量拉取")

	// 检查是否已有运行中或暂停的同类型任务
	var existCount int
	common.Db.QueryRow("SELECT COUNT(*) FROM tasks WHERE type = 'batch_pull' AND status IN ('running', 'paused')").Scan(&existCount)
	if existCount > 0 {
		c.JSON(409, gin.H{"code": 409, "message": "已有正在进行的批量拉取任务"})
		return
	}

	// 查询所有已克隆且有效的仓库
	rows, err := common.Db.Query("SELECT id, author, repo, url, source FROM repos WHERE is_cloned = 1 AND COALESCE(valid, 1) = 1")
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "查询数据库失败"})
		return
	}
	defer rows.Close()

	var repos []RepoInfo
	for rows.Next() {
		var r RepoInfo
		var url string
		if err := rows.Scan(&r.ID, &r.Author, &r.Repo, &url, &r.Source); err == nil {
			repos = append(repos, r)
		}
	}

	if len(repos) == 0 {
		c.JSON(200, gin.H{"code": 0, "message": "没有需要拉取的仓库", "data": gin.H{"total": 0}})
		return
	}

	// 创建 task 记录
	result, err := common.Db.Exec(`
		INSERT INTO tasks (type, status, total, success_count, fail_count, created_at)
		VALUES ('batch_pull', 'running', ?, 0, 0, strftime('%Y-%m-%d %H:%M:%S', 'now','localtime'))`, len(repos))
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "创建任务失败"})
		return
	}
	taskID, _ := result.LastInsertId()

	// 为每个仓库创建 task_item
	for _, r := range repos {
		repoName := fmt.Sprintf("%s/%s", r.Author, r.Repo)
		common.Db.Exec(`INSERT INTO task_items (task_id, repo_id, repo_name, status, created_at)
			VALUES (?, ?, ?, 'pending', strftime('%Y-%m-%d %H:%M:%S', 'now','localtime'))`, taskID, r.ID, repoName)
	}

	// 启动 goroutine 执行实际拉取
	cancelChan := make(chan struct{})
	taskCancelChans.Store("batch_pull", cancelChan)
	go executeBatchPullTask(taskID, repos, cancelChan)

	c.JSON(200, gin.H{
		"code":    0,
		"message": fmt.Sprintf("开始批量拉取 %d 个仓库", len(repos)),
		"data": gin.H{
			"task_id": taskID,
			"total":   len(repos),
		},
	})
}

func batchPullSSE(c *gin.Context) {
	log.Printf("🔄 [batchPullSSE] 客户端连接")

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 查找正在运行或暂停的 batch_pull 任务
	var taskID int64
	err := common.Db.QueryRow("SELECT id FROM tasks WHERE type = 'batch_pull' AND status IN ('running', 'paused') ORDER BY id DESC LIMIT 1").Scan(&taskID)
	if err != nil {
		c.SSEvent("error", gin.H{"message": "没有正在进行的拉取任务"})
		c.Writer.Flush()
		return
	}

	c.SSEvent("start", gin.H{"task_id": taskID, "message": "连接成功"})
	c.Writer.Flush()

	// 轮询 DB 读取进度
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	lastProcessed := 0

	for range ticker.C {
		// 检查客户端是否断开
		select {
		case <-c.Request.Context().Done():
			return
		default:
		}

		// 查询任务状态
		var status string
		var total, successCount, failCount int
		err := common.Db.QueryRow(`SELECT status, total, success_count, fail_count FROM tasks WHERE id = ?`, taskID).Scan(&status, &total, &successCount, &failCount)
		if err != nil {
			continue
		}

		// 查询新处理的 task_items（从 lastProcessed 开始）
		rows, err := common.Db.Query(`SELECT id, repo_name, status, message FROM task_items WHERE task_id = ? AND id > ? ORDER BY id`, taskID, lastProcessed)
		if err != nil {
			continue
		}

		for rows.Next() {
			var itemID int
			var repoName, itemStatus string
			var message sql.NullString
			if err := rows.Scan(&itemID, &repoName, &itemStatus, &message); err == nil {
				msg := ""
				if message.Valid {
					msg = message.String
				}
				c.SSEvent("progress", gin.H{
					"task_id":  taskID,
					"repo":     repoName,
					"status":   itemStatus,
					"message":  msg,
					"current":  successCount + failCount,
					"total":    total,
				})
				c.Writer.Flush()
				if itemID > lastProcessed {
					lastProcessed = itemID
				}
			}
		}
		rows.Close()

		// 推送整体进度
		c.SSEvent("summary", gin.H{
			"task_id":       taskID,
			"total":         total,
			"success_count": successCount,
			"fail_count":    failCount,
			"status":        status,
		})
		c.Writer.Flush()

		// 任务完成，发送 complete 事件后关闭
		if status == "completed" || status == "failed" || status == "cancelled" {
			c.SSEvent("complete", gin.H{
				"task_id":       taskID,
				"total":         total,
				"success_count": successCount,
				"fail_count":    failCount,
				"message":       fmt.Sprintf("批量拉取完成! 成功: %d, 失败: %d, 总计: %d", successCount, failCount, total),
			})
			c.Writer.Flush()
			return
		}
	}
}

// executeBatchPullTask 执行批量拉取任务的核心逻辑
func executeBatchPullTask(taskID int64, repos []RepoInfo, cancelChan chan struct{}) {
	defer taskCancelChans.Delete("batch_pull")

	config, err := utils.ConfigInstance.ReadConfig()
	if err != nil {
		common.Db.Exec(`UPDATE tasks SET status = 'failed', updated_at = strftime('%Y-%m-%d %H:%M:%S', 'now','localtime') WHERE id = ?`, taskID)
		return
	}

	for i, r := range repos {
		repoName := fmt.Sprintf("%s/%s", r.Author, r.Repo)

		// 检查任务是否被取消
		select {
		case <-cancelChan:
			common.Db.Exec(`UPDATE task_items SET status = 'cancelled', message = '任务已取消' WHERE task_id = ? AND status IN ('pending', 'running')`, taskID)
			common.Db.Exec(`UPDATE tasks SET status = 'cancelled', updated_at = strftime('%Y-%m-%d %H:%M:%S', 'now','localtime') WHERE id = ?`, taskID)
			return
		default:
		}

		// 从 DB 检查状态（处理暂停）
		var currentStatus string
		common.Db.QueryRow("SELECT status FROM tasks WHERE id = ?", taskID).Scan(&currentStatus)
		if currentStatus == "paused" {
			// 阻塞等待：要么恢复，要么取消
			for {
				time.Sleep(2 * time.Second)
				select {
				case <-cancelChan:
					common.Db.Exec(`UPDATE task_items SET status = 'cancelled', message = '任务已取消' WHERE task_id = ? AND status IN ('pending', 'running')`, taskID)
					common.Db.Exec(`UPDATE tasks SET status = 'cancelled', updated_at = strftime('%Y-%m-%d %H:%M:%S', 'now','localtime') WHERE id = ?`, taskID)
					return
				default:
				}
				common.Db.QueryRow("SELECT status FROM tasks WHERE id = ?", taskID).Scan(&currentStatus)
				if currentStatus == "running" {
					break // 恢复执行
				}
				if currentStatus == "cancelled" {
					common.Db.Exec(`UPDATE task_items SET status = 'cancelled', message = '任务已取消' WHERE task_id = ? AND status IN ('pending', 'running')`, taskID)
					return
				}
			}
		}
		if currentStatus == "cancelled" {
			return
		}

		// 检查该 item 是否已完成（retry 场景跳过已成功的）
		var itemStatus string
		common.Db.QueryRow("SELECT status FROM task_items WHERE task_id = ? AND repo_name = ?", taskID, repoName).Scan(&itemStatus)
		if itemStatus == "success" {
			continue // 跳过已成功的 item
		}

		repoPath := filepath.Join(config.StoreRootPath, r.Source, r.Author, r.Repo)

		// 更新 task_item 为 running
		common.Db.Exec(`UPDATE task_items SET status = 'running', message = ? WHERE task_id = ? AND repo_name = ?`,
			fmt.Sprintf("正在拉取 %s ...", repoName), taskID, repoName)

		// 检查本地仓库是否存在
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			common.Db.Exec(`UPDATE task_items SET status = 'failed', message = ? WHERE task_id = ? AND repo_name = ?`,
				"本地仓库不存在，跳过", taskID, repoName)
			common.Db.Exec(`UPDATE tasks SET fail_count = fail_count + 1, updated_at = strftime('%Y-%m-%d %H:%M:%S', 'now','localtime') WHERE id = ?`, taskID)
			continue
		}

		// 设置代理
		_, restoreProxy := applyGitProxyForPlatform(r.Source)

		// fetch + reset 镜像模式
		cmd := exec.Command("git", "fetch", "origin")
		cmd.Dir = repoPath
		logGitCmd("fetch", "origin")
		fetchOutput, fetchErr := cmd.CombinedOutput()

		if fetchErr != nil {
			restoreProxy()
			errMsg := strings.TrimSpace(string(fetchOutput))
			if errMsg == "" {
				errMsg = fetchErr.Error()
			}
			common.Db.Exec(`UPDATE task_items SET status = 'failed', message = ? WHERE task_id = ? AND repo_name = ?`,
				"fetch 失败: "+errMsg, taskID, repoName)
			common.Db.Exec(`UPDATE tasks SET fail_count = fail_count + 1, updated_at = strftime('%Y-%m-%d %H:%M:%S', 'now','localtime') WHERE id = ?`, taskID)
			continue
		}

		// 获取当前分支
		branchCmd := exec.Command("git", "branch", "--show-current")
		branchCmd.Dir = repoPath
		branchOutput, branchErr := branchCmd.Output()
		if branchErr != nil {
			restoreProxy()
			common.Db.Exec(`UPDATE task_items SET status = 'failed', message = ? WHERE task_id = ? AND repo_name = ?`,
				"获取分支失败", taskID, repoName)
			common.Db.Exec(`UPDATE tasks SET fail_count = fail_count + 1, updated_at = strftime('%Y-%m-%d %H:%M:%S', 'now','localtime') WHERE id = ?`, taskID)
			continue
		}
		currentBranch := strings.TrimSpace(string(branchOutput))

		// reset --hard origin/branch
		resetCmd := exec.Command("git", "reset", "--hard", fmt.Sprintf("origin/%s", currentBranch))
		resetCmd.Dir = repoPath
		logGitCmd("reset", "--hard", fmt.Sprintf("origin/%s", currentBranch))
		resetCmd.CombinedOutput()

		restoreProxy()

		// 更新 last_pulled_at
		common.Db.Exec("UPDATE repos SET last_pulled_at = strftime('%Y-%m-%d %H:%M:%S', 'now','localtime') WHERE id = ?", r.ID)

		common.Db.Exec(`UPDATE task_items SET status = 'success', message = ? WHERE task_id = ? AND repo_name = ?`,
			"拉取完成", taskID, repoName)
		common.Db.Exec(`UPDATE tasks SET success_count = success_count + 1, updated_at = strftime('%Y-%m-%d %H:%M:%S', 'now','localtime') WHERE id = ?`, taskID)

		// 每个仓库之间间隔 1 秒，避免触发平台限流
		if i < len(repos)-1 {
			time.Sleep(1 * time.Second)
		}
	}

	// 任务完成
	common.Db.Exec(`UPDATE tasks SET status = 'completed', updated_at = strftime('%Y-%m-%d %H:%M:%S', 'now','localtime') WHERE id = ?`, taskID)
}

func restartBatchPullTask(taskID int64) {
	restartTask(taskID, "batch_pull", executeBatchPullTask)
}

// ============================================
//   BATCH UPDATE INFO - 批量更新仓库信息
//   ============================================

func batchUpdateInfoRepos(c *gin.Context) {
	log.Printf("🔄 [batchUpdateInfoRepos] 开始批量更新信息")

	// 检查是否已有运行中或暂停的同类型任务
	var existCount int
	common.Db.QueryRow("SELECT COUNT(*) FROM tasks WHERE type = 'batch_update_info' AND status IN ('running', 'paused')").Scan(&existCount)
	if existCount > 0 {
		c.JSON(409, gin.H{"code": 409, "message": "已有正在进行的批量更新信息任务"})
		return
	}

	// 查询所有有效仓库（valid = 1 或 IS NULL，排除已标记失效的）
	rows, err := common.Db.Query("SELECT id, author, repo, url, source FROM repos WHERE COALESCE(valid, 1) = 1 ORDER BY id")
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "查询数据库失败"})
		return
	}
	defer rows.Close()

	var repos []RepoInfo
	for rows.Next() {
		var r RepoInfo
		var url string
		if err := rows.Scan(&r.ID, &r.Author, &r.Repo, &url, &r.Source); err == nil {
			repos = append(repos, r)
		}
	}

	if len(repos) == 0 {
		c.JSON(200, gin.H{"code": 0, "message": "没有需要更新的仓库", "data": gin.H{"total": 0}})
		return
	}

	// 创建 task 记录
	result, err := common.Db.Exec(`
		INSERT INTO tasks (type, status, total, success_count, fail_count, created_at)
		VALUES ('batch_update_info', 'running', ?, 0, 0, strftime('%Y-%m-%d %H:%M:%S', 'now','localtime'))`, len(repos))
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "创建任务失败"})
		return
	}
	taskID, _ := result.LastInsertId()

	// 为每个仓库创建 task_item
	for _, r := range repos {
		repoName := fmt.Sprintf("%s/%s", r.Author, r.Repo)
		common.Db.Exec(`INSERT INTO task_items (task_id, repo_id, repo_name, status, created_at)
			VALUES (?, ?, ?, 'pending', strftime('%Y-%m-%d %H:%M:%S', 'now','localtime'))`, taskID, r.ID, repoName)
	}

	// 生成临时 token
	tempToken := generateToken()
	tempTokenMutex.Lock()
	tempTokenStore["batch_update_info"] = tempToken
	tempTokenMutex.Unlock()

	// 启动 goroutine 执行实际更新
	cancelChan := make(chan struct{})
	taskCancelChans.Store("batch_update_info", cancelChan)
	go executeBatchUpdateInfoTask(taskID, repos, cancelChan)

	c.JSON(200, gin.H{
		"code":    0,
		"message": fmt.Sprintf("开始批量更新 %d 个仓库信息", len(repos)),
		"data": gin.H{
			"task_id":   taskID,
			"tempToken": tempToken,
			"total":     len(repos),
		},
	})
}

func batchUpdateInfoSSE(c *gin.Context) {
	log.Printf("🔄 [batchUpdateInfoSSE] 客户端连接")

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 查找正在运行或暂停的 batch_update_info 任务
	var taskID int64
	err := common.Db.QueryRow("SELECT id FROM tasks WHERE type = 'batch_update_info' AND status IN ('running', 'paused') ORDER BY id DESC LIMIT 1").Scan(&taskID)
	if err != nil {
		c.SSEvent("error", gin.H{"message": "没有正在进行的更新信息任务"})
		c.Writer.Flush()
		return
	}

	c.SSEvent("start", gin.H{"task_id": taskID, "message": "连接成功"})
	c.Writer.Flush()

	// 轮询 DB 读取进度
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	lastProcessed := 0

	for range ticker.C {
		// 检查客户端是否断开
		select {
		case <-c.Request.Context().Done():
			return
		default:
		}

		// 查询任务状态
		var status string
		var total, successCount, failCount int
		err := common.Db.QueryRow(`SELECT status, total, success_count, fail_count FROM tasks WHERE id = ?`, taskID).Scan(&status, &total, &successCount, &failCount)
		if err != nil {
			continue
		}

		// 查询新处理的 task_items
		rows, err := common.Db.Query(`SELECT id, repo_name, status, message FROM task_items WHERE task_id = ? AND id > ? ORDER BY id`, taskID, lastProcessed)
		if err != nil {
			continue
		}

		for rows.Next() {
			var itemID int
			var repoName, itemStatus string
			var message sql.NullString
			if err := rows.Scan(&itemID, &repoName, &itemStatus, &message); err == nil {
				msg := ""
				if message.Valid {
					msg = message.String
				}
				c.SSEvent("progress", gin.H{
					"task_id":  taskID,
					"repo":     repoName,
					"status":   itemStatus,
					"message":  msg,
					"current":  successCount + failCount,
					"total":    total,
				})
				c.Writer.Flush()
				if itemID > lastProcessed {
					lastProcessed = itemID
				}
			}
		}
		rows.Close()

		// 推送整体进度
		c.SSEvent("summary", gin.H{
			"task_id":       taskID,
			"total":         total,
			"success_count": successCount,
			"fail_count":    failCount,
			"status":        status,
		})
		c.Writer.Flush()

		// 任务完成
		if status == "completed" || status == "failed" || status == "cancelled" {
			c.SSEvent("complete", gin.H{
				"task_id":       taskID,
				"total":         total,
				"success_count": successCount,
				"fail_count":    failCount,
				"message":       fmt.Sprintf("批量更新信息完成! 成功: %d, 失败: %d, 总计: %d", successCount, failCount, total),
			})
			c.Writer.Flush()
			return
		}
	}
}

// executeBatchUpdateInfoTask 执行批量更新信息任务的核心逻辑
func executeBatchUpdateInfoTask(taskID int64, repos []RepoInfo, cancelChan chan struct{}) {
	defer taskCancelChans.Delete("batch_update_info")

	for i, r := range repos {
		repoName := fmt.Sprintf("%s/%s", r.Author, r.Repo)

		// 检查任务是否被取消
		select {
		case <-cancelChan:
			common.Db.Exec(`UPDATE task_items SET status = 'cancelled', message = '任务已取消' WHERE task_id = ? AND status IN ('pending', 'running')`, taskID)
			common.Db.Exec(`UPDATE tasks SET status = 'cancelled', updated_at = strftime('%Y-%m-%d %H:%M:%S', 'now','localtime') WHERE id = ?`, taskID)
			return
		default:
		}

		// 从 DB 检查状态（处理暂停）
		var currentStatus string
		common.Db.QueryRow("SELECT status FROM tasks WHERE id = ?", taskID).Scan(&currentStatus)
		if currentStatus == "paused" {
			for {
				time.Sleep(2 * time.Second)
				select {
				case <-cancelChan:
					common.Db.Exec(`UPDATE task_items SET status = 'cancelled', message = '任务已取消' WHERE task_id = ? AND status IN ('pending', 'running')`, taskID)
					common.Db.Exec(`UPDATE tasks SET status = 'cancelled', updated_at = strftime('%Y-%m-%d %H:%M:%S', 'now','localtime') WHERE id = ?`, taskID)
					return
				default:
				}
				common.Db.QueryRow("SELECT status FROM tasks WHERE id = ?", taskID).Scan(&currentStatus)
				if currentStatus == "running" {
					break
				}
				if currentStatus == "cancelled" {
					common.Db.Exec(`UPDATE task_items SET status = 'cancelled', message = '任务已取消' WHERE task_id = ? AND status IN ('pending', 'running')`, taskID)
					return
				}
			}
		}
		if currentStatus == "cancelled" {
			return
		}

		// 检查该 item 是否已完成（retry 场景跳过已成功的）
		var itemStatus string
		common.Db.QueryRow("SELECT status FROM task_items WHERE task_id = ? AND repo_name = ?", taskID, repoName).Scan(&itemStatus)
		if itemStatus == "success" {
			continue
		}

		// 更新 task_item 为 running
		common.Db.Exec(`UPDATE task_items SET status = 'running', message = ? WHERE task_id = ? AND repo_name = ?`,
			fmt.Sprintf("正在更新 %s 信息...", repoName), taskID, repoName)

		// 获取仓库 URL
		var repoURL string
		if err := common.Db.QueryRow("SELECT url FROM repos WHERE id = ?", r.ID).Scan(&repoURL); err != nil {
			errMsg := "查询仓库 URL 失败"
			common.Db.Exec(`UPDATE task_items SET status = 'failed', message = ? WHERE task_id = ? AND repo_name = ?`,
				errMsg, taskID, repoName)
			common.Db.Exec(`UPDATE tasks SET fail_count = fail_count + 1, updated_at = strftime('%Y-%m-%d %H:%M:%S', 'now','localtime') WHERE id = ?`, taskID)
			continue
		}

		// 根据平台获取仓库信息
		platform, err := utils.GetPlatform(repoURL)
		if err != nil {
			common.Db.Exec(`UPDATE task_items SET status = 'failed', message = ? WHERE task_id = ? AND repo_name = ?`,
				"不支持的平台类型", taskID, repoName)
			common.Db.Exec(`UPDATE tasks SET fail_count = fail_count + 1, updated_at = strftime('%Y-%m-%d %H:%M:%S', 'now','localtime') WHERE id = ?`, taskID)
			continue
		}

		// 获取最新仓库信息
		var updatedRepo *models.GitRepoInfo
		valid := 1
		fetchFailed := false
		if giteePlatform, ok := platform.(*utils.Gitee); ok {
			updatedRepo, err = giteePlatform.FetchRepoInfo(r.Author, r.Repo)
			if err != nil {
				log.Printf("❌ [批量更新] Gitee %s 失败: %v", repoName, err)
				valid = 0
				fetchFailed = true
			}
		} else {
			doc, docErr := platform.FetchDocs(repoURL)
			if docErr != nil {
				log.Printf("❌ [批量更新] %s 抓取页面失败: %v", repoName, docErr)
				valid = 0
				fetchFailed = true
			} else {
				updatedRepo = platform.ParseDoc(doc)
			}
		}

		if fetchFailed {
			common.Db.Exec("UPDATE repos SET valid = 0, updated_at = strftime('%Y-%m-%d %H:%M:%S', 'now','localtime') WHERE id = ?", r.ID)
			common.Db.Exec(`UPDATE task_items SET status = 'failed', message = ? WHERE task_id = ? AND repo_name = ?`,
				"远端仓库不可访问，已标记为失效", taskID, repoName)
			common.Db.Exec(`UPDATE tasks SET fail_count = fail_count + 1, updated_at = strftime('%Y-%m-%d %H:%M:%S', 'now','localtime') WHERE id = ?`, taskID)
		} else {
			// 更新数据库
			_, dbErr := common.Db.Exec(`UPDATE repos SET
				description = ?,
				stars = ?,
				forks = ?,
				topics = ?,
				license = ?,
				languages = ?,
				valid = ?,
				updated_at = strftime('%Y-%m-%d %H:%M:%S', 'now','localtime')
				WHERE id = ?`,
				updatedRepo.Description,
				updatedRepo.Stars,
				updatedRepo.Fork,
				updatedRepo.Topics,
				updatedRepo.License,
				updatedRepo.Languages,
				valid,
				r.ID)

			if dbErr != nil {
				common.Db.Exec(`UPDATE task_items SET status = 'failed', message = ? WHERE task_id = ? AND repo_name = ?`,
					"更新数据库失败", taskID, repoName)
				common.Db.Exec(`UPDATE tasks SET fail_count = fail_count + 1, updated_at = strftime('%Y-%m-%d %H:%M:%S', 'now','localtime') WHERE id = ?`, taskID)
			} else {
				log.Printf("✅ [批量更新] %s 完成: stars=%d, forks=%d", repoName, updatedRepo.Stars, updatedRepo.Fork)
				common.Db.Exec(`UPDATE task_items SET status = 'success', message = ? WHERE task_id = ? AND repo_name = ?`,
					fmt.Sprintf("更新完成 (stars=%d, forks=%d)", updatedRepo.Stars, updatedRepo.Fork), taskID, repoName)
				common.Db.Exec(`UPDATE tasks SET success_count = success_count + 1, updated_at = strftime('%Y-%m-%d %H:%M:%S', 'now','localtime') WHERE id = ?`, taskID)
			}
		}

		// 每个仓库之间间隔 2 秒，避免触发平台限流
		if i < len(repos)-1 {
			time.Sleep(2 * time.Second)
		}
	}

	// 任务完成
	common.Db.Exec(`UPDATE tasks SET status = 'completed', updated_at = strftime('%Y-%m-%d %H:%M:%S', 'now','localtime') WHERE id = ?`, taskID)

	// 记录活动
	var successCount, failCount int
	common.Db.QueryRow("SELECT success_count, fail_count FROM tasks WHERE id = ?", taskID).Scan(&successCount, &failCount)
	addActivityRecord("success", "批量更新信息完成", fmt.Sprintf("成功 %d 个, 失败 %d 个", successCount, failCount), 0, "")
}

func restartBatchUpdateInfoTask(taskID int64) {
	restartTask(taskID, "batch_update_info", executeBatchUpdateInfoTask)
}

// restartTask 通用任务重启逻辑
func restartTask(taskID int64, taskType string, executeFn func(int64, []RepoInfo, chan struct{})) {
	// 将 running 的 task_items 回退为 pending（上次中断的）
	common.Db.Exec(`UPDATE task_items SET status = 'pending', message = '' WHERE task_id = ? AND status = 'running'`, taskID)

	// 从 task_items 加载 pending 的 repo_id，再从 repos 表查详情
	rows, err := common.Db.Query(`SELECT ti.repo_id, r.author, r.repo, r.source FROM task_items ti
		JOIN repos r ON ti.repo_id = r.id
		WHERE ti.task_id = ? AND ti.status = 'pending' ORDER BY ti.id`, taskID)
	if err != nil {
		log.Printf("❌ [恢复任务] 查询 task_items 失败: %v", err)
		common.Db.Exec(`UPDATE tasks SET status = 'failed', updated_at = strftime('%Y-%m-%d %H:%M:%S', 'now','localtime') WHERE id = ?`, taskID)
		return
	}
	var repos []RepoInfo
	for rows.Next() {
		var r RepoInfo
		if err := rows.Scan(&r.ID, &r.Author, &r.Repo, &r.Source); err == nil {
			repos = append(repos, r)
		}
	}
	rows.Close()

	if len(repos) == 0 {
		common.Db.Exec(`UPDATE tasks SET status = 'completed', updated_at = strftime('%Y-%m-%d %H:%M:%S', 'now','localtime') WHERE id = ?`, taskID)
		return
	}

	// 重置任务状态，从 task_items 重新统计计数
	common.Db.Exec(`UPDATE tasks SET status = 'running',
		success_count = (SELECT COUNT(*) FROM task_items WHERE task_id = ? AND status = 'success'),
		fail_count = (SELECT COUNT(*) FROM task_items WHERE task_id = ? AND status = 'failed'),
		updated_at = strftime('%Y-%m-%d %H:%M:%S', 'now','localtime') WHERE id = ?`, taskID, taskID, taskID)

	cancelChan := make(chan struct{})
	taskCancelChans.Store(taskType, cancelChan)
	go executeFn(taskID, repos, cancelChan)
}
