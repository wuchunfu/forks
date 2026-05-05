<template>
  <div class="trending-view">
    <div class="page-header">
      <h1 class="page-title">GitHub 趋势</h1>
      <p class="page-description">
        发现 GitHub 上最受欢迎的开源项目
        <span v-if="isHistoryMode" class="history-badge">历史数据 · {{ selectedDate }}</span>
      </p>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <div class="filter-left">
        <n-date-picker
          v-model:value="datePickerValue"
          type="date"
          size="small"
          clearable
          :placeholder="'选择日期'"
          style="width: 150px"
          :is-date-disabled="disableFutureDates"
        />
        <n-select
          v-model:value="selectedLanguage"
          :options="programmingLanguageOptions"
          placeholder="编程语言"
          clearable
          filterable
          size="small"
          style="width: 180px"
        />
        <n-select
          v-model:value="selectedSpokenLanguage"
          :options="spokenLanguageOptions"
          placeholder="自然语言"
          clearable
          filterable
          size="small"
          style="width: 160px"
        />
        <n-select
          v-model:value="selectedSince"
          :options="sinceOptions"
          size="small"
          style="width: 120px"
        />
      </div>
      <div class="filter-right">
        <n-button
          v-if="!isHistoryMode"
          size="small"
          @click="handleRefresh"
          :loading="loading"
        >
          <template #icon>
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="23 4 23 10 17 10"/>
              <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
            </svg>
          </template>
          刷新
        </n-button>
      </div>
    </div>

    <!-- 加载骨架屏 -->
    <div v-if="loading" class="trending-skeleton">
      <div class="skeleton-card" v-for="i in 6" :key="i">
        <div class="skeleton-line skeleton-title"></div>
        <div class="skeleton-line skeleton-desc"></div>
        <div class="skeleton-line skeleton-meta"></div>
      </div>
    </div>

    <!-- 错误提示 -->
    <div v-else-if="error" class="trending-error">
      <div class="error-icon">
        <svg viewBox="0 0 24 24" width="40" height="40" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"/>
          <line x1="12" y1="8" x2="12" y2="12"/>
          <line x1="12" y1="16" x2="12.01" y2="16"/>
        </svg>
      </div>
      <p class="error-text">{{ error }}</p>
      <n-button size="small" @click="loadTrending">{{ isHistoryMode ? '返回今天' : '重试' }}</n-button>
    </div>

    <!-- 空状态 -->
    <div v-else-if="trendingRepos.length === 0" class="trending-empty">
      <p>{{ isHistoryMode ? '该日期没有历史数据' : '暂无趋势数据' }}</p>
    </div>

    <!-- 项目列表 -->
    <div v-else class="trending-list">
      <div v-for="repo in trendingRepos" :key="repo.url" class="trending-card" @contextmenu.prevent="handleContextMenu($event, repo)">
        <div class="card-header">
          <div class="card-title-row">
            <a :href="repo.url" target="_blank" rel="noopener" class="repo-name">
              {{ repo.author }} / {{ repo.repo }}
            </a>
            <n-button
              v-if="repo._exists || repo._added"
              size="tiny"
              type="success"
              ghost
              disabled
            >
              已添加
            </n-button>
            <n-button
              v-else
              size="tiny"
              type="default"
              ghost
              @click="handleAddRepo(repo)"
              :loading="repo._adding"
            >
              + 添加
            </n-button>
          </div>
          <p class="repo-desc">{{ repo.description }}</p>
        </div>
        <div class="card-meta">
          <span v-if="repo.language" class="meta-item">
            <span class="lang-dot" :style="{ backgroundColor: repo.language_color || '#ccc' }"></span>
            {{ repo.language }}
          </span>
          <span class="meta-item">
            <svg viewBox="0 0 16 16" width="14" height="14" fill="currentColor">
              <path d="M8 .25a.75.75 0 01.673.418l1.882 3.815 4.21.612a.75.75 0 01.416 1.279l-3.046 2.97.719 4.192a.75.75 0 01-1.088.791L8 12.347l-3.766 1.98a.75.75 0 01-1.088-.79l.72-4.194L.818 6.374a.75.75 0 01.416-1.28l4.21-.611L7.327.668A.75.75 0 018 .25z"/>
            </svg>
            {{ formatNum(repo.stars) }}
          </span>
          <span class="meta-item">
            <svg viewBox="0 0 16 16" width="14" height="14" fill="currentColor">
              <path d="M5 3.25a.75.75 0 11-1.5 0 .75.75 0 011.5 0zm0 2.122a2.25 2.25 0 10-1.5 0v.878A2.25 2.25 0 005.75 8.5h1.5v2.128a2.251 2.251 0 101.5 0V8.5h1.5a2.25 2.25 0 002.25-2.25v-.878a2.25 2.25 0 10-1.5 0v.878a.75.75 0 01-.75.75h-4.5A.75.75 0 015 6.25v-.878zm3.75 7.378a.75.75 0 11-1.5 0 .75.75 0 011.5 0zm3-8.75a.75.75 0 100-1.5.75.75 0 000 1.5z"/>
            </svg>
            {{ formatNum(repo.forks) }}
          </span>
          <span v-if="repo.current_period_stars > 0" class="meta-item stars-today">
            +{{ formatNum(repo.current_period_stars) }} {{ sinceLabel }}
          </span>
          <a
            :href="'https://deepwiki.com/' + repo.author + '/' + repo.repo"
            target="_blank"
            rel="noopener"
            class="meta-item ext-link"
            title="在 DeepWiki 中查看"
          >
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/>
              <path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/>
            </svg>
            DeepWiki
          </a>
          <a
            :href="'https://zread.ai/' + repo.author + '/' + repo.repo"
            target="_blank"
            rel="noopener"
            class="meta-item ext-link"
            title="在 ZRead 中查看"
          >
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/>
              <path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/>
            </svg>
            ZRead
          </a>
        </div>
        <div v-if="repo.built_by && repo.built_by.length > 0" class="card-built-by">
          <img
            v-for="contributor in repo.built_by.slice(0, 5)"
            :key="contributor.username"
            :src="contributor.avatar"
            :alt="contributor.username"
            :title="contributor.username"
            class="contributor-avatar"
            loading="lazy"
          />
        </div>
      </div>
    </div>

    <!-- 右键菜单 -->
    <n-dropdown
      :options="contextMenuOptions"
      @select="handleContextMenuAction"
      trigger="manual"
      placement="bottom-start"
      :x="contextMenuX"
      :y="contextMenuY"
      :show="showCtxMenu"
      @update:show="showCtxMenu = $event"
      @clickoutside="showCtxMenu = false"
    />

    <!-- 仓库详情抽屉 -->
    <RepoDetailDrawer
      :show="detailDrawer.show"
      :repo="detailDrawer.repo"
      :update-loading="updateLoading"
      @update:show="detailDrawer.show = $event"
      @open-repo="handleOpenRepo"
      @view-code="handleViewCode"
      @update-info="handleUpdateInfo"
      @delete-repo="handleDelete"
      @toggle-valid="handleToggleValid"
    />
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { NDropdown, useMessage } from 'naive-ui'
import { getTrending, getTrendingLanguages } from '@/api/trending'
import { addRepo, getRepo, toggleValid } from '@/api/repos'
import RepoDetailDrawer from '@/components/RepoDetailDrawer.vue'

