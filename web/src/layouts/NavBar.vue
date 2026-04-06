<template>
  <n-layout-header bordered class="n-layout-header header-no-padding">
    <div class="nav-bar-center">
      <div class="nav-bar-juejin">
        <div class="nav-left">
          <div class="brand" @click="goHome" style="cursor:pointer;">
            <div class="logo-text">
              <span class="logo-main">Git<span class="brand-accent">Plus</span></span>
              <span class="logo-shadow">Git<span class="brand-accent">Plus</span></span>
            </div>
          </div>
          <template v-if="isCreatorPage">
            <span class="center-title">创作者中心</span>
          </template>
          <template v-else>
            <n-menu
              class="menu-center"
              mode="horizontal"
              :options="menuOptions"
              :value="activeKey"
              @update:value="handleUpdateValue"
            />
            <div class="search-inline" ref="searchRef">
              <n-input 
                v-model:value="searchKeyword"
                round 
                placeholder="搜索仓库名称、作者或描述..." 
                class="search-input"
                clearable
                @keyup.enter="handleSearch"
                @update:value="handleSearchChange"
                @clear="handleSearchClear"
              >
                <template #prefix>
                  <n-icon><SearchOutline /></n-icon>
                </template>
              </n-input>
              <!-- 搜索建议下拉框 -->
              <div v-if="showSuggestions && searchSuggestions.length > 0" class="search-suggestions">
                <div 
                  v-for="suggestion in searchSuggestions" 
                  :key="suggestion.key"
                  class="suggestion-item"
                  @click="handleSuggestionClick(suggestion)"
                >
                  <n-icon size="16" class="suggestion-icon">
                    <SearchOutline />
                  </n-icon>
                  <span class="suggestion-text">{{ suggestion.label }}</span>
                  <span class="suggestion-type">{{ suggestion.type }}</span>
                </div>
              </div>
            </div>
          </template>
        </div>
        <div class="nav-right">
          <!-- 任务图标 -->
          <n-popover trigger="click" placement="bottom-end" :width="380" @update:show="handleTaskPopoverChange">
            <template #trigger>
              <n-badge :value="tasksStore.runningCount" :max="99" :show="tasksStore.runningCount > 0">
                <n-button quaternary circle class="task-icon-btn">
                  <template #icon>
                    <n-icon size="20"><ListOutline /></n-icon>
                  </template>
                </n-button>
              </n-badge>
            </template>
            <div class="task-popover">
              <div class="task-popover-header">
                <span class="task-popover-title">后台任务</span>
                <n-button v-if="tasksStore.taskList.some(t => t.status === 'completed' || t.status === 'failed')" text size="tiny" @click="tasksStore.clearCompleted()">清空已完成</n-button>
              </div>
              <div v-if="tasksStore.loading && tasksStore.taskList.length === 0" class="task-popover-empty">加载中...</div>
              <div v-else-if="tasksStore.taskList.length === 0" class="task-popover-empty">暂无任务</div>
              <div v-else class="task-popover-list">
                <div v-for="task in tasksStore.taskList.slice(0, 10)" :key="task.id" class="task-popover-item" @click="handleTaskClick(task)">
                  <div class="task-item-left">
                    <span class="task-item-type">{{ taskTypeName(task.type) }}</span>
                    <n-progress
                      v-if="task.status === 'running'"
                      type="line"
                      :percentage="task.total > 0 ? Math.round(((task.success_count + task.fail_count) / task.total) * 100) : 0"
                      :show-indicator="false"
                      status="info"
                      style="width: 80px; margin-left: 8px;"
                    />
                  </div>
                  <div class="task-item-right">
                    <span :class="['task-item-status', `status-${task.status}`]">{{ taskStatusName(task.status) }}</span>
                    <span class="task-item-time">{{ task.created_at }}</span>
                  </div>
                </div>
              </div>
            </div>
          </n-popover>

          <!-- 主题切换按钮 -->
          <span class="theme-toggle-btn" @click="themeStore.toggleTheme()">
            <n-icon size="20">
              <SunnyOutline v-if="isDark" />
              <MoonOutline v-else />
            </n-icon>
          </span>

          <template v-if="!userInfo">
            <n-space align="center">
              <n-avatar
                round
                size="medium"
                style="cursor: pointer; background: #4CAF50; color: white; font-weight: bold;"
                @click="goLogin"
              >
                G+
              </n-avatar>
              <n-button type="primary" round @click="goLogin">登录</n-button>
            </n-space>
          </template>
          <template v-else>
            <n-dropdown
              trigger="hover"
              :options="dropdownOptions"
              @select="onSelect"
            >
              <n-avatar
                round
                :src="userInfo.avatar"
                size="medium"
                style="cursor: pointer; background: #4CAF50; color: white; font-weight: bold;"
              >
                {{ userInfo.username ? userInfo.username.charAt(0).toUpperCase() : 'G+' }}
              </n-avatar>
            </n-dropdown>
          </template>
        </div>
      </div>
    </div>
  </n-layout-header>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NLayoutHeader, NMenu, NInput, NButton, NAvatar, NBadge, NIcon, NDropdown, NSpace, NPopover, NProgress, useMessage } from 'naive-ui'
