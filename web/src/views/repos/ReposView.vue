<template>
  <div class="repos-view">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="header-row">
        <h1 class="page-title">仓库管理</h1>
        <div class="header-stats">
          <span class="header-stat-item">
            <span class="header-stat-label">总计</span>
            <span class="header-stat-value">{{ reposStore.total }}</span>
          </span>
          <span class="header-stat-dot"></span>
          <span class="header-stat-item">
            <span class="header-stat-label">已克隆</span>
            <span class="header-stat-value stat-success">{{ reposStore.statistics.cloned }}</span>
          </span>
          <span class="header-stat-dot"></span>
          <span class="header-stat-item">
            <span class="header-stat-label">未克隆</span>
            <span class="header-stat-value stat-warning">{{ reposStore.statistics.notCloned }}</span>
          </span>
        </div>
      </div>
      <p class="page-description">管理和查看所有 Git 仓库</p>
    </div>

    <!-- 工具栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <!-- 搜索框 -->
        <div class="search-input">
          <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8"/>
            <line x1="21" y1="21" x2="16.65" y2="16.65"/>
          </svg>
          <input
            v-model="searchQuery"
            type="text"
            placeholder="搜索仓库..."
            @input="handleSearch"
          />
        </div>

        <!-- 筛选器 -->
        <div class="filters">
          <n-select
            v-model:value="selectedStatus"
            :options="statusOptions"
            placeholder="状态"
            clearable
            size="small"
            style="width: 120px"
            @update:value="handleStatusFilter"
          />

          <n-select
            v-model:value="selectedAuthor"
            :options="authorOptions"
            placeholder="作者"
            clearable
            filterable
            size="small"
            style="width: 140px"
            @update:value="handleAuthorFilter"
          />

          <n-select
            v-model:value="selectedSource"
            :options="sourceOptions"
            placeholder="平台"
            clearable
            size="small"
            style="width: 110px"
            @update:value="handleSourceFilter"
          />

          <n-button v-if="hasActiveFilters" quaternary size="small" @click="clearFilters">
            <template #icon>
              <n-icon><Close /></n-icon>
            </template>
            清除筛选
          </n-button>
        </div>
      </div>

      <div class="toolbar-right">
        <!-- 视图切换 -->
        <div class="view-toggle">
          <button
            class="toggle-btn"
            :class="{ active: viewMode === 'grid' }"
            @click="viewMode = 'grid'"
            title="网格视图"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="3" width="7" height="7"/>
              <rect x="14" y="3" width="7" height="7"/>
              <rect x="14" y="14" width="7" height="7"/>
              <rect x="3" y="14" width="7" height="7"/>
            </svg>
          </button>
          <button
            class="toggle-btn"
            :class="{ active: viewMode === 'list' }"
            @click="viewMode = 'list'"
            title="列表视图"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="8" y1="6" x2="21" y2="6"/>
              <line x1="8" y1="12" x2="21" y2="12"/>
              <line x1="8" y1="18" x2="21" y2="18"/>
              <line x1="3" y1="6" x2="3.01" y2="6"/>
              <line x1="3" y1="12" x2="3.01" y2="12"/>
              <line x1="3" y1="18" x2="3.01" y2="18"/>
            </svg>
          </button>
        </div>

        <!-- 操作按钮 -->
        <n-button type="primary" size="small" @click="showAddRepoModal = true">
          <template #icon>
            <n-icon><Add /></n-icon>
          </template>
          添加仓库
        </n-button>
      </div>
    </div>

    <!-- 仓库列表 -->
    <div v-if="reposStore.loading" class="repos-loading">
      <div class="loading-skeleton" v-for="i in 6" :key="i">
        <div class="skeleton-icon"></div>
        <div class="skeleton-content">
          <div class="skeleton-line skeleton-title"></div>
          <div class="skeleton-line skeleton-desc"></div>
        </div>
      </div>
    </div>

    <div v-else-if="paginatedRepos.length > 0">
      <!-- 全局右键菜单 -->
      <n-dropdown
        :options="contextMenuOptions"
        @select="handleContextMenuAction"
        trigger="manual"
        placement="bottom-start"
        :x="contextMenuX"
        :y="contextMenuY"
        :show="showContextMenu"
        @update:show="showContextMenu = $event"
        @clickoutside="showContextMenu = false"
      />

      <!-- 网格视图 -->
      <div v-if="viewMode === 'grid'" class="repos-grid">
        <div
          v-for="(repo, idx) in paginatedRepos"
          :key="repo.id"
          class="repo-card"
          :class="{
            'card-uncloned': repo.valid !== 0 && !repo.is_cloned,
            'card-invalid': repo.valid === 0,
          }"
          @click="handleRepoClick(repo)"
          @contextmenu.prevent="(e) => handleContextMenu(e, repo)"
        >
          <!-- 拉取进度覆盖层 -->
          <div v-if="pullingRepoId === repo.id" class="repo-card-overlay">
            <n-spin size="small" />
            <span class="overlay-progress">{{ pullingProgress }}</span>
          </div>

          <!-- 平台角标 - 右上角 -->
          <span
            class="card-badge"
            :class="`badge-${repo.source || 'github'}`"
          >{{ repo.source === 'gitee' ? 'Gitee' : 'GitHub' }}</span>

          <!-- 仓库信息 -->
          <div class="repo-info">
            <div class="repo-name">
              {{ repo.repo }}
            </div>
            <div class="repo-author">{{ repo.author }}</div>
            <div class="repo-meta">
              <span v-if="repo.languages" class="repo-language">
                {{ getFirstLanguage(repo.languages) }}
              </span>
            </div>
          </div>

          <!-- 更新时间 - 右下角绝对定位 -->
          <span
            v-if="repo.last_pulled_at"
            class="repo-card-time"
            :class="getPullTimeClass(repo.last_pulled_at)"
          >
            <svg class="time-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/>
              <polyline points="12 6 12 12 16 14"/>
            </svg>
            {{ formatPullTime(repo.last_pulled_at) }}
          </span>
        </div>
      </div>

      <!-- 列表视图 -->
      <div v-else class="repos-list">
        <div
          v-for="(repo, idx) in paginatedRepos"
          :key="repo.id"
          class="repo-list-item"
          @click="handleRepoClick(repo)"
          @contextmenu.prevent="(e) => handleContextMenu(e, repo)"
        >
          <!-- 拉取进度覆盖层 -->
          <div v-if="pullingRepoId === repo.id" class="repo-card-overlay">
            <n-spin size="small" />
            <span class="overlay-progress">{{ pullingProgress }}</span>
          </div>

          <div class="repo-list-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
            </svg>
          </div>

          <div class="repo-list-info">
            <div class="repo-list-name">{{ repo.repo }}</div>
            <div class="repo-list-meta">
              <span>{{ repo.author }}</span>
              <span v-if="repo.languages">· {{ getFirstLanguage(repo.languages) }}</span>
              <span v-if="repo.last_pulled_at" class="repo-pulled-time" :class="getPullTimeClass(repo.last_pulled_at)">
                · {{ formatPullTime(repo.last_pulled_at) }}
              </span>
            </div>
          </div>

          <span
            class="card-badge"
            :class="`badge-${repo.source || 'github'}`"
          >{{ repo.source === 'gitee' ? 'Gitee' : 'GitHub' }}</span>

          <div class="repo-list-status">
            <n-tag
              v-if="repo.valid === 0"
              type="error"
              size="small"
              round
            >
              已失效
            </n-tag>
            <n-tag
              v-else
              :type="repo.is_cloned ? 'success' : 'default'"
              size="small"
              round
            >
              {{ repo.is_cloned ? '已克隆' : '未克隆' }}
            </n-tag>
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
          :page-sizes="[20, 40, 60, 100]"
          @update:page="handlePageChange"
          @update:page-size="handlePageSizeChange"
        />
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else class="repos-empty">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
        <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
      </svg>
      <div class="empty-title">{{ emptyDescription }}</div>
      <n-space v-if="!hasActiveFilters">
        <n-button @click="showAddRepoModal = true">添加仓库</n-button>
        <n-button @click="handleScanRepos">扫描本地</n-button>
      </n-space>
      <n-button v-else @click="clearFilters">清除筛选</n-button>
    </div>

    <!-- 添加仓库对话框 -->
    <AddRepoModal
      :show="showAddRepoModal"
      @update:show="showAddRepoModal = $event"
      @success="handleAddRepoSuccess"
    />

    <!-- 仓库详情抽屉 -->
    <RepoDetailDrawer
      :show="showDetailDrawer"
      :repo="selectedRepo"
      :update-loading="updateLoading"
      @update:show="showDetailDrawer = $event"
      @open-repo="handleOpenRepo"
      @view-code="handleViewCode"
      @update-info="handleUpdateInfo"
      @delete-repo="handleDelete"
      @toggle-valid="handleToggleValid"
    />
  </div>
