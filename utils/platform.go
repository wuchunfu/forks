package utils

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/cicbyte/forks/common"
	"github.com/cicbyte/forks/models"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// Platform 平台接口，定义仓库平台需要实现的方法
type Platform interface {
	Name() string
	MatchURL(url string) bool
	ParseURL(url string) (author, repo, pageURL string, err error)
	FetchDocs(url string) (*html.Node, error)
	ParseDoc(doc *html.Node) *models.GitRepoInfo
	CheckExist(url string) (bool, error)
	SaveRecords(repo *models.GitRepoInfo) error
}

// registeredPlatforms 注册的所有平台
var registeredPlatforms []Platform

func init() {
	registeredPlatforms = []Platform{
		&Github{},
		&Gitee{},
	}
}

// GetPlatform 根据 URL 获取匹配的平台，无匹配返回 error
func GetPlatform(url string) (Platform, error) {
	for _, p := range registeredPlatforms {
		if p.MatchURL(url) {
			return p, nil
		}
	}
	return nil, fmt.Errorf("不支持的仓库URL: %s", url)
}

// DefaultCheckExist 默认的数据库存在检查实现
func DefaultCheckExist(url string) (bool, error) {
	querySQL := `SELECT * FROM repos WHERE url = ?`
	rows, err := common.Db.Query(querySQL, url)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	if rows.Next() {
		return true, nil
	}
	return false, nil
}

// DefaultSaveRecords 默认的保存记录实现
func DefaultSaveRecords(repo *models.GitRepoInfo) error {
	insertSQL := `INSERT INTO repos (
		author, repo, url, git_url, topics, license, stars, watching, forks, description, languages, source, created_at, updated_at
	) VALUES (
		?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, strftime('%Y-%m-%d %H:%M:%S', 'now','localtime'), strftime('%Y-%m-%d %H:%M:%S', 'now','localtime')
	);`
	_, err := common.Db.Exec(insertSQL,
		repo.Author,
		repo.Repo,
		repo.URL,
		fmt.Sprintf("%s.git", repo.URL),
		repo.Topics,
		repo.License,
		repo.Stars,
		repo.Watchers,
		repo.Fork,
		repo.Description,
		repo.Languages,
		repo.Source,
	)
	return err
}

// GetHTTPClient 根据平台是否启用代理返回对应的 http.Client
// platformName 为平台名（如 "github"、"gitee"）
func GetHTTPClient(platformName string) *http.Client {
	proxyConfig := getCurrentProxyConfigForHTTP()
	if !proxyConfig.Enabled || proxyConfig.Type == "none" {
		log.Printf("🌐 [HTTP] 平台 %s | 直连（代理未启用）", platformName)
		return &http.Client{}
	}

	// 检查平台级开关
	if proxyConfig.Platforms != nil {
		if enabled, ok := proxyConfig.Platforms[platformName]; ok {
			if !enabled {
				log.Printf("🌐 [HTTP] 平台 %s | 直连（平台级代理已关闭）", platformName)
				return &http.Client{}
			}
		}
	}

	// 需要走代理
	var proxyURL string
	if proxyConfig.Type == "socks5" {
		proxyURL = fmt.Sprintf("socks5://%s:%d", proxyConfig.Host, proxyConfig.Port)
	} else {
		proxyURL = fmt.Sprintf("http://%s:%d", proxyConfig.Host, proxyConfig.Port)
	}

	log.Printf("🌐 [HTTP] 平台 %s | 代理: %s", platformName, proxyURL)

	proxyParsed, err := url.Parse(proxyURL)
	if err != nil {
		return &http.Client{}
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyParsed),
	}
	return &http.Client{Transport: transport}
}

// HTTPProxyConfig 用于传递代理配置的轻量结构
type HTTPProxyConfig struct {
	Enabled   bool
	Type      string
	Host      string
	Port      int
	Platforms map[string]bool
}

// getCurrentProxyConfigForHTTP 从内存中获取代理配置（供 HTTP 层使用）
// 由于 utils 包不能直接访问 cmd 包的变量，通过回调方式实现
var getProxyConfigCallback func() models.ProxyConfig

// SetProxyConfigCallback 设置代理配置回调（由 cmd 包初始化时调用）
func SetProxyConfigCallback(cb func() models.ProxyConfig) {
	getProxyConfigCallback = cb
}

func getCurrentProxyConfigForHTTP() models.ProxyConfig {
	if getProxyConfigCallback != nil {
		return getProxyConfigCallback()
	}
	return models.ProxyConfig{}
}

// fetchHTTPPage 公共 HTTP 请求 + HTML 解析
func fetchHTTPPage(pageURL string) (*html.Node, error) {
	// 根据 URL 推断平台名
	platformName := ""
	if strings.HasPrefix(pageURL, "https://github.com/") || strings.HasPrefix(pageURL, "http://github.com/") {
		platformName = "github"
	} else if strings.HasPrefix(pageURL, "https://gitee.com/") || strings.HasPrefix(pageURL, "http://gitee.com/") {
		platformName = "gitee"
	}

	client := GetHTTPClient(platformName)
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	doc, err := htmlquery.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	return doc, nil
}
