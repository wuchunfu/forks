<template>
  <div class="trending-view">
    <div class="page-header">
      <h1 class="page-title">GitHub Trending</h1>
      <p class="page-description">发现 GitHub 上最受欢迎的开源项目</p>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <div class="filter-left">
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
        <n-button size="small" @click="loadTrending" :loading="loading">
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
      <n-button size="small" @click="loadTrending">重试</n-button>
    </div>

    <!-- 空状态 -->
    <div v-else-if="trendingRepos.length === 0" class="trending-empty">
      <p>暂无趋势数据</p>
    </div>

    <!-- 项目列表 -->
    <div v-else class="trending-list">
      <div v-for="repo in trendingRepos" :key="repo.url" class="trending-card">
        <div class="card-header">
          <div class="card-title-row">
            <a :href="repo.url" target="_blank" rel="noopener" class="repo-name">
              {{ repo.author }} / {{ repo.repo }}
            </a>
            <n-button
              size="tiny"
              :type="repo._added ? 'success' : 'default'"
              ghost
              @click="handleAddRepo(repo)"
              :loading="repo._adding"
            >
              {{ repo._added ? '已添加' : '+ 添加' }}
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
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { getTrending, getTrendingLanguages } from '@/api/trending'
import { addRepo } from '@/api/repos'

const selectedLanguage = ref(null)
const selectedSpokenLanguage = ref(null)
const selectedSince = ref('daily')

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

const sinceLabel = computed(() => {
  const map = { daily: 'today', weekly: 'this week', monthly: 'this month' }
  return map[selectedSince.value] || 'today'
})

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

async function loadTrending() {
  loading.value = true
  error.value = null
  try {
    const params = {
      since: selectedSince.value
    }
    if (selectedLanguage.value) params.language = selectedLanguage.value
    if (selectedSpokenLanguage.value) params.spoken_language_code = selectedSpokenLanguage.value

    const res = await getTrending(params)
    trendingRepos.value = (res.data.data.items || []).map(r => ({
      ...r,
      _adding: false,
      _added: false
    }))
  } catch (e) {
    error.value = e.response?.data?.message || e.message || '加载失败'
  } finally {
    loading.value = false
  }
}

async function handleAddRepo(repo) {
  if (repo._added) return
  repo._adding = true
  try {
    await addRepo({ url: repo.url })
    repo._added = true
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

watch([selectedLanguage, selectedSpokenLanguage, selectedSince], () => {
  loadTrending()
})

onMounted(() => {
  loadLanguages()
  loadTrending()
})
</script>

<style scoped>
.trending-view {
  padding: var(--space-6);
  max-width: 100%;
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

  .filter-left :deep(.n-select) {
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