</template>

<script setup>
/**
 * ReposView - 仓库管理页
 *
 * Grid Design System - Card/List View
 * Chromatic Team - Pixel Implementation
 *
 * 功能特性:
 * - 搜索和筛选仓库
 * - 视图切换（网格/列表）
 * - 仓库卡片展示
 * - 分页组件
 * - 仓库操作（查看、克隆、删除等）
 */
import { ref, computed, onMounted, watch, h } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage, useDialog, NButton, NIcon, NSelect, NDropdown, NTag, NPagination, NSpace, NSpin } from 'naive-ui'
import { Add, Close } from '@vicons/ionicons5'
import { useReposStore } from '@/stores/repos'
import { pullRepo, pullRepoSSE, toggleValid } from '@/api/repos'
import AddRepoModal from '@/components/AddRepoModal.vue'
import RepoDetailDrawer from '@/components/RepoDetailDrawer.vue'

const props = defineProps({
  refreshKey: { type: Number, default: 0 }
})

const router = useRouter()
const message = useMessage()
const dialog = useDialog()
const reposStore = useReposStore()

// 响应式数据
const viewMode = ref('grid') // 'grid' | 'list'
const showAddRepoModal = ref(false)
const showDetailDrawer = ref(false)
const selectedRepo = ref(null)
const updateLoading = ref(false)
const searchQuery = ref('')
const selectedStatus = ref(null)
const selectedAuthor = ref(null)
const selectedSource = ref(null)

