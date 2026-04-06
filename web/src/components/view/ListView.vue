<template>
  <div class="list-view">
    <!-- 工具栏 -->
    <div class="list-view__toolbar" v-if="showToolbar">
      <div class="list-view__info">
        <n-text depth="3">
          共 {{ total }} 个仓库
          <template v-if="selectedRepos.length > 0">
            ，已选择 {{ selectedRepos.length }} 个
          </template>
        </n-text>
      </div>

      <div class="list-view__actions">
        <!-- 批量操作 -->
        <template v-if="selectedRepos.length > 0">
          <n-space size="small">
            <n-button size="small" @click="clearSelection">
              取消选择
            </n-button>
            <n-button size="small" type="primary" @click="batchClone">
              批量克隆
            </n-button>
            <n-button size="small" @click="batchDelete">
              批量删除
            </n-button>
          </n-space>
        </template>

        <!-- 排序选项 -->
        <n-select
          :value="sortBy"
          :options="sortOptions"
          size="small"
          style="width: 140px"
          @update:value="handleSortChange"
        />

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
      </div>
    </div>

    <!-- 表格 -->
    <div class="list-view__table">
      <!-- 表头 -->
      <div class="list-view__header">
        <!-- 选择列 -->
        <div
          v-if="showSelectColumn"
          key="select"
          class="list-view__header-cell"
          style="width: 50px;"
        >
          <n-checkbox
            :checked="isAllSelected"
            :indeterminate="isIndeterminate"
            @update:checked="handleSelectAll"
          />
        </div>
        <!-- 其他列 -->
        <div
          v-for="column in columns.filter(col => col.visible)"
          :key="column.key"
          class="list-view__header-cell"
          :style="{ width: column.width ? `${column.width}px` : 'auto' }"
        >
          <div class="header-cell-content" @click="handleColumnSort(column.key)">
            <n-text strong>{{ column.label }}</n-text>
            <n-icon v-if="sortKey === column.key" :size="14">
              <ChevronUpOutline v-if="sortOrder === 'asc'" />
              <ChevronDownOutline v-else />
            </n-icon>
          </div>
        </div>
      </div>

      <!-- 数据行 -->
      <div class="list-view__body">
        <!-- 加载状态 -->
        <template v-if="loading">
          <div
            v-for="i in pageSize"
            :key="`skeleton-${i}`"
            class="list-view__row list-view__row--skeleton"
          >
            <div
              v-for="column in columns"
              :key="`skeleton-${column.key}`"
              class="list-view__cell"
            >
              <n-skeleton height="20px" width="80%" />
            </div>
          </div>
        </template>

        <!-- 仓库行 -->
        <template v-else>
          <div
            v-for="repo in repos"
            :key="repo.id"
            class="list-view__row"
            :class="{
              'list-view__row--selected': isRepoSelected(repo),
              'list-view__row--clickable': !batchMode
            }"
            @click="handleRowClick(repo)"
          >
            <!-- 选择列 -->
            <div v-if="showSelectColumn" class="list-view__cell" @click.stop>
              <n-checkbox
                :checked="isRepoSelected(repo)"
                @update:checked="handleRepoSelect(repo, $event)"
              />
            </div>

            <!-- 仓库名列 -->
            <div class="list-view__cell" v-if="showColumn('repo')">
              <div class="repo-name-cell">
                <n-ellipsis style="max-width: 200px;">
                  <n-text strong>
                    {{ repo.author }}/<span class="repo-name">{{ repo.repo }}</span>
                  </n-text>
                </n-ellipsis>
                <div class="repo-actions" v-if="!batchMode">
                  <n-button-group size="tiny">
                    <n-button
                      quaternary
                      @click.stop="$emit('view-code', repo)"
                      title="查看代码"
                    >
                      <template #icon>
                        <n-icon size="14">
                          <CodeSlashOutline />
                        </n-icon>
                      </template>
                    </n-button>
                    <n-button
                      v-if="getRepoStatus(repo.id)?.exists"
                      quaternary
                      @click.stop="$emit('open-folder', repo)"
                      title="打开文件夹"
                    >
                      <template #icon>
                        <n-icon size="14">
                          <FolderOutline />
                        </n-icon>
                      </template>
                    </n-button>
                  </n-button-group>
                </div>
              </div>
            </div>

            <!-- 描述列 -->
            <div class="list-view__cell" v-if="showColumn('description')">
              <n-ellipsis style="max-width: 300px;">
                <n-text depth="3">
                  {{ repo.description || '暂无描述' }}
                </n-text>
              </n-ellipsis>
            </div>

            <!-- 作者列 -->
            <div class="list-view__cell" v-if="showColumn('author')">
              <n-text>{{ repo.author }}</n-text>
            </div>

            <!-- 星标列 -->
            <div class="list-view__cell" v-if="showColumn('stars')">
              <div class="stat-cell">
                <n-icon size="14">
                  <StarOutline />
                </n-icon>
                <span>{{ formatNumber(repo.stars) }}</span>
              </div>
            </div>

            <!-- Fork列 -->
            <div class="list-view__cell" v-if="showColumn('forks')">
              <div class="stat-cell">
                <n-icon size="14">
                  <GitBranchOutline />
                </n-icon>
                <span>{{ formatNumber(repo.forks) }}</span>
              </div>
            </div>

            <!-- 编程语言列 -->
            <div class="list-view__cell" v-if="showColumn('language')">
              <div class="language-cell" v-if="mainLanguage(repo)">
                <div
                  class="language-dot"
                  :style="{ backgroundColor: getLanguageColor(mainLanguage(repo)) }"
                ></div>
                <span>{{ mainLanguage(repo) }}</span>
              </div>
              <n-text depth="3" v-else>-</n-text>
            </div>

            <!-- 许可证列 -->
            <div class="list-view__cell" v-if="showColumn('license')">
              <n-text v-if="repo.license">{{ repo.license }}</n-text>
              <n-text depth="3" v-else>-</n-text>
            </div>

            <!-- 状态列 -->
            <div class="list-view__cell" v-if="showColumn('status')">
              <div class="status-cell">
                <n-tag
                  v-if="repoStatus?.exists"
                  type="success"
                  size="small"
                  round
                >
                  已克隆
                </n-tag>
                <n-tag
                  v-if="repoStatus?.hasBehind"
                  type="warning"
                  size="small"
                  round
                >
                  有更新
                </n-tag>
                <n-tag
                  v-if="repoStatus?.hasChanges"
                  type="info"
                  size="small"
                  round
                >
                  有变更
                </n-tag>
                <n-text depth="3" v-if="!repoStatus?.exists">未克隆</n-text>
              </div>
            </div>

            <!-- 大小列 -->
            <div class="list-view__cell" v-if="showColumn('size')">
              <n-text v-if="repoStatus?.repoSize">{{ repoStatus.repoSize }}</n-text>
              <n-text depth="3" v-else>-</n-text>
            </div>

            <!-- 创建时间列 -->
            <div class="list-view__cell" v-if="showColumn('created_at')">
              <n-text depth="2">
                {{ formatDate(repo.created_at) }}
              </n-text>
            </div>

            <!-- 操作列 -->
            <div class="list-view__cell" v-if="showColumn('actions')">
              <n-dropdown
                trigger="click"
                :options="actionMenuOptions(repo)"
                @select="handleAction"
              >
                <n-button circle size="small" quaternary>
                  <template #icon>
                    <n-icon>
                      <EllipsisVerticalOutline />
                    </n-icon>
                  </template>
                </n-button>
              </n-dropdown>
            </div>
          </div>
        </template>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-if="!loading && repos.length === 0" class="list-view__empty">
      <n-empty description="暂无仓库数据" />
    </div>

    <!-- 分页 -->
    <div class="list-view__pagination" v-if="showPagination && totalPages > 1">
      <n-pagination
        :page="currentPage"
        :page-count="totalPages"
        :page-size="pageSize"
        :item-count="total"
        show-size-picker
        :page-sizes="[10, 20, 50, 100]"
        @update:page="$emit('page-change', $event)"
        @update:page-size="$emit('page-size-change', $event)"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import {
  CheckboxOutline,
  ChevronUpOutline,
  ChevronDownOutline,
  StarOutline,
  GitBranchOutline,
  CodeSlashOutline,
  FolderOutline,
  EllipsisVerticalOutline,
  PlayOutline,
  DownloadOutline,
  RefreshOutline,
  TrashOutline,
  EyeOutline,
  ShareSocialOutline
} from '@vicons/ionicons5'

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
  selectedRepos: {
    type: Array,
    default: () => []
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
    default: 20
  },

  // 列配置
  columns: {
    type: Array,
    default: () => []
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
  batchMode: {
    type: Boolean,
    default: false
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
  'delete',
  'update-info',
  'share',
  'batch-clone',
  'batch-delete',
  'page-change',
  'page-size-change'
])

