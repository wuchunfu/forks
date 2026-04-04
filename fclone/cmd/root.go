package cmd

import (
	"fclone/internal"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	flagToken  string
	flagServer string
	flagForce  bool
)

var rootCmd = &cobra.Command{
	Use:   "fclone [flags] <仓库地址> [目标目录]",
	Short: "通过 Forks 镜像加速克隆 Git 仓库",
	Long: `fclone - 通过 Forks 镜像加速克隆 Git 仓库

仓库地址支持三种格式:
  1) 镜像 URL:  http://host:port/git/github/author/repo.git
  2) 原始 URL:  https://github.com/author/repo
  3) 简写:      author/repo 或 github/author/repo

使用简写或原始 URL 时，需先配置镜像服务器:
  fclone config server http://192.168.1.100:8080

Token 优先级: --token 参数 > FORKS_TOKEN 环境变量 > 配置文件`,
	Version: internal.Version,
	Args:    cobra.MinimumNArgs(1),
	RunE:    runClone,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本号",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("fclone %s\n", internal.Version)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVarP(&flagToken, "token", "t", "", "本次使用的 API Token（不保存）")
	rootCmd.Flags().StringVarP(&flagServer, "server", "s", "", "本次使用的镜像服务器（不保存）")
	rootCmd.Flags().BoolVarP(&flagForce, "force", "f", false, "强制更新镜像缓存")
	rootCmd.AddCommand(versionCmd)
}
