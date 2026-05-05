<template>
  <div class="authors-view">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="header-row">
        <h1 class="page-title">作者管理</h1>
        <div class="header-stats">
          <span class="header-stat-item">
            <span class="header-stat-label">总作者</span>
            <span class="header-stat-value">{{ totalAuthors }}</span>
          </span>
          <span class="header-stat-dot"></span>
          <span class="header-stat-item">
            <span class="header-stat-label">总仓库</span>
            <span class="header-stat-value">{{ totalRepos }}</span>
          </span>
        </div>
      </div>
      <p class="page-description">管理和查看所有仓库作者</p>
    </div>

    <!-- 搜索和筛选栏 -->
    <div class="filters-bar">
      <div class="search-input">
        <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="11" cy="11" r="8"/>
          <line x1="21" y1="21" x2="16.65" y2="16.65"/>
        </svg>
        <input
          v-model="searchQuery"
          type="text"
          placeholder="搜索作者..."
          @input="handleSearch"
        />
        <button v-if="searchQuery" class="clear-btn" @click="clearSearch">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"/>
            <line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>

      <div class="filter-options">
        <n-select
          v-model:value="selectedSource"
          :options="sourceOptions"
          placeholder="平台"
          clearable
          size="small"
          style="width: 110px"
          @update:value="() => { currentPage = 1; loadAuthors() }"
        />

        <n-select
          v-model:value="sortBy"
          :options="sortByOptions"
          placeholder="排序方式"
          size="small"
          style="width: 130px"
          @update:value="handleSort"
        />

        <n-select
          v-model:value="sortOrder"
          :options="sortOrderOptions"
          placeholder="排序顺序"
          size="small"
          style="width: 100px"
          @update:value="handleSort"
        />
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="authors-loading">
      <div class="loading-skeleton" v-for="i in 8" :key="i">
        <div class="skeleton-avatar"></div>
        <div class="skeleton-content">
          <div class="skeleton-line skeleton-title"></div>
          <div class="skeleton-line skeleton-desc"></div>
        </div>
      </div>
    </div>

    <!-- 作者卡片网格 -->
    <div v-else-if="authors.length > 0">
      <div class="authors-grid">
        <div
          v-for="author in authors"
          :key="author.author"
          class="author-card"
          @click="handleViewRepos(author)"
        >
          <!-- 平台角标 -->
          <span
            class="card-badge"
            :class="`badge-${author.source || 'github'}`"
          >{{ author.source === 'gitee' ? 'Gitee' : 'GitHub' }}</span>

          <!-- 作者头像 -->
          <div class="author-avatar">
            <div class="avatar-placeholder">
              {{ author.author.charAt(0).toUpperCase() }}
            </div>
          </div>

          <!-- 作者信息 -->
          <div class="author-info">
            <div class="author-name">{{ author.author }}</div>
            <div class="author-meta">
              <span class="repo-count">{{ author.repo_count }} 个仓库</span>
              <span v-if="author.cloned_count < author.repo_count" class="cloned-info">
                已克隆 {{ author.cloned_count }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- 分页 -->
      <div v-if="total > pageSize" class="pagination">
        <n-pagination
          v-model:page="currentPage"
          :page-count="totalPages"
          :page-size="pageSize"
          show-size-picker
          :page-sizes="[20, 40, 60, 100]"
          @update:page="handlePageChange"
          @update:page-size="handlePageSizeChange"
        />
      </div>
    </div>
    <div v-else class="authors-empty">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
        <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
        <circle cx="9" cy="7" r="4"/>
        <path d="M23 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75"/>
      </svg>
      <div class="empty-title">{{ searchQuery ? '未找到相关作者' : '暂无作者' }}</div>
      <div class="empty-desc">
        {{ searchQuery ? '尝试使用其他关键词搜索' : '添加仓库后，作者将自动显示在这里' }}
      </div>
      <button v-if="!searchQuery" class="add-repo-btn" @click="showAddRepoModal = true">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="12" y1="5" x2="12" y2="19"/>
          <line x1="5" y1="12" x2="19" y2="12"/>
        </svg>
        添加仓库
      </button>
    </div>

    <!-- 添加仓库对话框 -->
    <AddRepoModal
      :show="showAddRepoModal"
      @update:show="showAddRepoModal = $event"
      @success="handleAddRepoSuccess"
    />
  </div>
</template>

<script setup>
/**
 * AuthorsView - 作者管理页
 *
 * Grid Design System - Card Grid Layout
 * Chromatic Team - Pixel Implementation
 *
 * 功能特性:
 * - 搜索和筛选作者
 * - 作者卡片网格展示（3-4列响应式）
 * - 作者头像、名称、仓库数
 * - 查看作者所有仓库
 * - 响应式设计
 */
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage, NSelect, NPagination } from 'naive-ui'
import { getAuthors } from '@/api/repos'
import AddRepoModal from '@/components/AddRepoModal.vue'