// 响应式数据
const sortBy = ref('created_at')
const sortOrder = ref('desc')

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

const isAllSelected = computed(() => {
  return props.repos.length > 0 && props.selectedRepos.length === props.repos.length
})

const isIndeterminate = computed(() => {
  return props.selectedRepos.length > 0 && props.selectedRepos.length < props.repos.length
})

const showSelectColumn = computed(() => {
  return props.batchMode || props.selectedRepos.length > 0
})

const sortKey = computed(() => sortBy.value)

// 方法
const getRepoStatus = (repoId) => {
  return props.repoStatuses[repoId] || null
}

const isRepoSelected = (repo) => {
  return props.selectedRepos.some(r => r.id === repo.id)
}

const showColumn = (key) => {
  return props.columns.some(col => col.key === key && col.visible)
}

const mainLanguage = (repo) => {
  if (repo.languages) {
    try {
      const languages = JSON.parse(repo.languages)
      return languages[0]
    } catch {
      return null
    }
  }
  return null
}

const getLanguageColor = (language) => {
  const colors = {
    'JavaScript': '#f1e05a',
    'TypeScript': '#2b7489',
    'Python': '#3572A5',
    'Java': '#b07219',
    'Go': '#00ADD8',
    'Rust': '#dea584',
    'C++': '#f34b7d',
    'C#': '#239120',
    'PHP': '#4F5D95',
    'Ruby': '#701516',
    'Swift': '#ffac45',
    'Kotlin': '#F18E33',
    'Vue': '#4FC08D',
    'React': '#61DAFB',
    'HTML': '#e34c26',
    'CSS': '#1572b6',
    'Shell': '#89e051'
  }
  return colors[language] || '#666'
}

