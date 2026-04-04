<template>
  <div class="card-view" :class="`card-view--${cardSize}`">
    <!-- 视图顶部工具栏 -->
    <div class="card-view__toolbar" v-if="showToolbar">
      <div class="card-view__info">
        <n-text depth="3">
          共 {{ total }} 个仓库
          <template v-if="selectedRepos.length > 0">
            ，已选择 {{ selectedRepos.length }} 个
          </template>
        </n-text>
      </div>

      <div class="card-view__actions">
        <!-- 批量选择模式切换 -->
        <n-button
          size="small"
          :type="batchMode ? 'primary' : 'default'"
          @click="toggleBatchMode"
        >
          <template #icon>
            <n-icon>
              <CheckboxOutline />
            </n-icon>
          </template>
          批量选择
        </n-button>

        <!-- 排序选项 -->
        <n-select
          :value="sortBy"
          :options="sortOptions"
          size="small"
          style="width: 140px"
          @update:value="handleSortChange"
        />

        <!-- 批量操作提示 -->
        <n-text v-if="selectedRepos.length > 0" depth="3" style="font-size: 12px;">
          已选择 {{ selectedRepos.length }} 个仓库
        </n-text>
      </div>

      <!-- 批量操作浮动栏 -->
      <div v-if="selectedRepos.length > 0 && batchMode" class="card-view__batch-actions">
        <n-space size="small">
          <n-button size="small" @click="exitBatchMode">
            取消选择
          </n-button>
          <n-button size="small" type="primary" @click="batchClone">
            批量克隆
          </n-button>
          <n-button size="small" @click="batchDelete">
            批量删除
          </n-button>
        </n-space>
      </div>
    </div>

    <!-- 卡片网格 -->
    <div class="card-view__grid" :style="gridStyle">
      <!-- 加载状态 -->
      <template v-if="loading">
        <div
          v-for="i in skeletonCount"
          :key="`skeleton-${i}`"
          class="repo-card-skeleton"
        >
          <n-skeleton height="220px" :sharp="false" />
        </div>
      </template>

      <!-- 仓库卡片 -->
      <template v-else>
        <RepoCard
          v-for="repo in repos"
          :key="repo.id"
          :repo="repo"
          :repo-status="getRepoStatus(repo)"
          :selected="isRepoSelected(repo)"
          :batch-mode="batchMode"
          :card-size="cardSize"
          @click="handleRepoClick"
          @select="handleRepoSelect"
          @view-code="$emit('view-code', $event)"
          @open-folder="$emit('open-folder', $event)"
          @clone="$emit('clone', $event)"
          @pull="$emit('pull', $event)"
          @reset="$emit('reset', $event)"
          @delete="$emit('delete', $event)"
          @update-info="$emit('update-info', $event)"
          @share="$emit('share', $event)"
        />
      </template>
    </div>

    <!-- 空状态 -->
    <div v-if="!loading && repos.length === 0" class="card-view__empty">
      <n-empty
        :description="emptyDescription"
        size="large"
      >
        <template #extra>
          <n-button type="primary" @click="$emit('add-repo')">
            添加第一个仓库
          </n-button>
        </template>
      </n-empty>
    </div>

    <!-- 分页 -->
    <div class="card-view__pagination" v-if="showPagination && totalPages > 1">
      <n-pagination
        :page="currentPage"
        :page-count="totalPages"
        :page-size="pageSize"
        :item-count="total"
        show-size-picker
        :page-sizes="[4, 8, 12, 24, 48]"
        @update:page="$emit('page-change', $event)"
        @update:page-size="$emit('page-size-change', $event)"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { CheckboxOutline } from '@vicons/ionicons5'
import RepoCard from '../repo/RepoCard.vue'

const props = defineProps({
  // 数据
  repos: {
    type: Array,
    default: () => []
  },
  repoStatuses: {
    type: Object,
    default: () => ({})
  },
  loading: {
    type: Boolean,
    default: false
  },

  // 分页
  total: {
    type: Number,
    default: 0
  },
  currentPage: {
    type: Number,
    default: 1
  },
  pageSize: {
    type: Number,
    default: 12
  },

  // 视图配置
  cardSize: {
    type: String,
    default: 'medium',
    validator: (value) => ['small', 'medium', 'large'].includes(value)
  },

  // 功能开关
  showToolbar: {
    type: Boolean,
    default: true
  },
  showPagination: {
    type: Boolean,
    default: true
  },
  selectable: {
    type: Boolean,
    default: true
  }
})

const emit = defineEmits([
  'repo-click',
  'repo-select',
  'batch-select',
  'view-code',
  'open-folder',
  'clone',
  'pull',
  'reset',
  'delete',
  'update-info',
  'share',
  'add-repo',
  'sort-change',
  'page-change',
  'page-size-change'
])

// 响应式数据
const sortBy = ref('created_at')
const sortOrder = ref('desc')
const batchMode = ref(false)
const selectedRepos = ref([])

// 排序选项
const sortOptions = [
  { label: '最新添加', value: 'created_at' },
  { label: '最多星标', value: 'stars' },
  { label: '最多Fork', value: 'forks' },
  { label: '仓库名称', value: 'repo' },
  { label: '作者名称', value: 'author' }
]

// 计算属性
const totalPages = computed(() => {
  return Math.ceil(props.total / props.pageSize)
})