const router = useRouter()
const message = useMessage()

// 响应式数据
const loading = ref(false)
const searchQuery = ref('')
const sortBy = ref('repo_count')
const sortOrder = ref('desc')
const showAddRepoModal = ref(false)
const selectedSource = ref(null)
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)

const sourceOptions = [
  { label: 'GitHub', value: 'github' },
  { label: 'Gitee', value: 'gitee' }
]

// 作者数据（当前页）
const authors = ref([])

const sortByOptions = [
  { label: '按名称排序', value: 'name' },
  { label: '按仓库数排序', value: 'repo_count' },
]

const sortOrderOptions = [
  { label: '升序', value: 'asc' },
  { label: '降序', value: 'desc' }
]

// 计算属性
const totalAuthors = computed(() => total.value)

const totalRepos = computed(() => {
  return authors.value.reduce((sum, author) => sum + author.repo_count, 0)
})

const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

// 方法
const loadAuthors = async () => {
  loading.value = true
  try {
    const params = {
      page: currentPage.value,
      page_size: pageSize.value,
      sort_by: sortBy.value,
      sort_order: sortOrder.value,
    }
    if (searchQuery.value) params.search = searchQuery.value
    if (selectedSource.value) params.source = selectedSource.value

    const response = await getAuthors(params)
    const res = response.data
    if (res.code === 0 && res.data) {
      authors.value = res.data.list || []
      total.value = res.data.total || 0
    } else {
      message.error('加载作者列表失败')
    }
  } catch (error) {
    message.error('加载作者列表失败：' + error.message)
  } finally {
    loading.value = false
  }
}

let searchTimer = null
const handleSearch = () => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    currentPage.value = 1
    loadAuthors()
  }, 300)
}

const clearSearch = () => {
  searchQuery.value = ''
  currentPage.value = 1
  loadAuthors()
}

const handleSort = () => {
  currentPage.value = 1
  loadAuthors()
}

const handlePageChange = (page) => {
  currentPage.value = page
  loadAuthors()
}

const handlePageSizeChange = (size) => {
  pageSize.value = size
  currentPage.value = 1
  loadAuthors()
}

const handleViewRepos = (author) => {
  router.push({
    path: '/repos',
    query: { author: author.author }
  })
}

const handleAddRepoSuccess = () => {
  showAddRepoModal.value = false
  loadAuthors()
}

// 生命周期
onMounted(() => {
  loadAuthors()
})
</script>

<style scoped>
/* ============================================
   PAGE HEADER - 页面标题
   ============================================ */

.authors-view {
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

.header-row {
  display: flex;
  align-items: baseline;
  gap: var(--space-4);
}

.header-stats {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
}

.header-stat-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.header-stat-label {
  color: var(--color-text-tertiary);
}

.header-stat-value {
  font-weight: var(--font-semibold);
  color: var(--color-text-primary);
}

.header-stat-dot {
  width: 3px;
  height: 3px;
  border-radius: 50%;
  background-color: var(--color-text-quaternary);
}

/* ============================================
   FILTERS BAR - 搜索和筛选栏
   ============================================ */

.filters-bar {
  display: flex;
  gap: var(--space-4);
  align-items: center;
  flex-wrap: wrap;
}

.search-input {
  position: relative;
  flex: 1;
  min-width: 280px;
  max-width: 400px;
}

.search-icon {
  position: absolute;
  left: var(--space-3);
  top: 50%;
  transform: translateY(-50%);
  width: 18px;
  height: 18px;
  color: var(--color-text-tertiary);
  pointer-events: none;
}

.search-input input {
  width: 100%;
  height: 28px;
  padding: 0 var(--space-10) 0 var(--space-10);
  font-size: var(--text-sm);
  color: var(--color-text-primary);
  background-color: var(--color-bg-card);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  outline: none;
  transition: all 0.2s ease;
}

.search-input input:focus {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-50);
}

.search-input input::placeholder {
  color: var(--color-text-tertiary);
}

