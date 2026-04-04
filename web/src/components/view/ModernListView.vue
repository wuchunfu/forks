<template>
  <div class="modern-list-view">
    <!-- 工具栏 -->
    <div class="modern-list-view__toolbar" v-if="showToolbar">
      <div class="modern-list-view__info">
        <n-text depth="3">
          共 {{ total }} 个仓库
          <template v-if="selectedRepos.length > 0">
            ，已选择 {{ selectedRepos.length }} 个
          </template>
        </n-text>
      </div>

      <div class="modern-list-view__actions">
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

    <!-- 现代化卡片列表 -->
    <div class="modern-list-container">
      <!-- 加载状态 -->
      <template v-if="loading">
        <div class="list-skeleton">
          <div v-for="i in pageSize" :key="`skeleton-${i}`" class="list-card-skeleton">
            <n-skeleton height="120px" :sharp="false" style="margin-bottom: 12px;" />
          </div>
        </div>
      </template>

      <!-- 仓库卡片列表 -->
      <template v-else>
        <!-- 选择列头部 -->
        <div class="list-header" v-if="showSelectColumn">
          <div class="list-header__selection">
            <n-checkbox
              :checked="isAllSelected"
              :indeterminate="isIndeterminate"
              @update:checked="handleSelectAll"
            >
              全选 ({{ selectedRepos.length }}/{{ repos.length }})
            </n-checkbox>
          </div>
        </div>

        <!-- 卡片列表 -->
        <div class="list-cards">
          <div
            v-for="(repo, index) in repos"
            :key="repo.id"
            class="list-card"
            :class="{
              'list-card--selected': isRepoSelected(repo),
              'list-card--first': index === 0
            }"
            @click="handleRepoClick(repo)"
          >
            <!-- 选择框 -->
            <div v-if="showSelectColumn" class="list-card__selection" @click.stop>
              <n-checkbox
                :checked="isRepoSelected(repo)"
                @update:checked="handleRepoSelect(repo, $event)"
              />
            </div>

            <!-- 仓库信息 -->
            <div class="list-card__content">
              <!-- 卡片头部 -->
              <div class="list-card__header">
                <div class="list-card__title">
                  <n-ellipsis style="max-width: 300px;">
                    <n-text strong style="font-size: 16px;">
                      {{ repo.author }}/<span class="repo-name">{{ repo.repo }}</span>
                    </n-text>
                  </n-ellipsis>
                </div>
                <div class="list-card__time" v-if="showColumn('created_at')">
                  <n-text depth="3" style="font-size: 12px;">
                    {{ formatDate(repo.created_at) }}
                  </n-text>
                </div>
              </div>

              <!-- 描述 -->
              <div class="list-card__description" v-if="showColumn('description')">
                <n-ellipsis style="max-height: 40px; line-height: 20px;">
                  <n-text depth="2">
                    {{ repo.description || '暂无描述' }}
                  </n-text>
                </n-ellipsis>
              </div>

              <!-- 统计信息 -->
              <div class="list-card__stats">
                <div class="stat-item" v-if="showColumn('stars')">
                  <n-icon size="14">
                    <StarOutline />
                  </n-icon>
                  <span>{{ formatNumber(repo.stars) }}</span>
                </div>
                <div class="stat-item" v-if="showColumn('forks')">
                  <n-icon size="14">
                    <GitBranchOutline />
                  </n-icon>
                  <span>{{ formatNumber(repo.forks) }}</span>
                </div>
                <div class="stat-item" v-if="showColumn('language') && mainLanguage(repo)">
                  <div
                    class="language-dot"
                    :style="{ backgroundColor: getLanguageColor(mainLanguage(repo)) }"
                  ></div>
                  <span>{{ mainLanguage(repo) }}</span>
                </div>
                <div class="stat-item" v-if="showColumn('license') && repo.license">
                  <span>{{ repo.license }}</span>
                </div>
                <div class="stat-item" v-if="showColumn('status')">
                  <n-tag
                    v-if="getRepoStatus(repo.id)?.exists"
                    type="success"
                    size="small"
                    round
                  >
                    已克隆
                  </n-tag>
                  <n-tag
                    v-else
                    type="default"
                    size="small"
                    round
                  >
                    未克隆
                  </n-tag>
                </div>
              </div>

              <!-- 快速操作 -->
              <div class="list-card__actions">
                <n-button-group size="small">
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
                  <n-button
                    quaternary
                    @click.stop="handleMoreActions(repo)"
                    title="更多操作"
                  >
                    <template #icon>
                      <n-icon size="14">
                        <EllipsisVerticalOutline />
                      </n-icon>
                    </template>
                  </n-button>
                </n-button-group>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>

    <!-- 空状态 -->
    <div v-if="!loading && repos.length === 0" class="modern-list-view__empty">
      <n-empty description="暂无仓库数据" />
    </div>

    <!-- 分页 -->
    <div class="modern-list-view__pagination" v-if="showPagination && totalPages > 1">
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
  'reset',
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
      label: '拉取更新',
      key: 'pull',
      show: repoStatus?.exists
    },
    {
      label: '重置仓库',
      key: 'reset',
      show: repoStatus?.exists
    },
    {
      label: '删除仓库',
      key: 'delete'
    },
    {
      label: '更新信息',
      key: 'update-info'
    },
    {
      label: '分享仓库',
      key: 'share'
    }
  ].filter(option => option.show !== false)
}