// 拉取进度
const pullingRepoId = ref(null)
const pullingProgress = ref('')

// 右键菜单
const contextMenuRepo = ref(null)
const contextMenuX = ref(0)
const contextMenuY = ref(0)
const showContextMenu = ref(false)

// 右键菜单选项
const contextMenuOptions = computed(() => {
  if (!contextMenuRepo.value) return []
  return getRepoActions(contextMenuRepo.value)
})

// 分页
const currentPage = ref(1)
const pageSize = ref(20)

// 计算属性
const paginatedRepos = computed(() => reposStore.filteredRepos)

const totalPages = computed(() => reposStore.totalPages)

const hasActiveFilters = computed(() => {
  return searchQuery.value || selectedStatus.value || selectedAuthor.value || selectedSource.value
})

const emptyDescription = computed(() => {
  return hasActiveFilters.value ? '没有找到符合条件的仓库' : '还没有添加任何仓库'
})

// 筛选选项
const authorOptions = computed(() => {
  return reposStore.uniqueAuthors.map(author => ({
    label: author,
    value: author
  }))
})

const statusOptions = [
  { label: '已克隆', value: 'cloned' },
  { label: '未克隆', value: 'not-cloned' }
]

const sourceOptions = [
  { label: 'GitHub', value: 'github' },
  { label: 'Gitee', value: 'gitee' }
]

// 方法
const getFirstLanguage = (languages) => {
  if (!languages) return '未知'
  try {
    const langArray = JSON.parse(languages)
    return langArray[0] || '未知'
  } catch {
    return '未知'
  }
}

const formatPullTime = (timeStr) => {
  if (!timeStr) return ''
  const date = new Date(timeStr.replace(/-/g, '/'))
  if (isNaN(date.getTime())) return ''
  const now = new Date()
  const diffMs = now - date
  const minutes = Math.floor(diffMs / 60000)
  const hours = Math.floor(diffMs / 3600000)
  const days = Math.floor(diffMs / 86400000)

  if (minutes < 30) return '半小时内'
  if (hours < 1) return '1小时内'
  if (hours < 24) return `${hours}小时前`
  if (days === 1) return '1天前'
  if (days < 30) return `${days}天前`
  if (days < 365) return `${Math.floor(days / 30)}个月前`
  return `${Math.floor(days / 365)}年前`
}

const getPullTimeClass = (timeStr) => {
  if (!timeStr) return ''
  const date = new Date(timeStr.replace(/-/g, '/'))
  if (isNaN(date.getTime())) return ''
  const days = Math.floor((Date.now() - date.getTime()) / 86400000)
  if (days <= 1) return 'pull-fresh'
  if (days <= 3) return 'pull-recent'
  if (days <= 7) return 'pull-normal'
  return 'pull-stale'
}

