<template>
  <div class="timeline-view">
    <!-- 顶部时间导航条 -->
    <div class="timeline-nav" v-if="!loading && repos.length > 0">
      <div class="timeline-nav__content">
        <button
          v-for="(group, index) in timelineGroups"
          :key="group.key"
          class="timeline-nav__item"
          :class="{ 'timeline-nav__item--active': activeGroup === group.key }"
          @click="scrollToGroup(group.key)"
        >
          <span class="timeline-nav__label">{{ group.label }}</span>
          <span class="timeline-nav__count">{{ group.count }}</span>
        </button>
      </div>
    </div>

    <!-- 主内容区域 -->
    <div class="timeline-view__content">
      <!-- 加载状态 -->
      <template v-if="loading">
        <div class="timeline-skeleton">
          <div v-for="i in 3" :key="i" class="timeline-group-skeleton">
            <n-skeleton height="40px" width="200px" style="margin-bottom: 24px;" />
            <div class="timeline-items-skeleton">
              <div v-for="j in 2" :key="j" class="timeline-item-skeleton">
                <n-skeleton height="180px" :sharp="false" style="margin-bottom: 16px;" />
              </div>
            </div>
          </div>
        </div>
      </template>

      <!-- 时间轴内容 -->
      <template v-else-if="repos.length > 0">
        <div class="timeline-container">
          <div
            v-for="(group, groupKey) in groupedRepos"
            :key="groupKey"
            :id="`group-${groupKey}`"
            class="timeline-group"
            :class="{ 'timeline-group--active': activeGroup === groupKey }"
          >
            <!-- 时间标题 -->
            <div class="timeline-header">
              <div class="timeline-header__line"></div>
              <div class="timeline-header__content">
                <h2 class="timeline-header__title">{{ formatGroupTitle(groupKey) }}</h2>
                <n-tag size="large" round type="info" class="timeline-header__count">
                  {{ group.length }} 个项目
                </n-tag>
              </div>
              <div class="timeline-header__line"></div>
            </div>

            <!-- 仓库列表 -->
            <div class="timeline-items">
              <div
                v-for="(repo, index) in group"
                :key="repo.id"
                class="timeline-item"
                :class="{ 'timeline-item--first': index === 0 }"
              >
                <!-- 时间指示器 -->
                <div class="timeline-indicator">
                  <div class="timeline-dot">
                    <div class="timeline-dot__core"></div>
                  </div>
                  <div v-if="index < group.length - 1" class="timeline-connector"></div>
                </div>

                <!-- 仓库卡片 -->
                <div class="timeline-card">
                  <!-- 日期标签 -->
                  <div class="timeline-card__date">
                    <n-text depth="3" style="font-size: 13px;">
                      {{ formatDateDetail(repo.created_at) }}
                    </n-text>
                  </div>

                  <!-- 卡片主体 -->
                  <div class="timeline-card__body" @click="handleRepoClick(repo)">
                    <!-- 头部 -->
                    <div class="card-header">
                      <div class="card-header__main">
                        <n-icon size="20" color="#18a058" class="card-icon">
                          <GitBranchOutline />
                        </n-icon>
                        <n-ellipsis style="max-width: 320px;">
                          <n-text strong style="font-size: 17px;">
                            {{ repo.author }}/<span class="repo-name">{{ repo.repo }}</span>
                          </n-text>
                        </n-ellipsis>
                      </div>
                      <div class="card-header__status">
                        <n-tag
                          v-if="getRepoStatus(repo.id)?.exists"
                          type="success"
                          size="small"
                          round
                          bordered
                        >
                          已克隆
                        </n-tag>
                      </div>
                    </div>

                    <!-- 描述 -->
                    <div class="card-description">
                      <n-ellipsis :line-clamp="2" :tooltip="false">
                        <n-text depth="2">
                          {{ repo.description || '暂无描述' }}
                        </n-text>
                      </n-ellipsis>
                    </div>

                    <!-- 元数据 -->
                    <div class="card-meta">
                      <div class="meta-item">
                        <n-icon size="15"><StarOutline /></n-icon>
                        <span>{{ formatNumber(repo.stars) }}</span>
                      </div>
                      <div class="meta-item">
                        <n-icon size="15"><GitBranchOutline /></n-icon>
                        <span>{{ formatNumber(repo.forks) }}</span>
                      </div>
                      <div class="meta-item" v-if="mainLanguage(repo)">
                        <div
                          class="language-dot"
                          :style="{ backgroundColor: getLanguageColor(mainLanguage(repo)) }"
                        ></div>
                        <span>{{ mainLanguage(repo) }}</span>
                      </div>
                      <div class="meta-divider"></div>
                      <div class="meta-actions">
                        <n-button-group size="small">
                          <n-button
                            quaternary
                            size="small"
                            @click.stop="$emit('view-code', repo)"
                          >
                            <template #icon>
                              <n-icon size="15"><CodeSlashOutline /></n-icon>
                            </template>
                            代码
                          </n-button>
                          <n-button
                            v-if="getRepoStatus(repo.id)?.exists"
                            quaternary
                            size="small"
                            @click.stop="$emit('open-folder', repo)"
                          >
                            <template #icon>
                              <n-icon size="15"><FolderOutline /></n-icon>
                            </template>
                            文件夹
                          </n-button>
                          <n-dropdown
                            trigger="click"
                            :options="getActionOptions(repo)"
                            @select="(key) => handleAction(key, repo)"
                          >
                            <n-button quaternary size="small">
                              <template #icon>
                                <n-icon size="15"><EllipsisVerticalOutline /></n-icon>
                              </template>
                            </n-button>
                          </n-dropdown>
                        </n-button-group>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 返回顶部按钮 -->
        <transition name="fade">
          <button
            v-if="showBackToTop"
            class="back-to-top"
            @click="scrollToTop"
          >
            <n-icon size="20">
              <ArrowUpOutline />
            </n-icon>
          </button>
        </transition>
      </template>

      <!-- 空状态 -->
      <div v-else class="timeline-empty">
        <n-empty description="暂无仓库数据" size="large">
          <template #icon>
            <n-icon size="80" depth="3">
              <TimeOutline />
            </n-icon>
          </template>
        </n-empty>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import {
  TimeOutline,
  StarOutline,
  GitBranchOutline,
  CodeSlashOutline,
  FolderOutline,
  EllipsisVerticalOutline,
  ArrowUpOutline,
  PlayOutline,
  DownloadOutline,
  RefreshOutline,
  TrashOutline
} from '@vicons/ionicons5'
import { h } from 'vue'

