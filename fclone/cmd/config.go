package cmd

import (
	"fclone/internal"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config [子命令]",
	Short: "查看或修改配置",
	Long: `查看或修改 fclone 配置。

子命令:
  server <url>     设置镜像服务器地址
  token <value>    设置 API Token
  show             显示当前配置（默认行为）`,
	Args: cobra.MaximumNArgs(3),
	RunE: runConfigShow,
}

var configServerCmd = &cobra.Command{
	Use:   "server <url>",
	Short: "设置镜像服务器地址",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigServer,
}

var configTokenCmd = &cobra.Command{
	Use:   "token <value>",
	Short: "设置 API Token",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigToken,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "显示当前配置",
	Args:  cobra.NoArgs,
	RunE:  runConfigShow,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configServerCmd)
	configCmd.AddCommand(configTokenCmd)
	configCmd.AddCommand(configShowCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cfg := internal.LoadConfig()
	path := internal.GetConfigPath()

	fmt.Printf("配置文件: %s\n", path)
	if cfg.Server != "" {
		fmt.Printf("server: %s\n", cfg.Server)
	} else {
		fmt.Println("server: (未设置)")
	}
	if cfg.Token != "" {
		fmt.Printf("token:  %s\n", internal.MaskToken(cfg.Token))
	} else {
		fmt.Println("token:  (未设置)")
	}
	return nil
}

func runConfigServer(cmd *cobra.Command, args []string) error {
	cfg := internal.LoadConfig()
	cfg.Server = strings.TrimSuffix(args[0], "/")
	if err := internal.SaveConfig(cfg); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}
	fmt.Printf("已保存 server: %s\n", cfg.Server)
	return nil
}

func runConfigToken(cmd *cobra.Command, args []string) error {
	cfg := internal.LoadConfig()
	cfg.Token = args[0]
	if err := internal.SaveConfig(cfg); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}
	fmt.Printf("已保存 token: %s\n", internal.MaskToken(cfg.Token))
	return nil
}
