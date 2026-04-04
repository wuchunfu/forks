<template>
  <div class="activities-view">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">活动记录</h1>
      <p class="page-description">查看所有操作历史记录</p>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <div class="filter-left">
        <n-select
          v-model:value="selectedType"
          :options="typeOptions"
          placeholder="活动类型"
          clearable
          size="small"
          style="width: 120px"
          @update:value="handleFilter"
        />
      </div>
      <div class="filter-right">
        <n-button
          v-if="activities.length > 0"
          type="error"
          size="small"
          ghost
          @click="handleClearAll"
        >
          <template #icon>
            <n-icon><TrashOutline /></n-icon>
          </template>
          清空记录
        </n-button>
      </div>
    </div>

    <!-- 活动列表 -->
    <div v-if="loading" class="activities-loading">
      <div class="loading-item" v-for="i in 5" :key="i">
        <div class="loading-icon"></div>
        <div class="loading-content">
          <div class="loading-line loading-title"></div>
          <div class="loading-line loading-desc"></div>
        </div>
      </div>
    </div>

    <div v-else-if="filteredActivities.length > 0" class="activities-list">
      <div
        v-for="activity in filteredActivities"
        :key="activity.id"
        class="activity-card"
      >
        <div class="activity-icon" :class="`icon-${activity.type}`">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="20 6 9 17 4 12" v-if="activity.type === 'success'"/>
            <circle cx="12" cy="12" r="10" v-else-if="activity.type === 'warning'"/>
            <path d="M18 6L6 18M6 6l12 12" v-else-if="activity.type === 'error'"/>
            <circle cx="12" cy="12" r="10" v-else/>
            <line x1="12" y1="8" x2="12" y2="12" v-else/>
            <line x1="12" y1="16" x2="12.01" y2="16" v-else/>
          </svg>
        </div>
        <div class="activity-content">
          <div class="activity-header">
            <span class="activity-title">{{ activity.title }}</span>
            <n-tag
              :type="getTypeTag(activity.type)"
              size="small"
              round
            >
              {{ getTypeLabel(activity.type) }}
            </n-tag>
          </div>
          <div class="activity-desc">{{ activity.description }}</div>
          <div class="activity-meta">
            <span v-if="activity.repo_name" class="repo-name">
              <n-icon size="14"><GitBranchOutline /></n-icon>
              {{ activity.repo_name }}
            </span>
            <span class="activity-time">{{ formatTime(activity.created_at) }}</span>
          </div>
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
    <div v-else class="activities-empty">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
        <path d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"/>
      </svg>
      <div class="empty-title">暂无活动记录</div>
      <div class="empty-desc">操作记录将在这里显示</div>
    </div>
  </div>
</template>

<script setup>
/**
 * ActivitiesView - 活动记录页面
 *
 * 显示所有操作历史记录，支持筛选和分页
 */
import { ref, computed, onMounted, watch } from 'vue'
import { useMessage, NSelect, NButton, NIcon, NTag, NPagination, useDialog } from 'naive-ui'
import { TrashOutline, GitBranchOutline } from '@vicons/ionicons5'
import { getActivities, clearActivities } from '@/api/repos'

const props = defineProps({
  refreshKey: { type: Number, default: 0 }
})

const message = useMessage()
const dialog = useDialog()

// 响应式数据
const loading = ref(false)
const activities = ref([])
const selectedType = ref(null)
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)

// 活动类型选项
const typeOptions = [
  { label: '成功', value: 'success' },
  { label: '信息', value: 'info' },
  { label: '警告', value: 'warning' },
  { label: '错误', value: 'error' }
]

// 计算属性
const totalPages = computed(() => Math.ceil(total.value / pageSize.value) || 1)

const filteredActivities = computed(() => {
  if (!selectedType.value) {
    return activities.value
  }
  return activities.value.filter(a => a.type === selectedType.value)
})

// 方法
const loadActivities = async () => {
  try {
    loading.value = true
    const response = await getActivities({
      page: currentPage.value,
      page_size: pageSize.value
    })
    const res = response.data || response
    if (res.code === 0 && res.data) {
      activities.value = res.data.list || []
      total.value = res.data.total || 0
    }
  } catch (error) {
    console.error('加载活动记录失败:', error)
    message.error('加载活动记录失败')
  } finally {
    loading.value = false
  }
}

const handleFilter = () => {
  // 筛选是在前端进行的，不需要重新请求
}

const handlePageChange = (page) => {
  currentPage.value = page
  loadActivities()
}

const handlePageSizeChange = (size) => {
  pageSize.value = size
  currentPage.value = 1
  loadActivities()
}

const handleClearAll = () => {
  dialog.warning({
    title: '确认清空',
    content: '确定要清空所有活动记录吗？此操作不可撤销。',
    positiveText: '确认清空',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const res = await clearActivities()
        const data = res.data || res
        if (data.code === 0) {
          message.success('活动记录已清空')
          activities.value = []
          total.value = 0
        } else {
          message.error(data.message || '清空失败')
        }
      } catch (error) {
        console.error('清空活动记录失败:', error)
        message.error('清空失败')
      }
    }
  })
}

const getTypeTag = (type) => {
  const map = {
    success: 'success',
    info: 'info',
    warning: 'warning',
    error: 'error'
  }
  return map[type] || 'default'
}

const getTypeLabel = (type) => {
  const map = {
    success: '成功',
    info: '信息',
    warning: '警告',
    error: '错误'
  }
  return map[type] || type
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

// 生命周期
onMounted(() => {
  loadActivities()
})

// 监听刷新信号
watch(() => props.refreshKey, (newVal, oldVal) => {
  if (newVal !== oldVal && newVal > 0) {
    loadActivities()
  }
})
</script>

<style scoped>
/* ============================================
   PAGE HEADER
   ============================================ */

.activities-view {
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

/* ============================================
   ACTIVITIES LIST
   ============================================ */

.activities-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.activity-card {
  display: flex;
  gap: var(--space-4);
  padding: var(--space-4);
  background-color: var(--color-bg-card);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg);
  transition: all 0.2s ease;
}

.activity-card:hover {
  border-color: var(--color-border);
  box-shadow: var(--shadow-sm);
}

.activity-icon {
  width: 40px;
  height: 40px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-full);
}

.activity-icon svg {
  width: 20px;
  height: 20px;
}

.icon-success {
  background-color: var(--color-green-50);
  color: var(--color-green-600);
}

.icon-info {
  background-color: var(--color-blue-50);
  color: var(--color-blue-600);
}

.icon-warning {
  background-color: var(--color-amber-50);
  color: var(--color-amber-600);
}

.icon-error {
  background-color: var(--color-red-50);
  color: var(--color-red-600);
}

.activity-content {
  flex: 1;
  min-width: 0;
}

.activity-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-1);
}

.activity-title {
  font-size: var(--text-base);
  font-weight: var(--font-semibold);
  color: var(--color-text-primary);
}

.activity-desc {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  margin-bottom: var(--space-2);
}

.activity-meta {
  display: flex;
  gap: var(--space-4);
  font-size: var(--text-xs);
  color: var(--color-text-tertiary);
}

.repo-name {
  display: flex;
  align-items: center;
  gap: var(--space-1);
}

/* ============================================
   LOADING STATE
   ============================================ */

.activities-loading {
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

.activities-empty {
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

.activities-empty svg {
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