const formatNumber = (num) => {
  if (!num) return '0'
  if (num >= 1000) {
    return (num / 1000).toFixed(1) + 'k'
  }
  return num.toString()
}

const formatDate = (dateString) => {
  if (!dateString) return ''
  return new Date(dateString).toLocaleDateString('zh-CN')
}

const actionMenuOptions = (repo) => {
  const repoStatus = getRepoStatus(repo.id)
  return [
    {
      label: '克隆仓库',
      key: 'clone',
      show: !repoStatus?.exists
    },
    {
      label: '拉取最新',
      key: 'pull',
      show: repoStatus?.exists
    },
    {
      type: 'divider',
      show: repoStatus?.exists
    },
    {
      label: '打开文件夹',
      key: 'open-folder',
      show: repoStatus?.exists
    },
    {
      label: '查看代码',
      key: 'view-code'
    },
    {
      label: '更新信息',
      key: 'update-info'
    },
    {
      type: 'divider'
    },
    {
      label: '复制链接',
      key: 'copy-link'
    },
    {
      label: '分享',
      key: 'share'
    },
    {
      type: 'divider'
    },
    {
      label: '删除仓库',
      key: 'delete',
      props: {
        style: 'color: #e74c3c;'
      }
    }
  ].filter(item => item.show !== false)
}

