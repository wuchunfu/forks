package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"forks.com/m/common"
)

// resumeRunningTasks 服务启动时将未完成任务标记为暂停，由用户手动恢复
func resumeRunningTasks() {
	// 将所有 running 的任务改为 paused（运行中的 item 回退为 pending）
	result, err := common.Db.Exec(`UPDATE tasks SET status = 'paused', updated_at = strftime('%Y-%m-%d %H:%M:%S','now','localtime') WHERE status = 'running'`)
	if err == nil {
		if n, _ := result.RowsAffected(); n > 0 {
			log.Printf("⏸️ [启动恢复] 已将 %d 个运行中任务标记为暂停", n)
		}
	}
	// running 的 task_items 也回退为 pending
	common.Db.Exec(`UPDATE task_items SET status = 'pending', message = '' WHERE status = 'running'`)
}

// pauseTask 暂停运行中的任务
func pauseTask(c *gin.Context) {
	id := c.Param("id")
	var status, taskType string
	err := common.Db.QueryRow("SELECT status, type FROM tasks WHERE id = ?", id).Scan(&status, &taskType)
	if err != nil {
		c.JSON(404, gin.H{"code": 404, "message": "任务不存在"})
		return
	}
	if status != "running" {
		c.JSON(400, gin.H{"code": 400, "message": "只能暂停运行中的任务"})
		return
	}
	// 把当前 running 的 task_item 回退为 pending
	common.Db.Exec(`UPDATE task_items SET status = 'pending', message = '' WHERE task_id = ? AND status = 'running'`, id)
	common.Db.Exec(`UPDATE tasks SET status = 'paused', updated_at = strftime('%Y-%m-%d %H:%M:%S', 'now','localtime') WHERE id = ?`, id)
	c.JSON(200, gin.H{"code": 0, "message": "任务已暂停"})
}

// resumeTask 恢复暂停的任务
func resumeTask(c *gin.Context) {
	id := c.Param("id")
	var status, taskType string
	err := common.Db.QueryRow("SELECT status, type FROM tasks WHERE id = ?", id).Scan(&status, &taskType)
	if err != nil {
		c.JSON(404, gin.H{"code": 404, "message": "任务不存在"})
		return
	}
	if status != "paused" {
		c.JSON(400, gin.H{"code": 400, "message": "只能恢复暂停中的任务"})
		return
	}

	taskID, _ := strconv.ParseInt(id, 10, 64)

	// 检查是否有活跃的 goroutine
	if _, loaded := taskCancelChans.Load(taskType); loaded {
		// goroutine 还在，只需改 DB 状态，goroutine 会自动检测
		common.Db.Exec(`UPDATE tasks SET status = 'running', updated_at = strftime('%Y-%m-%d %H:%M:%S', 'now','localtime') WHERE id = ?`, id)
	} else {
		// goroutine 不存在（重启后），重新启动
		if taskType == "batch_update_info" {
			restartBatchUpdateInfoTask(taskID)
		} else {
			restartBatchPullTask(taskID)
		}
	}

	c.JSON(200, gin.H{"code": 0, "message": "任务已恢复"})
}


// cancelTask 取消运行中或暂停的任务
func cancelTask(c *gin.Context) {
	id := c.Param("id")
	var status, taskType string
	err := common.Db.QueryRow("SELECT status, type FROM tasks WHERE id = ?", id).Scan(&status, &taskType)
	if err != nil {
		c.JSON(404, gin.H{"code": 404, "message": "任务不存在"})
		return
	}
	if status != "running" && status != "paused" {
		c.JSON(400, gin.H{"code": 400, "message": "只能取消运行中或暂停的任务"})
		return
	}
	// 更新 DB 状态
	common.Db.Exec(`UPDATE tasks SET status = 'cancelled', updated_at = strftime('%Y-%m-%d %H:%M:%S', 'now','localtime') WHERE id = ?`, id)
	common.Db.Exec(`UPDATE task_items SET status = 'cancelled', message = '任务已取消' WHERE task_id = ? AND status IN ('pending', 'running')`, id)
	// 发送取消信号给 goroutine
	if ch, ok := taskCancelChans.Load(taskType); ok {
		close(ch.(chan struct{}))
		taskCancelChans.Delete(taskType)
	}
	c.JSON(200, gin.H{"code": 0, "message": "任务已取消"})
}

