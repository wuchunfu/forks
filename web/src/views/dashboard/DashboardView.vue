<template>
  <div class="dashboard-view">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">仪表盘</h1>
      <p class="page-description">欢迎使用 Forks 后台管理系统</p>
    </div>

    <!-- 统计卡片行 - Bento Grid 风格 -->
    <div class="stats-grid">
      <!-- 仓库总数 -->
      <div class="stat-card stat-card-primary">
        <div class="stat-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-label">仓库总数</div>
          <div class="stat-value">{{ statistics.total }}</div>
        </div>
      </div>

      <!-- 已克隆 -->
      <div class="stat-card stat-card-success">
        <div class="stat-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="20 6 9 17 4 12"/>
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-label">已克隆</div>
          <div class="stat-value">{{ statistics.cloned }}</div>
        </div>
      </div>

      <!-- 未克隆 -->
      <div class="stat-card stat-card-warning">
        <div class="stat-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <line x1="12" y1="8" x2="12" y2="12"/>
            <line x1="12" y1="16" x2="12.01" y2="16"/>
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-label">未克隆</div>
          <div class="stat-value">{{ statistics.notCloned }}</div>
        </div>
      </div>

      <!-- 今日操作 -->
      <div class="stat-card stat-card-info">
        <div class="stat-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/>
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-label">今日操作</div>
          <div class="stat-value">{{ todayOperations }}</div>
        </div>
      </div>
    </div>

    <!-- 快捷操作区 -->
    <div class="quick-actions">
      <div class="section-title">快捷操作</div>
      <div class="actions-grid">
        <!-- 添加仓库 -->
        <button class="action-card" @click="showAddRepoModal = true">
          <div class="action-icon action-icon-primary">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="12" y1="5" x2="12" y2="19"/>
              <line x1="5" y1="12" x2="19" y2="12"/>
            </svg>
          </div>
          <div class="action-text">
            <div class="action-title">添加仓库</div>
            <div class="action-desc">添加新的 Git 仓库</div>
          </div>
        </button>

        <!-- 扫描本地 -->
        <button class="action-card" @click="handleScanRepos" :disabled="scanLoading">
          <div class="action-icon action-icon-success">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M3 7V5a2 2 0 0 1 2-2h2"/>
              <path d="M17 3h2a2 2 0 0 1 2 2v2"/>
              <path d="M21 17v2a2 2 0 0 1-2 2h-2"/>
              <path d="M7 21H5a2 2 0 0 1-2-2v-2"/>
            </svg>
          </div>
          <div class="action-text">
            <div class="action-title">扫描本地</div>
            <div class="action-desc">扫描本地 Git 仓库</div>
          </div>
          <div v-if="scanLoading" class="action-loading">
            <svg class="loading-spinner" viewBox="0 0 24 24">
              <circle cx="12" cy="12" r="10" fill="none" stroke="currentColor" stroke-width="2" stroke-dasharray="32" stroke-dashoffset="32"/>
            </svg>
          </div>
        </button>

        <!-- 批量克隆 -->
        <button class="action-card" @click="handleBatchClone" :disabled="batchCloneLoading">
          <div class="action-icon action-icon-warning">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
              <polyline points="7 10 12 15 17 10"/>
              <line x1="12" y1="15" x2="12" y2="3"/>
            </svg>
          </div>
          <div class="action-text">
            <div class="action-title">批量克隆</div>
            <div class="action-desc">克隆所有未克隆仓库</div>
          </div>
          <div v-if="batchCloneLoading" class="action-loading">
            <svg class="loading-spinner" viewBox="0 0 24 24">
              <circle cx="12" cy="12" r="10" fill="none" stroke="currentColor" stroke-width="2" stroke-dasharray="32" stroke-dashoffset="32"/>
            </svg>
          </div>
        </button>

        <!-- 刷新列表 -->
        <button class="action-card" @click="refreshRepos" :disabled="loading">
          <div class="action-icon action-icon-info">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="23 4 23 10 17 10"/>
              <polyline points="1 20 1 14 7 14"/>
              <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/>
            </svg>
          </div>
          <div class="action-text">
            <div class="action-title">刷新列表</div>
            <div class="action-desc">更新仓库数据</div>
          </div>
        </button>

        <!-- 一键拉取 -->
        <button class="action-card" @click="handleBatchPull" :disabled="batchPullLoading">
          <div class="action-icon action-icon-success">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
              <polyline points="7 10 12 15 17 10"/>
              <line x1="12" y1="15" x2="12" y2="3"/>
            </svg>
          </div>
          <div class="action-text">
            <div class="action-title">一键拉取</div>
            <div class="action-desc">拉取所有已克隆仓库</div>
          </div>
          <div v-if="batchPullLoading" class="action-loading">
            <svg class="loading-spinner" viewBox="0 0 24 24">
              <circle cx="12" cy="12" r="10" fill="none" stroke="currentColor" stroke-width="2" stroke-dasharray="32" stroke-dashoffset="32"/>
            </svg>
          </div>
        </button>
      </div>
    </div>

    <!-- 最近活动列表 -->
    <div class="recent-activities">
      <div class="section-header">
        <div class="section-title">最近活动</div>
        <button class="view-all-btn" @click="viewAllActivities">查看全部</button>
      </div>

      <div v-if="activitiesLoading" class="activities-loading">
        <div class="loading-skeleton" v-for="i in 5" :key="i"></div>
      </div>

      <div v-else-if="recentActivities.length > 0" class="activities-list">
        <div
          v-for="activity in recentActivities"
          :key="activity.id"
          class="activity-item"
        >
          <div class="activity-icon" :class="`activity-icon-${activity.type}`">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="20 6 9 17 4 12" v-if="activity.type === 'success'"/>
              <circle cx="12" cy="12" r="10" v-else-if="activity.type === 'warning'"/>
              <path d="M18 6L6 18M6 6l12 12" v-else-if="activity.type === 'error'"/>
              <circle cx="12" cy="12" r="10" v-else/>
              <polyline points="12 6 12 12 16 14" v-else/>
            </svg>
          </div>
          <div class="activity-content">
            <div class="activity-title">{{ activity.title }}</div>
            <div class="activity-desc">{{ activity.description }}</div>
            <div class="activity-time">{{ formatTime(activity.time) }}</div>
          </div>
        </div>
      </div>

      <div v-else class="activities-empty">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M9 17H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2v10a2 2 0 0 1-2 2h-4"/>
          <polyline points="9 12 12 15 15 12"/>
        </svg>
        <div>暂无活动记录</div>
      </div>
    </div>

    <!-- 添加仓库对话框 -->
    <AddRepoModal
      :show="showAddRepoModal"
      @update:show="showAddRepoModal = $event"
      @success="handleAddRepoSuccess"
    />

    <!-- 批量克隆选择弹窗 -->
    <BatchCloneModal
      :show="showBatchCloneModal"
      @update:show="showBatchCloneModal = $event"
      @confirm="handleBatchCloneConfirm"
    />

    <!-- 批量克隆进度弹窗 -->
    <n-modal v-model:show="showBatchCloneProgress" :style="{ width: '500px' }" :closable="false" :maskClosable="false">
      <n-card title="批量克隆进度" :bordered="false">
        <div v-if="batchCloneProgress">
          <n-progress
            type="line"
            :percentage="batchCloneProgress.total > 0 ? Math.round((batchCloneProgress.current / batchCloneProgress.total) * 100) : 0"
            :status="batchCloneProgress.completed ? 'success' : 'default'"
          />

          <n-space vertical style="margin-top: 16px">
            <n-text>
              {{ batchCloneProgress.current || 0 }} / {{ batchCloneProgress.total || 0 }}
              <template v-if="batchCloneProgress.repo">
                - {{ batchCloneProgress.repo }}
              </template>
            </n-text>
            <n-text depth="3">{{ batchCloneProgress.status }}</n-text>

            <template v-if="batchCloneProgress.completed">
              <n-divider />
              <n-space>
                <n-tag type="success">成功: {{ batchCloneProgress.cloned }}</n-tag>
                <n-tag type="warning">失败: {{ batchCloneProgress.failed }}</n-tag>
                <n-tag type="error">失效: {{ batchCloneProgress.invalid }}</n-tag>
              </n-space>
            </template>
          </n-space>
        </div>

        <template #footer>
          <n-space justify="end">
            <n-button
              v-if="batchCloneProgress?.completed"
              type="primary"
              @click="showBatchCloneProgress = false"
            >
              完成
            </n-button>
          </n-space>
        </template>
      </n-card>
    </n-modal>

    <!-- 当前运行中的任务 -->
    <div v-if="currentRunningTask" class="running-task-section">
      <div class="section-header">
        <div class="section-title">当前任务</div>
        <span :class="['task-status-badge', `status-${currentRunningTask.status}`]">
          {{ taskStatusName(currentRunningTask.status) }}
        </span>
      </div>
      <div class="running-task-card">
        <div class="running-task-info">
          <span class="running-task-type">{{ taskTypeName(currentRunningTask.type) }}</span>
          <span class="running-task-progress">
            {{ currentRunningTask.success_count + currentRunningTask.fail_count }} / {{ currentRunningTask.total }}
          </span>
        </div>
        <n-progress
          type="line"
          :percentage="currentRunningTask.total > 0 ? Math.round(((currentRunningTask.success_count + currentRunningTask.fail_count) / currentRunningTask.total) * 100) : 0"
          :show-indicator="false"
          status="info"
        />
        <div class="running-task-counts">
          <n-tag type="success" size="small">成功: {{ currentRunningTask.success_count }}</n-tag>
          <n-tag type="warning" size="small">失败: {{ currentRunningTask.fail_count }}</n-tag>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
