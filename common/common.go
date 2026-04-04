package common

import (
	"database/sql"
	"os"
)

var (
	// 数据库连接
	Db      *sql.DB
	LogFile *os.File
)