const router = useRouter()
const message = useMessage()

const selectedLanguage = ref(null)
const selectedSpokenLanguage = ref(null)
const selectedSince = ref('daily')
const datePickerValue = ref(null) // null = today

const programmingLanguageOptions = ref([])
const spokenLanguageOptions = ref([])
const sinceOptions = [
  { label: 'Today', value: 'daily' },
  { label: 'This Week', value: 'weekly' },
  { label: 'This Month', value: 'monthly' }
]

const trendingRepos = ref([])
const loading = ref(false)
const error = ref(null)

const selectedDate = computed(() => {
  if (!datePickerValue.value) return ''
  const d = new Date(datePickerValue.value)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
})

const isHistoryMode = computed(() => {
  if (!datePickerValue.value) return false
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  const selected = new Date(datePickerValue.value)
  return selected < today
})

const sinceLabel = computed(() => {
  const map = { daily: 'today', weekly: 'this week', monthly: 'this month' }
  return map[selectedSince.value] || 'today'
})

function disableFutureDates(ts) {
  return ts > Date.now()
}

async function loadLanguages() {
  try {
    const res = await getTrendingLanguages()
    const data = res.data.data
    programmingLanguageOptions.value = [
      { label: '全部语言', value: null },
      ...data.programming_languages.map(l => ({ label: l.name, value: l.slug }))
    ]
    spokenLanguageOptions.value = [
      { label: '全部语言', value: null },
      ...data.spoken_languages.map(l => ({ label: `${l.name} (${l.code})`, value: l.code }))
    ]
  } catch (e) {
    console.error('加载语言映射失败:', e)
  }
}