const getRepoActions = (repo) => {
  const isInvalid = repo.valid === 0
  const actions = [
    {
      label: '查看详情',
      key: 'view'
    },
    {
      label: '查看代码',
      key: 'code'
    },
    {
      type: 'divider',
      key: 'd1'
    }
  ]

  if (!isInvalid) {
    actions.push({
      label: repo.is_cloned ? '拉取更新' : '克隆仓库',
      key: repo.is_cloned ? 'pull' : 'clone'
    })
  }

  actions.push(
    {
      label: '打开文件夹',
      key: 'open'
    },
    {
      type: 'divider',
      key: 'd2'
    },
    {
      label: isInvalid ? '取消失效标记' : '标记为失效',
      key: 'toggle-valid'
    },
    {
      type: 'divider',
      key: 'd3'
    },
    {
      label: '删除仓库',
      key: 'delete',
      props: {
        style: {
          color: 'var(--color-error-600)'
        }
      }
    }
  )

  return actions
}

const handleAction = async (key, repo) => {
  switch (key) {
    case 'view':
      handleRepoClick(repo)
      break
    case 'code':
      handleViewCode(repo)
      break
    case 'clone':
      message.success(`开始克隆 ${repo.repo}...`)
      break
    case 'pull':
      try {
        pullingRepoId.value = repo.id
        pullingProgress.value = '正在发起拉取...'
        const response = await pullRepo(repo.id)
        const apiData = response.data
        if (apiData && apiData.code === 0 && apiData.data?.useSSE) {
          pullingProgress.value = '连接中...'
          pullRepoSSE(
            repo.id,
            apiData.data.tempToken,
            (data) => {
              pullingProgress.value = data.message || '拉取中...'
            },
            (data) => {
              pullingRepoId.value = null
              pullingProgress.value = ''
              message.success(data.message || `${repo.repo} 拉取完成`)
              reposStore.fetchRepos()
            },
            (data) => {
              pullingRepoId.value = null
              pullingProgress.value = ''
              message.error(data.message || `${repo.repo} 拉取失败`)
            }
          )
        } else {
          throw new Error(apiData?.message || '拉取请求失败')
        }
      } catch (error) {
        pullingRepoId.value = null
        pullingProgress.value = ''
        message.error('拉取失败：' + error.message)
      }
      break
    case 'open':
      message.info('打开文件夹功能开发中...')
      break
    case 'delete':
      await handleDelete(repo)
      break
    case 'toggle-valid':
      try {
        const res = await toggleValid(repo.id)
        const apiData = res.data
        if (apiData && apiData.code === 0) {
          message.success(apiData.message)
          reposStore.fetchRepos()
        } else {
          throw new Error(apiData?.message || '操作失败')
        }
      } catch (error) {
        message.error('操作失败：' + error.message)
      }
      break
  }
}

const handleSearch = () => {
  reposStore.setSearchQuery(searchQuery.value)
  currentPage.value = 1
  reposStore.fetchRepos()
}

const handleStatusFilter = (value) => {
  selectedStatus.value = value
  reposStore.setActiveFilter('status', value)
  currentPage.value = 1
  reposStore.fetchRepos()
}

const handleAuthorFilter = (value) => {
  selectedAuthor.value = value
  reposStore.setActiveFilter('author', value)
  currentPage.value = 1
  reposStore.fetchRepos()
}

const handleSourceFilter = (value) => {
  selectedSource.value = value
  reposStore.setActiveFilter('source', value)
  currentPage.value = 1
  reposStore.fetchRepos({ source: value || undefined })
}

const clearFilters = () => {
  searchQuery.value = ''
  selectedStatus.value = null
  selectedAuthor.value = null
  selectedSource.value = null
  reposStore.clearFilters()
  currentPage.value = 1
  // 清除 URL 中的筛选参数
  if (router.currentRoute.value.query.author) {
    router.replace({ path: '/repos' })
  }
  reposStore.fetchRepos()
}

const handleRepoClick = (repo) => {
  selectedRepo.value = repo
  showDetailDrawer.value = true
}

// 右键菜单处理
const handleContextMenu = (e, repo) => {
  e.preventDefault()
  contextMenuX.value = e.clientX
  contextMenuY.value = e.clientY
  contextMenuRepo.value = repo
  showContextMenu.value = true
}

