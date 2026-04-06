<template>
  <div class="tasks-view">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">任务管理</h1>
      <p class="page-description">查看和管理所有后台任务</p>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <div class="filter-left">
        <n-select
          v-model:value="selectedStatus"
          :options="statusOptions"
          placeholder="任务状态"
          clearable
          size="small"
          style="width: 130px"
          @update:value="handleStatusFilter"
        />
      </div>
      <div class="filter-right">
        <n-button
          size="small"
          ghost
          @click="loadTasks"
        >
          <template #icon>
            <n-icon><RefreshOutline /></n-icon>
          </template>
          刷新
        </n-button>
        <n-button
          type="warning"
          size="small"
          ghost
          @click="handleClearCompleted"
        >
          <template #icon>
            <n-icon><TrashOutline /></n-icon>
          </template>
          清空已完成
        </n-button>
      </div>
    </div>

    <!-- 加载骨架屏 -->
    <div v-if="loading" class="tasks-loading">
      <div class="loading-item" v-for="i in 5" :key="i">
        <div class="loading-icon"></div>
        <div class="loading-content">
          <div class="loading-line loading-title"></div>
          <div class="loading-line loading-desc"></div>
        </div>
      </div>
    </div>

    <!-- 任务列表 -->
    <div v-else-if="filteredTasks.length > 0" class="tasks-list">
      <div
        v-for="task in filteredTasks"
        :key="task.id"
        class="task-card"
      >
        <div class="task-header">
          <div class="task-header-left">
            <span class="task-type">{{ getTaskTypeName(task.type) }}</span>
            <n-tag
              :type="getStatusTagType(task.status)"
              size="small"
              round
            >
              {{ getStatusLabel(task.status) }}
            </n-tag>
          </div>
          <div class="task-header-right">
            <span class="task-time">{{ formatTime(task.created_at) }}</span>
            <n-button
              v-if="task.status === 'running'"
              text
              size="tiny"
              @click="handlePause(task.id)"
            >
              暂停
            </n-button>
            <n-button
              v-if="task.status === 'paused'"
              text
              size="tiny"
              @click="handleResume(task.id)"
            >
              继续
            </n-button>
            <n-button
              v-if="task.status === 'running' || task.status === 'paused'"
              text
              size="tiny"
              @click="handleCancel(task.id)"
            >
              取消
            </n-button>
            <n-button
              v-if="['completed', 'failed', 'cancelled'].includes(task.status)"
              text
              size="tiny"
              @click="handleRetry(task.id)"
            >
              重跑
            </n-button>
            <n-button
              text
              size="tiny"
              @click="handleExpand(task.id)"
            >
              详情
            </n-button>
            <n-button
              v-if="task.status !== 'running' && task.status !== 'paused'"
              text
              size="tiny"
              @click="handleDelete(task.id)"
            >
              删除
            </n-button>
          </div>
        </div>

        <!-- 进度条（running/paused 时显示） -->
        <div v-if="task.status === 'running' || task.status === 'paused'" class="task-progress">
          <n-progress
            :percentage="getProgress(task)"
            :show-indicator="false"
            :height="6"
            :border-radius="3"
          />
        </div>

        <!-- 统计信息 -->
        <div class="task-stats">
          <span class="stat-item">
            <span class="stat-label">总数</span>
            <span class="stat-value">{{ task.total || 0 }}</span>
          </span>
          <span class="stat-item stat-success">
            <span class="stat-label">成功</span>
            <span class="stat-value">{{ task.success_count || 0 }}</span>
          </span>
          <span class="stat-item stat-fail">
            <span class="stat-label">失败</span>
            <span class="stat-value">{{ task.fail_count || 0 }}</span>
          </span>
        </div>
      </div>

      <!-- 分页 -->
      <div class="pagination">
        <n-pagination
          v-model:page="currentPage"
          :page-count="totalPages"
          :page-size="pageSize"
          show-size-picker
          :page-sizes="[10, 20, 50]"
          @update:page="handlePageChange"
          @update:page-size="handlePageSizeChange"
        />
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else class="tasks-empty">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
        <rect x="3" y="3" width="18" height="18" rx="2"/>
        <line x1="8" y1="8" x2="16" y2="8"/>
        <line x1="8" y1="12" x2="16" y2="12"/>
        <line x1="8" y1="16" x2="12" y2="16"/>
      </svg>
      <div class="empty-title">暂无任务</div>
      <div class="empty-desc">后台任务将在这里显示</div>
    </div>

    <!-- 任务详情抽屉 -->
    <n-drawer v-model:show="drawerVisible" :width="480" placement="right">
      <n-drawer-content :title="drawerTitle" closable>
        <div v-if="detailLoading" class="drawer-loading">加载中...</div>
        <template v-else-if="taskDetail">
          <!-- 任务概要 -->
          <div class="drawer-summary">
            <div class="drawer-summary-row">
              <span class="drawer-summary-label">类型</span>
              <span class="drawer-summary-value">{{ getTaskTypeName(taskDetail.type) }}</span>
            </div>
            <div class="drawer-summary-row">
              <span class="drawer-summary-label">状态</span>
              <n-tag :type="getStatusTagType(taskDetail.status)" size="small" round>
                {{ getStatusLabel(taskDetail.status) }}
              </n-tag>
            </div>
            <div class="drawer-summary-row">
              <span class="drawer-summary-label">进度</span>
              <span class="drawer-summary-value">{{ taskDetail.success_count || 0 }} / {{ taskDetail.total || 0 }}</span>
            </div>
            <div class="drawer-summary-row" v-if="taskDetail.created_at">
              <span class="drawer-summary-label">创建时间</span>
              <span class="drawer-summary-value">{{ taskDetail.created_at }}</span>
            </div>
          </div>

          <!-- 子任务列表 -->
          <div class="drawer-section-title">子任务列表</div>
          <div v-if="taskDetail.items && taskDetail.items.length > 0" class="drawer-list">
            <div
              v-for="item in taskDetail.items"
              :key="item.id"
              class="drawer-item"
            >
              <span class="drawer-item-name">{{ item.repo_name || item.name || '-' }}</span>
              <n-tag :type="getStatusTagType(item.status)" size="tiny" round>
                {{ getStatusLabel(item.status) }}
              </n-tag>
              <span v-if="item.message" class="drawer-item-msg">{{ item.message }}</span>
            </div>
          </div>
          <div v-else class="drawer-empty">暂无子任务</div>
        </template>
      </n-drawer-content>
    </n-drawer>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useMessage, NSelect, NButton, NIcon, NTag, NPagination, NProgress, NDrawer, NDrawerContent, useDialog } from 'naive-ui'