async function loadTrending(forceRefresh = false) {
  loading.value = true
  error.value = null
  try {
    const params = {
      since: selectedSince.value
    }
    if (selectedLanguage.value) params.language = selectedLanguage.value
    if (selectedSpokenLanguage.value) params.spoken_language_code = selectedSpokenLanguage.value
    if (selectedDate.value) params.date = selectedDate.value
    if (forceRefresh) params.refresh = 'true'

    const res = await getTrending(params)
    trendingRepos.value = (res.data.data.items || []).map(r => ({
      ...r,
      _exists: r._exists || false,
      _adding: false,
      _added: false
    }))
  } catch (e) {
    error.value = e.response?.data?.message || e.message || '加载失败'
  } finally {
    loading.value = false
  }
}

function handleRefresh() {
  loadTrending(true)
}

async function handleAddRepo(repo) {
  if (repo._added) return
  repo._adding = true
  try {
    await addRepo({ url: repo.url })
    repo._added = true
    // 刷新数据以获取后端返回的 _exists 和 repo_id
    await loadTrending()
  } catch (e) {
    // ignore
  } finally {
    repo._adding = false
  }
}

function formatNum(n) {
  if (!n) return '0'
  if (n >= 1000) return (n / 1000).toFixed(n % 1000 === 0 ? 0 : 1) + 'k'
  return n.toLocaleString()
}

// 右键菜单
const showCtxMenu = ref(false)
const contextMenuX = ref(0)
const contextMenuY = ref(0)
const contextMenuRepo = ref(null)

function handleContextMenu(e, repo) {
  e.preventDefault()
  contextMenuX.value = e.clientX
  contextMenuY.value = e.clientY
  contextMenuRepo.value = repo
  showCtxMenu.value = true
}

const contextMenuOptions = computed(() => {
  const repo = contextMenuRepo.value
  if (!repo) return []
  const options = []
  if (repo._exists || repo.repo_id) {
    options.push({ label: '查看详情', key: 'detail' })
    options.push({ type: 'divider', key: 'd1' })
  }
  if (!(repo._exists || repo._added)) {
    options.push({ label: '添加仓库', key: 'add' })
  } else {
    options.push({ label: '已添加', key: 'added', disabled: true })
  }
  options.push(
    { label: '在 GitHub 打开', key: 'github' },
    { label: '在 DeepWiki 打开', key: 'deepwiki' },
    { label: '在 ZRead 打开', key: 'zread' }
  )
  return options
})

function handleContextMenuAction(key) {
  const repo = contextMenuRepo.value
  showCtxMenu.value = false
  contextMenuRepo.value = null
  if (!repo) return
  switch (key) {
    case 'detail':
      openRepoDetail(repo)
      break
    case 'add':
      handleAddRepo(repo)
      break
    case 'github':
      window.open(repo.url, '_blank')
      break
    case 'deepwiki':
      window.open(`https://deepwiki.com/${repo.author}/${repo.repo}`, '_blank')
      break
    case 'zread':
      window.open(`https://zread.ai/${repo.author}/${repo.repo}`, '_blank')
      break
  }
}

// 仓库详情抽屉
const detailDrawer = reactive({ show: false, repo: null })

async function openRepoDetail(repo) {
  if (!repo?.repo_id) return
  try {
    const res = await getRepo(repo.repo_id)
    const resp = res.data
    if (resp && resp.code === 0 && resp.data) {
      detailDrawer.repo = resp.data
      detailDrawer.show = true
    }
  } catch (e) {
    console.error('加载仓库详情失败:', e)
  }
}

const updateLoading = ref(false)

function handleOpenRepo(url) {
  window.open(url, '_blank')
}

function handleViewCode(repo) {
  if (repo?.id) router.push(`/code/${repo.id}`)
}

async function handleUpdateInfo(repo) {
  try {
    updateLoading.value = true
    const { updateRepoInfo } = await import('@/api/repos')
    const response = await updateRepoInfo(repo.id)
    const apiData = response.data
    if (apiData && apiData.code === 0) {
      message.success('更新成功')
      detailDrawer.repo = { ...detailDrawer.repo, ...apiData.data }
    } else {
      message.error(apiData?.message || '更新失败')
    }
  } catch (e) {
    message.error('更新失败：' + e.message)
  } finally {
    updateLoading.value = false
  }
}

