package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"forks.com/m/common"
	"forks.com/m/models"
	"forks.com/m/utils"
)

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
	querySQL := "SELECT id, author, repo, url, COALESCE(description, ''), COALESCE(stars, 0), COALESCE(forks, 0), COALESCE(topics, ''), COALESCE(license, ''), created_at, COALESCE(updated_at, ''), COALESCE(is_cloned, 0), COALESCE(source, 'github'), COALESCE(valid, 1), COALESCE(last_pulled_at, '') FROM repos" + whereSQL + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
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
		var valid int
		var lastPulledAt string
		err := rows.Scan(&id, &repo.Author, &repo.Repo, &repo.URL, &repo.Description, &repo.Stars, &repo.Fork, &repo.Topics, &repo.License, &createdAt, &updatedAt, &isCloned, &repo.Source, &valid, &lastPulledAt)
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
			"source":          repo.Source,
			"valid":           valid,
			"last_pulled_at":  lastPulledAt,
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

	// 先获取仓库信息（删除数据库前需要这些信息）
	var author, repo, source string
	common.Db.QueryRow("SELECT author, repo, COALESCE(source, 'github') FROM repos WHERE id = ?", id).Scan(&author, &repo, &source)

	repoName := author + "/" + repo

	// 构建本地仓库路径
	config, _ := utils.ConfigInstance.ReadConfig()
	repoPath := filepath.Join(config.StoreRootPath, source, author, repo)

	_, err := common.Db.Exec("DELETE FROM repos WHERE id = ?", id)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "删除仓库失败"})
		return
	}

	// 删除本地 git 仓库文件
	if _, err := os.Stat(repoPath); err == nil {
		if removeErr := os.RemoveAll(repoPath); removeErr != nil {
			log.Printf("⚠️ [删除] 删除本地仓库文件失败: %s, 错误: %v", repoPath, removeErr)
		} else {
			log.Printf("🗑️ [删除] 已删除本地仓库文件: %s", repoPath)
		}
	}

	// 记录活动（repo_id 为 0 因为已删除）
	addActivityRecord("warning", "删除仓库", fmt.Sprintf("已删除仓库 %s", repoName), 0, repoName)

	c.JSON(200, gin.H{"code": 0, "message": "删除成功"})
}

func getRepo(c *gin.Context) {
	id := c.Param("id")
	var repo models.GitRepoInfo
	var createdAt, updatedAt string
	var isCloned int
	var source string
	var valid int
	err := common.Db.QueryRow("SELECT author, repo, url, COALESCE(description, ''), COALESCE(stars, 0), COALESCE(forks, 0), COALESCE(topics, ''), COALESCE(license, ''), created_at, COALESCE(updated_at, ''), COALESCE(is_cloned, 0), COALESCE(source, 'github'), COALESCE(valid, 1) FROM repos WHERE id = ?", id).
		Scan(&repo.Author, &repo.Repo, &repo.URL, &repo.Description, &repo.Stars, &repo.Fork, &repo.Topics, &repo.License, &createdAt, &updatedAt, &isCloned, &source, &valid)
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
		"valid":       valid,
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
	valid := 1 // 默认有效
	if giteePlatform, ok := platform.(*utils.Gitee); ok {
		updatedRepo, err = giteePlatform.FetchRepoInfo(repo.Author, repo.Repo)
		if err != nil {
			// Gitee API 返回 404 等错误，标记为失效
			log.Printf("❌ [更新] Gitee API 请求失败，标记为失效: %v", err)
			valid = 0
			common.Db.Exec("UPDATE repos SET valid = 0, updated_at = strftime('%Y-%m-%d %H:%M:%S', 'now','localtime') WHERE id = ?", id)
			c.JSON(200, gin.H{"code": 0, "message": "远端仓库不可访问，已标记为失效", "data": gin.H{"valid": 0}})
			return
		}
	} else {
		log.Printf("🔄 [更新] 正在抓取页面: %s", repo.URL)
		doc, err := platform.FetchDocs(repo.URL)
		if err != nil {
			// 页面抓取失败（404等），标记为失效
			log.Printf("❌ [更新] 抓取页面失败，标记为失效: %v", err)
			valid = 0
			common.Db.Exec("UPDATE repos SET valid = 0, updated_at = strftime('%Y-%m-%d %H:%M:%S', 'now','localtime') WHERE id = ?", id)
			c.JSON(200, gin.H{"code": 0, "message": "远端仓库不可访问，已标记为失效", "data": gin.H{"valid": 0}})
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
			"valid":       valid,
		},
	})
}

// toggleValid 切换仓库的 valid 状态
func toggleValid(c *gin.Context) {
	id := c.Param("id")

	var currentValid int
	err := common.Db.QueryRow("SELECT COALESCE(valid, 1) FROM repos WHERE id = ?", id).Scan(&currentValid)
	if err != nil {
		c.JSON(404, gin.H{"code": 404, "message": "仓库不存在"})
		return
	}

	newValid := 1
	if currentValid == 1 {
		newValid = 0
	}

	_, err = common.Db.Exec("UPDATE repos SET valid = ? WHERE id = ?", newValid, id)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "更新状态失败"})
		return
	}

	msg := "已恢复为有效"
	if newValid == 0 {
		msg = "已标记为失效"
	}

	c.JSON(200, gin.H{
		"code":    0,
		"message": msg,
		"data": gin.H{
			"valid": newValid,
		},
	})
}
