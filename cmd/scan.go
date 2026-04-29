package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/cicbyte/forks/common"
	"github.com/cicbyte/forks/utils"
)

var (
	// 扫描任务管理
	scanTaskMutex sync.RWMutex
	scanTasks     = make(map[string]*ScanTask)
)

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

	// 将新仓库插入数据库
	for _, newRepo := range newRepos {
		authorVal, _ := newRepo["author"].(string)
		repoVal, _ := newRepo["repo"].(string)
		urlVal, _ := newRepo["url"].(string)
		sourceVal, _ := newRepo["source"].(string)

		_, err := common.Db.Exec(`INSERT INTO repos (
			author, repo, url, git_url, description, stars, forks, topics, license, source, is_cloned, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, '', 0, 0, '', '', ?, 1, strftime('%Y-%m-%d %H:%M:%S', 'now','localtime'), strftime('%Y-%m-%d %H:%M:%S', 'now','localtime')
		)`,
			authorVal, repoVal, urlVal,
			fmt.Sprintf("%s.git", urlVal),
			sourceVal,
		)
		if err != nil {
			log.Printf("⚠️ [扫描] 插入新仓库失败: %s/%s, 错误: %v", authorVal, repoVal, err)
		} else {
			log.Printf("✅ [扫描] 已入库: %s/%s (%s)", authorVal, repoVal, sourceVal)
		}
	}

	task.mu.Lock()
	task.MissingRepos = missingRepos
	task.NewRepos = newRepos
	task.Status = "completed"
	task.ScannedDirs = scannedCount
	task.mu.Unlock()

	log.Printf("✅ [扫描] 扫描完成! 扫描=%d, 新仓库=%d, 已入库=%d, 未克隆=%d, 已存在=%d, 状态同步=%d",
		scannedCount, len(newRepos), len(newRepos), len(missingRepos), existingCount, len(statusUpdated))

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
