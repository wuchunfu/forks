# Forks

一个本地 Git 仓库管理与镜像加速工具，提供 Web 界面管理 GitHub/Gitee 仓库的克隆、更新和浏览，并支持作为 Git 镜像服务器为局域网提供加速克隆服务。

## 界面预览

**首页仪表盘** — 仓库统计与快捷操作
![首页](images/首页.png)

**仓库列表** — 搜索、筛选、批量操作
![仓库列表](images/仓库列表.png)

**仓库详情** — 仓库信息与克隆地址
![仓库详情](images/仓库详情.png)

**作者** — 按作者聚合查看仓库
![作者](images/作者.png)

**活动记录** — 操作日志追踪
![活动](images/活动.png)

**设置** — 代理、Token 等系统配置
![设置](images/设置.png)

## 功能特性

- **仓库管理** — 添加、克隆、拉取、重置 GitHub/Gitee 仓库
- **批量操作** — 批量克隆未克隆的仓库，扫描本地目录自动导入
- **文件浏览** — 在线浏览仓库文件树和文件内容（支持多语言语法高亮、Markdown 渲染、图片/音视频预览）
- **状态监控** — 实时查看克隆进度、更新差异、提交历史
- **代理支持** — HTTP/SOCKS5 代理，支持按平台独立配置
- **镜像加速** — 内置 Git Smart HTTP 服务，局域网内可从本服务加速克隆
- **MCP 集成** — 暴露 MCP 工具接口，支持 AI 客户端（Claude Code、Cursor 等）直接操作仓库
- **活动记录** — 操作日志追踪
- **数据统计** — 仓库和作者的统计仪表盘

## 技术栈

| 层 | 技术 |
|---|---|
| 后端 | Go 1.23、Gin、SQLite、Cobra |
| 前端 | Vue 3、Naive UI、CodeMirror、Pinia |
| 构建 | Vite 7、go:embed |

## 快速开始

### 环境要求

- Go 1.23+（需要 CGO，用于 SQLite 驱动）
- Node.js `^20.19.0` 或 `>=22.12.0`

### 构建

```bash
# 构建前端
cd web
npm install
npm run build
cd ..

# 构建后端（会将前端嵌入二进制）
go build -o forks
```

### 运行

```bash
./forks serve                # 默认监听 :8080
./forks serve -p 9090        # 指定端口
./forks serve --token xxx    # 启用 Token 认证
```

**环境变量：**

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `FORKS_PORT` | 服务端口 | 8080 |
| `FORKS_ADDRESS` | 监听地址 | 0.0.0.0 |
| `FORKS_HOME` | 数据目录 | `~/.cicbyte/forks/` |
| `FORKS_REPO_PATH` | 仓库存储路径 | 交互式输入 |
| `TZ` | 时区 | UTC |

优先级：命令行参数 > 环境变量 > 默认值。

首次运行会提示输入仓库存储路径，数据保存在 `~/.cicbyte/forks/` 目录下。

### 前端开发

```bash
cd web
npm install
npm run dev       # 开发服务器 http://localhost:3000，自动代理 API 到 :8080
```

## 镜像加速克隆

Forks 内置 Git Smart HTTP 协议服务，部署在服务器后，局域网内其他机器可直接从本服务克隆已缓存的仓库，无需访问外网。

### 使用方式

```bash
# 直接 git clone（克隆后 origin 会指向镜像地址）
git clone http://<server-ip>:8080/git/github/author/repo.git

# 推荐使用 fclone 工具（自动修正 remote 地址）
fclone http://<server-ip>:8080/git/github/author/repo.git
```

### fclone 工具

独立 CLI 工具，位于 `fclone/` 目录，零依赖。

```bash
cd fclone && go build -o fclone .

# 克隆并自动修正 remote
fclone http://<server-ip>:8080/git/github/torvalds/linux.git

# 指定目标目录
fclone http://<server-ip>:8080/git/github/torvalds/linux.git my-linux
```

克隆完成后 remote 配置：
- `origin` → `https://github.com/author/repo.git`（原始仓库，支持 push/pull）

## Docker 部署

### 使用预构建镜像（推荐）

```bash
docker pull ghcr.io/cicbyte/forks:latest
```

```bash
# 运行（默认端口 8080）
docker run -d \
  -p 8080:8080 \
  -e TZ=Asia/Shanghai \
  -v ./data:/data \
  ghcr.io/cicbyte/forks:latest

# 自定义端口
docker run -d \
  -p 9090:9090 \
  -e FORKS_PORT=9090 \
  -e TZ=Asia/Shanghai \
  -v ./data:/data \
  ghcr.io/cicbyte/forks:latest
```

使用 `docker-compose.yml` 部署：

```yaml
services:
  forks:
    image: ghcr.io/cicbyte/forks:latest
    container_name: forks
    restart: unless-stopped
    network_mode: host
    environment:
      - TZ=Asia/Shanghai
      - FORKS_HOME=/data
      - FORKS_PORT=8083
    volumes:
      - ./data:/data
```

```bash
docker compose up -d
```

### 自行构建

```bash
# 构建镜像
docker build -t forks .

# 运行（默认端口 8080）
docker run -d \
  -p 8080:8080 \
  -e TZ=Asia/Shanghai \
  -v ./data:/data \
  forks

# 使用 docker-compose
docker compose up -d
```

## fbackup 批量备份工具

独立 CLI 工具，位于 `fbackup/` 目录，从 Forks 服务端并发备份仓库到本地。