.clear-btn {
  position: absolute;
  right: var(--space-2);
  top: 50%;
  transform: translateY(-50%);
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: none;
  border: none;
  cursor: pointer;
  color: var(--color-text-tertiary);
  border-radius: var(--radius-sm);
  transition: all 0.2s ease;
}

.clear-btn:hover {
  background-color: var(--color-gray-100);
  color: var(--color-text-secondary);
}

.clear-btn svg {
  width: 14px;
  height: 14px;
}

.filter-options {
  display: flex;
  gap: var(--space-3);
}

/* ============================================
   AUTHORS LOADING - 加载状态
   ============================================ */

.authors-loading {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: var(--space-4);
}

@media (max-width: 1400px) {
  .authors-loading {
    grid-template-columns: repeat(4, 1fr);
  }
}

@media (max-width: 1024px) {
  .authors-loading {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 768px) {
  .authors-loading {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 480px) {
  .authors-loading {
    grid-template-columns: 1fr;
  }
}

.loading-skeleton {
  display: flex;
  gap: var(--space-3);
  padding: var(--space-4);
  background-color: var(--color-bg-card);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg);
}

.skeleton-avatar {
  width: 48px;
  height: 48px;
  flex-shrink: 0;
  background: linear-gradient(90deg, var(--color-gray-200) 25%, var(--color-gray-300) 50%, var(--color-gray-200) 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
  border-radius: var(--radius-full);
}

.skeleton-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  padding-top: var(--space-1);
}

.skeleton-line {
  background: linear-gradient(90deg, var(--color-gray-200) 25%, var(--color-gray-300) 50%, var(--color-gray-200) 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
  border-radius: var(--radius-sm);
}

.skeleton-title {
  width: 60%;
  height: 16px;
}

.skeleton-desc {
  width: 40%;
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
   AUTHORS GRID - 作者卡片网格
   ============================================ */

.authors-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: var(--space-4);
}

/* 响应式网格布局 */
@media (max-width: 1400px) {
  .authors-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}

@media (max-width: 1024px) {
  .authors-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 768px) {
  .authors-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 480px) {
  .authors-grid {
    grid-template-columns: 1fr;
  }
}

.author-card {
  position: relative;
  background-color: var(--color-bg-card);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-md);
  padding: var(--space-3) var(--space-3);
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  gap: var(--space-3);
  overflow: hidden;
}

.author-card:hover {
  border-color: var(--color-primary);
  box-shadow: var(--shadow-md);
  transform: translateY(-1px);
}

.author-avatar {
  width: 36px;
  height: 36px;
  flex-shrink: 0;
}

.avatar-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-primary-600) 100%);
  color: white;
  font-size: 15px;
  font-weight: var(--font-bold);
  border-radius: var(--radius-full);
}

.author-info {
  flex: 1;
  min-width: 0;
}

.author-name {
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--color-text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.author-meta {
  font-size: var(--text-xs);
  color: var(--color-text-tertiary);
  display: flex;
  align-items: center;
  gap: var(--space-1);
}

/* 平台角标 */
.card-badge {
  position: absolute;
  top: 0;
  right: 0;
  font-size: 9px;
  padding: 1px 6px;
  text-align: center;
  border-radius: 0 var(--radius-md) 0 var(--radius-sm);
  font-weight: var(--font-medium);
  line-height: 1.5;
  z-index: 1;
}

.badge-github {
  background-color: rgba(110, 110, 110, 0.12);
  color: #8b949e;
}

.badge-gitee {
  background-color: rgba(139, 195, 74, 0.12);
  color: #8bc34a;
}

.repo-count {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
}

/* ============================================
   AUTHORS EMPTY - 空状态
   ============================================ */

.authors-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-12) var(--space-6);
  background-color: var(--color-bg-card);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg);
  text-align: center;
}

.authors-empty svg {
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
  margin-bottom: var(--space-6);
  max-width: 400px;
}

.add-repo-btn {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2_5) var(--space-4);
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  color: var(--color-primary);
  background-color: var(--color-primary-50);
  border: none;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all 0.2s ease;
}

.add-repo-btn:hover {
  background-color: var(--color-primary-100);
}

.add-repo-btn svg {
  width: 16px;
  height: 16px;
}

/* ============================================
   PAGINATION - 分页
   ============================================ */

.pagination {
  display: flex;
  justify-content: center;
  margin-top: var(--space-6);
}
</style>