// 事件处理
const handleRowClick = (repo) => {
  if (props.batchMode) {
    handleRepoSelect(repo, !isRepoSelected(repo))
  } else {
    emit('repo-click', repo)
  }
}

const handleRepoSelect = (repo, selected) => {
  emit('repo-select', repo, selected)
}

const handleSelectAll = (checked) => {
  if (checked) {
    emit('batch-select', [...props.repos])
  } else {
    emit('batch-select', [])
  }
}

const clearSelection = () => {
  emit('batch-select', [])
}

const toggleBatchMode = () => {
  // 切换批量模式需要由父组件处理
  console.log('Toggle batch mode')
}

const handleColumnSort = (key) => {
  if (sortBy.value === key) {
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortBy.value = key
    sortOrder.value = 'desc'
  }
}

const handleSortChange = (value) => {
  sortBy.value = value
  sortOrder.value = 'desc'
  // 可以触发排序事件
}

const handleAction = (key) => {
  // 这里需要获取对应的仓库
  // 实际实现中需要在actionMenuOptions中包含repo信息
  console.log('Action:', key)
}

const batchClone = () => {
  emit('batch-clone', props.selectedRepos)
}

const batchDelete = () => {
  emit('batch-delete', props.selectedRepos)
}
</script>

<style scoped>
.list-view {
  width: 100%;
}

.list-view__toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding: 12px 0;
}

.list-view__info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.list-view__actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.list-view__table {
  border: 1px solid #f0f0f0;
  border-radius: 6px;
  overflow: hidden;
}

.list-view__header {
  display: flex;
  background: #fafafa;
  border-bottom: 1px solid #f0f0f0;
}

.list-view__header-cell {
  padding: 12px 16px;
  font-weight: 600;
  color: #333;
  border-right: 1px solid #f0f0f0;
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.list-view__header-cell:last-child {
  border-right: none;
}

.header-cell-content {
  display: flex;
  align-items: center;
  gap: 4px;
  cursor: pointer;
  user-select: none;
}

.header-cell-content:hover {
  color: #18a058;
}

.list-view__body {
  max-height: 600px;
  overflow-y: auto;
}

.list-view__row {
  display: flex;
  border-bottom: 1px solid #f0f0f0;
  transition: background-color 0.2s;
}

.list-view__row:hover {
  background-color: #f8f9fa;
}

.list-view__row--selected {
  background-color: #e6f7ff;
}

.list-view__row--clickable {
  cursor: pointer;
}

.list-view__row:last-child {
  border-bottom: none;
}

.list-view__cell {
  padding: 12px 16px;
  border-right: 1px solid #f0f0f0;
  display: flex;
  align-items: center;
  min-height: 48px;
  flex-shrink: 0;
}

.list-view__cell:last-child {
  border-right: none;
}

.list-view__row--skeleton .list-view__cell {
  padding: 16px;
}

.repo-name-cell {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  flex: 1;
}

.repo-name {
  color: #18a058;
}

.repo-actions {
  opacity: 0;
  transition: opacity 0.2s;
}

.list-view__row:hover .repo-actions {
  opacity: 1;
}

.stat-cell {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 14px;
  color: #666;
}

.language-cell {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
}

.language-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-cell {
  display: flex;
  gap: 4px;
}

.list-view__empty {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 200px;
}

.list-view__pagination {
  display: flex;
  justify-content: center;
  padding: 24px 0;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .list-view__toolbar {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }

  .list-view__actions {
    justify-content: space-between;
  }

  .list-view__header {
    display: none;
  }

  .list-view__cell {
    padding: 8px 12px;
  }

  .repo-name-cell {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>