// 事件处理
const handleRepoClick = (repo) => {
  emit('repo-click', repo)
}

const handleRepoSelect = (repo, selected) => {
  emit('repo-select', repo, selected)
}

const handleSelectAll = (checked) => {
  emit('batch-select', props.repos, checked)
}

const clearSelection = () => {
  emit('batch-select', props.selectedRepos, false)
}

const handleSortChange = (value) => {
  sortBy.value = value
  // 可以触发排序事件
}

const toggleBatchMode = () => {
  // 切换批量选择模式
  console.log('Toggle batch mode')
}

const batchClone = () => {
  emit('batch-clone', props.selectedRepos)
}

const batchDelete = () => {
  emit('batch-delete', props.selectedRepos)
}

const handleMoreActions = (repo) => {
  console.log('More actions for:', repo)
}

const handleAction = (key, repo) => {
  switch (key) {
    case 'clone':
      emit('clone', repo)
      break
    case 'pull':
      emit('pull', repo)
      break
    case 'reset':
      emit('reset', repo)
      break
    case 'delete':
      emit('delete', repo)
      break
    case 'update-info':
      emit('update-info', repo)
      break
    case 'share':
      emit('share', repo)
      break
  }
}
</script>

<style scoped>
.modern-list-view {
  width: 100%;
}

.modern-list-view__toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding: 12px 0;
}

.modern-list-view__info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.modern-list-view__actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.modern-list-container {
  min-height: 200px;
}

.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding: 12px 16px;
  background: #f8f9fa;
  border-radius: 8px;
  border: 1px solid #f0f0f0;
}

.list-header__selection {
  display: flex;
  align-items: center;
}

.list-cards {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.list-card {
  background: white;
  border: 1px solid #f0f0f0;
  border-radius: 8px;
  padding: 16px;
  cursor: pointer;
  transition: all 0.2s ease;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  display: flex;
  align-items: flex-start;
  gap: 12px;
  min-height: 160px;
  max-height: 160px;
}

.list-card:hover {
  border-color: #18a058;
  box-shadow: 0 4px 12px rgba(24, 160, 88, 0.15);
  transform: translateY(-1px);
}

.list-card--selected {
  background: #f0f9ff;
  border-color: #18a058;
}

.list-card--first {
  margin-top: 0;
}

.list-card__selection {
  flex-shrink: 0;
  padding-top: 2px;
}

.list-card__content {
  flex: 1;
  min-width: 0;
}

.list-card__header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 8px;
}

.list-card__title {
  flex: 1;
  min-width: 0;
}

.repo-name {
  color: #18a058;
}

.list-card__time {
  flex-shrink: 0;
  margin-left: 12px;
}

.list-card__description {
  margin-bottom: 12px;
  line-height: 20px;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  max-height: 40px;
  flex: 1;
}

.list-card__stats {
  display: flex;
  gap: 16px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 14px;
  color: #666;
}

.language-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.list-card__actions {
  display: flex;
  justify-content: flex-end;
  opacity: 0;
  transition: opacity 0.2s;
  flex-shrink: 0;
}

.list-card:hover .list-card__actions {
  opacity: 1;
}

.list-skeleton {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.list-card-skeleton {
  margin-bottom: 12px;
}

.modern-list-view__empty {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 200px;
}

.modern-list-view__pagination {
  display: flex;
  justify-content: center;
  padding: 24px 0;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .modern-list-view__toolbar {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }

  .modern-list-view__actions {
    justify-content: space-between;
  }

  .list-card {
    padding: 12px;
  }

  .list-card__header {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .list-card__time {
    margin-left: 0;
  }

  .list-card__stats {
    gap: 12px;
  }

  .list-card__actions {
    opacity: 1;
  }
}
</style>