// retryTask 重跑失败/取消/完成的任务，跳过已成功的 item
func retryTask(c *gin.Context) {
	id := c.Param("id")
	var taskType, status string
	var total, successCount int
	var failCount int
	err := common.Db.QueryRow("SELECT type, status, total, success_count, fail_count FROM tasks WHERE id = ?", id).Scan(&taskType, &status, &total, &successCount, &failCount)
	if err != nil {
		c.JSON(404, gin.H{"code": 404, "message": "任务不存在"})
		return
	}
	if taskType != "batch_pull" && taskType != "batch_update_info" {
		c.JSON(400, gin.H{"code": 400, "message": "不支持的任务类型"})
		return
	}
	if status == "running" || status == "paused" {
		c.JSON(400, gin.H{"code": 400, "message": "任务正在进行中，无法重跑"})
		return
	}

	// 检查是否有 running/paused 的同类型任务
	var existCount int
	common.Db.QueryRow("SELECT COUNT(*) FROM tasks WHERE type = ? AND status IN ('running', 'paused') AND id != ?", taskType, id).Scan(&existCount)
	if existCount > 0 {
		c.JSON(409, gin.H{"code": 409, "message": "已有同类型的运行中任务"})
		return
	}

	// 将 failed/cancelled 的 task_items 重置为 pending
	common.Db.Exec(`UPDATE task_items SET status = 'pending', message = '' WHERE task_id = ? AND status IN ('failed', 'cancelled')`, id)

	// 重新计算 fail_count：查询当前 failed/cancelled 数量
	var pendingCount int
	common.Db.QueryRow("SELECT COUNT(*) FROM task_items WHERE task_id = ? AND status = 'pending'", id).Scan(&pendingCount)
	if pendingCount == 0 {
		c.JSON(200, gin.H{"code": 0, "message": "所有子任务已完成，无需重跑"})
		return
	}

	// 重置任务状态，重新统计成功/失败数
	taskIDInt, _ := strconv.ParseInt(id, 10, 64)
	common.Db.Exec(`UPDATE tasks SET status = 'running',
		success_count = (SELECT COUNT(*) FROM task_items WHERE task_id = ? AND status = 'success'),
		fail_count = (SELECT COUNT(*) FROM task_items WHERE task_id = ? AND status = 'failed'),
		updated_at = strftime('%Y-%m-%d %H:%M:%S', 'now','localtime') WHERE id = ?`, taskIDInt, taskIDInt, taskIDInt)

	// 加载该任务对应的仓库信息
	rows, err := common.Db.Query(`SELECT ti.repo_id, r.author, r.repo, r.source FROM task_items ti
		JOIN repos r ON ti.repo_id = r.id
		WHERE ti.task_id = ? AND ti.status = 'pending' ORDER BY ti.id`, id)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "查询任务项失败"})
		return
	}
	var repos []RepoInfo
	for rows.Next() {
		var r RepoInfo
		if err := rows.Scan(&r.ID, &r.Author, &r.Repo, &r.Source); err == nil {
			repos = append(repos, r)
		}
	}
	rows.Close()

	if len(repos) == 0 {
		c.JSON(200, gin.H{"code": 0, "message": "没有需要重跑的子任务"})
		return
	}

	// 启动 goroutine 执行
	cancelChan := make(chan struct{})
	taskCancelChans.Store(taskType, cancelChan)

	if taskType == "batch_update_info" {
		go executeBatchUpdateInfoTask(taskIDInt, repos, cancelChan)
	} else {
		go executeBatchPullTask(taskIDInt, repos, cancelChan)
	}

	c.JSON(200, gin.H{"code": 0, "message": fmt.Sprintf("任务已重新开始，待处理 %d 项", len(repos))})
}