// import { useUserStore } from '@/stores/user'
// import { storeToRefs } from 'pinia'
import { SearchOutline, ListOutline, SunnyOutline, MoonOutline } from '@vicons/ionicons5'
import { useThemeStore } from '@/stores/theme'
import { getRepos } from '@/api/repos'
import { useTasksStore } from '@/stores/tasks'

const tasksStore = useTasksStore()
const themeStore = useThemeStore()
const isDark = computed(() => themeStore.getCurrentTheme() === 'dark')

// 防抖函数
function debounce(func, wait) {
  let timeout
  return function executedFunction(...args) {
    const later = () => {
      clearTimeout(timeout)
      func(...args)
    }
    clearTimeout(timeout)
    timeout = setTimeout(later, wait)
  }
}

const route = useRoute()
const router = useRouter()
const isCreatorPage = computed(() => route.path.startsWith('/creator'))

const menuOptions = [
  { label: '首页', key: 'home' },
]

const activeKey = ref('home')
function handleUpdateValue(key) {
  activeKey.value = key
  if (key === 'home') router.push('/')
}
// const userStore = useUserStore()
// const { userInfo } = storeToRefs(userStore)
const userInfo = ref(null)

// 检查登录状态
function checkLoginStatus() {
  const token = localStorage.getItem('token')
  const savedUserInfo = localStorage.getItem('userInfo')
  
  if (token && savedUserInfo) {
    try {
      userInfo.value = JSON.parse(savedUserInfo)
    } catch (e) {
      console.error('解析用户信息失败:', e)
      userInfo.value = null
    }
  } else {
    userInfo.value = null
  }
}

// 任务相关辅助函数
function taskTypeName(type) {
  const map = { batch_pull: '一键拉取', batch_clone: '批量克隆', scan: '扫描仓库' }
  return map[type] || type
}

function taskStatusName(status) {
  const map = { pending: '等待中', running: '运行中', completed: '已完成', failed: '失败' }
  return map[status] || status
}

function handleTaskPopoverChange(show) {
  if (show) {
    tasksStore.fetchTasks({ page_size: 10 })
    tasksStore.startSSE()
  } else {
    tasksStore.stopSSE()
  }
}

function handleTaskClick(task) {
  // 如果有 running 任务，导航到首页仪表盘
  if (task.status === 'running') {
    router.push('/')
  }
}

// 组件挂载时检查登录状态
onMounted(() => {
  checkLoginStatus()

  // 监听存储变化（跨标签页同步）
  window.addEventListener('storage', checkLoginStatus)

  // 监听自定义登录事件
  window.addEventListener('userLoggedIn', checkLoginStatus)

  // 初始加载任务列表
  tasksStore.fetchTasks({ page_size: 10 })
  tasksStore.startSSE()
})

onUnmounted(() => {
  window.removeEventListener('storage', checkLoginStatus)
  window.removeEventListener('userLoggedIn', checkLoginStatus)
  tasksStore.stopSSE()
})

function goLogin() {
  router.push('/login')
}

function logout() {
  // userStore.logout()
  userInfo.value = null
  localStorage.removeItem('token')
  localStorage.removeItem('userInfo')
  router.push('/login')
}

function handleDropdown(key) {
  if (key === 'logout') logout()
}

function goHome() {
  router.push('/')
}

// 生成下拉菜单选项
const dropdownOptions = computed(() => {
  return [
    { label: '退出登录', key: 'logout' }
  ]
})

function onSelect(key) {
  if (key === 'logout') {
    logout()
  }
}

// 搜索相关
const searchKeyword = ref('')
const searchSuggestions = ref([])
const showSuggestions = ref(false)
const searchRef = ref(null)
const message = useMessage()