const handleContextMenuAction = (key) => {
  if (contextMenuRepo.value) {
    handleAction(key, contextMenuRepo.value)
  }
  showContextMenu.value = false
  contextMenuRepo.value = null
}

const handleViewCode = (repo) => {
  if (!repo || !repo.id) {
    message.error('仓库信息无效')
    return
  }
  router.push(`/code/${repo.id}`)
}

const handleOpenRepo = (url) => {
  window.open(url, '_blank')
}

const handleUpdateInfo = async (repo) => {
  try {
    updateLoading.value = true
    const { updateRepoInfo } = await import('@/api/repos')
    const response = await updateRepoInfo(repo.id)
    const apiData = response.data
    if (apiData && apiData.code === 0) {
      message.success(`更新 ${repo.repo} 信息成功`)
      // 即时更新详情抽屉数据
      if (selectedRepo.value && selectedRepo.value.id === repo.id) {
        selectedRepo.value = { ...selectedRepo.value, ...apiData.data }
      }
      reposStore.fetchRepos()
    } else {
      throw new Error(apiData?.message || '更新失败')
    }
  } catch (error) {
    message.error('更新失败：' + error.message)
  } finally {
    updateLoading.value = false
  }
}

const handleDelete = (repo) => {
  if (!repo || !repo.id) {
    message.error('仓库信息无效')
    return
  }

  const repoName = `${repo.author}/${repo.repo}`
  let confirmInput = ''

  dialog.warning({
    title: '删除仓库',
    content: () => {
      return h('div', {}, [
        h('p', { style: 'margin-bottom: 12px;' }, [
          '确定要删除仓库 ',
          h('strong', { style: 'color: var(--color-error-600);' }, repoName),
          ' 吗？此操作不可恢复。'
        ]),
        h('p', { style: 'margin-bottom: 8px; font-size: 13px;' },
          `请输入 ${repoName} 以确认删除：`
        ),
        h('input', {
          style: 'width: 100%; padding: 6px 10px; border: 1px solid var(--border-color); border-radius: 6px; background: var(--input-bg); color: inherit; font-size: 14px; outline: none; box-sizing: border-box;',
          placeholder: repoName,
          onInput: (e) => { confirmInput = e.target.value },
          onKeydown: (e) => {
            if (e.key === 'Enter') {
              const positiveBtn = e.target.closest('.n-dialog').querySelector('.n-dialog__action .n-button--warning-type')
              positiveBtn?.click()
            }
          }
        })
      ])
    },
    positiveText: '删除仓库',
    negativeText: '取消',
    onPositiveClick: async () => {
      if (confirmInput !== repoName) {
        message.error('输入的仓库名称不匹配')
        return false
      }
      try {
        await reposStore.removeRepo(repo.id)
        message.success('仓库删除成功')
        showDetailDrawer.value = false
      } catch (error) {
        message.error('删除失败：' + error.message)
      }
    }
  })
}

const handleToggleValid = async (repo) => {
  try {
    const res = await toggleValid(repo.id)
    const apiData = res.data
    if (apiData && apiData.code === 0) {
      message.success(apiData.message)
      if (selectedRepo.value && selectedRepo.value.id === repo.id) {
        selectedRepo.value = { ...selectedRepo.value, valid: apiData.data.valid }
      }
      reposStore.fetchRepos()
    } else {
      throw new Error(apiData?.message || '操作失败')
    }
  } catch (error) {
    message.error('操作失败：' + error.message)
  }
}

const handleAddRepoSuccess = () => {
  showAddRepoModal.value = false
  reposStore.fetchRepos()
}

const handleScanRepos = async () => {
  message.info('扫描功能开发中...')
}

const handlePageChange = async (page) => {
  currentPage.value = page
  reposStore.setCurrentPage(page)
  await reposStore.fetchRepos()
}

const handlePageSizeChange = async (size) => {
  pageSize.value = size
  currentPage.value = 1
  reposStore.setPageSize(size)
  await reposStore.fetchRepos()
}

// 生命周期
onMounted(async () => {
  // 重置筛选状态，避免从其他页面返回时残留旧条件
  searchQuery.value = ''
  selectedStatus.value = null
  selectedAuthor.value = null
  selectedSource.value = null
  reposStore.clearFilters()

  // 从 URL 查询参数获取筛选条件（从作者页跳转过来时携带）
  const route = router.currentRoute.value
  if (route.query.author) {
    selectedAuthor.value = route.query.author
    reposStore.setActiveFilter('author', route.query.author)
    // 读取后立即清除 URL 参数，避免刷新页面残留
    router.replace({ path: '/repos' })
  }

  // 并行加载数据
  await Promise.all([
    reposStore.fetchRepos(),
    reposStore.fetchStats()
  ])
})

