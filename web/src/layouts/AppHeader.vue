<template>
  <header class="app-header" :class="{ collapsed: collapsed }">
    <div class="header-content">
      <!-- 左侧: 折叠按钮 + 当前页面名称 -->
      <div class="header-left">
        <!-- 侧边栏折叠按钮 -->
        <button
          class="header-action-btn"
          @click="handleToggleSidebar"
          aria-label="切换侧边栏"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="3" y1="12" x2="21" y2="12"/>
            <line x1="3" y1="6" x2="21" y2="6"/>
            <line x1="3" y1="18" x2="21" y2="18"/>
          </svg>
        </button>

        <!-- 当前页面名称 -->
        <span class="page-title">{{ currentPageTitle }}</span>
      </div>

      <!-- 右侧: 主题切换 + 用户 -->
      <div class="header-right">
        <!-- 主题切换按钮 -->
        <button
          class="header-action-btn theme-toggle-btn"
          @click="themeStore.toggleTheme()"
          :aria-label="isDark ? '切换到亮色模式' : '切换到暗色模式'"
        >
          <n-icon size="20">
            <SunnyOutline v-if="isDark" />
            <MoonOutline v-else />
          </n-icon>
        </button>

        <!-- 用户信息 -->
        <div class="header-user">
          <n-dropdown
            :options="userMenuOptions"
            @select="handleUserMenuSelect"
            trigger="click"
            placement="bottom-end"
          >
            <button class="user-btn">
              <!-- Forks 头像 -->
              <div class="user-avatar">
                <svg viewBox="0 0 24 24" fill="none">
                  <circle cx="12" cy="12" r="12" fill="currentColor"/>
                  <path
                    d="M8 9.5C8 8.67 8.67 8 9.5 8H12C13.66 8 15 9.34 15 11C15 12.66 13.66 14 12 14H10V16H8V9.5Z"
                    fill="white"
                  />
                  <path
                    d="M10 10V12H12C12.55 12 13 11.55 13 11C13 10.45 12.55 10 12 10H10Z"
                    fill="currentColor"
                  />
                  <circle cx="15.5" cy="15.5" r="2" fill="white"/>
                </svg>
              </div>
              <span class="user-name">{{ userName }}</span>
              <svg class="user-dropdown-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="6 9 12 15 18 9"/>
              </svg>
            </button>
          </n-dropdown>
        </div>
      </div>
    </div>
  </header>
</template>

<script setup>
/**
 * AppHeader - 顶部标题栏
 *
 * 功能特性:
 * - 面包屑导航
 * - 用户信息下拉菜单
 * - 响应式设计
 */
import { ref, computed, h } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NDropdown, NIcon, useMessage } from 'naive-ui'
import { LogOutOutline, SunnyOutline, MoonOutline } from '@vicons/ionicons5'
import { useThemeStore } from '@/stores/theme'

// Props
const props = defineProps({
  collapsed: {
    type: Boolean,
    default: false
  }
})

// Emits
const emit = defineEmits(['toggle-sidebar', 'toggle-mobile-sidebar'])

// 依赖
const route = useRoute()
const router = useRouter()
const message = useMessage()
const themeStore = useThemeStore()
const isDark = computed(() => themeStore.getCurrentTheme() === 'dark')

// 状态
const userName = ref('Forks')

// 当前页面标题
const currentPageTitle = computed(() => {
  const titleMap = {
    '/': '首页',
    '/authors': '作者',
    '/repos': '仓库',
    '/settings': '设置',
    '/activities': '活动'
  }
  return titleMap[route.path] || route.meta?.title || '首页'
})

// 用户菜单选项
const userMenuOptions = [
  {
    label: '退出登录',
    key: 'logout',
    icon: () => h(NIcon, null, { default: () => h(LogOutOutline) }),
    props: {
      style: {
        color: 'var(--color-error-500)'
      }
    }
  }
]

// 面包屑导航
const breadcrumbItems = computed(() => {
  const pathSegments = route.path.split('/').filter(Boolean)
  const items = []

  // 构建面包屑路径
  let currentPath = ''
  for (let i = 0; i < pathSegments.length; i++) {
    const segment = pathSegments[i]
    currentPath += `/${segment}`

    // 路径标题映射
    const titleMap = {
      'authors': '作者管理',
      'repos': '仓库管理',
      'settings': '系统设置',
      'activities': '活动记录',
      'detail': '详情'
    }

    const title = titleMap[segment] || segment
    const isLast = i === pathSegments.length - 1

    items.push({
      title,
      path: isLast ? null : currentPath
    })
  }

  return items
})

// 侧边栏切换
const handleToggleSidebar = () => {
  emit('toggle-sidebar')
}

// 用户菜单选择
const handleUserMenuSelect = (key) => {
  if (key === 'logout') {
    handleLogout()
  }
}

// 退出登录
const handleLogout = () => {
  // 清除本地存储的认证信息
  localStorage.removeItem('token')
  localStorage.removeItem('userInfo')

  // 提示用户
  message.success('已退出登录')

  // 跳转到登录页
  router.push('/login')
}
</script>

<style scoped>
/* ============================================
   HEADER STRUCTURE - 头部结构
   ============================================ */

.app-header {
  position: fixed;
  top: 0;
  right: 0;
  left: var(--sidebar-width-expanded);
  height: var(--navbar-height);
  background-color: var(--color-bg-card);
  border-bottom: 1px solid var(--color-border-light);
  z-index: calc(var(--z-sticky) - 1);
  transition: left 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

/* 折叠状态 */
@media (min-width: 1024px) {
  .app-header.collapsed {
    left: var(--sidebar-width-collapsed);
  }
}

/* 移动端 */
@media (max-width: 1023px) {
  .app-header {
    left: 0;
  }
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  padding: 0 var(--space-6);
}

@media (max-width: 768px) {
  .header-content {
    padding: 0 var(--space-4);
  }
}

/* ============================================
   HEADER LEFT - 左侧区域
   ============================================ */

.header-left {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  flex: 1;
  min-width: 0;
}

.header-action-btn {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: transparent;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.header-action-btn:hover {
  background-color: var(--color-gray-100);
  color: var(--color-text-primary);
}

.header-action-btn svg {
  width: 20px;
  height: 20px;
}

/* 当前页面标题 */
.page-title {
  font-size: var(--text-base);
  font-weight: var(--font-semibold);
  color: var(--color-text-primary);
  white-space: nowrap;
}

/* ============================================
   HEADER RIGHT - 右侧区域
   ============================================ */

.header-right {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

/* 用户信息 */
.header-user {
  display: flex;
  align-items: center;
}

.user-btn {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-1) var(--space-3);
  background-color: transparent;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all 0.2s ease;
}

.user-btn:hover {
  background-color: var(--color-gray-100);
}

.user-avatar {
  width: 32px;
  height: 32px;
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-primary-500) 100%);
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-primary);
  flex-shrink: 0;
  box-shadow: 0 2px 4px rgba(37, 99, 235, 0.2);
}

.user-avatar svg {
  width: 20px;
  height: 20px;
}

.user-name {
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  color: var(--color-text-primary);
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-dropdown-icon {
  width: 16px;
  height: 16px;
  color: var(--color-text-tertiary);
  flex-shrink: 0;
}

@media (max-width: 640px) {
  .user-name {
    display: none;
  }

  .user-btn {
    padding: var(--space-1);
  }
}
</style>