async function handleDelete(repo) {
  try {
    const { deleteRepo } = await import('@/api/repos')
    await deleteRepo(repo.id)
    message.success('仓库删除成功')
    detailDrawer.show = false
    loadTrending()
  } catch (e) {
    message.error('删除失败：' + e.message)
  }
}

async function handleToggleValid(repo) {
  try {
    const res = await toggleValid(repo.id)
    const apiData = res.data
    if (apiData && apiData.code === 0) {
      message.success(apiData.message)
      // 刷新详情
      await openRepoDetail({ repo_id: repo.id })
    } else {
      message.error(apiData?.message || '操作失败')
    }
  } catch (e) {
    message.error('操作失败：' + e.message)
  }
}

onMounted(() => {
  loadLanguages()
  loadTrending()
})
</script>

<style scoped>
.trending-view {
  padding: var(--space-6);
}

.page-header {
  margin-bottom: var(--space-5);
}

.page-title {
  font-size: var(--text-2xl);
  font-weight: var(--font-bold);
  color: var(--color-text-primary);
  margin: 0;
}

.page-description {
  font-size: var(--text-sm);
  color: var(--color-text-tertiary);
  margin-top: var(--space-1);
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.history-badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  background-color: var(--color-warning-50, #fff8e1);
  color: var(--color-warning-600, #e65100);
  font-size: var(--text-xs);
  font-weight: var(--font-medium);
}

.filter-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--space-3);
  margin-bottom: var(--space-5);
  flex-wrap: wrap;
}

.filter-left {
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.filter-right {
  display: flex;
  gap: var(--space-2);
}

/* 骨架屏 */
.trending-skeleton {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--space-3);
}

.skeleton-card {
  padding: var(--space-4);
  border-radius: var(--radius-lg);
  background-color: var(--color-bg-card);
  border: 1px solid var(--color-border-light);
}

.skeleton-line {
  height: 14px;
  background: linear-gradient(90deg, var(--color-gray-100) 25%, var(--color-gray-200) 50%, var(--color-gray-100) 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
  border-radius: 4px;
}

.skeleton-title {
  width: 40%;
  height: 18px;
  margin-bottom: 12px;
}

.skeleton-desc {
  width: 80%;
  margin-bottom: 12px;
}

.skeleton-meta {
  width: 50%;
}

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

/* 错误提示 */
.trending-error {
  text-align: center;
  padding: var(--space-10);
  color: var(--color-text-tertiary);
}

.error-icon {
  margin-bottom: var(--space-3);
  color: var(--color-text-quaternary);
}

.error-text {
  margin-bottom: var(--space-3);
}

/* 空状态 */
.trending-empty {
  text-align: center;
  padding: var(--space-10);
  color: var(--color-text-tertiary);
}

/* 项目列表 */
.trending-list {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--space-3);
}

.trending-card {
  padding: var(--space-4);
  border-radius: var(--radius-lg);
  background-color: var(--color-bg-card);
  border: 1px solid var(--color-border-light);
  transition: border-color 0.2s, box-shadow 0.2s;
}

.trending-card:hover {
  border-color: var(--color-primary-200);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.card-header {
  margin-bottom: var(--space-2);
}

.card-title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-1_5);
}

.repo-name {
  font-size: var(--text-base);
  font-weight: var(--font-semibold);
  color: var(--color-primary);
  text-decoration: none;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.repo-name:hover {
  text-decoration: underline;
}

.repo-desc {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  margin: 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  line-height: 1.5;
}

.card-meta {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  font-size: var(--text-xs);
  color: var(--color-text-tertiary);
  margin-top: var(--space-2);
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.lang-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.stars-today {
  color: var(--color-success);
  font-weight: var(--font-medium);
}

.ext-link {
  color: var(--color-text-tertiary);
  text-decoration: none;
  margin-left: auto;
  transition: color 0.2s;
}

.ext-link + .ext-link {
  margin-left: 0;
}

.ext-link:hover {
  color: var(--color-primary);
}

.card-built-by {
  display: flex;
  gap: 2px;
  margin-top: var(--space-2);
}

.contributor-avatar {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  border: 2px solid var(--color-bg-card);
}

/* 移动端 */
@media (max-width: 640px) {
  .trending-view {
    padding: var(--space-4);
  }

  .filter-left {
    width: 100%;
  }

  .filter-left :deep(.n-select),
  .filter-left :deep(.n-date-picker) {
    flex: 1;
    min-width: 0;
  }

  .card-title-row {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-1);
  }
}

</style>