func getTaskList(c *gin.Context) {
	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if val, err := strconv.Atoi(ps); err == nil && val > 0 && val <= 100 {
			pageSize = val
		}
	}

	offset := (page - 1) * pageSize

	// 构建 WHERE 条件
	args := []interface{}{}
	where := "WHERE 1=1"
	if status := c.Query("status"); status != "" {
		where += " AND status = ?"
		args = append(args, status)
	}

	// 获取总数
	var total int
	countSQL := "SELECT COUNT(*) FROM tasks " + where
	err := common.Db.QueryRow(countSQL, args...).Scan(&total)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "查询任务总数失败"})
		return
	}

	// 获取列表
	querySQL := `SELECT id, type, status, total, success_count, fail_count, error, created_at, updated_at
		FROM tasks ` + where + ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	queryArgs := append(args, pageSize, offset)
	rows, err := common.Db.Query(querySQL, queryArgs...)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "查询任务列表失败"})
		return
	}
	defer rows.Close()

	tasks := []map[string]interface{}{}
	for rows.Next() {
		var id int64
		var taskType, status string
		var totalItems, successCount, failCount int
		var errMsg sql.NullString
		var createdAt string
		var updatedAt sql.NullString

		if err := rows.Scan(&id, &taskType, &status, &totalItems, &successCount, &failCount, &errMsg, &createdAt, &updatedAt); err != nil {
			continue
		}

		task := map[string]interface{}{
			"id":            id,
			"type":          taskType,
			"status":        status,
			"total":         totalItems,
			"success_count": successCount,
			"fail_count":    failCount,
			"created_at":    createdAt,
		}
		if errMsg.Valid {
			task["error"] = errMsg.String
		}
		if updatedAt.Valid {
			task["updated_at"] = updatedAt.String
		}
		tasks = append(tasks, task)
	}

	c.JSON(200, gin.H{
		"code": 0,
		"data": gin.H{
			"list":  tasks,
			"total": total,
			"page":  page,
			"page_size": pageSize,
		},
	})
}

func getTaskDetail(c *gin.Context) {
	id := c.Param("id")

	// 查询任务
	var taskID int64
	var taskType, status string
	var totalItems, successCount, failCount int
	var errMsg sql.NullString
	var createdAt string
	var updatedAt sql.NullString

	err := common.Db.QueryRow(`
		SELECT id, type, status, total, success_count, fail_count, error, created_at, updated_at
		FROM tasks WHERE id = ?`, id).Scan(&taskID, &taskType, &status, &totalItems, &successCount, &failCount, &errMsg, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(404, gin.H{"code": 404, "message": "任务不存在"})
		} else {
			c.JSON(500, gin.H{"code": 500, "message": "查询任务失败"})
		}
		return
	}

	task := map[string]interface{}{
		"id":            taskID,
		"type":          taskType,
		"status":        status,
		"total":         totalItems,
		"success_count": successCount,
		"fail_count":    failCount,
		"created_at":    createdAt,
	}
	if errMsg.Valid {
		task["error"] = errMsg.String
	}
	if updatedAt.Valid {
		task["updated_at"] = updatedAt.String
	}

	// 查询 task_items
	rows, err := common.Db.Query(`
		SELECT id, task_id, repo_id, repo_name, status, message, created_at
		FROM task_items WHERE task_id = ? ORDER BY id`, id)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "查询任务子项失败"})
		return
	}
	defer rows.Close()

	items := []map[string]interface{}{}
	for rows.Next() {
		var itemID int64
		var taskIDFk int64
		var repoID sql.NullInt64
		var repoName, itemStatus string
		var message sql.NullString
		var itemCreatedAt string

		if err := rows.Scan(&itemID, &taskIDFk, &repoID, &repoName, &itemStatus, &message, &itemCreatedAt); err != nil {
			continue
		}

		item := map[string]interface{}{
			"id":         itemID,
			"task_id":    taskIDFk,
			"repo_name":  repoName,
			"status":     itemStatus,
			"created_at": itemCreatedAt,
		}
		if repoID.Valid {
			item["repo_id"] = repoID.Int64
		}
		if message.Valid {
			item["message"] = message.String
		}
		items = append(items, item)
	}

	task["items"] = items

	c.JSON(200, gin.H{
		"code": 0,
		"data": task,
	})
}

// deleteTask 删除任务及其 items
func deleteTask(c *gin.Context) {
	id := c.Param("id")

	result, err := common.Db.Exec("DELETE FROM task_items WHERE task_id = ?", id)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "删除任务子项失败"})
		return
	}

	result, err = common.Db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "删除任务失败"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(404, gin.H{"code": 404, "message": "任务不存在"})
		return
	}

	c.JSON(200, gin.H{"code": 0, "message": "删除成功"})
}

func clearCompletedTasks(c *gin.Context) {
	// 先删除 task_items
	common.Db.Exec(`DELETE FROM task_items WHERE task_id IN (SELECT id FROM tasks WHERE status IN ('completed', 'failed'))`)

	result, err := common.Db.Exec(`DELETE FROM tasks WHERE status IN ('completed', 'failed')`)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "清空任务失败"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	c.JSON(200, gin.H{"code": 0, "message": fmt.Sprintf("已清空 %d 个已完成任务", rowsAffected)})
}