/**
 * DashboardView - 首页仪表盘
 *
 * Grid Design System - Bento Grid Layout
 * Chromatic Team - Pixel Implementation
 *
 * 功能特性:
 * - 统计卡片展示 (仓库总数、已克隆、未克隆、今日操作)
 * - 快捷操作区 (添加仓库、扫描本地、批量克隆、刷新列表)
 * - 最近活动列表 (展示最近的 git 操作记录)
 * - 响应式设计，适配移动端和桌面端
 */
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage, useModal } from 'naive-ui'
import { useReposStore } from '@/stores/repos'
import { useTasksStore } from '@/stores/tasks'
import { getStats, getActivities } from '@/api/repos'
import AddRepoModal from '@/components/AddRepoModal.vue'
import BatchCloneModal from '@/components/BatchCloneModal.vue'

const props = defineProps({
  refreshKey: { type: Number, default: 0 }
})

const router = useRouter()
const message = useMessage()
const modal = useModal()
const reposStore = useReposStore()
const tasksStore = useTasksStore()

// 响应式数据
const showAddRepoModal = ref(false)
const showBatchCloneModal = ref(false)
const loading = ref(false)
const scanLoading = ref(false)
const batchCloneLoading = ref(false)
const showBatchCloneProgress = ref(false)
const batchCloneProgress = ref(null)
const batchPullLoading = ref(false)

