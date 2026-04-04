package cmd

import (
	"fbackup/internal"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	flagToken       string
	flagServer      string
	flagDir         string
	flagConcurrency int
)

var rootCmd = &cobra.Command{
	Use:   "fbackup [flags]",
	Short: "从 Forks 服务端批量备份仓库到本地",
	Long: `fbackup - 从 Forks 服务端批量备份仓库到本地

从 Forks 服务端读取已克隆的仓库列表，然后批量备份到本地。
已存在的仓库执行 git pull --ff-only，不存在的仓库执行 git clone。

使用前需先配置服务端地址:
  fbackup config server http://192.168.1.100:8080

Token 优先级: --token 参数 > FORKS_TOKEN 环境变量 > 配置文件`,
	Version: internal.Version,
	Args:    cobra.NoArgs,
	RunE:    runBackup,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本号",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("fbackup %s\n", internal.Version)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVarP(&flagToken, "token", "t", "", "本次使用的 API Token（不保存）")
	rootCmd.Flags().StringVarP(&flagServer, "server", "s", "", "本次使用的服务端地址（不保存）")
	rootCmd.Flags().StringVarP(&flagDir, "dir", "d", "", "本地备份目录（默认 ./backup）")
	rootCmd.Flags().IntVarP(&flagConcurrency, "concurrency", "c", 5, "并发数")
	rootCmd.AddCommand(versionCmd)
}

func runBackup(cmd *cobra.Command, args []string) error {
	cfg := internal.LoadConfig()

	// 命令行 --server 临时覆盖
	server := cfg.Server
	if flagServer != "" {
		server = strings.TrimSuffix(flagServer, "/")
	}

	if server == "" {
		return fmt.Errorf("请先配置服务端地址: fbackup config server <url>")
	}

	// token 优先级: 命令行 > 环境变量 > 配置文件
	token := flagToken
	if token == "" {
		token = os.Getenv("FORKS_TOKEN")
	}
	if token == "" {
		token = cfg.Token
	}

	// 获取仓库列表
	fmt.Printf("正在从 %s 获取仓库列表...\n", server)
	repos, err := internal.FetchRepoList(server, token)
	if err != nil {
		return err
	}
	fmt.Printf("获取到 %d 个仓库\n", len(repos))

	// 备份目录优先级: 命令行 > 配置文件 > 默认值
	dir := flagDir
	if dir == "" {
		dir = cfg.Dir
	}
	if dir == "" {
		dir = "./backup"
	}

	// 执行备份
	return internal.BackupRepos(repos, server, dir, flagConcurrency)
}
