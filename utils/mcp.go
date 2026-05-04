package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/cicbyte/forks/common"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPToolInfo MCP 工具元信息（用于 API 返回）
type MCPToolInfo struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Params      []MCPToolParamInfo `json:"params"`
}

// MCPToolParamInfo 工具参数信息
type MCPToolParamInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

// GetMCPToolInfos 返回所有工具的元信息（供 API 使用）
func GetMCPToolInfos() []MCPToolInfo {
	return []MCPToolInfo{
		{
			Name:        "list_repos",
			Description: "列出仓库，支持搜索和筛选。可按关键词、作者、克隆状态、平台来源筛选，支持分页。",
			Params: []MCPToolParamInfo{
				{Name: "search", Type: "string", Required: false, Description: "搜索关键词，匹配作者/仓库名/描述"},
				{Name: "status", Type: "string", Required: false, Description: "克隆状态筛选：cloned 或 not-cloned"},
				{Name: "author", Type: "string", Required: false, Description: "按作者筛选"},
				{Name: "source", Type: "string", Required: false, Description: "平台来源：github 或 gitee"},
				{Name: "page", Type: "number", Required: false, Description: "页码，默认1"},
				{Name: "page_size", Type: "number", Required: false, Description: "每页条数，默认10，最大100"},
			},
		},
		{
			Name:        "add_repo",
			Description: "通过 URL 添加仓库到收藏列表。支持 GitHub 和 Gitee 平台。",
			Params: []MCPToolParamInfo{
				{Name: "url", Type: "string", Required: true, Description: "仓库 URL，如 https://github.com/owner/repo"},
			},
		},
		{
			Name:        "get_repo",
			Description: "根据 ID 获取单个仓库的详细信息。",
			Params: []MCPToolParamInfo{
				{Name: "id", Type: "string", Required: true, Description: "仓库 ID"},
			},
		},
		{
			Name:        "update_repo_info",
			Description: "从远程平台获取并更新仓库的最新信息（stars、forks、描述等）。",
			Params: []MCPToolParamInfo{
				{Name: "id", Type: "string", Required: true, Description: "仓库 ID"},
			},
		},
		{
			Name:        "get_stats",
			Description: "获取仓库统计信息，包括总数、已克隆数、未克隆数、作者数。",
			Params:      nil,
		},
		{
			Name:        "list_repo_files",
			Description: "获取仓库的文件目录树结构。仅对已克隆的仓库有效。可通过 depth 控制遍历深度，sub_path 指定子目录。",
			Params: []MCPToolParamInfo{
				{Name: "id", Type: "string", Required: true, Description: "仓库 ID"},
				{Name: "depth", Type: "number", Required: false, Description: "目录遍历深度，默认3，最大10"},
				{Name: "sub_path", Type: "string", Required: false, Description: "子目录路径，为空表示仓库根目录"},
			},
		},
		{
			Name:        "read_repo_file",
			Description: "读取仓库中指定文件的文本内容。仅支持文本文件，二进制文件会返回错误。",
			Params: []MCPToolParamInfo{
				{Name: "id", Type: "string", Required: true, Description: "仓库 ID"},
				{Name: "path", Type: "string", Required: true, Description: "文件在仓库中的相对路径"},
			},
		},
		{
			Name:        "get_trending",
			Description: "获取 GitHub Trending 趋势仓库列表。支持按编程语言、时间范围（daily/weekly/monthly）、自然语言筛选。返回仓库名、描述、Stars、Forks 等信息，并标注哪些仓库已在 Forks 中。",
			Params: []MCPToolParamInfo{
				{Name: "language", Type: "string", Required: false, Description: "编程语言，如 python、go、rust。为空表示全部语言"},
				{Name: "since", Type: "string", Required: false, Description: "时间范围：daily（每天）、weekly（每周）、monthly（每月），默认 daily"},
				{Name: "spoken_language", Type: "string", Required: false, Description: "自然语言代码，如 zh（中文）、en（英文）。为空表示全部"},
				{Name: "date", Type: "string", Required: false, Description: "指定日期，格式 2026-05-04。为空表示今天"},
			},
		},
	}
}