const props = defineProps({
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
  total: {
    type: Number,
    default: 0
  },
  groupBy: {
    type: String,
    default: 'month',
    validator: (value) => ['year', 'month', 'week', 'day'].includes(value)
  },
  showToolbar: {
    type: Boolean,
    default: true
  }
})

const emit = defineEmits([
  'repo-click',
  'view-code',
  'open-folder',
  'clone',
  'pull',
  'delete',
  'update-info',
  'share',
  'group-by-change'
])

// 响应式数据
const activeGroup = ref(null)
const showBackToTop = ref(false)
const observer = ref(null)

// 计算属性 - 分组后的仓库
const groupedRepos = computed(() => {
  const groups = {}

  props.repos.forEach(repo => {
    const date = new Date(repo.created_at)
    let groupKey = ''

    switch (props.groupBy) {
      case 'year':
        groupKey = date.getFullYear().toString()
        break
      case 'month':
        groupKey = `${date.getFullYear()}-${(date.getMonth() + 1).toString().padStart(2, '0')}`
        break
      case 'week':
        const weekStart = new Date(date)
        weekStart.setDate(date.getDate() - date.getDay())
        groupKey = `${weekStart.getFullYear()}-W${Math.ceil((weekStart - new Date(weekStart.getFullYear(), 0, 1)) / (7 * 24 * 60 * 60 * 1000))}`
        break
      case 'day':
        groupKey = `${date.getFullYear()}-${(date.getMonth() + 1).toString().padStart(2, '0')}-${date.getDate().toString().padStart(2, '0')}`
        break
    }

    if (!groups[groupKey]) {
      groups[groupKey] = []
    }
    groups[groupKey].push(repo)
  })

  // 对每个分组内的仓库按创建时间排序
  Object.keys(groups).forEach(key => {
    groups[key].sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
  })

  // 返回按时间倒序的分组
  const sortedGroups = {}
  Object.keys(groups)
    .sort((a, b) => b.localeCompare(a))
    .forEach(key => {
      sortedGroups[key] = groups[key]
    })

  return sortedGroups
})

// 时间导航分组
const timelineGroups = computed(() => {
  return Object.keys(groupedRepos.value).map(key => ({
    key,
    label: formatGroupTitle(key),
    count: groupedRepos.value[key].length
  }))
})

// 方法
const getRepoStatus = (repoId) => {
  return props.repoStatuses[repoId] || null
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
  return colors[language] || 'var(--color-text-secondary)'
}

