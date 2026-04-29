package cmd

import (
	"bufio"
	"compress/gzip"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/cicbyte/forks/common"
	"github.com/cicbyte/forks/models"
	"github.com/cicbyte/forks/utils"
)

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