// 防抖搜索建议
const debouncedSearchSuggestions = debounce(async (keyword) => {
  if (!keyword || keyword.length < 2) {
    showSuggestions.value = false
    searchSuggestions.value = []
    return
  }
  
  try {
    // 获取搜索建议
    const response = await getRepos({
      search: keyword,
      page_size: 5,
      page: 1
    })
    
    if (!response || !response.data || !response.data.data) {
      throw new Error('Invalid response format')
    }
    
    const repos = response.data.data.list || []
    
    const suggestions = repos.map((repo, index) => ({
      key: `repo-${repo.id || index}`,
      label: `${repo.author || 'Unknown'}/${repo.repo || 'Unknown'}`,
      type: '仓库',
      repo: repo
    }))
    
    // 添加全局搜索选项
    suggestions.unshift({
      key: 'search-all',
      label: `搜索 "${keyword}"`,
      type: '全部结果',
      keyword: keyword
    })
    
    searchSuggestions.value = suggestions
    showSuggestions.value = suggestions.length > 0
  } catch (error) {
    console.warn('搜索建议获取失败:', error)
    // 只显示全局搜索
    const fallbackSuggestions = [{
      key: 'search-all',
      label: `搜索 "${keyword}"`,
      type: '全部结果',
      keyword: keyword
    }]
    searchSuggestions.value = fallbackSuggestions
    showSuggestions.value = true
  }
}, 300)

function handleSearchChange(value) {
  searchKeyword.value = value
  if (value) {
    debouncedSearchSuggestions(value)
  } else {
    showSuggestions.value = false
    searchSuggestions.value = []
    
    // 当搜索框清空时，触发空搜索以显示所有仓库
    if (route.path === '/') {
      window.dispatchEvent(new CustomEvent('navbar-search', {
        detail: { keyword: '' }
      }))
    }
  }
}

function handleSearchClear() {
  searchKeyword.value = ''
  showSuggestions.value = false
  searchSuggestions.value = []
  
  // 清空搜索时，显示所有仓库
  if (route.path === '/') {
    window.dispatchEvent(new CustomEvent('navbar-search', {
      detail: { keyword: '' }
    }))
  }
}

function handleSearch() {
  const keyword = searchKeyword.value?.trim()
  if (keyword) {
    showSuggestions.value = false
    try {
      // 触发搜索事件，通知HomeView组件
      window.dispatchEvent(new CustomEvent('navbar-search', {
        detail: { keyword }
      }))
      
      // 如果不在首页，则跳转到首页
      if (route.path !== '/') {
        router.push('/')
      }
    } catch (error) {
      console.error('搜索事件触发失败:', error)
    }
  }
}

function handleSuggestionClick(suggestion) {
  showSuggestions.value = false
  
  try {
    if (suggestion?.repo?.id) {
      // 点击仓库建议，在新标签页打开代码查看页面
      const url = router.resolve({ name: 'code', params: { id: suggestion.repo.id } }).href
      window.open(url, '_blank')
    } else if (suggestion?.keyword) {
      // 点击全部结果，触发搜索
      searchKeyword.value = suggestion.keyword
      window.dispatchEvent(new CustomEvent('navbar-search', {
        detail: { keyword: suggestion.keyword }
      }))
      
      // 如果不在首页，则跳转到首页
      if (route.path !== '/') {
        router.push('/')
      }
    }
  } catch (error) {
    console.error('处理搜索建议点击失败:', error)
  }
}