const formatNumber = (num) => {
  if (!num) return '0'
  if (num >= 1000) {
    return (num / 1000).toFixed(1) + 'k'
  }
  return num.toString()
}

const formatGroupTitle = (groupKey) => {
  if (props.groupBy === 'year') {
    return groupKey + '年'
  } else if (props.groupBy === 'month') {
    const [year, month] = groupKey.split('-')
    return `${year}年${month}月`
  } else if (props.groupBy === 'week') {
    const [year, week] = groupKey.split('-W')
    return `${year}年 第${week}周`
  } else {
    const [year, month, day] = groupKey.split('-')
    return `${year}年${month}月${day}日`
  }
}

const formatDateDetail = (dateString) => {
  if (!dateString) return ''
  const date = new Date(dateString)
  return date.toLocaleDateString('zh-CN', {
    month: 'long',
    day: 'numeric',
    weekday: 'short'
  })
}

const handleRepoClick = (repo) => {
  emit('repo-click', repo)
}

const scrollToGroup = (groupKey) => {
  const element = document.getElementById(`group-${groupKey}`)
  if (element) {
    element.scrollIntoView({ behavior: 'smooth', block: 'start' })
    activeGroup.value = groupKey
  }
}

const scrollToTop = () => {
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

const getActionOptions = (repo) => {
  const repoStatus = getRepoStatus(repo.id)
  return [
    {
      label: '克隆仓库',
      key: 'clone',
      icon: () => h('span', '⬇️'),
      show: !repoStatus?.exists
    },
    {
      label: '拉取最新',
      key: 'pull',
      icon: () => h('span', '🔄'),
      show: repoStatus?.exists
    },
    { type: 'divider' },
    {
      label: '更新信息',
      key: 'update-info',
      icon: () => h('span', '📋')
    },
    {
      label: '删除仓库',
      key: 'delete',
      icon: () => h('span', '🗑️'),
      props: {
        style: 'color: #e74c3c;'
      }
    }
  ].filter(option => option.show !== false)
}

const handleAction = (key, repo) => {
  switch (key) {
    case 'clone':
      emit('clone', repo)
      break
    case 'pull':
      emit('pull', repo)
      break
    case 'delete':
      emit('delete', repo)
      break
    case 'update-info':
      emit('update-info', repo)
      break
  }
}

// 滚动监听
const handleScroll = () => {
  const scrollTop = window.pageYOffset || document.documentElement.scrollTop
  showBackToTop.value = scrollTop > 400

  // 检测当前可见的分组
  const groupElements = document.querySelectorAll('.timeline-group')
  groupElements.forEach(el => {
    const rect = el.getBoundingClientRect()
    if (rect.top <= 150 && rect.bottom >= 150) {
      const groupKey = el.id.replace('group-', '')
      if (activeGroup.value !== groupKey) {
        activeGroup.value = groupKey
      }
    }
  })
}

// 生命周期
onMounted(() => {
  window.addEventListener('scroll', handleScroll)
  // 初始化激活第一个分组
  if (timelineGroups.value.length > 0) {
    activeGroup.value = timelineGroups.value[0].key
  }
})

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
  if (observer.value) {
    observer.value.disconnect()
  }
})
</script>

<style scoped>
.timeline-view {
  width: 100%;
  position: relative;
}

/* 时间导航条 */
.timeline-nav {
  position: sticky;
  top: 0;
  z-index: 100;
  background: var(--color-bg-navbar);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid var(--color-border);
  padding: 12px 0;
  box-shadow: var(--shadow-sm);
}

.timeline-nav__content {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  overflow-y: hidden;
  padding: 0 16px;
  scrollbar-width: none;
}

.timeline-nav__content::-webkit-scrollbar {
  display: none;
}

.timeline-nav__item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 8px 16px;
  border: none;
  background: transparent;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
  min-width: 80px;
}

.timeline-nav__item:hover {
  background: var(--color-gray-100);
}

.timeline-nav__item--active {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.timeline-nav__label {
  font-size: 13px;
  font-weight: 500;
}

.timeline-nav__count {
  font-size: 11px;
  opacity: 0.7;
}

/* 内容区域 */
.timeline-view__content {
  min-height: 400px;
  padding: 24px 16px;
}

/* 时间轴容器 */
.timeline-container {
  max-width: 900px;
  margin: 0 auto;
}

/* 时间组 */
.timeline-group {
  margin-bottom: 64px;
  scroll-margin-top: 80px;
  transition: opacity 0.3s ease;
}

.timeline-group--active {
  opacity: 1;
}

/* 时间标题 */
.timeline-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 32px;
}

