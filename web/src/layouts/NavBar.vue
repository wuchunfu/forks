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
import { NLayoutHeader, NMenu, NInput, NButton, NAvatar, NBadge, NIcon, NDropdown, NSpace, useMessage } from 'naive-ui'
// import { useUserStore } from '@/stores/user'
// import { storeToRefs } from 'pinia'
import { SearchOutline } from '@vicons/ionicons5'
import { getRepos } from '@/api/repos'

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

// 组件挂载时检查登录状态
onMounted(() => {
  checkLoginStatus()
  
  // 监听存储变化（跨标签页同步）
  window.addEventListener('storage', checkLoginStatus)
  
  // 监听自定义登录事件
  window.addEventListener('userLoggedIn', checkLoginStatus)
})

onUnmounted(() => {
  window.removeEventListener('storage', checkLoginStatus)
  window.removeEventListener('userLoggedIn', checkLoginStatus)
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
  background: #fff;
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
  color: #333;
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
  color: rgba(51, 51, 51, 0.2);
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
  background: #f5f6fa;
}
.search-suggestions {
  position: absolute;
  top: 100%; /* Position below the input */
  left: 0;
  width: 100%;
  background: #fff;
  border: 1px solid #eee;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  z-index: 10;
  max-height: 200px; /* Limit height for suggestions */
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
  background-color: #f0f0f0;
}
.suggestion-icon {
  margin-right: 8px;
  color: #666;
}
.suggestion-text {
  flex-grow: 1;
  font-size: 0.9rem;
  color: #333;
}
.suggestion-type {
  font-size: 0.7rem;
  color: #999;
  margin-left: 8px;
}
.nav-right {
  display: flex;
  align-items: center;
  gap: 18px;
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
  color: #333;
}
.logout-btn {
  margin-left: 4px;
  color: #888;
}
.n-layout-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 200;
  background: #fff;
  box-shadow: 0 2px 8px 0 rgba(60,60,60,0.04);
}
.center-title {
  font-size: 1.25rem;
  color: #1E80FF;
  font-weight: 700;
  margin-left: 18px;
}
.navbar {
  display: flex;
  align-items: center;
  height: 56px;
  padding: 0 24px;
  background: #fff;
  box-shadow: 0 2px 8px rgba(0,0,0,0.04);
}
</style> 