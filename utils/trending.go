package utils

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"

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

type LanguageMappings struct {
	SpokenLanguages      []LanguageEntry    `json:"spoken_languages"`
	ProgrammingLanguages []ProgLangEntry    `json:"programming_languages"`
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
	// 仓库名：h2 > a 的 href
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

	// 描述
	descNode := htmlquery.FindOne(article, ".//p[contains(@class,'col-9')]")
	if descNode != nil {
		repo.Description = strings.TrimSpace(htmlquery.InnerText(descNode))
	}

	// 底部信息行
	row := htmlquery.FindOne(article, ".//div[contains(@class,'f6') and contains(@class,'color-fg-muted')]")

	// 编程语言 + 颜色
	if row != nil {
		langSpan := htmlquery.FindOne(row, ".//span[.//span[contains(@class,'ml-0')]]")
		if langSpan != nil {
			repo.Language = strings.TrimSpace(htmlquery.InnerText(langSpan))
		}
		// 尝试 itemprop 方式
		if repo.Language == "" {
			langItem := htmlquery.FindOne(row, ".//span[@itemprop='programmingLanguage']")
			if langItem != nil {
				repo.Language = strings.TrimSpace(htmlquery.InnerText(langItem))
			}
		}
		// 语言颜色
		colorSpan := htmlquery.FindOne(row, ".//span[contains(@class,'ml-0')]")
		if colorSpan != nil {
			repo.LanguageColor = htmlquery.SelectAttr(colorSpan, "style")
			repo.LanguageColor = strings.TrimPrefix(repo.LanguageColor, "background-color:")
		}
	}

	// Stars
	starsLink := htmlquery.FindOne(article, ".//a[contains(@href,'/stargazers')]")
	if starsLink != nil {
		repo.Stars = parseNum(htmlquery.InnerText(starsLink))
	}

	// Forks
	forksLink := htmlquery.FindOne(article, ".//a[contains(@href,'/forks')]")
	if forksLink != nil {
		repo.Forks = parseNum(htmlquery.InnerText(forksLink))
	}

	// 本周期新增 stars
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

	// 贡献者头像
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

	// 处理 k 后缀（如 1.2k → 1200）
	if strings.HasSuffix(s, "k") {
		s = strings.TrimSuffix(s, "k")
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			return int(v * 1000)
		}
	}

	n, _ := strconv.Atoi(s)
	return n
}
