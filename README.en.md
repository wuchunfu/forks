# Forks

> Local Git repository manager and mirror acceleration tool — Web UI for managing GitHub/Gitee repos with LAN-accelerated cloning.

**English** | [中文](./README.md)

[![Docker Image CI](https://github.com/cicbyte/forks/actions/workflows/docker-image.yml/badge.svg)](https://github.com/cicbyte/forks/actions/workflows/docker-image.yml)
[![Release](https://img.shields.io/github/v/release/cicbyte/forks?style=flat&logo=github&color=green)](https://github.com/cicbyte/forks/releases/latest)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE)
[![Docker Pulls](https://img.shields.io/badge/ghcr.io-cicbyte%2Fforks-24292f?logo=docker&logoColor=white)](https://github.com/cicbyte/forks/pkgs/container/forks)

## Table of Contents

- [Screenshots](#screenshots)
- [Features](#features)
- [Tech Stack](#tech-stack)
- [Quick Start](#quick-start)
- [Mirror Accelerated Cloning](#mirror-accelerated-cloning)
- [Docker Deployment](#docker-deployment)
- [MCP Integration](#mcp-integration)
- [Companion Tools](#companion-tools)
- [Project Structure](#project-structure)
- [License](#license)

## Screenshots

**Dashboard** — Repository statistics and quick actions
![Dashboard](images/首页.png)

**Repository List** — Search, filter, batch operations
![Repository List](images/仓库列表.png)

**Repository Detail** — Repository info and clone URLs
![Repository Detail](images/仓库详情.png)

**Authors** — Browse repositories grouped by author
![Authors](images/作者.png)

**Activities** — Operation log tracking
![Activities](images/活动.png)

**Settings** — Proxy, token, and system configuration
![Settings](images/设置.png)

## Features

- **Repository Management** — Add, clone, pull, and reset GitHub/Gitee repositories
- **Batch Operations** — Batch clone uncloned repos, scan local directories for auto-import
- **File Browser** — Browse repository file trees and contents (multi-language syntax highlighting, Markdown rendering, image/audio/video preview)
- **Status Monitoring** — Real-time clone progress, update diffs, and commit history
- **Mirror Acceleration** — Built-in Git Smart HTTP server for LAN-accelerated cloning
- **MCP Integration** — Expose MCP tool interfaces for AI clients (Claude Code, Cursor, etc.) to operate repos directly
- **Proxy Support** — HTTP/SOCKS5 proxy with per-platform configuration
- **Statistics Dashboard** — Repository and author statistics with operation log tracking

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.25, Gin, SQLite (modernc), Cobra |
| Frontend | Vue 3, Naive UI, CodeMirror, Pinia, ECharts |
| Build | Vite 7, go:embed |
| Deployment | Docker, GitHub Actions |

## Quick Start

### Prerequisites

- Go 1.25+
- Node.js `^20.19.0` or `>=22.12.0` (frontend dev only)

### Build

```bash
# Build frontend
cd web && npm install && npm run build && cd ..

# Build backend (embeds frontend into binary)
go build -o forks
```

### Run

```bash
./forks serve                # Default port :8080
./forks serve -p 9090        # Custom port
./forks serve --token xxx    # Enable Bearer Token auth
```

**Environment Variables:**

| Variable | Description | Default |
|---|---|---|
| `FORKS_PORT` | Service port | 8080 |
| `FORKS_ADDRESS` | Listen address | 0.0.0.0 |
| `FORKS_HOME` | Data directory | `~/.cicbyte/forks/` |
| `FORKS_REPO_PATH` | Repository storage path | Interactive prompt |
| `TZ` | Timezone | UTC |

Priority: CLI flags > environment variables > defaults. On first run, you'll be prompted for the repository storage path.

### Frontend Development

```bash
cd web && npm install && npm run dev
# Dev server at http://localhost:3000, auto-proxies API to :8080
```

## Mirror Accelerated Cloning

Forks includes a built-in Git Smart HTTP server. Once deployed, other machines on the LAN can clone cached repositories directly without accessing the internet.

```bash
# Direct git clone (origin will point to mirror address)
git clone http://<server-ip>:8080/git/github/author/repo.git

# Recommended: use forks-cli (auto-fixes remote URL)
forks clone http://<server-ip>:8080/git/github/author/repo.git
```

## Docker Deployment

### Pre-built Image (Recommended)

```bash
docker run -d \
  -p 8080:8080 \
  -e TZ=Asia/Shanghai \
  -v ./data:/data \
  ghcr.io/cicbyte/forks:latest
```

Using `docker-compose.yml`:

```bash
docker compose up -d
```

### Build from Source

```bash
docker build -t forks .
docker run -d -p 8080:8080 -v ./data:/data forks
```

## MCP Integration

Expose tool interfaces via [MCP (Model Context Protocol)](https://modelcontextprotocol.io) Streamable HTTP, enabling AI clients to operate repositories directly.

**Endpoint:** `http://<server>:8080/mcp`

**Available Tools:**

| Tool | Description |
|---|---|
| `list_repos` | List repositories (search/filter/pagination) |
| `add_repo` | Add a repository |
| `get_repo` | Get single repository details |
| `update_repo_info` | Update repository remote info |
| `get_stats` | Get repository statistics |
| `list_repo_files` | List repository file structure |
| `read_repo_file` | Read repository file content |

Authentication uses Bearer Token, same as the REST API.

## Companion Tools

### forks-cli

Standalone CLI tool providing mirror-accelerated cloning, batch backup, and more. See [cicbyte/forks-cli](https://github.com/cicbyte/forks-cli).

```bash
# Mirror-accelerated clone (auto-fixes remote URL)
forks clone http://<server-ip>:8080/git/github/torvalds/linux.git

# Batch backup
forks backup --server http://<server-ip>:8080 --dir /data/backup
```

## Project Structure

```
├── main.go              # Entry point, embeds web/dist
├── cmd/
│   ├── root.go          # CLI root command, DB and log initialization
│   └── serve.go         # HTTP routes and business logic
├── common/              # Global variables (DB, log file)
├── models/              # Data structure definitions
├── utils/               # Config, GitHub API, MCP service, utilities
├── assets/              # Embedded resource bridge
└── web/                 # Vue 3 frontend
    └── src/
        ├── api/         # API call wrappers
        ├── views/       # Page components
        ├── components/  # Shared components
        ├── stores/      # Pinia state management
        └── router/      # Router configuration
```

## License

[MIT](./LICENSE)