const gridStyle = computed(() => {
  return {
    gridTemplateColumns: 'repeat(2, 1fr)',
    gap: props.cardSize === 'small' ? '12px' : '16px',
    maxWidth: '1200px',
    margin: '0 auto 24px auto'
  }
})

const emptyDescription = computed(() => {
  return batchMode.value ? '请先取消批量选择模式' : '还没有添加任何仓库'
})

const skeletonCount = computed(() => {
  return props.pageSize || 12
})

// 方法
const getRepoStatus = (repo) => {
  // 优先使用 repo 对象中的 is_cloned 字段
  const isCloned = repo.is_cloned === 1 || repo.is_cloned === true

  // 如果有额外的状态信息，合并返回
  const extraStatus = props.repoStatuses[repo.id] || {}

  return {
    exists: isCloned,
    hasBehind: extraStatus.hasBehind || false,
    hasChanges: extraStatus.hasChanges || false,
    status: isCloned ? '已克隆' : '未克隆',
    ...extraStatus
  }
}

const isRepoSelected = (repo) => {
  return selectedRepos.value.some(r => r.id === repo.id)
}

const handleRepoClick = (repo) => {
  if (batchMode.value) {
    handleRepoSelect(repo, !isRepoSelected(repo))
  } else {
    emit('repo-click', repo)
  }
}

const handleRepoSelect = (repo, selected) => {
  if (selected) {
    if (!isRepoSelected(repo)) {
      selectedRepos.value.push(repo)
    }
  } else {
    selectedRepos.value = selectedRepos.value.filter(r => r.id !== repo.id)
  }
  emit('repo-select', repo, selected)
  emit('batch-select', selectedRepos.value)
}

const toggleBatchMode = () => {
  batchMode.value = !batchMode.value
  if (!batchMode.value) {
    selectedRepos.value = []
    emit('batch-select', [])
  }
}

const exitBatchMode = () => {
  batchMode.value = false
  selectedRepos.value = []
  emit('batch-select', [])
}

const handleSortChange = (value) => {
  sortBy.value = value
  emit('sort-change', { sortBy: value, sortOrder: sortOrder.value })
}

const batchClone = () => {
  if (selectedRepos.value.length === 0) return

  emit('batch-clone', selectedRepos.value)
  exitBatchMode()
}

const batchDelete = () => {
  if (selectedRepos.value.length === 0) return

  emit('batch-delete', selectedRepos.value)
  exitBatchMode()
}

// 监听器
watch(() => props.repos, () => {
  // 当数据变化时，清除不存在的选中项
  selectedRepos.value = selectedRepos.value.filter(
    selected => props.repos.some(repo => repo.id === selected.id)
  )
})

// 暴露方法给父组件
defineExpose({
  toggleBatchMode,
  exitBatchMode,
  selectedRepos,
  batchMode,
  sortBy,
  sortOrder
})
</script>

<style scoped>
.card-view {
  width: 100%;
}

.card-view__toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding: 12px 0;
}

.card-view__info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.card-view__actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.card-view__grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  max-width: 1200px;
  margin: 0 auto 24px auto;
  width: 100%;
}

/* 卡片大小样式 */
.card-view--small .repo-card,
.card-view--small .repo-card-skeleton {
  height: 160px;
}

.card-view--medium .repo-card,
.card-view--medium .repo-card-skeleton {
  height: 160px;
}

.card-view--large .repo-card,
.card-view--large .repo-card-skeleton {
  height: 160px;
}

/* 骨架屏样式 */
.repo-card-skeleton {
  border-radius: 8px;
  overflow: hidden;
}

/* 空状态 */
.card-view__empty {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
  text-align: center;
}

/* 批量操作浮动栏 */
.card-view__batch-actions {
  position: fixed;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  background: white;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  padding: 12px 16px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  z-index: 1000;
  animation: slideUp 0.3s ease-out;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateX(-50%) translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateX(-50%) translateY(0);
  }
}

/* 分页 */
.card-view__pagination {
  display: flex;
  justify-content: center;
  padding: 24px 0;
}

/* 响应式设计 */
@media (max-width: 1400px) {
  .card-view__grid {
    max-width: 1000px;
    gap: 14px;
  }
}

@media (max-width: 1200px) {
  .card-view__grid {
    max-width: 900px;
    gap: 12px;
  }
}

@media (max-width: 900px) {
  .card-view__grid {
    max-width: 100%;
    gap: 16px;
  }
}

@media (max-width: 768px) {
  .card-view__grid {
    grid-template-columns: 1fr !important;
    gap: 12px;
    max-width: 100%;
  }

  .card-view__toolbar {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }

  .card-view__actions {
    justify-content: space-between;
  }

  .card-view__batch-actions {
    bottom: 10px;
    left: 10px;
    right: 10px;
    transform: none;
    width: auto;
  }
}

@media (max-width: 480px) {
  .card-view__grid {
    gap: 8px;
  }

  .card-view__batch-actions {
    padding: 10px 12px;
  }

  .card-view__batch-actions .n-space {
    flex-wrap: wrap;
  }
}

/* 卡片悬停效果 */
.card-view__grid > * {
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.card-view__grid > *:hover {
  transform: translateY(-2px);
  z-index: 1;
}

/* 选中状态动画 */
.repo-card {
  transition: all 0.2s ease;
}

.repo-card--selected {
  transform: scale(0.98);
  box-shadow: 0 0 0 2px #18a058;
}
</style>