// 统计数据
const statsData = ref({ total: 0, cloned: 0, notCloned: 0 })

// 最近活动数据
const recentActivities = ref([])
const activitiesLoading = ref(false)

// 加载活动记录
const loadActivities = async () => {
  try {
    activitiesLoading.value = true
    const response = await getActivities({ page: 1, page_size: 10 })
    const res = response.data || response
    if (res.code === 0 && res.data?.list) {
      recentActivities.value = res.data.list.map(item => ({
        id: item.id,
        type: item.type,
        title: item.title,
        description: item.description,
        repoName: item.repo_name,
        time: item.created_at
      }))
    }
  } catch (error) {
    console.error('加载活动记录失败:', error)
  } finally {
    activitiesLoading.value = false
  }
}

// 计算属性
const statistics = computed(() => statsData.value)

// 今日操作数（从活动记录计算）
const todayOperations = computed(() => {
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  return recentActivities.value.filter(activity => {
    const activityDate = new Date(activity.time)
    activityDate.setHours(0, 0, 0, 0)
    return activityDate.getTime() === today.getTime()
  }).length
})

// 方法
const loadStatistics = async () => {
  try {
    loading.value = true
    const response = await getStats()
    const res = response.data || response
    if (res.code === 0 && res.data) {
      statsData.value = {
        total: res.data.total_repos || 0,
        cloned: res.data.cloned_count || 0,
        notCloned: res.data.not_cloned_count || 0
      }
    }
  } catch (error) {
    console.error('加载统计数据失败:', error)
  } finally {
    loading.value = false
  }
}