// SetupMCPServer 创建并配置 MCP Server，注册所有工具
func SetupMCPServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "forks",
		Version: Version,
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_repos",
		Description: "列出仓库，支持搜索和筛选。可按关键词、作者、克隆状态、平台来源筛选，支持分页。",
	}, listReposTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "add_repo",
		Description: "通过 URL 添加仓库到收藏列表。支持 GitHub 和 Gitee 平台。",
	}, addRepoTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_repo",
		Description: "根据 ID 获取单个仓库的详细信息。",
	}, getRepoTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_repo_info",
		Description: "从远程平台获取并更新仓库的最新信息（stars、forks、描述等）。",
	}, updateRepoInfoTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_stats",
		Description: "获取仓库统计信息，包括总数、已克隆数、未克隆数、作者数。",
	}, getStatsTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_repo_files",
		Description: "获取仓库的文件目录树结构。仅对已克隆的仓库有效。可通过 depth 控制遍历深度，sub_path 指定子目录。",
	}, listRepoFilesTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "read_repo_file",
		Description: "读取仓库中指定文件的文本内容。仅支持文本文件，二进制文件会返回错误。",
	}, readRepoFileTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_trending",
		Description: "获取 GitHub Trending 趋势仓库列表。支持按编程语言、时间范围（daily/weekly/monthly）、自然语言筛选。返回仓库名、描述、Stars、Forks 等信息。",
	}, getTrendingTool)

	return server
}

// ===================== 工具输入/输出类型 =====================

type ListReposInput struct {
	Search   string `json:"search,omitempty"   jsonschema:"搜索关键词，匹配作者/仓库名/描述"`
	Status   string `json:"status,omitempty"   jsonschema:"克隆状态筛选：cloned 或 not-cloned"`
	Author   string `json:"author,omitempty"   jsonschema:"按作者筛选"`
	Source   string `json:"source,omitempty"   jsonschema:"平台来源：github 或 gitee"`
	Page     int    `json:"page,omitempty"     jsonschema:"页码，默认1"`
	PageSize int    `json:"page_size,omitempty" jsonschema:"每页条数，默认10，最大100"`
}