.timeline-header__line {
  flex: 1;
  height: 2px;
  background: linear-gradient(to right, transparent, var(--color-border), transparent);
}

.timeline-header__content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 0 16px;
}

.timeline-header__title {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: var(--color-text-primary);
}

.timeline-header__count {
  font-size: 12px;
}

/* 时间项目 */
.timeline-items {
  position: relative;
  padding-left: 40px;
}

.timeline-item {
  position: relative;
  margin-bottom: 24px;
}

.timeline-item--first {
  animation: slideIn 0.4s ease-out;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 时间指示器 */
.timeline-indicator {
  position: absolute;
  left: 0;
  top: 16px;
  width: 20px;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.timeline-dot {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  background: var(--color-bg-card);
  border: 3px solid var(--color-success);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2;
  transition: all 0.3s ease;
  box-shadow: 0 2px 8px rgba(24, 160, 88, 0.2);
}

.timeline-dot__core {
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: var(--color-success);
}

.timeline-item:hover .timeline-dot {
  transform: scale(1.3);
  box-shadow: 0 4px 12px rgba(24, 160, 88, 0.3);
}

.timeline-connector {
  flex: 1;
  width: 2px;
  background: linear-gradient(to bottom, var(--color-success) 0%, transparent 100%);
  margin-top: 4px;
  min-height: 40px;
}

/* 卡片 */
.timeline-card {
  position: relative;
  margin-left: 20px;
  padding: 20px;
  background: var(--color-bg-card);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  box-shadow: var(--shadow-sm);
  transition: all 0.3s ease;
}

.timeline-card:hover {
  border-color: var(--color-success);
  box-shadow: 0 8px 24px rgba(24, 160, 88, 0.12);
  transform: translateX(4px);
}

.timeline-card::before {
  content: '';
  position: absolute;
  left: -8px;
  top: 20px;
  width: 0;
  height: 0;
  border-style: solid;
  border-width: 6px 8px 6px 0;
  border-color: transparent var(--color-border) transparent transparent;
  transition: border-color 0.3s ease;
}

.timeline-card:hover::before {
  border-color: transparent var(--color-success) transparent transparent;
}

.timeline-card__date {
  position: absolute;
  top: -12px;
  left: 16px;
  background: var(--color-bg-card);
  padding: 0 8px;
}

/* 卡片内容 */
.timeline-card__body {
  cursor: pointer;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 12px;
  padding-top: 8px;
}

.card-header__main {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  min-width: 0;
}

.card-icon {
  flex-shrink: 0;
}

.repo-name {
  color: #18a058;
}

.card-description {
  margin-bottom: 16px;
  line-height: 1.6;
  min-height: 40px;
}

.card-meta {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: var(--color-text-secondary);
}

.language-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.meta-divider {
  flex: 1;
  min-width: 8px;
}

.meta-actions {
  flex-shrink: 0;
}

/* 骨架屏 */
.timeline-skeleton {
  max-width: 900px;
  margin: 0 auto;
  padding: 24px 16px;
}

.timeline-group-skeleton {
  margin-bottom: 64px;
}

/* 空状态 */
.timeline-empty {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 500px;
  padding: 48px 24px;
}

/* 返回顶部 */
.back-to-top {
  position: fixed;
  bottom: 32px;
  right: 32px;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  color: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
  transition: all 0.3s ease;
  z-index: 90;
}

.back-to-top:hover {
  transform: translateY(-4px);
  box-shadow: 0 6px 20px rgba(102, 126, 234, 0.5);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .timeline-view__content {
    padding: 16px 12px;
  }

  .timeline-container {
    max-width: 100%;
  }

  .timeline-group {
    margin-bottom: 48px;
  }

  .timeline-header__title {
    font-size: 20px;
  }

  .timeline-items {
    padding-left: 32px;
  }

  .timeline-card {
    padding: 16px;
  }

  .card-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .card-meta {
    gap: 12px;
  }

  .meta-actions {
    width: 100%;
    margin-top: 8px;
  }

  .back-to-top {
    bottom: 20px;
    right: 20px;
    width: 40px;
    height: 40px;
  }
}

@media (max-width: 480px) {
  .timeline-nav__item {
    padding: 6px 12px;
    min-width: 70px;
  }

  .timeline-nav__label {
    font-size: 12px;
  }

  .timeline-items {
    padding-left: 24px;
  }

  .timeline-indicator {
    left: -8px;
  }

  .timeline-card {
    margin-left: 12px;
  }
}
</style>
