package utils

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

//go:embed github_language_mappings.json
var languageMappingsJSON []byte

type TrendingRepo struct {
	Author             string                `json:"author"`
	Repo               string                `json:"repo"`
	URL                string                `json:"url"`
	Description        string                `json:"description"`
	Language           string                `json:"language"`
	LanguageColor      string                `json:"language_color,omitempty"`
	Stars              int                   `json:"stars"`
	Forks              int                   `json:"forks"`
	CurrentPeriodStars int                   `json:"current_period_stars"`
	BuiltBy            []TrendingContributor `json:"built_by"`
}

type TrendingContributor struct {
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

type TrendingFileData struct {
	FetchedAt string         `json:"fetched_at"`
	Params    TrendingParams `json:"params"`
	Count     int            `json:"count"`
	Items     []TrendingRepo `json:"items"`
}

type TrendingParams struct {
	Language          string `json:"language"`
	Since             string `json:"since"`
	SpokenLanguageCode string `json:"spoken_language_code"`
}

type LanguageMappings struct {
	SpokenLanguages      []LanguageEntry `json:"spoken_languages"`
	ProgrammingLanguages []ProgLangEntry `json:"programming_languages"`
}

type LanguageEntry struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type ProgLangEntry struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

var (
	cachedMappings     *LanguageMappings
	cachedMappingsOnce sync.Once
)

func GetLanguageMappings() (*LanguageMappings, error) {
	var err error
	cachedMappingsOnce.Do(func() {
		var m LanguageMappings
		if e := json.Unmarshal(languageMappingsJSON, &m); e != nil {
			err = e
			return
		}
		cachedMappings = &m
	})
	if err != nil {
		return nil, err
	}
	return cachedMappings, nil
}

// GetTrendingDir 返回 trending 数据存储根目录
func GetTrendingDir() string {
	return filepath.Join(ConfigInstance.GetAppDir(), "trending")
}

// trendingFileName 构建文件名：2026-05-04_daily_python_none.json
func trendingFileName(date, since, language, spokenLanguageCode string) string {
	lang := "none"
	if language != "" {
		lang = language
	}
	spoken := "none"
	if spokenLanguageCode != "" {
		spoken = spokenLanguageCode
	}
	return fmt.Sprintf("%s_%s_%s_%s.json", date, since, lang, spoken)
}

// trendingFilePath 构建完整文件路径：{trendingDir}/2026-05/2026-05-04_daily_python_none.json
func trendingFilePath(date, since, language, spokenLanguageCode string) string {
	// date 格式 2026-05-04，取前 7 位作为月份目录
	monthDir := date[:7]
	dir := filepath.Join(GetTrendingDir(), monthDir)
	return filepath.Join(dir, trendingFileName(date, since, language, spokenLanguageCode))
}

// SaveTrendingData 将趋势数据保存为 JSON 文件
func SaveTrendingData(date, since, language, spokenLanguageCode string, repos []TrendingRepo) error {
	path := trendingFilePath(date, since, language, spokenLanguageCode)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	data := TrendingFileData{
		FetchedAt: time.Now().Format(time.RFC3339),
		Params: TrendingParams{
			Language:          language,
			Since:             since,
			SpokenLanguageCode: spokenLanguageCode,
		},
		Count: len(repos),
		Items: repos,
	}

	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}

	if err := os.WriteFile(path, raw, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}
	return nil
}

// LoadTrendingData 从文件加载趋势数据
func LoadTrendingData(date, since, language, spokenLanguageCode string) ([]TrendingRepo, error) {
	path := trendingFilePath(date, since, language, spokenLanguageCode)
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var data TrendingFileData
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	return data.Items, nil
}

// ListTrendingDates 列出某月有数据的日期（去重）
func ListTrendingDates(year, month int) ([]string, error) {
	monthDir := filepath.Join(GetTrendingDir(), fmt.Sprintf("%04d-%02d", year, month))
	entries, err := os.ReadDir(monthDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	dateSet := make(map[string]bool)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// 文件名格式：2026-05-04_daily_python_none.json
		// 取前 10 位作为日期
		if len(name) >= 10 {
			dateStr := name[:10]
			if _, err := time.Parse("2006-01-02", dateStr); err == nil {
				dateSet[dateStr] = true
			}
		}
	}

	dates := make([]string, 0, len(dateSet))
	for d := range dateSet {
		dates = append(dates, d)
	}
	sort.Strings(dates)
	return dates, nil
}

// GetTrending 主逻辑：带存储的趋势数据获取
// date 为空表示当天，非空表示历史日期
// refresh 为 true 时跳过缓存重新采集（仅当天有效）
func GetTrending(language, since, spokenLanguageCode, date string, refresh bool) ([]TrendingRepo, error) {
	today := time.Now().Format("2006-01-02")

	// 历史日期：只读归档
	if date != "" && date != today {
		repos, err := LoadTrendingData(date, since, language, spokenLanguageCode)
		if err != nil {
			return nil, fmt.Errorf("历史数据不存在: %s", date)
		}
		return repos, nil
	}

	// 当天：先看文件是否存在
	filePath := trendingFilePath(today, since, language, spokenLanguageCode)
	if !refresh {
		if _, err := os.Stat(filePath); err == nil {
			repos, err := LoadTrendingData(today, since, language, spokenLanguageCode)
			if err == nil {
				return repos, nil
			}
		}
	}

	// 需要采集
	repos, err := FetchTrendingData(language, since, spokenLanguageCode)
	if err != nil {
		return nil, err
	}

	// 保存到文件（忽略保存错误，不影响返回）
	_ = SaveTrendingData(today, since, language, spokenLanguageCode, repos)

	return repos, nil
}