import { TrashOutline, RefreshOutline } from '@vicons/ionicons5'
import { getTaskList, getTaskDetail, deleteTask, clearCompletedTasks, pauseTask, resumeTask, cancelTask, retryTask, tasksStreamSSE } from '@/api/repos'

const message = useMessage()
const dialog = useDialog()

// 响应式数据
const loading = ref(false)
const tasks = ref([])
const selectedStatus = ref(null)
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const drawerVisible = ref(false)
const drawerTitle = ref('')
const taskDetail = ref(null)
const detailLoading = ref(false)

// SSE 连接
let sseConnection = null

// 状态筛选选项
const statusOptions = [
  { label: '运行中', value: 'running' },
  { label: '已暂停', value: 'paused' },
  { label: '已完成', value: 'completed' },
  { label: '失败', value: 'failed' },
  { label: '已取消', value: 'cancelled' }
]

// 计算属性
const totalPages = computed(() => Math.ceil(total.value / pageSize.value) || 1)

const filteredTasks = computed(() => {
  if (!selectedStatus.value) {
    return tasks.value
  }
  return tasks.value.filter(t => t.status === selectedStatus.value)
})

const hasRunning = computed(() => tasks.value.some(t => t.status === 'running' || t.status === 'paused'))

// SSE 消息处理：用 SSE 推送的数据更新本地列表
const handleSSETasks = (sseTasks) => {
  // SSE 推送的是最新任务列表，直接替换
  tasks.value = sseTasks
  total.value = sseTasks.length
}

// 方法
const loadTasks = async () => {
  try {
    loading.value = true
    const response = await getTaskList({
      page: currentPage.value,
      page_size: pageSize.value
    })
    const res = response.data || response
    if (res.code === 0 && res.data) {
      tasks.value = res.data.list || []
      total.value = res.data.total || 0
    }
  } catch (error) {
    console.error('加载任务列表失败:', error)
    message.error('加载任务列表失败')
  } finally {
    loading.value = false
  }
}

const handleStatusFilter = () => {
  // 前端筛选，无需重新请求
}

const handlePageChange = (page) => {
  currentPage.value = page
  loadTasks()
}

const handlePageSizeChange = (size) => {
  pageSize.value = size
  currentPage.value = 1
  loadTasks()
}

const handleExpand = async (id) => {
  // 从 tasks 中找到对应任务获取类型名
  const task = tasks.value.find(t => t.id === id)
  drawerTitle.value = task ? getTaskTypeName(task.type) + ' #' + id : '任务详情'
  drawerVisible.value = true
  taskDetail.value = null
  detailLoading.value = true
  try {
    const response = await getTaskDetail(id)
    const res = response.data || response
    if (res.code === 0 && res.data) {
      taskDetail.value = res.data
    }
  } catch (error) {
    console.error('加载任务详情失败:', error)
    message.error('加载任务详情失败')
  } finally {
    detailLoading.value = false
  }
}

