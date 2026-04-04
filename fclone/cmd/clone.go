package cmd

import (
	"fclone/internal"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func runClone(cmd *cobra.Command, args []string) error {
	repoArg := args[0]
	targetDir := ""
	if len(args) > 1 {
		targetDir = args[1]
	}

	cfg := internal.LoadConfig()

	// 命令行 --server 临时覆盖（不保存）
	if flagServer != "" {
		cfg.Server = strings.TrimSuffix(flagServer, "/")
	}

	// token 优先级: 命令行 > 环境变量 > 配置文件
	token := flagToken
	if token == "" {
		token = os.Getenv("FORKS_TOKEN")
	}
	if token == "" {
		token = cfg.Token
	}

	// 解析输入参数
	info, err := internal.ResolveRepoInfo(repoArg, cfg.Server)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	// 默认目标目录
	if targetDir == "" {
		targetDir = strings.TrimSuffix(info.Repo, ".git")
	}

	flagForce, _ := cmd.Flags().GetBool("force")

	// 启动带动画的克隆流程
	return internal.RunCloneUI(info, targetDir, token, flagForce)
}