// --- 以下为已有的爬取和解析逻辑 ---

func BuildTrendingURL(language, since, spokenLanguageCode string) string {
	base := "https://github.com/trending"
	if language != "" {
		base += "/" + url.PathEscape(language)
	}
	params := url.Values{}
	if since == "" {
		since = "daily"
	}
	params.Set("since", since)
	if spokenLanguageCode != "" {
		params.Set("spoken_language_code", spokenLanguageCode)
	}
	return base + "?" + params.Encode()
}

func FetchTrendingData(language, since, spokenLanguageCode string) ([]TrendingRepo, error) {
	u := BuildTrendingURL(language, since, spokenLanguageCode)
	doc, err := fetchHTTPPage(u)
	if err != nil {
		return nil, fmt.Errorf("获取 GitHub Trending 页面失败: %w", err)
	}
	return ParseTrendingPage(doc), nil
}

func ParseTrendingPage(doc *html.Node) []TrendingRepo {
	articles := htmlquery.Find(doc, "//article[contains(@class,'Box-row')]")
	repos := make([]TrendingRepo, 0, len(articles))

	for _, article := range articles {
		repo := TrendingRepo{}
		parseTrendingArticle(article, &repo)
		repos = append(repos, repo)
	}
	return repos
}

func parseTrendingArticle(article *html.Node, repo *TrendingRepo) {
	nameLink := htmlquery.FindOne(article, ".//h2//a")
	if nameLink != nil {
		href := htmlquery.SelectAttr(nameLink, "href")
		href = strings.Trim(href, "/")
		parts := strings.Split(href, "/")
		if len(parts) >= 2 {
			repo.Author = parts[0]
			repo.Repo = parts[1]
			repo.URL = "https://github.com/" + parts[0] + "/" + parts[1]
		}
	}

	descNode := htmlquery.FindOne(article, ".//p[contains(@class,'col-9')]")
	if descNode != nil {
		repo.Description = strings.TrimSpace(htmlquery.InnerText(descNode))
	}

	row := htmlquery.FindOne(article, ".//div[contains(@class,'f6') and contains(@class,'color-fg-muted')]")

	if row != nil {
		langSpan := htmlquery.FindOne(row, ".//span[.//span[contains(@class,'ml-0')]]")
		if langSpan != nil {
			repo.Language = strings.TrimSpace(htmlquery.InnerText(langSpan))
		}
		if repo.Language == "" {
			langItem := htmlquery.FindOne(row, ".//span[@itemprop='programmingLanguage']")
			if langItem != nil {
				repo.Language = strings.TrimSpace(htmlquery.InnerText(langItem))
			}
		}
		colorSpan := htmlquery.FindOne(row, ".//span[contains(@class,'ml-0')]")
		if colorSpan != nil {
			repo.LanguageColor = htmlquery.SelectAttr(colorSpan, "style")
			repo.LanguageColor = strings.TrimPrefix(repo.LanguageColor, "background-color:")
		}
	}

	starsLink := htmlquery.FindOne(article, ".//a[contains(@href,'/stargazers')]")
	if starsLink != nil {
		repo.Stars = parseNum(htmlquery.InnerText(starsLink))
	}

	forksLink := htmlquery.FindOne(article, ".//a[contains(@href,'/forks')]")
	if forksLink != nil {
		repo.Forks = parseNum(htmlquery.InnerText(forksLink))
	}

	if row != nil {
		spans := htmlquery.Find(row, ".//span")
		for _, span := range spans {
			text := strings.TrimSpace(htmlquery.InnerText(span))
			if strings.Contains(text, "stars") {
				repo.CurrentPeriodStars = parseNum(text)
				break
			}
		}
	}

	imgs := htmlquery.Find(article, ".//img[contains(@class,'avatar')]")
	for _, img := range imgs {
		alt := htmlquery.SelectAttr(img, "alt")
		src := htmlquery.SelectAttr(img, "src")
		username := strings.TrimPrefix(alt, "@")
		if username != "" {
			repo.BuiltBy = append(repo.BuiltBy, TrendingContributor{
				Username: username,
				Avatar:   src,
			})
		}
	}
}

func parseNum(s string) int {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "stars", "")
	s = strings.ReplaceAll(s, "today", "")
	s = strings.ReplaceAll(s, "this", "")
	s = strings.ReplaceAll(s, "week", "")
	s = strings.ReplaceAll(s, "month", "")
	s = strings.TrimSpace(s)

	if strings.HasSuffix(s, "k") {
		s = strings.TrimSuffix(s, "k")
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			return int(v * 1000)
		}
	}

	n, _ := strconv.Atoi(s)
	return n
}