const handleDelete = (id) => {
  dialog.warning({
    title: '确认删除',
    content: '确定要删除该任务吗？',
    positiveText: '确认删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await deleteTask(id)
        message.success('任务已删除')
        if (drawerVisible.value) {
          drawerVisible.value = false
          taskDetail.value = null
        }
        loadTasks()
      } catch (error) {
        console.error('删除任务失败:', error)
        message.error('删除任务失败')
      }
    }
  })
}

const handleClearCompleted = () => {
  dialog.warning({
    title: '确认清空',
    content: '确定要清空所有已完成的任务吗？',
    positiveText: '确认清空',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await clearCompletedTasks()
        message.success('已完成任务已清空')
        loadTasks()
      } catch (error) {
        console.error('清空任务失败:', error)
        message.error('清空失败')
      }
    }
  })
}

const handlePause = async (id) => {
  try {
    await pauseTask(id)
    message.success('任务已暂停')
    loadTasks()
  } catch (error) {
    console.error('暂停任务失败:', error)
    message.error('暂停失败')
  }
}

const handleResume = async (id) => {
  try {
    await resumeTask(id)
    message.success('任务已恢复')
    loadTasks()
    startSSE()
  } catch (error) {
    console.error('恢复任务失败:', error)
    message.error('恢复失败')
  }
}

const handleCancel = (id) => {
  dialog.warning({
    title: '确认取消',
    content: '确定要取消该任务吗？未完成的项目将被标记为已取消。',
    positiveText: '确认取消',
    negativeText: '返回',
    onPositiveClick: async () => {
      try {
        await cancelTask(id)
        message.success('任务已取消')
        loadTasks()
      } catch (error) {
        console.error('取消任务失败:', error)
        message.error('取消失败')
      }
    }
  })
}

const handleRetry = async (id) => {
  try {
    const response = await retryTask(id)
    const res = response.data || response
    if (res.code === 0) {
      message.success(res.message || '任务已重新开始')
      loadTasks()
      startSSE()
    } else {
      message.warning(res.message || '重跑失败')
    }
  } catch (error) {
    console.error('重跑任务失败:', error)
    message.error('重跑失败')
  }
}

const getProgress = (task) => {
  const total = task.total || 0
  if (total === 0) return 0
  const done = (task.success_count || 0) + (task.fail_count || 0)
  return Math.round((done / total) * 100)
}

const getTaskTypeName = (type) => {
  const map = {
    batch_clone: '批量克隆',
    batch_pull: '批量拉取',
    scan: '扫描仓库',
    pull: '拉取仓库',
    clone: '克隆仓库'
  }
  return map[type] || type || '未知任务'
}

const getStatusTagType = (status) => {
  const map = {
    running: 'info',
    paused: 'warning',
    completed: 'success',
    failed: 'error',
    cancelled: 'warning',
    pending: 'default'
  }
  return map[status] || 'default'
}

const getStatusLabel = (status) => {
  const map = {
    running: '运行中',
    paused: '已暂停',
    completed: '已完成',
    failed: '失败',
    cancelled: '已取消',
    pending: '等待中'
  }
  return map[status] || status
}

