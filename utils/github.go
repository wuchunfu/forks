package utils

import (
	"fmt"
	"strings"

	"github.com/cicbyte/forks/models"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

type Github struct {
}

func (g *Github) Name() string {
	return "github"
}

func (g *Github) MatchURL(url string) bool {
	return strings.HasPrefix(url, "https://github.com/") || strings.HasPrefix(url, "http://github.com/")
}

func (g *Github) ParseURL(url string) (string, string, string, error) {
	// 去除https://github.com/
	url = strings.TrimPrefix(url, "https://github.com/")
	url = strings.TrimPrefix(url, "http://github.com/")
	// 去除后缀
	url = strings.TrimSuffix(url, ".git")
	// 去除首尾斜杠和空格
	url = strings.Trim(url, "/ ")
	// 分割
	arr := strings.Split(url, "/")
	if len(arr) < 2 || arr[0] == "" || arr[1] == "" {
		return "", "", "", fmt.Errorf("无效的仓库URL格式: %s", url)
	}
	author := arr[0]
	repo := arr[1]
	pageURL := fmt.Sprintf("https://github.com/%s/%s", author, repo)
	return author, repo, pageURL, nil
}

func (g *Github) CheckExist(url string) (bool, error) {
	return DefaultCheckExist(url)
}

func (g *Github) FetchDocs(url string) (*html.Node, error) {
	return fetchHTTPPage(url)
}

func (g *Github) ParseDoc(doc *html.Node) *models.GitRepoInfo {
	repo := &models.GitRepoInfo{}

	descriptionNode := htmlquery.FindOne(doc, "//h2[text()='About']/following-sibling::p[1]")
	if descriptionNode != nil {
		repo.Description = strings.TrimSpace(htmlquery.InnerText(descriptionNode))
	} else {
		repo.Description = ""
	}
	//Topics string list
	topics := []string{}
	topicNodes := htmlquery.Find(doc, "//h3[text()='Topics']/following-sibling::div[1]//a")
	for _, topicNode := range topicNodes {
		topic := strings.TrimSpace(htmlquery.InnerText(topicNode))
		topics = append(topics, topic)
	}
	repo.Topics, _ = StructToJSON(topics)
	//License
	licenseNode := htmlquery.FindOne(doc, "//h3[text()='License']/following-sibling::div[1]")
	if licenseNode != nil {
		repo.License = strings.TrimSpace(htmlquery.InnerText(licenseNode))
	} else {
		repo.License = ""
	}
	//Stars
	starsNode := htmlquery.FindOne(doc, "//h3[text()='Stars']/following-sibling::div[1]")
	if starsNode != nil {
		stars := strings.TrimSpace(htmlquery.InnerText(starsNode))
		stars = strings.ReplaceAll(stars, "stars", "")
		stars = strings.TrimSpace(stars)
		repo.Stars, _ = ParseGitHubNum(stars)
	} else {
		repo.Stars = 0
	}
	//Watchers
	watchersNode := htmlquery.FindOne(doc, "//h3[text()='Watchers']/following-sibling::div[1]")
	if watchersNode != nil {
		watchers := strings.TrimSpace(htmlquery.InnerText(watchersNode))
		watchers = strings.ReplaceAll(watchers, "watching", "")
		watchers = strings.TrimSpace(watchers)
		repo.Watchers, _ = ParseGitHubNum(watchers)
	} else {
		fmt.Println("未找到watchers节点")
	}
	//Forks
	forksNode := htmlquery.FindOne(doc, "//h3[text()='Forks']/following-sibling::div[1]")
	if forksNode != nil {
		forks := strings.TrimSpace(htmlquery.InnerText(forksNode))
		forks = strings.ReplaceAll(forks, "forks", "")
		forks = strings.TrimSpace(forks)
		repo.Fork, _ = ParseGitHubNum(forks)
	} else {
		fmt.Println("未找到forks节点")
	}
	// Languages
	languages := []string{}
	languageNodes := htmlquery.Find(doc, "//h2[text()='Languages']/following-sibling::ul[1]//a")
	for _, languageNode := range languageNodes {
		language := strings.TrimSpace(htmlquery.InnerText(languageNode))
		language = strings.Replace(language, "\n", "", -1)
		language = strings.Replace(language, " ", "", -1)
		languages = append(languages, language)
	}
	repo.Languages, _ = StructToJSON(languages)

	return repo
}

func (g *Github) SaveRecords(repo *models.GitRepoInfo) error {
	return DefaultSaveRecords(repo)
}
