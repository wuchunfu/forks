package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"forks.com/m/common"
	"forks.com/m/models"
	"forks.com/m/utils"
)

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

	// 更新 last_pulled_at（克隆完成也算一次拉取）
	common.Db.Exec("UPDATE repos SET last_pulled_at = strftime('%Y-%m-%d %H:%M:%S', 'now','localtime') WHERE id = ?", id)

	log.Printf("✅ [cloneRepoSSE] 克隆成功, id=%s", id)
	c.SSEvent("complete", gin.H{"message": "仓库克隆成功!"})
	c.Writer.Flush()
}

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
	c.SSEvent("start", gin.H{"message": "开始拉取最新..."})
	c.Writer.Flush()

	// 按平台动态设置代理
	_, restorePullProxy := applyGitProxyForPlatform(repoSource)
	defer restorePullProxy()

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
	c.SSEvent("progress", gin.H{"message": "验证结果..."})
	c.Writer.Flush()

	cmd = exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoPath
	statusOutput, err := cmd.Output()
	if err != nil {
		c.SSEvent("progress", gin.H{"message": "检查状态时出现问题，但拉取可能已成功"})
	} else {
		if len(strings.TrimSpace(string(statusOutput))) == 0 {
			c.SSEvent("progress", gin.H{"message": "✅ 工作区干净，已同步到远程最新状态"})
		} else {
			c.SSEvent("progress", gin.H{"message": "⚠️ 仍有未提交的更改"})
		}
	}
	c.Writer.Flush()

	// 更新 last_pulled_at
	common.Db.Exec("UPDATE repos SET last_pulled_at = strftime('%Y-%m-%d %H:%M:%S', 'now','localtime') WHERE id = ?", id)

	c.SSEvent("complete", gin.H{"message": "✅ 仓库已同步到远程最新状态!"})
	c.Writer.Flush()
}