const refreshRepos = async () => {
  await loadStatistics()
}

// 扫描本地仓库
const handleScanRepos = async () => {
  try {
    scanLoading.value = true
    const { scanRepos, scanReposSSE } = await import('@/api/repos')
    const response = await scanRepos()
    const apiData = response.data

    if (apiData && apiData.code === 0 && apiData.data && apiData.data.useSSE) {
      const { taskId, tempToken } = apiData.data

      scanReposSSE(
        taskId,
        tempToken,
        () => {},
        (result) => {
          scanLoading.value = false
          const newRepos = result.new_repos?.length || 0
          const missingRepos = result.missing_repos?.length || 0

          if (newRepos > 0) {
            message.success(`扫描完成，发现 ${newRepos} 个新仓库`)
          } else if (missingRepos > 0) {
            message.warning(`发现 ${missingRepos} 个未克隆仓库`)
          } else {
            message.success('扫描完成，数据已是最新')
          }

          refreshRepos()
          loadActivities() // 刷新活动记录
        },
        (error) => {
          scanLoading.value = false
          message.error('扫描失败：' + (error.message || '未知错误'))
        }
      )
    } else {
      scanLoading.value = false
      message.error('启动扫描失败')
    }
  } catch (error) {
    scanLoading.value = false
    message.error('扫描失败：' + error.message)
  }
}

// 批量克隆 - 打开选择弹窗
const handleBatchClone = () => {
  showBatchCloneModal.value = true
}

// 批量克隆 - 用户确认选择后执行
const handleBatchCloneConfirm = async (ids) => {
  try {
    batchCloneLoading.value = true
    const { batchCloneRepos, batchCloneSSE } = await import('@/api/repos')
    const response = await batchCloneRepos({ ids })
    const apiData = response.data

    if (apiData && apiData.code === 0) {
      // 检查是否有需要克隆的仓库
      if (apiData.data?.total === 0 || (!apiData.data?.useSSE && apiData.data?.cloned === 0)) {
        message.info(apiData.message || '没有需要克隆的仓库')
        batchCloneLoading.value = false
        return
      }

      // 使用 SSE 获取进度
      if (apiData.data?.useSSE) {
        const { tempToken, total } = apiData.data

        showBatchCloneProgress.value = true
        batchCloneProgress.value = { total, current: 0, status: '准备中...' }

        batchCloneSSE(
          tempToken,
          (progress) => {
            if (progress.type === 'start') {
              batchCloneProgress.value = { ...batchCloneProgress.value, status: '开始克隆...', total: progress.total }
            } else if (progress.type === 'progress') {
              batchCloneProgress.value = {
                ...batchCloneProgress.value,
                current: progress.current,
                repo: progress.repo,
                status: progress.message
              }
            }
          },
          (result) => {
            batchCloneLoading.value = false
            batchCloneProgress.value = {
              ...batchCloneProgress.value,
              status: '完成',
              cloned: result.cloned,
              failed: result.failed,
              invalid: result.invalid,
              completed: true
            }
            message.success(`克隆完成: 成功 ${result.cloned}, 失败 ${result.failed}, 失效 ${result.invalid}`)

            refreshRepos()
            loadActivities() // 刷新活动记录
          },
          (error) => {
            batchCloneLoading.value = false
            message.error('批量克隆失败：' + (error.message || '未知错误'))
          }
        )
      } else {
        // 不使用 SSE 的情况
        batchCloneLoading.value = false
        message.info(apiData.message || '操作完成')
      }
    } else {
      batchCloneLoading.value = false
      message.error(apiData?.message || '启动批量克隆失败')
    }
  } catch (error) {
    batchCloneLoading.value = false
    message.error('批量克隆失败：' + error.message)
  }
}

const handleAddRepoSuccess = () => {
  showAddRepoModal.value = false
  refreshRepos()
  loadActivities() // 刷新活动记录
}

const viewAllActivities = () => {
  // 跳转到活动记录页面
  router.push('/activities')
}