const formatTime = (timeStr) => {
  if (!timeStr) return ''
  const date = new Date(timeStr.replace(/-/g, '/'))
  const now = new Date()
  const diff = now - date
  const minutes = Math.floor(diff / 1000 / 60)
  const hours = Math.floor(diff / 1000 / 60 / 60)
  const days = Math.floor(diff / 1000 / 60 / 60 / 24)

  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes} 分钟前`
  if (hours < 24) return `${hours} 小时前`
  if (days < 7) return `${days} 天前`

  return timeStr.replace('T', ' ').substring(0, 16)
}

// 启动 SSE 连接（替代轮询）
const startSSE = () => {
  stopSSE()
  sseConnection = tasksStreamSSE(
    (sseTasks) => {
      handleSSETasks(sseTasks)
    },
    (err) => {
      console.error('[SSE] 任务列表连接错误:', err)
    }
  )
}

const stopSSE = () => {
  if (sseConnection) {
    sseConnection.close()
    sseConnection = null
  }
}

// 生命周期
onMounted(() => {
  loadTasks()
  startSSE()
})

onUnmounted(() => {
  stopSSE()
})
</script>

<style scoped>
/* ============================================
   PAGE HEADER
   ============================================ */

.tasks-view {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
  animation: fadeIn 0.3s ease-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.page-header {
  margin-bottom: var(--space-2);
}

.page-title {
  font-size: 28px;
  font-weight: var(--font-bold);
  color: var(--color-text-primary);
  margin: 0 0 var(--space-1) 0;
}

.page-description {
  font-size: var(--text-base);
  color: var(--color-text-secondary);
  margin: 0;
}

/* ============================================
   FILTER BAR
   ============================================ */

.filter-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-3) var(--space-4);
  background-color: var(--color-bg-card);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg);
}

.filter-right {
  display: flex;
  gap: var(--space-2);
}

/* ============================================
   TASKS LIST
   ============================================ */

.tasks-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.task-card {
  padding: var(--space-4);
  background-color: var(--color-bg-card);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg);
  transition: all 0.2s ease;
}

.task-card:hover {
  border-color: var(--color-border);
  box-shadow: var(--shadow-sm);
}

.task-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.task-header-left {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.task-header-right {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--text-xs);
  color: var(--color-text-tertiary);
}

.task-type {
  font-size: var(--text-base);
  font-weight: var(--font-semibold);
  color: var(--color-text-primary);
}

.task-time {
  color: var(--color-text-tertiary);
}

/* ============================================
   TASK PROGRESS
   ============================================ */

.task-progress {
  margin-top: var(--space-3);
}

/* ============================================
   TASK STATS
   ============================================ */

.task-stats {
  display: flex;
  gap: var(--space-4);
  margin-top: var(--space-3);
  font-size: var(--text-sm);
}

.stat-item {
  display: flex;
  align-items: center;
  gap: var(--space-1);
}

.stat-label {
  color: var(--color-text-tertiary);
}

.stat-value {
  font-weight: var(--font-semibold);
  color: var(--color-text-secondary);
}

.stat-success .stat-value {
  color: var(--color-success-600);
}

.stat-fail .stat-value {
  color: var(--color-error-500);
}

/* ============================================
   DRAWER DETAIL
   ============================================ */

.drawer-loading {
  text-align: center;
  color: var(--color-text-tertiary);
  font-size: var(--text-sm);
  padding: var(--space-8) 0;
}

.drawer-summary {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  padding: var(--space-3) var(--space-4);
  background-color: var(--color-bg-card);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg);
  margin-bottom: var(--space-4);
}

.drawer-summary-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: var(--text-sm);
}

.drawer-summary-label {
  color: var(--color-text-tertiary);
}

.drawer-summary-value {
  color: var(--color-text-primary);
  font-weight: var(--font-semibold);
}

.drawer-section-title {
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--color-text-primary);
  margin-bottom: var(--space-2);
}

.drawer-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.drawer-item {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--text-sm);
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-md);
  transition: background-color 0.15s;
}

.drawer-item:hover {
  background-color: var(--color-bg-hover, var(--color-gray-50));
}

.drawer-item-name {
  flex: 1;
  color: var(--color-text-primary);
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.drawer-item-msg {
  font-size: var(--text-xs);
  color: var(--color-text-tertiary);
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.drawer-empty {
  text-align: center;
  color: var(--color-text-tertiary);
  font-size: var(--text-sm);
  padding: var(--space-8) 0;
}

/* ============================================
   LOADING STATE
   ============================================ */

.tasks-loading {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.loading-item {
  display: flex;
  gap: var(--space-4);
  padding: var(--space-4);
  background-color: var(--color-bg-card);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg);
}

.loading-icon {
  width: 40px;
  height: 40px;
  flex-shrink: 0;
  background: linear-gradient(90deg, var(--color-gray-200) 25%, var(--color-gray-300) 50%, var(--color-gray-200) 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
  border-radius: var(--radius-full);
}

.loading-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  padding-top: var(--space-1);
}

.loading-line {
  background: linear-gradient(90deg, var(--color-gray-200) 25%, var(--color-gray-300) 50%, var(--color-gray-200) 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
  border-radius: var(--radius-sm);
}

.loading-title {
  width: 60%;
  height: 16px;
}

.loading-desc {
  width: 80%;
  height: 12px;
}

@keyframes shimmer {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}

/* ============================================
   EMPTY STATE
   ============================================ */

.tasks-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-16) var(--space-6);
  background-color: var(--color-bg-card);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg);
  text-align: center;
}

.tasks-empty svg {
  width: 64px;
  height: 64px;
  color: var(--color-text-tertiary);
  margin-bottom: var(--space-4);
  opacity: 0.6;
}

.empty-title {
  font-size: var(--text-lg);
  font-weight: var(--font-semibold);
  color: var(--color-text-primary);
  margin-bottom: var(--space-2);
}

.empty-desc {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
}

/* ============================================
   PAGINATION
   ============================================ */

.pagination {
  display: flex;
  justify-content: center;
  padding: var(--space-4) 0;
}

</style>
