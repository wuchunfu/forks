package models

type GitPlusConfig struct {
	StoreRootPath string `json:"store_root_path"`
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	Enabled   bool            `json:"enabled"`
	Type      string          `json:"type"` // none, http, socks5
	Host      string          `json:"host"`
	Port      int             `json:"port"`
	NoProxy   string          `json:"no_proxy"`
	Platforms map[string]bool `json:"platforms"` // 各平台独立开关，如 {"github": true, "gitee": false}，为空则跟随全局 enabled
}

/*
*
  - Git仓库信息
    description := ""
    topics_str := ""
    license := ""
    stars_num := 0
    watchers_num := 0
    fork_num := 0
    languages_str := ""
    redme_str := ""
    source := "github"
*/
type GitRepoInfo struct {
	Author      string `json:"author"`
	Repo        string `json:"repo"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Topics      string `json:"topics"`
	License     string `json:"license"`
	Stars       int    `json:"stars"`
	Watchers    int    `json:"watchers"`
	Fork        int    `json:"fork"`
	Languages   string `json:"languages"`
	Source      string `json:"source"`
}