// 点击外部关闭建议框
function handleClickOutside(event) {
  if (searchRef.value && !searchRef.value.contains(event.target)) {
    showSuggestions.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.header-no-padding {
  padding: 0 !important;
}
.nav-bar-center {
  width: 80vw;
  max-width: 1280px;
  margin: 0 auto;
  display: flex;
  justify-content: center;
}
.nav-bar-juejin {
  width: 100%;
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--color-bg-navbar);
  box-sizing: border-box;
  padding: 0 32px;
  position: relative;
}
.nav-left {
  display: flex;
  align-items: center;
}
.brand {
  display: flex;
  align-items: center;
  margin-right: 36px;
  margin-left: 8px;
  position: relative;
  cursor: pointer;
  transition: all 0.3s ease;
}

.brand:hover {
  transform: scale(1.05) rotate(-1deg);
}

.brand:hover .logo-main {
  text-shadow: 0 0 20px rgba(76, 175, 80, 0.8);
  transform: translateY(-2px) scale(1.1);
}

.brand:hover .logo-shadow {
  opacity: 0.8;
  transform: translateY(4px) scale(1.1);
}

.logo-text {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.logo-main {
  font-size: 1.7rem;
  font-weight: 800;
  letter-spacing: 1.5px;
  color: var(--color-text-primary);
  font-family: 'Fira Code', 'Lato', 'Segoe UI', 'Arial', sans-serif;
  background: linear-gradient(90deg, #4CAF50 0%, #2196F3 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  position: relative;
  z-index: 2;
  animation: float 3s ease-in-out infinite;
  transition: all 0.3s ease;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.logo-shadow {
  font-size: 1.7rem;
  font-weight: 800;
  letter-spacing: 1.5px;
  color: rgba(100, 100, 100, 0.2);
  font-family: 'Fira Code', 'Lato', 'Segoe UI', 'Arial', sans-serif;
  position: absolute;
  top: 2px;
  left: 0;
  z-index: 1;
  animation: float-shadow 3s ease-in-out infinite;
  transition: all 0.3s ease;
  filter: blur(1px);
}

.brand-accent {
  color: #4CAF50;
  -webkit-text-fill-color: #4CAF50;
  background: none;
  font-weight: 900;
}

@keyframes float {
  0%, 100% {
    transform: translateY(0px);
  }
  50% {
    transform: translateY(-3px);
  }
}

@keyframes float-shadow {
  0%, 100% {
    transform: translateY(2px);
    opacity: 0.2;
  }
  50% {
    transform: translateY(5px);
    opacity: 0.3;
  }
}

/* 添加背景光效 */
.brand::before {
  content: '';
  position: absolute;
  top: -20px;
  left: -20px;
  right: -20px;
  bottom: -20px;
  background: radial-gradient(circle, rgba(76, 175, 80, 0.1) 0%, transparent 70%);
  opacity: 0;
  transition: opacity 0.3s ease;
  pointer-events: none;
}

.brand:hover::before {
  opacity: 1;
}
.menu-center {
  min-width: 360px;
}
.search-inline {
  margin-left: 32px;
  display: flex;
  align-items: center;
  position: relative; /* Added for suggestions positioning */
}
.search-input {
  width: 260px;
  background: var(--color-bg-page);
}
.search-suggestions {
  position: absolute;
  top: 100%;
  left: 0;
  width: 100%;
  background: var(--color-bg-card);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  box-shadow: var(--shadow-md);
  z-index: 10;
  max-height: 200px;
  overflow-y: auto;
  padding: 8px 0;
}
.suggestion-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  cursor: pointer;
  transition: background-color 0.2s ease;
}
.suggestion-item:hover {
  background-color: var(--color-gray-100);
}
.suggestion-icon {
  margin-right: 8px;
  color: var(--color-text-secondary);
}
.suggestion-text {
  flex-grow: 1;
  font-size: 0.9rem;
  color: var(--color-text-primary);
}
.suggestion-type {
  font-size: 0.7rem;
  color: var(--color-text-tertiary);
  margin-left: 8px;
}
.nav-right {
  display: flex;
  align-items: center;
  gap: 18px;
}

.task-icon-btn {
  color: var(--color-text-secondary);
}
.task-icon-btn:hover {
  color: var(--color-text-primary);
}
.theme-toggle-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  cursor: pointer;
  color: var(--color-text-secondary);
  transition: color 0.2s, background-color 0.2s;
}
.theme-toggle-btn:hover {
  color: var(--color-text-primary);
  background-color: var(--color-gray-100);
}

.task-popover {
  max-height: 400px;
  overflow-y: auto;
}
.task-popover-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--color-border-light);
}
.task-popover-title {
  font-weight: 600;
  font-size: 14px;
  color: var(--color-text-primary);
}
.task-popover-empty {
  text-align: center;
  color: var(--color-text-tertiary);
  padding: 20px 0;
  font-size: 13px;
}
.task-popover-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.task-popover-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 10px;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.15s;
}
.task-popover-item:hover {
  background: var(--color-bg-page);
}
.task-item-left {
  display: flex;
  align-items: center;
  gap: 4px;
}
.task-item-type {
  font-size: 13px;
  color: var(--color-text-primary);
  font-weight: 500;
}
.task-item-right {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 2px;
}
.task-item-status {
  font-size: 12px;
  font-weight: 500;
}
.task-item-status.status-running { color: var(--color-info-500); }
.task-item-status.status-completed { color: var(--color-success); }
.task-item-status.status-failed { color: var(--color-error-500); }
.task-item-status.status-pending { color: var(--color-text-tertiary); }
.task-item-time {
  font-size: 11px;
  color: var(--color-text-quaternary);
}

.dropdown-arrow {
  margin-left: 2px;
  vertical-align: middle;
  display: inline-block;
}
.icon-badge {
  margin-right: 8px;
}
:deep(.menu-center .n-menu--horizontal .n-menu__content) {
  justify-content: flex-start !important;
}
:deep(.n-layout-header__content) {
  padding: 0 !important;
}
.user-name {
  margin: 0 10px;
  font-weight: 500;
  color: var(--color-text-primary);
}
.logout-btn {
  margin-left: 4px;
  color: var(--color-text-tertiary);
}
.n-layout-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 200;
  background: var(--color-bg-navbar);
  box-shadow: var(--shadow-sm);
}
.center-title {
  font-size: 1.25rem;
  color: var(--color-primary);
  font-weight: 700;
  margin-left: 18px;
}
.navbar {
  display: flex;
  align-items: center;
  height: 56px;
  padding: 0 24px;
  background: var(--color-bg-navbar);
  box-shadow: var(--shadow-sm);
}
</style> 