type RepoOutput struct {
	ID          int    `json:"id"`
	Author      string `json:"author"`
	Repo        string `json:"repo"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Stars       int    `json:"stars"`
	Forks       int    `json:"forks"`
	Topics      string `json:"topics"`
	License     string `json:"license"`
	IsCloned    int    `json:"is_cloned"`
	Source      string `json:"source"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type ListReposOutput struct {
	Repos       []RepoOutput `json:"repos"`
	Total       int          `json:"total"`
	Page        int          `json:"page"`
	PageSize    int          `json:"page_size"`
	TotalPages  int          `json:"total_pages"`
}

type AddRepoInput struct {
	URL string `json:"url" jsonschema:"仓库 URL，如 https://github.com/owner/repo"`
}

type IDInput struct {
	ID string `json:"id" jsonschema:"仓库 ID"`
}

type StatsOutput struct {
	TotalRepos     int `json:"total_repos"`
	ClonedCount    int `json:"cloned_count"`
	NotClonedCount int `json:"not_cloned_count"`
	AuthorCount    int `json:"author_count"`
}

type ListRepoFilesInput struct {
	ID      string `json:"id"                jsonschema:"仓库 ID"`
	Depth   int    `json:"depth,omitempty"   jsonschema:"目录遍历深度，默认3，最大10"`
	SubPath string `json:"sub_path,omitempty" jsonschema:"子目录路径，为空表示仓库根目录"`
}

type FileEntry struct {
	Name    string      `json:"name"`
	Path    string      `json:"path"`
	IsDir   bool        `json:"is_dir"`
	Size    int64       `json:"size,omitempty"`
	Children []FileEntry `json:"children,omitempty"`
}

type ReadRepoFileInput struct {
	ID   string `json:"id"   jsonschema:"仓库 ID"`
	Path string `json:"path" jsonschema:"文件在仓库中的相对路径"`
}

type GetTrendingInput struct {
	Language         string `json:"language,omitempty"          jsonschema:"编程语言，如 python、go、rust。为空表示全部语言"`
	Since            string `json:"since,omitempty"             jsonschema:"时间范围：daily（每天）、weekly（每周）、monthly（每月），默认 daily"`
	SpokenLanguage   string `json:"spoken_language,omitempty"   jsonschema:"自然语言代码，如 zh（中文）、en（英文）。为空表示全部"`
	Date             string `json:"date,omitempty"              jsonschema:"指定日期，格式 2026-05-04。为空表示今天"`
}

type TrendingRepoOutput struct {
	Author             string `json:"author"`
	Repo               string `json:"repo"`
	URL                string `json:"url"`
	Description        string `json:"description"`
	Language           string `json:"language"`
	Stars              int    `json:"stars"`
	Forks              int    `json:"forks"`
	CurrentPeriodStars int    `json:"current_period_stars"`
	Exists             bool   `json:"exists_in_forks"`
}

// ===================== 工具实现 =====================

func listReposTool(ctx context.Context, req *mcp.CallToolRequest, input ListReposInput) (*mcp.CallToolResult, ListReposOutput, error) {
	page := input.Page
	if page < 1 {
		page = 1
	}
	pageSize := input.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	var whereClauses []string
	var args []interface{}
	var countArgs []interface{}

	if input.Search != "" {
		pattern := "%" + input.Search + "%"
		whereClauses = append(whereClauses, "(author LIKE ? OR repo LIKE ? OR description LIKE ?)")
		args = append(args, pattern, pattern, pattern)
		countArgs = append(countArgs, pattern, pattern, pattern)
	}
	if input.Author != "" {
		whereClauses = append(whereClauses, "author = ?")
		args = append(args, input.Author)
		countArgs = append(countArgs, input.Author)
	}
	if input.Status == "cloned" {
		whereClauses = append(whereClauses, "is_cloned = 1")
	} else if input.Status == "not-cloned" {
		whereClauses = append(whereClauses, "(is_cloned != 1)")
	}
	if input.Source != "" {
		whereClauses = append(whereClauses, "source = ?")
		args = append(args, input.Source)
		countArgs = append(countArgs, input.Source)
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// 总数
	var total int
	err := common.Db.QueryRow("SELECT COUNT(*) FROM repos"+whereSQL, countArgs...).Scan(&total)
	if err != nil {
		return nil, ListReposOutput{}, fmt.Errorf("查询总数失败: %w", err)
	}

	// 数据
	querySQL := "SELECT id, author, repo, url, description, stars, forks, topics, license, created_at, COALESCE(updated_at,''), COALESCE(is_cloned, 0), source FROM repos" +
		whereSQL + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)

	rows, err := common.Db.Query(querySQL, args...)
	if err != nil {
		return nil, ListReposOutput{}, fmt.Errorf("查询仓库失败: %w", err)
	}
	defer rows.Close()

	var repos []RepoOutput
	for rows.Next() {
		var r RepoOutput
		if err := rows.Scan(&r.ID, &r.Author, &r.Repo, &r.URL, &r.Description,
			&r.Stars, &r.Forks, &r.Topics, &r.License, &r.CreatedAt, &r.UpdatedAt, &r.IsCloned, &r.Source); err != nil {
			continue
		}
		repos = append(repos, r)
	}
	if repos == nil {
		repos = []RepoOutput{}
	}

	totalPages := (total + pageSize - 1) / pageSize

	return nil, ListReposOutput{
		Repos:      repos,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func addRepoTool(ctx context.Context, req *mcp.CallToolRequest, input AddRepoInput) (*mcp.CallToolResult, any, error) {
	platform, err := GetPlatform(input.URL)
	if err != nil {
		return errorResult("不支持该 URL"), nil, nil
	}

	author, repo, pageURL, err := platform.ParseURL(input.URL)
	if err != nil {
		return errorResult("URL 解析失败: " + err.Error()), nil, nil
	}

	exists, err := platform.CheckExist(pageURL)
	if err != nil {
		return errorResult("检查仓库状态失败"), nil, nil
	}
	if exists {
		return errorResult("该仓库已经添加过了"), nil, nil
	}

	var repoInfo *RepoOutput
	if giteePlatform, ok := platform.(*Gitee); ok {
		info, err := giteePlatform.FetchRepoInfo(author, repo)
		if err != nil {
			return errorResult("无法获取仓库信息"), nil, nil
		}
		info.URL = pageURL
		info.Author = author
		info.Repo = repo
		info.Source = platform.Name()
		if err := platform.SaveRecords(info); err != nil {
			return errorResult("保存到数据库失败"), nil, nil
		}
		repoInfo = repoModelToOutput(info)
	} else {
		doc, err := platform.FetchDocs(pageURL)
		if err != nil {
			return errorResult("无法获取仓库信息"), nil, nil
		}
		info := platform.ParseDoc(doc)
		info.URL = pageURL
		info.Author = author
		info.Repo = repo
		info.Source = platform.Name()
		if err := platform.SaveRecords(info); err != nil {
			return errorResult("保存到数据库失败"), nil, nil
		}
		repoInfo = repoModelToOutput(info)
	}

	// 获取插入后的 ID
	var repoId int64
	common.Db.QueryRow("SELECT id FROM repos WHERE author = ? AND repo = ? AND source = ? ORDER BY created_at DESC LIMIT 1",
		author, repo, platform.Name()).Scan(&repoId)
	repoInfo.ID = int(repoId)

	addActivity("success", "MCP 添加仓库", fmt.Sprintf("通过 MCP 添加仓库 %s/%s", author, repo), repoId, author+"/"+repo)

	return nil, repoInfo, nil
}

func getRepoTool(ctx context.Context, req *mcp.CallToolRequest, input IDInput) (*mcp.CallToolResult, RepoOutput, error) {
	var r RepoOutput
	err := common.Db.QueryRow(
		"SELECT id, author, repo, url, description, stars, forks, topics, license, COALESCE(is_cloned,0), source, created_at, COALESCE(updated_at,'') FROM repos WHERE id = ?",
		input.ID,
	).Scan(&r.ID, &r.Author, &r.Repo, &r.URL, &r.Description, &r.Stars, &r.Forks, &r.Topics, &r.License, &r.IsCloned, &r.Source, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, RepoOutput{}, fmt.Errorf("仓库不存在 (id=%s)", input.ID)
	}
	return nil, r, nil
}

func updateRepoInfoTool(ctx context.Context, req *mcp.CallToolRequest, input IDInput) (*mcp.CallToolResult, any, error) {
	var author, repo, url, source string
	err := common.Db.QueryRow("SELECT author, repo, url, source FROM repos WHERE id = ?", input.ID).
		Scan(&author, &repo, &url, &source)
	if err != nil {
		return errorResult("仓库不存在 (id=" + input.ID + ")"), nil, nil
	}

	platform, err := GetPlatform(url)
	if err != nil {
		return errorResult("不支持该平台"), nil, nil
	}

	var description, topics, license, languages string
	var stars, forks int

	if giteePlatform, ok := platform.(*Gitee); ok {
		info, err := giteePlatform.FetchRepoInfo(author, repo)
		if err != nil {
			return errorResult("获取远程信息失败"), nil, nil
		}
		description = info.Description
		topics = info.Topics
		license = info.License
		languages = info.Languages
		stars = info.Stars
		forks = info.Fork
	} else {
		doc, err := platform.FetchDocs(url)
		if err != nil {
			return errorResult("获取远程信息失败"), nil, nil
		}
		info := platform.ParseDoc(doc)
		description = info.Description
		topics = info.Topics
		license = info.License
		languages = info.Languages
		stars = info.Stars
		forks = info.Fork
	}

	_, err = common.Db.Exec(`UPDATE repos SET description=?, stars=?, forks=?, topics=?, license=?, languages=? WHERE id=?`,
		description, stars, forks, topics, license, languages, input.ID)
	if err != nil {
		return errorResult("更新数据库失败"), nil, nil
	}

	var repoId int64
	if id, parseErr := fmt.Sscanf(input.ID, "%d", &repoId); parseErr != nil || id != 1 {
		repoId = 0
	}
	addActivity("info", "MCP 更新信息", fmt.Sprintf("通过 MCP 更新仓库 %s/%s 的信息", author, repo), repoId, author+"/"+repo)

	return nil, map[string]string{"message": fmt.Sprintf("已更新仓库 %s/%s 的信息", author, repo)}, nil
}

func addActivity(activityType, title, description string, repoId int64, repoName string) {
	_, err := common.Db.Exec(`
		INSERT INTO activities (type, title, description, repo_id, repo_name, created_at)
		VALUES (?, ?, ?, ?, ?, datetime('now', 'localtime'))
	`, activityType, title, description, repoId, repoName)
	if err != nil {
		log.Printf("添加活动记录失败: %v", err)
	}
}

func getStatsTool(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, StatsOutput, error) {
	var total, cloned, notCloned, authors int
	common.Db.QueryRow("SELECT COUNT(*) FROM repos").Scan(&total)
	common.Db.QueryRow("SELECT COUNT(*) FROM repos WHERE is_cloned = 1").Scan(&cloned)
	common.Db.QueryRow("SELECT COUNT(*) FROM repos WHERE is_cloned != 1").Scan(&notCloned)
	common.Db.QueryRow("SELECT COUNT(DISTINCT author) FROM repos").Scan(&authors)

	return nil, StatsOutput{
		TotalRepos:     total,
		ClonedCount:    cloned,
		NotClonedCount: notCloned,
		AuthorCount:    authors,
	}, nil
}

func getTrendingTool(ctx context.Context, req *mcp.CallToolRequest, input GetTrendingInput) (*mcp.CallToolResult, any, error) {
	since := input.Since
	if since == "" {
		since = "daily"
	}

	repos, err := GetTrending(input.Language, since, input.SpokenLanguage, input.Date, false)
	if err != nil {
		return errorResult("获取趋势数据失败: " + err.Error()), nil, nil
	}

	// 批量查询哪些已在 forks 中
	urls := make([]string, 0, len(repos))
	for _, r := range repos {
		urls = append(urls, r.URL)
	}
	existSet := batchCheckRepoExists(urls)

	result := make([]TrendingRepoOutput, 0, len(repos))
	for _, r := range repos {
		result = append(result, TrendingRepoOutput{
			Author:             r.Author,
			Repo:               r.Repo,
			URL:                r.URL,
			Description:        r.Description,
			Language:           r.Language,
			Stars:              r.Stars,
			Forks:              r.Forks,
			CurrentPeriodStars: r.CurrentPeriodStars,
			Exists:             existSet[r.URL],
		})
	}

	return nil, result, nil
}

// batchCheckRepoExists 批量检查 URL 是否已在 repos 表中
func batchCheckRepoExists(urls []string) map[string]bool {
	m := make(map[string]bool, len(urls))
	if len(urls) == 0 {
		return m
	}

	placeholders := make([]string, 0, len(urls))
	args := make([]interface{}, 0, len(urls))
	for _, u := range urls {
		placeholders = append(placeholders, "?")
		args = append(args, u)
	}

	query := "SELECT url FROM repos WHERE url IN (" + strings.Join(placeholders, ",") + ")"
	rows, err := common.Db.Query(query, args...)
	if err != nil {
		return m
	}
	defer rows.Close()

	for rows.Next() {
		var u string
		if rows.Scan(&u) == nil {
			m[u] = true
		}
	}
	return m
}

func listRepoFilesTool(ctx context.Context, req *mcp.CallToolRequest, input ListRepoFilesInput) (*mcp.CallToolResult, any, error) {
	repoPath, err := resolveRepoPath(input.ID)
	if err != nil {
		return errorResult(err.Error()), nil, nil
	}

	depth := input.Depth
	if depth < 1 {
		depth = 3
	}
	if depth > 10 {
		depth = 10
	}

	root := filepath.Join(repoPath, input.SubPath)
	if err := validatePath(repoPath, root); err != nil {
		return errorResult(err.Error()), nil, nil
	}

	if _, err := os.Stat(root); os.IsNotExist(err) {
		return errorResult("路径不存在"), nil, nil
	}

	tree, err := buildMCPFileTree(root, "", depth)
	if err != nil {
		return errorResult("读取目录结构失败: " + err.Error()), nil, nil
	}

	return nil, tree, nil
}

func readRepoFileTool(ctx context.Context, req *mcp.CallToolRequest, input ReadRepoFileInput) (*mcp.CallToolResult, any, error) {
	repoPath, err := resolveRepoPath(input.ID)
	if err != nil {
		return errorResult(err.Error()), nil, nil
	}

	fullPath := filepath.Join(repoPath, input.Path)
	if err := validatePath(repoPath, fullPath); err != nil {
		return errorResult(err.Error()), nil, nil
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return errorResult("文件不存在"), nil, nil
	}
	if info.IsDir() {
		return errorResult("路径是目录，不是文件"), nil, nil
	}

	// 限制文件大小：最大 1MB
	if info.Size() > 1024*1024 {
		return errorResult(fmt.Sprintf("文件过大 (%d bytes)，最大支持 1MB", info.Size())), nil, nil
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return errorResult("读取文件失败: " + err.Error()), nil, nil
	}

	// 简单的二进制检测：检查前 512 字节是否包含 NULL
	checkLen := len(content)
	if checkLen > 512 {
		checkLen = 512
	}
	for i := 0; i < checkLen; i++ {
		if content[i] == 0 {
			return errorResult("二进制文件，不支持读取"), nil, nil
		}
	}

	return nil, map[string]string{
		"path":    input.Path,
		"content": string(content),
		"size":    fmt.Sprintf("%d", info.Size()),
	}, nil
}

// ===================== 辅助函数 =====================

// resolveRepoPath 根据仓库 ID 解析本地绝对路径
func resolveRepoPath(id string) (string, error) {
	var author, repo, source string
	err := common.Db.QueryRow("SELECT author, repo, source FROM repos WHERE id = ?", id).
		Scan(&author, &repo, &source)
	if err != nil {
		return "", fmt.Errorf("仓库不存在 (id=%s)", id)
	}

	var isCloned int
	common.Db.QueryRow("SELECT COALESCE(is_cloned,0) FROM repos WHERE id = ?", id).Scan(&isCloned)
	if isCloned != 1 {
		return "", fmt.Errorf("仓库尚未克隆")
	}

	config, err := ConfigInstance.ReadConfig()
	if err != nil {
		return "", fmt.Errorf("读取配置失败")
	}

	return filepath.Join(config.StoreRootPath, source, author, repo), nil
}

// validatePath 安全检查：确保路径在仓库目录内
func validatePath(repoPath, targetPath string) error {
	absRepo, err := filepath.Abs(repoPath)
	if err != nil {
		return fmt.Errorf("路径解析失败")
	}
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("路径解析失败")
	}
	rel, err := filepath.Rel(absRepo, absTarget)
	if err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return fmt.Errorf("路径越界，访问被拒绝")
	}
	return nil
}

// buildMCPFileTree 构建 MCP 用的文件树（支持深度限制）
func buildMCPFileTree(basePath, relativePath string, maxDepth int) ([]FileEntry, error) {
	if maxDepth <= 0 {
		return nil, nil
	}

	fullPath := filepath.Join(basePath, relativePath)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	var result []FileEntry
	for _, entry := range entries {
		name := entry.Name()
		// 跳过隐藏文件/目录
		if strings.HasPrefix(name, ".") {
			continue
		}

		entryRelPath := filepath.Join(relativePath, name)
		// 统一使用正斜杠
		entryRelPath = filepath.ToSlash(entryRelPath)

		info, err := entry.Info()
		if err != nil {
			continue
		}

		fe := FileEntry{
			Name:  name,
			Path:  entryRelPath,
			IsDir: entry.IsDir(),
			Size:  info.Size(),
		}

		if entry.IsDir() {
			children, err := buildMCPFileTree(basePath, entryRelPath, maxDepth-1)
			if err == nil {
				fe.Children = children
			}
		}

		result = append(result, fe)
	}
	return result, nil
}

func errorResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		IsError: true,
	}
}

func repoModelToOutput(info interface{}) *RepoOutput {
	// 通过 JSON 序列化转换
	data, err := json.Marshal(info)
	if err != nil {
		log.Printf("[MCP] repoModelToOutput 序列化失败: %v", err)
		return &RepoOutput{}
	}
	var out RepoOutput
	if err := json.Unmarshal(data, &out); err != nil {
		log.Printf("[MCP] repoModelToOutput 反序列化失败: %v", err)
		return &RepoOutput{}
	}
	return &out
}