// 一键拉取所有已克隆仓库（任务系统版）
const handleBatchPull = async () => {
  try {
    batchPullLoading.value = true
    const { batchPullRepos } = await import('@/api/repos')
    const response = await batchPullRepos()
    const apiData = response.data

    if (apiData && apiData.code === 0) {
      if (!apiData.data?.task_id) {
        message.info(apiData.message || '没有需要拉取的仓库')
        batchPullLoading.value = false
        return
      }

      message.success(apiData.message || '已创建拉取任务')
      batchPullLoading.value = false

      // 刷新任务列表，启动轮询
      await tasksStore.fetchTasks({ page_size: 10 })
      tasksStore.startSSE()
    } else {
      throw new Error(apiData?.message || '操作失败')
    }
  } catch (error) {
    batchPullLoading.value = false
    message.error('批量拉取失败：' + error.message)
  }
}

// 格式化时间
const formatTime = (time) => {
  if (!time) return ''

  // 如果是字符串，解析为 Date
  const date = typeof time === 'string' ? new Date(time.replace(/-/g, '/')) : new Date(time)
  const now = new Date()
  const diff = now - date
  const minutes = Math.floor(diff / 1000 / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)

  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes} 分钟前`
  if (hours < 24) return `${hours} 小时前`
  if (days < 7) return `${days} 天前`
  return date.toLocaleDateString()
}

// 任务辅助函数
function taskTypeName(type) {
  const map = { batch_pull: '一键拉取', batch_clone: '批量克隆', scan: '扫描仓库' }
  return map[type] || type
}

function taskStatusName(status) {
  const map = { pending: '等待中', running: '运行中', completed: '已完成', failed: '失败' }
  return map[status] || status
}

// 当前运行中的任务
const currentRunningTask = computed(() => {
  return tasksStore.taskList.find(t => t.status === 'running') || null
})

// 生命周期
onMounted(async () => {
  await Promise.all([
    loadStatistics(),
    loadActivities(),
    tasksStore.fetchTasks({ page_size: 10 })
  ])
  tasksStore.startSSE()
})

onUnmounted(() => {
  tasksStore.stopSSE()
})

// 监听刷新信号
watch(() => props.refreshKey, async (newVal, oldVal) => {
  if (newVal !== oldVal && newVal > 0) {
    await Promise.all([
      loadStatistics(),
      loadActivities()
    ])
  }
})

// 任务完成时自动刷新统计
watch(() => tasksStore.hasRunning, (newVal, oldVal) => {
  if (oldVal && !newVal) {
    // 从有 running 变为没有，说明任务完成了
    loadStatistics()
    loadActivities()
  }
})
</script>

<style scoped>
/* ============================================
   PAGE HEADER - 页面标题区
   ============================================ */

.dashboard-view {
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
   STATS GRID - 统计卡片网格
   ============================================ */

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--space-4);
}

/* 响应式网格布局 */
@media (max-width: 1200px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 640px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
}

.stat-card {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  padding: var(--space-5);
  background-color: var(--color-bg-card);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg);
  transition: all 0.2s ease;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.stat-icon {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-md);
  flex-shrink: 0;
}

.stat-icon svg {
  width: 24px;
  height: 24px;
}

.stat-card-primary .stat-icon {
  background-color: var(--color-primary-50);
  color: var(--color-primary);
}

.stat-card-success .stat-icon {
  background-color: var(--color-green-50);
  color: var(--color-green-600);
}

.stat-card-warning .stat-icon {
  background-color: var(--color-amber-50);
  color: var(--color-amber-600);
}

.stat-card-info .stat-icon {
  background-color: var(--color-cyan-50);
  color: var(--color-cyan-600);
}

.stat-content {
  flex: 1;
  min-width: 0;
}

.stat-label {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  margin-bottom: var(--space-1);
}

.stat-value {
  font-size: 28px;
  font-weight: var(--font-bold);
  color: var(--color-text-primary);
  line-height: 1;
}

/* ============================================
   QUICK ACTIONS - 快捷操作区
   ============================================ */

.quick-actions {
  background-color: var(--color-bg-card);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg);
  padding: var(--space-5);
}

.section-title {
  font-size: var(--text-lg);
  font-weight: var(--font-semibold);
  color: var(--color-text-primary);
  margin-bottom: var(--space-4);
}

.actions-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--space-4);
}

@media (max-width: 1200px) {
  .actions-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 640px) {
  .actions-grid {
    grid-template-columns: 1fr;
  }
}

.action-card {
  position: relative;
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-4);
  background-color: var(--color-bg-page);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all 0.2s ease;
  text-align: left;
}

.action-card:hover:not(:disabled) {
  background-color: var(--color-primary-50);
  border-color: var(--color-primary);
  transform: translateY(-2px);
  box-shadow: var(--shadow-sm);
}

.action-card:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.action-icon {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-md);
  flex-shrink: 0;
}

.action-icon svg {
  width: 20px;
  height: 20px;
}

.action-icon-primary {
  background-color: var(--color-primary-50);
  color: var(--color-primary);
}

.action-icon-success {
  background-color: var(--color-green-50);
  color: var(--color-green-600);
}

.action-icon-warning {
  background-color: var(--color-amber-50);
  color: var(--color-amber-600);
}

.action-icon-info {
  background-color: var(--color-cyan-50);
  color: var(--color-cyan-600);
}

.action-icon-success {
  background-color: var(--color-green-50);
  color: var(--color-green-600);
}

.action-text {
  flex: 1;
  min-width: 0;
}

.action-title {
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--color-text-primary);
  margin-bottom: var(--space-0_5);
}

.action-desc {
  font-size: var(--text-xs);
  color: var(--color-text-secondary);
}

.action-loading {
  position: absolute;
  top: var(--space-2);
  right: var(--space-2);
}

.loading-spinner {
  width: 16px;
  height: 16px;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

/* ============================================
   RECENT ACTIVITIES - 最近活动列表
   ============================================ */

.recent-activities {
  background-color: var(--color-bg-card);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg);
  padding: var(--space-5);
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-4);
}

.view-all-btn {
  padding: var(--space-1) var(--space-3);
  font-size: var(--text-sm);
  color: var(--color-primary);
  background: none;
  border: none;
  cursor: pointer;
  border-radius: var(--radius-md);
  transition: background-color 0.2s ease;
}

.view-all-btn:hover {
  background-color: var(--color-primary-50);
}

.activities-loading {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.loading-skeleton {
  height: 72px;
  background: linear-gradient(90deg, var(--color-gray-100) 25%, var(--color-gray-200) 50%, var(--color-gray-100) 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
  border-radius: var(--radius-md);
}

@keyframes shimmer {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}

.activities-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.activity-item {
  display: flex;
  gap: var(--space-3);
  padding: var(--space-3);
  border-radius: var(--radius-md);
  transition: background-color 0.2s ease;
}

.activity-item:hover {
  background-color: var(--color-gray-50);
}

.activity-icon {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-md);
  flex-shrink: 0;
}

.activity-icon svg {
  width: 18px;
  height: 18px;
}

.activity-icon-success {
  background-color: var(--color-green-50);
  color: var(--color-green-600);
}

.activity-icon-warning {
  background-color: var(--color-amber-50);
  color: var(--color-amber-600);
}

.activity-icon-error {
  background-color: var(--color-red-50);
  color: var(--color-red-600);
}

.activity-icon-info {
  background-color: var(--color-cyan-50);
  color: var(--color-cyan-600);
}

.activity-content {
  flex: 1;
  min-width: 0;
}

.activity-title {
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--color-text-primary);
  margin-bottom: var(--space-0_5);
}

.activity-desc {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  margin-bottom: var(--space-0_5);
}

.activity-time {
  font-size: var(--text-xs);
  color: var(--color-text-tertiary);
}

.activities-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-8);
  color: var(--color-text-secondary);
  gap: var(--space-3);
}

.activities-empty svg {
  width: 48px;
  height: 48px;
  opacity: 0.5;
}

/* ============================================
   RUNNING TASK - 当前运行任务
   ============================================ */

.running-task-section {
  background-color: var(--color-bg-card);
  border: 1px solid #e0e7ff;
  border-radius: var(--radius-lg);
  padding: var(--space-5);
}

.running-task-card {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.running-task-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.running-task-type {
  font-size: var(--text-base);
  font-weight: 600;
  color: var(--color-text-primary);
}

.running-task-progress {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
}

.running-task-counts {
  display: flex;
  gap: var(--space-2);
}

.task-status-badge {
  font-size: var(--text-xs);
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 4px;
}

.task-status-badge.status-running {
  color: #2080f0;
  background: #e8f4fd;
}
</style>
