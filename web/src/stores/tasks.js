import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getTaskList, getTaskDetail, deleteTask, clearCompletedTasks, tasksStreamSSE } from '@/api/repos'

export const useTasksStore = defineStore('tasks', () => {
  const taskList = ref([])
  const loading = ref(false)
  const total = ref(0)

  // 当前正在进行的任务详情
  const activeTaskDetail = ref(null)

  // SSE 连接实例
  let sseConnection = null
  // SSE 引用计数（多个组件可能同时需要）
  let sseRefCount = 0

  // running 任务数量
  const runningCount = computed(() => taskList.value.filter(t => t.status === 'running' || t.status === 'paused').length)

  // 是否有 running 任务
  const hasRunning = computed(() => runningCount.value > 0)

  // 加载任务列表
  async function fetchTasks(params = {}) {
    try {
      loading.value = true
      const response = await getTaskList(params)
      const res = response.data || response
      if (res.code === 0 && res.data) {
        taskList.value = res.data.list || []
        total.value = res.data.total || 0
      }
    } catch (e) {
      console.error('获取任务列表失败:', e)
    } finally {
      loading.value = false
    }
  }

  // 加载任务详情
  async function fetchTaskDetail(id) {
    try {
      const response = await getTaskDetail(id)
      const res = response.data || response
      if (res.code === 0 && res.data) {
        activeTaskDetail.value = res.data
      }
    } catch (e) {
      console.error('获取任务详情失败:', e)
    }
  }

  // 删除单个任务
  async function removeTask(id) {
    try {
      await deleteTask(id)
      taskList.value = taskList.value.filter(t => t.id !== id)
      if (activeTaskDetail.value?.id === id) {
        activeTaskDetail.value = null
      }
    } catch (e) {
      console.error('删除任务失败:', e)
    }
  }

  // 清空已完成任务
  async function clearCompleted() {
    try {
      await clearCompletedTasks()
      taskList.value = taskList.value.filter(t => t.status === 'running' || t.status === 'pending' || t.status === 'paused')
    } catch (e) {
      console.error('清空任务失败:', e)
    }
  }

  // 启动 SSE 连接
  function startSSE() {
    sseRefCount++
    if (sseConnection) return // 已有连接，复用

    sseConnection = tasksStreamSSE(
      (sseTasks) => {
        taskList.value = sseTasks
        total.value = sseTasks.length
      },
      (err) => {
        console.error('[SSE Store] 连接错误:', err)
      }
    )
  }

  // 停止 SSE 连接（引用计数归零时才真正关闭）
  function stopSSE() {
    sseRefCount = Math.max(0, sseRefCount - 1)
    if (sseRefCount === 0 && sseConnection) {
      sseConnection.close()
      sseConnection = null
    }
  }

  // 兼容旧接口
  function startPolling() {
    startSSE()
  }

  function stopPolling() {
    stopSSE()
  }

  return {
    taskList,
    loading,
    total,
    activeTaskDetail,
    runningCount,
    hasRunning,
    fetchTasks,
    fetchTaskDetail,
    removeTask,
    clearCompleted,
    startSSE,
    stopSSE,
    startPolling,
    stopPolling
  }
})
