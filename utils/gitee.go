package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/cicbyte/forks/models"

	"golang.org/x/net/html"
)

type Gitee struct {
}

func (g *Gitee) Name() string {
	return "gitee"
}

func (g *Gitee) MatchURL(url string) bool {
	return strings.HasPrefix(url, "https://gitee.com/") || strings.HasPrefix(url, "http://gitee.com/")
}

func (g *Gitee) ParseURL(url string) (string, string, string, error) {
	url = strings.TrimPrefix(url, "https://gitee.com/")
	url = strings.TrimPrefix(url, "http://gitee.com/")
	url = strings.TrimSuffix(url, ".git")
	url = strings.Trim(url, "/ ")
	arr := strings.Split(url, "/")
	if len(arr) < 2 || arr[0] == "" || arr[1] == "" {
		return "", "", "", fmt.Errorf("无效的仓库URL格式: %s", url)
	}
	author := arr[0]
	repo := arr[1]
	pageURL := fmt.Sprintf("https://gitee.com/%s/%s", author, repo)
	return author, repo, pageURL, nil
}

func (g *Gitee) CheckExist(url string) (bool, error) {
	return DefaultCheckExist(url)
}

// FetchDocs Gitee 使用 REST API 获取仓库信息，不需要 HTML 解析
// 这里返回 nil，实际数据通过 FetchRepoInfo 获取
func (g *Gitee) FetchDocs(url string) (*html.Node, error) {
	// Gitee 不使用 HTML 解析方式，返回 nil 表示不适用
	return nil, nil
}

// FetchRepoInfo 通过 Gitee API v5 获取仓库信息
func (g *Gitee) FetchRepoInfo(author, repo string) (*models.GitRepoInfo, error) {
	apiURL := fmt.Sprintf("https://gitee.com/api/v5/repos/%s/%s", author, repo)
	client := GetHTTPClient("gitee")
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("解析 Gitee API 响应失败: %v", err)
	}

	repoInfo := &models.GitRepoInfo{}

	if v, ok := data["description"].(string); ok {
		repoInfo.Description = v
	}
	if v, ok := data["stargazers_count"].(float64); ok {
		repoInfo.Stars = int(v)
	}
	if v, ok := data["forks_count"].(float64); ok {
		repoInfo.Fork = int(v)
	}
	if v, ok := data["watchers_count"].(float64); ok {
		repoInfo.Watchers = int(v)
	}
	if v, ok := data["license"].(string); ok {
		repoInfo.License = v
	}
	if v, ok := data["language"].(string); ok && v != "" {
		languages := []string{v}
		repoInfo.Languages, _ = StructToJSON(languages)
	}
	// Gitee API 没有 topics 字段
	repoInfo.Topics, _ = StructToJSON([]string{})

	return repoInfo, nil
}

// ParseDoc Gitee 不需要 HTML 解析，这里返回空对象
func (g *Gitee) ParseDoc(doc *html.Node) *models.GitRepoInfo {
	return &models.GitRepoInfo{}
}

func (g *Gitee) SaveRecords(repo *models.GitRepoInfo) error {
	return DefaultSaveRecords(repo)
}