// 监听刷新信号
watch(() => props.refreshKey, async (newVal, oldVal) => {
  if (newVal !== oldVal && newVal > 0) {
    await Promise.all([
      reposStore.fetchRepos(),
      reposStore.fetchStats()
    ])
  }
})
</script>

<style scoped>
/* ============================================
   PAGE HEADER - 页面标题
   ============================================ */

.repos-view {
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
   TOOLBAR - 工具栏
   ============================================ */

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--space-4);
  flex-wrap: wrap;
}

.toolbar-left {
  display: flex;
  gap: var(--space-3);
  align-items: center;
  flex: 1;
  flex-wrap: wrap;
}

.search-input {
  position: relative;
  width: 280px;
}

.search-icon {
  position: absolute;
  left: var(--space-3);
  top: 50%;
  transform: translateY(-50%);
  width: 16px;
  height: 16px;
  color: var(--color-text-tertiary);
  pointer-events: none;
}

.search-input input {
  width: 100%;
  padding: var(--space-2) var(--space-3) var(--space-2) var(--space-9);
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

.filters {
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.toolbar-right {
  display: flex;
  gap: var(--space-3);
  align-items: center;
}

/* 视图切换 */
.view-toggle {
  display: flex;
  background-color: var(--color-bg-card);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.toggle-btn {
  padding: var(--space-2);
  background: none;
  border: none;
  cursor: pointer;
  color: var(--color-text-secondary);
  transition: all 0.2s ease;
}

.toggle-btn:hover {
  color: var(--color-text-primary);
  background-color: var(--color-gray-100);
}

.toggle-btn.active {
  color: var(--color-primary);
  background-color: var(--color-primary-50);
}

.toggle-btn svg {
  width: 18px;
  height: 18px;
}

/* ============================================
   STATS BAR - 统计栏
   ============================================ */

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

.stat-success {
  color: var(--color-success-600);
}

.stat-warning {
  color: var(--color-warning-600);
}

/* ============================================
   REPOS LOADING - 加载状态
   ============================================ */

.repos-loading {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--space-4);
}

@media (max-width: 1200px) {
  .repos-loading {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 900px) {
  .repos-loading {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 640px) {
  .repos-loading {
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

.skeleton-icon {
  width: 40px;
  height: 40px;
  flex-shrink: 0;
  background: linear-gradient(90deg, var(--color-gray-200) 25%, var(--color-gray-300) 50%, var(--color-gray-200) 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
  border-radius: var(--radius-md);
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
  width: 70%;
  height: 16px;
}

.skeleton-desc {
  width: 50%;
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
   REPOS GRID - 网格视图
   ============================================ */

.repos-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--space-4);
}

@media (max-width: 1200px) {
  .repos-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 900px) {
  .repos-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 640px) {
  .repos-grid {
    grid-template-columns: 1fr;
  }
}

.repo-card {
  position: relative;
  display: flex;
  gap: var(--space-3);
  padding: var(--space-4);
  padding-bottom: 28px;
  background-color: var(--color-bg-card);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg);
  cursor: pointer;
  transition: all 0.2s ease;
}

.repo-card:hover {
  border-color: var(--color-primary);
  box-shadow: var(--shadow-md);
  transform: translateY(-2px);
}

/* 未克隆 - 虚线边框 + 淡色 */
.card-uncloned {
  border-style: dashed;
  border-color: var(--color-border);
  opacity: 0.75;
}

.card-uncloned:hover {
  opacity: 1;
  border-color: var(--color-primary);
}

/* 已失效 - 红色边框 */
.card-invalid {
  border-color: rgba(220, 38, 38, 0.3);
}

.card-invalid:hover {
  border-color: var(--color-error-500, #ef4444);
}

.repo-info {
  flex: 1;
  min-width: 0;
}

.repo-name {
  font-size: var(--text-base);
  font-weight: var(--font-semibold);
  color: var(--color-text-primary);
  margin-bottom: var(--space-0_5);
  word-break: break-word;
}

.repo-author {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  margin-bottom: var(--space-1);
}

.repo-meta {
  display: flex;
  gap: var(--space-2);
  align-items: center;
  flex-wrap: wrap;
}

.repo-language {
  font-size: var(--text-xs);
  color: var(--color-text-secondary);
}

/* 平台角标 */
.card-badge {
  position: absolute;
  top: 0;
  right: 0;
  font-size: 10px;
  padding: 2px 8px;
  min-width: 46px;
  text-align: center;
  border-radius: 0 var(--radius-lg) 0 var(--radius-md);
  font-weight: var(--font-medium);
  line-height: 1.5;
  z-index: 1;
}

.badge-github {
  background-color: rgba(110, 110, 110, 0.12);
  color: var(--color-gray-600);
}

.badge-gitee {
  background-color: rgba(139, 195, 74, 0.12);
  color: var(--color-success-500);
}

.repo-status {
  font-size: var(--text-xs);
  padding: var(--space-0_5) var(--space-1);
  border-radius: var(--radius-full);
}

.status-cloned {
  background-color: var(--color-success-50);
  color: var(--color-success-700);
}

.status-uncloned {
  background-color: var(--color-gray-100);
  color: var(--color-gray-700);
}

.status-invalid {
  background-color: rgba(220, 38, 38, 0.1);
  color: var(--color-error-600);
}

.repo-pulled-time {
  font-size: var(--text-xs);
  color: var(--color-text-tertiary);
}

.repo-pulled-time.pull-fresh {
  color: var(--color-success-600);
}

.repo-pulled-time.pull-recent {
  color: var(--color-text-secondary);
}

.repo-pulled-time.pull-stale {
  color: var(--color-orange-500, #f97316);
}

/* 网格卡片右下角更新时间 */
.repo-card-time {
  position: absolute;
  right: 8px;
  bottom: 8px;
  display: inline-flex;
  align-items: center;
  gap: 3px;
  font-size: 10px;
  padding: 2px 7px;
  border-radius: var(--radius-full);
  background-color: var(--color-gray-50);
  color: var(--color-text-tertiary);
  line-height: 1.5;
  white-space: nowrap;
  transition: opacity 0.2s;
}

.repo-card-time .time-icon {
  width: 11px;
  height: 11px;
  flex-shrink: 0;
}

.repo-card-time.pull-fresh {
  background-color: var(--color-success-50);
  color: var(--color-success-600);
}

.repo-card-time.pull-recent {
  background-color: var(--color-gray-50);
  color: var(--color-text-secondary);
}

.repo-card-time.pull-normal {
  background-color: var(--color-warning-50, #fffbeb);
  color: var(--color-warning-600, #d97706);
}

.repo-card-time.pull-stale {
  background-color: rgba(249, 115, 22, 0.08);
  color: var(--color-orange-500, #f97316);
}

/* ============================================
   PULL OVERLAY - 拉取进度覆盖层
   ============================================ */

.repo-card-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  background-color: rgba(0, 0, 0, 0.45);
  border-radius: var(--radius-lg);
  z-index: 5;
  animation: fadeInOverlay 0.2s ease-out;
}

.overlay-progress {
  font-size: var(--text-xs);
  color: #fff;
  max-width: 60%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@keyframes fadeInOverlay {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

/* ============================================
   REPOS LIST - 列表视图
   ============================================ */

.repos-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.repo-list-item {
  position: relative;
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-4);
  background-color: var(--color-bg-card);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg);
  cursor: pointer;
  transition: all 0.2s ease;
}

.repo-list-item:hover {
  border-color: var(--color-primary);
  box-shadow: var(--shadow-sm);
}

.repo-list-icon {
  width: 36px;
  height: 36px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: var(--color-primary-50);
  color: var(--color-primary);
  border-radius: var(--radius-md);
}

.repo-list-icon svg {
  width: 18px;
  height: 18px;
}

.repo-list-info {
  flex: 1;
  min-width: 0;
}

.repo-list-name {
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--color-text-primary);
  margin-bottom: var(--space-0_5);
}

.repo-list-meta {
  font-size: var(--text-xs);
  color: var(--color-text-secondary);
}

/* ============================================
   PAGINATION - 分页
   ============================================ */

.pagination {
  display: flex;
  justify-content: center;
  padding: var(--space-4) 0;
}

/* ============================================
   REPOS EMPTY - 空状态
   ============================================ */

.repos-empty {
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

.repos-empty svg {
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
  margin-bottom: var(--space-4);
}
</style>
