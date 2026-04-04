package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// RepoItem 服务端返回的仓库信息
type RepoItem struct {
	Source   string `json:"source"`
	Author   string `json:"author"`
	Repo     string `json:"repo"`
	URL      string `json:"url"`
	IsCloned int    `json:"is_cloned"`
}

// apiResponse 服务端统一响应格式
type apiResponse struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Data    pageData  `json:"data"`
}

// pageData 分页数据
type pageData struct {
	List       []RepoItem `json:"list"`
	Total      int        `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	TotalPages int        `json:"total_pages"`
}

// FetchRepoList 从服务端获取所有仓库列表（自动分页）
func FetchRepoList(server, token string) ([]RepoItem, error) {
	var allRepos []RepoItem
	page := 1
	pageSize := 100

	for {
		url := fmt.Sprintf("%s/api/repos?page=%d&page_size=%d", server, page, pageSize)

		client := &http.Client{Timeout: 30 * time.Second}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %w", err)
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("连接服务端失败: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("读取响应失败: %w", err)
		}

		if resp.StatusCode == 401 {
			return nil, fmt.Errorf("认证失败，请设置 token: fbackup config token <your-token>")
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("服务端返回错误 (%d): %s", resp.StatusCode, string(body))
		}

		var result apiResponse
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("解析响应失败: %w", err)
		}

		pageRepos := result.Data.List
		if len(pageRepos) == 0 {
			break
		}

		allRepos = append(allRepos, pageRepos...)

		// 不足一页或已到最后一页
		if len(pageRepos) < pageSize || page >= result.Data.TotalPages {
			break
		}
		page++
	}

	return allRepos, nil
}

// backupTask 单个备份任务
type backupTask struct {
	index  int
	repo   RepoItem
	server string
}

// backupResult 单个备份结果
type backupResult struct {
	index  int
	action string // "cloned", "pulled", "failed"
	repo   string // source/author/repo
	err    error
}

// BackupRepos 批量备份：并发 clone 或 pull，从服务端本地仓库复制
func BackupRepos(repos []RepoItem, server, targetDir string, concurrency int) error {
	absDir := ResolveAbsDir(targetDir)
	total := len(repos)

	if total == 0 {
		fmt.Println("没有需要备份的仓库")
		return nil
	}

	if err := os.MkdirAll(absDir, 0755); err != nil {
		return fmt.Errorf("创建备份目录失败: %w", err)
	}

	fmt.Printf("共 %d 个仓库，备份到 %s（并发 %d）\n\n", total, absDir, concurrency)

	// 任务通道
	tasks := make(chan backupTask, total)
	results := make(chan backupResult, total)

	// 启动 worker
	var wg sync.WaitGroup
	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range tasks {
				results <- processRepo(task, absDir, total)
			}
		}()
	}

	// 发送任务
	go func() {
		for i, repo := range repos {
			tasks <- backupTask{index: i, repo: repo, server: server}
		}
		close(tasks)
	}()

	// 等待所有 worker 完成，然后关闭结果通道
	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集并打印结果
	var cloned, pulled, failed int32
	var printMu sync.Mutex
	done := int32(0)

	for r := range results {
		atomic.AddInt32(&done, 1)
		printMu.Lock()
		switch r.action {
		case "cloned":
			atomic.AddInt32(&cloned, 1)
			fmt.Printf("[%d/%d] %s ... \x1b[32m✓ 已克隆\x1b[0m\n", atomic.LoadInt32(&done), total, r.repo)
		case "pulled":
			atomic.AddInt32(&pulled, 1)
			fmt.Printf("[%d/%d] %s ... \x1b[32m✓ 已更新\x1b[0m\n", atomic.LoadInt32(&done), total, r.repo)
		case "failed":
			atomic.AddInt32(&failed, 1)
			fmt.Printf("[%d/%d] %s ... \x1b[31m✗ %v\x1b[0m\n", atomic.LoadInt32(&done), total, r.repo, r.err)
		}
		printMu.Unlock()
	}

	fmt.Printf("\n完成: %d 已克隆, %d 已更新", cloned, pulled)
	if failed > 0 {
		fmt.Printf(", \x1b[31m%d 失败\x1b[0m", failed)
	}
	fmt.Println()

	return nil
}

func processRepo(task backupTask, absDir string, total int) backupResult {
	repo := task.repo
	repoDir := filepath.Join(absDir, repo.Source, repo.Author, repo.Repo)
	gitDir := filepath.Join(repoDir, ".git")
	repoName := fmt.Sprintf("%s/%s/%s", repo.Source, repo.Author, repo.Repo)

	// 克隆地址：优先服务端本地，回退原始 URL
	serverURL := fmt.Sprintf("%s/git/%s/%s/%s.git", task.server, repo.Source, repo.Author, repo.Repo)

	if _, err := os.Stat(gitDir); err == nil {
		// 已存在，pull
		if err := RunGitInDirSilent(repoDir, "pull", "--ff-only"); err != nil {
			return backupResult{index: task.index, action: "failed", repo: repoName, err: fmt.Errorf("pull 失败: %v", err)}
		}
		return backupResult{index: task.index, action: "pulled", repo: repoName}
	}

	// 不存在，从服务端 clone
	parentDir := filepath.Dir(repoDir)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return backupResult{index: task.index, action: "failed", repo: repoName, err: fmt.Errorf("创建目录失败: %v", err)}
	}

	if err := RunGitClone(serverURL, repoDir); err != nil {
		return backupResult{index: task.index, action: "failed", repo: repoName, err: fmt.Errorf("clone 失败: %v", err)}
	}

	return backupResult{index: task.index, action: "cloned", repo: repoName}
}

// RunGitClone 执行 git clone（静默），失败时返回 git stderr 内容
func RunGitClone(url, dir string) error {
	cmd := exec.Command("git", "clone", url, dir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := string(out)
		// 只取最后几行关键错误信息
		lines := splitLastN(msg, 3)
		for _, l := range lines {
			l = trimSpace(l)
			if l != "" {
				return fmt.Errorf("%s", l)
			}
		}
		return err
	}
	return nil
}

func splitLastN(s string, n int) []string {
	lines := make([]string, 0, n)
	start := len(s)
	for i := len(s) - 1; i >= 0 && len(lines) < n; i-- {
		if s[i] == '\n' {
			if start > i+1 {
				lines = append(lines, s[i+1:start])
			}
			start = i
		}
	}
	if start > 0 && len(lines) < n {
		lines = append(lines, s[:start])
	}
	// reverse
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}
	return lines
}

func trimSpace(s string) string {
	result := make([]byte, 0, len(s))
	for _, c := range s {
		if c == '\r' || c == '\n' {
			continue
		}
		result = append(result, byte(c))
	}
	return string(result)
}