```bash
cd fbackup && go build -o fbackup

fbackup config server http://<server-ip>:8080     # 配置服务端
fbackup config token xxx                          # 配置 Token
fbackup config dir /data/backup                   # 配置备份目录
fbackup                                           # 执行备份（默认 5 并发）
fbackup -c 10 -d /backup                         # 指定并发数和目录
```

- 通过服务端 Git HTTP 接口（`/git/...`）从局域网复制，不访问外网
- 已存在的仓库执行 `git pull --ff-only`，不存在则 `git clone`
- 配置文件：`~/.fbackup.json`

## MCP 集成

Forks 通过 MCP (Model Context Protocol) Streamable HTTP 暴露工具接口，AI 客户端可直接操作仓库。

**端点：** `http://<server>:8080/mcp`

**可用工具：**

| 工具 | 功能 |
|---|---|
| `list_repos` | 列出仓库（支持搜索/筛选/分页） |
| `add_repo` | 添加仓库 |
| `get_repo` | 获取单个仓库详情 |
| `update_repo_info` | 更新仓库远程信息 |
| `get_stats` | 获取仓库统计信息 |
| `list_repo_files` | 列出仓库文件结构 |
| `read_repo_file` | 读取仓库文件内容 |

认证方式与 API 相同，使用 Bearer Token。

## 项目结构

```
├── main.go              # 入口，嵌入 web/dist
├── cmd/
│   ├── root.go          # CLI 根命令，数据库和日志初始化
│   └── serve.go         # HTTP 路由和业务逻辑
├── common/              # 全局变量（DB、日志文件）
├── models/              # 数据结构定义
├── utils/               # 配置、GitHub API、MCP 服务、工具函数
├── assets/              # 嵌入资源桥接
├── fclone/              # 独立 CLI 工具（镜像加速克隆）
│   ├── main.go
│   └── go.mod
├── fbackup/             # 独立 CLI 工具（批量备份）
│   ├── main.go
│   └── go.mod
├── web/                 # Vue 3 前端
│   └── src/
│       ├── api/         # API 调用封装
│       ├── views/       # 页面组件
│       ├── components/  # 通用组件
│       ├── stores/      # Pinia 状态管理
│       └── router/      # 路由配置
└── API.md               # API 文档
```

## API

所有接口在 `/api` 前缀下，完整文档见 [API.md](./API.md)。

## License

[MIT](./LICENSE)

[![zread](https://img.shields.io/badge/Ask_Zread-_.svg?style=flat&color=00b0aa&labelColor=000000&logo=data%3Aimage%2Fsvg%2Bxml%3Bbase64%2CPHN2ZyB3aWR0aD0iMTYiIGhlaWdodD0iMTYiIHZpZXdCb3g9IjAgMCAxNiAxNiIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZD0iTTQuOTYxNTYgMS42MDAxSDIuMjQxNTZDMS44ODgxIDEuNjAwMSAxLjYwMTU2IDEuODg2NjQgMS42MDE1NiAyLjI0MDFWNC45NjAxQzEuNjAxNTYgNS4zMTM1NiAxLjg4ODEgNS42MDAxIDIuMjQxNTYgNS42MDAxSDQuOTYxNTZDNS4zMTUwMiA1LjYwMDEgNS42MDE1NiA1LjMxMzU2IDUuNjAxNTYgNC45NjAxVjIuMjQwMUM1LjYwMTU2IDEuODg2NjQgNS4zMTUwMiAxLjYwMDEgNC45NjE1NiAxLjYwMDFaIiBmaWxsPSIjZmZmIi8%2BCjxwYXRoIGQ9Ik00Ljk2MTU2IDEwLjM5OTlIMi4yNDE1NkMxLjg4ODEgMTAuMzk5OSAxLjYwMTU2IDEwLjY4NjQgMS42MDE1NiAxMS4wMzk5VjEzLjc1OTlDMS42MDE1NiAxNC4xMTM0IDEuODg4MSAxNC4zOTk5IDIuMjQxNTYgMTQuMzk5OUg0Ljk2MTU2QzUuMzE1MDIgMTQuMzk5OSA1LjYwMTU2IDE0LjExMzQgNS42MDE1NiAxMy43NTk5VjExLjAzOTlDNS42MDE1NiAxMC42ODY0IDUuMzE1MDIgMTAuMzk5OSA0Ljk2MTU2IDEwLjM5OTlaIiBmaWxsPSIjZmZmIi8%2BCjxwYXRoIGQ9Ik0xMy43NTg0IDEuNjAwMUgxMS4wMzg0QzEwLjY4NSAxLjYwMDEgMTAuMzk4NCAxLjg4NjY0IDEwLjM5ODQgMi4yNDAxVjQuOTYwMUMxMC4zOTg0IDUuMzEzNTYgMTAuNjg1IDUuNjAwMSAxMS4wMzg0IDUuNjAwMUgxMy43NTg0QzE0LjExMTkgNS42MDAxIDE0LjM5ODQgNS4zMTM1NiAxNC4zOTg0IDQuOTYwMVYyLjI0MDFDMTQuMzk4NCAxLjg4NjY0IDE0LjExMTkgMS42MDAxIDEzLjc1ODQgMS42MDAxWiIgZmlsbD0iI2ZmZiIvPgo8cGF0aCBkPSJNNCAxMkwxMiA0TDQgMTJaIiBmaWxsPSIjZmZmIi8%2BCjxwYXRoIGQ9Ik00IDEyTDEyIDQiIHN0cm9rZT0iI2ZmZiIgc3Ryb2tlLXdpZHRoPSIxLjUiIHN0cm9rZS1saW5lY2FwPSJyb3VuZCIvPgo8L3N2Zz4K&logoColor=ffffff)](https://zread.ai/cicbyte/forks)
