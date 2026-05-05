package cmd

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	_ "modernc.org/sqlite"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/cicbyte/forks/common"
	"github.com/cicbyte/forks/models"
	"github.com/cicbyte/forks/utils"
)

func close_resource() {
	// 关闭数据库连接
	if common.Db != nil {
		common.Db.Close()
	}
	// 关闭日志文件
	if common.LogFile != nil {
		common.LogFile.Close()
	}
}

var rootCmd = &cobra.Command{
	Use:   "forks",
	Short: "forks",
	Long:  `Git repository manager`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// 初始化数据库连接
		db_path := utils.ConfigInstance.GetDbPath()
		var err error
		common.Db, err = sql.Open("sqlite", db_path)
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}

		// 检查数据库连接是否正常
		err = common.Db.Ping()
		if err != nil {
			log.Fatalf("Failed to ping database: %v", err)
		}
		// 创建表
		createTableSQL := `CREATE TABLE IF NOT EXISTS repos (
			"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
			"author" TEXT,
			"repo" TEXT,
			"url" TEXT,
			"git_url" TEXT,
			"topics" TEXT,
			"license" TEXT,
			"stars" INTEGER,
			"watching" INTEGER,
			"forks" INTEGER,
			"description" TEXT,
			"README" TEXT,
			"languages" TEXT,
			"source" TEXT,
			"is_cloned" INTEGER DEFAULT 0,
			"last_pulled_at" TEXT,
			"created_at" TEXT,
			"updated_at" TEXT
		);`
		_, err = common.Db.Exec(createTableSQL)
		if err != nil {
			fmt.Printf("Failed to create table: %v\n", err)
			return
		}

		// 创建活动记录表
		createActivitiesTableSQL := `CREATE TABLE IF NOT EXISTS activities (
			"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
			"type" TEXT NOT NULL,
			"title" TEXT NOT NULL,
			"description" TEXT,
			"repo_id" INTEGER,
			"repo_name" TEXT,
			"metadata" TEXT,
			"created_at" TEXT DEFAULT CURRENT_TIMESTAMP
		);`
		_, err = common.Db.Exec(createActivitiesTableSQL)
		if err != nil {
			fmt.Printf("Failed to create activities table: %v\n", err)
			return
		}

		// 兼容旧数据：添加 valid 字段
		_, _ = common.Db.Exec(`ALTER TABLE repos ADD COLUMN valid INTEGER DEFAULT 1`)

		// 创建任务表
		createTasksTableSQL := `CREATE TABLE IF NOT EXISTS tasks (
			"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
			"type" TEXT NOT NULL,
			"status" TEXT DEFAULT 'pending',
			"total" INTEGER DEFAULT 0,
			"success_count" INTEGER DEFAULT 0,
			"fail_count" INTEGER DEFAULT 0,
			"error" TEXT,
			"created_at" TEXT DEFAULT (strftime('%Y-%m-%d %H:%M:%S', 'now','localtime')),
			"updated_at" TEXT
		);`
		_, err = common.Db.Exec(createTasksTableSQL)
		if err != nil {
			fmt.Printf("Failed to create tasks table: %v\n", err)
			return
		}

		// 创建任务子项表
		createTaskItemsTableSQL := `CREATE TABLE IF NOT EXISTS task_items (
			"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
			"task_id" INTEGER NOT NULL,
			"repo_id" INTEGER,
			"repo_name" TEXT NOT NULL,
			"status" TEXT DEFAULT 'pending',
			"message" TEXT,
			"created_at" TEXT DEFAULT (strftime('%Y-%m-%d %H:%M:%S', 'now','localtime'))
		);`
		_, err = common.Db.Exec(createTaskItemsTableSQL)
		if err != nil {
			fmt.Printf("Failed to create task_items table: %v\n", err)
			return
		}

		fmt.Println("Table created successfully")

		fmt.Println("Database connected successfully")

	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		close_resource()
	},
}

func init() {
	// 调用 setupCloseHandler 监听退出信号
	setupCloseHandler()
	// 检查是否已经初始化
	appDir := utils.ConfigInstance.GetAppDir()
	initized_file := filepath.Join(appDir, "initized")
	_, err := os.Stat(initized_file)
	if err != nil && os.IsNotExist(err) {
		// 创建app目录
		if _, err := os.Stat(appDir); os.IsNotExist(err) {
			os.MkdirAll(appDir, 0755)
		}
		// 创建config目录
		configDir := utils.ConfigInstance.GetAppConfigDir()
		if _, err = os.Stat(configDir); os.IsNotExist(err) {
			os.MkdirAll(configDir, 0755)
		}
		// 读取仓库存储路径：环境变量优先，否则交互式输入
		config := models.GitPlusConfig{}
		if storeRootPath := os.Getenv("FORKS_REPO_PATH"); storeRootPath != "" {
			config.StoreRootPath = storeRootPath
		} else {
			reader := bufio.NewReader(os.Stdin)
			fmt.Println("请输入repos存储位置:")
			storeRootPath, _ := reader.ReadString('\n')
			config.StoreRootPath = strings.TrimSpace(storeRootPath)
		}
		// 确保目录存在
		if err = os.MkdirAll(config.StoreRootPath, 0755); err != nil {
			fmt.Printf("创建目录失败: %v\n", err)
			os.Exit(1)
		}
		// 转换为json字符串
		jsonstr, _ := utils.StructToPrettyJSON(config)
		configpath := utils.ConfigInstance.GetAppConfigFile()
		// 写入配置文件
		os.WriteFile(configpath, []byte(jsonstr), 0644)
		// 创建db目录
		dbDir := utils.ConfigInstance.GetDbDir()
		if _, err = os.Stat(dbDir); os.IsNotExist(err) {
			os.MkdirAll(dbDir, 0755)
		}
		// 创建logs目录
		logDir := utils.ConfigInstance.GetLogDir()
		if _, err = os.Stat(logDir); os.IsNotExist(err) {
			os.MkdirAll(logDir, 0755)
		}
		// 创建initized文件, 避免重复初始化
		os.Create(initized_file)

	}
	// 获取当前日期
	today := time.Now().Format("2006-01-02")
	homeDir := utils.ConfigInstance.GetLogDir()
	// 获取今天的日期
	logFilePath := filepath.Join(homeDir, "log_"+today+".log")

	// 打开（或创建）日志文件
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("open log file error:", err)
		os.Exit(1)
	}
	common.LogFile = file
	// 设置Logrus的全局输出为文件
	//logrus.SetOutput(file)
	// 设置日志格式（可选）
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})
}

func Execute() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "显示版本号",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(utils.Version)
		},
	})
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}

// setupCloseHandler 监听程序退出信号
func setupCloseHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nProgram is exiting...")
		close_resource()
		os.Exit(0)
	}()
}
