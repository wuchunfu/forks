<template>
  <aside
    class="app-sidebar"
    :class="{
      'sidebar-collapsed': collapsed,
      'sidebar-mobile-open': mobileOpen
    }"
    :style="{
      width: collapsed ? 'var(--sidebar-width-collapsed)' : 'var(--sidebar-width-expanded)'
    }"
  >
    <!-- 顶部区域：Logo -->
    <div class="sidebar-header">
      <router-link to="/" class="sidebar-brand">
        <svg class="brand-icon" viewBox="0 0 40 40" fill="none">
          <rect width="40" height="40" rx="8" fill="#2563EB"/>
          <path d="M20 10L28 18M20 10L12 18M20 10V28" stroke="white" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
        <transition name="brand-text">
          <span v-show="!collapsed" class="brand-text">Forks</span>
        </transition>
      </router-link>
    </div>

    <!-- 导航菜单 -->
    <nav class="sidebar-nav">
      <ul class="nav-list">
        <!-- 首页 -->
        <li class="nav-item">
          <router-link
            to="/"
            class="nav-link"
            :class="{ 'nav-link-active': isActiveRoute('/') }"
            :title="collapsed ? '首页' : ''"
          >
            <div class="nav-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>
                <polyline points="9 22 9 12 15 12 15 22"/>
              </svg>
            </div>
            <transition name="nav-text">
              <span v-show="!collapsed" class="nav-text">首页</span>
            </transition>
          </router-link>
        </li>

        <!-- 作者 -->
        <li class="nav-item">
          <router-link
            to="/authors"
            class="nav-link"
            :class="{ 'nav-link-active': isActiveRoute('/authors') }"
            :title="collapsed ? '作者' : ''"
          >
            <div class="nav-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
                <circle cx="9" cy="7" r="4"/>
                <path d="M23 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75"/>
              </svg>
            </div>
            <transition name="nav-text">
              <span v-show="!collapsed" class="nav-text">作者</span>
            </transition>
          </router-link>
        </li>

        <!-- 仓库 -->
        <li class="nav-item">
          <router-link
            to="/repos"
            class="nav-link"
            :class="{ 'nav-link-active': isActiveRoute('/repos') }"
            :title="collapsed ? '仓库' : ''"
          >
            <div class="nav-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
              </svg>
            </div>
            <transition name="nav-text">
              <span v-show="!collapsed" class="nav-text">仓库</span>
            </transition>
          </router-link>
        </li>

        <!-- 任务 -->
        <li class="nav-item">
          <router-link
            to="/tasks"
            class="nav-link"
            :class="{ 'nav-link-active': isActiveRoute('/tasks') }"
            :title="collapsed ? '任务' : ''"
          >
            <div class="nav-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="3" y="3" width="18" height="18" rx="2"/>
                <line x1="8" y1="8" x2="16" y2="8"/>
                <line x1="8" y1="12" x2="16" y2="12"/>
                <line x1="8" y1="16" x2="12" y2="16"/>
              </svg>
            </div>
            <transition name="nav-text">
              <span v-show="!collapsed" class="nav-text">任务</span>
            </transition>
          </router-link>
        </li>

        <!-- 活动记录 -->
        <li class="nav-item">
          <router-link
            to="/activities"
            class="nav-link"
            :class="{ 'nav-link-active': isActiveRoute('/activities') }"
            :title="collapsed ? '活动' : ''"
          >
            <div class="nav-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"/>
                <polyline points="12 6 12 12 16 14"/>
              </svg>
            </div>
            <transition name="nav-text">
              <span v-show="!collapsed" class="nav-text">活动</span>
            </transition>
          </router-link>
        </li>

        <!-- 设置 -->
        <li class="nav-item">
          <router-link
            to="/settings"
            class="nav-link"
            :class="{ 'nav-link-active': isActiveRoute('/settings') }"
            :title="collapsed ? '设置' : ''"
          >
            <div class="nav-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="3"/>
                <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/>
              </svg>
            </div>
            <transition name="nav-text">
              <span v-show="!collapsed" class="nav-text">设置</span>
            </transition>
          </router-link>
        </li>
      </ul>
    </nav>
  </aside>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const props = defineProps({
  collapsed: {
    type: Boolean,
    default: false
  },
  mobileOpen: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:collapsed', 'update:mobile-open'])

const route = useRoute()

const isActiveRoute = (path) => {
  return route.path === path || route.path.startsWith(path + '/')
}

const handleCollapse = () => {
  emit('update:collapsed', !props.collapsed)
}
</script>

<style scoped>
.app-sidebar {
  position: fixed;
  left: 0;
  top: 0;
  height: 100vh;
  background-color: var(--color-bg-card);
  border-right: 1px solid var(--color-border-light);
  display: flex;
  flex-direction: column;
  transition: width var(--sidebar-transition);
  z-index: var(--z-sticky);
  will-change: width;
}

/* 移动端处理 */
@media (max-width: 1023px) {
  .app-sidebar {
    width: var(--sidebar-width-expanded) !important;
    transform: translateX(-100%);
    transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .app-sidebar.sidebar-mobile-open {
    transform: translateX(0);
  }
}

/* ============================================
   HEADER - 顶部区域（Logo + 折叠按钮）
   ============================================ */

.sidebar-header {
  height: var(--sidebar-header-height);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--space-3);
  border-bottom: 1px solid var(--color-border-light);
  flex-shrink: 0;
}

/* Brand Logo */
.sidebar-brand {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  text-decoration: none;
  color: var(--color-text-primary);
  transition: opacity 0.2s ease;
  overflow: hidden;
}

.sidebar-brand:hover {
  opacity: 0.8;
}

.brand-icon {
  width: 28px;
  height: 28px;
  flex-shrink: 0;
  border-radius: 6px;
}

.brand-text {
  font-size: var(--text-lg);
  font-weight: var(--font-bold);
  color: var(--color-text-primary);
  white-space: nowrap;
  letter-spacing: -0.5px;
}

/* 过渡动画 */
.brand-text-enter-active,
.brand-text-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.brand-text-enter-from,
.brand-text-leave-to {
  opacity: 0;
  transform: translateX(-10px);
}

/* ============================================
   NAVIGATION - 导航菜单
   ============================================ */

.sidebar-nav {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: var(--space-3);
}

.sidebar-nav::-webkit-scrollbar {
  width: 4px;
}

.sidebar-nav::-webkit-scrollbar-track {
  background: transparent;
}

.sidebar-nav::-webkit-scrollbar-thumb {
  background: var(--color-gray-300);
  border-radius: 2px;
}

.nav-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.nav-item {
  display: flex;
}

.nav-link {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-2_5) var(--space-3);
  border-radius: var(--radius-md);
  text-decoration: none;
  color: var(--color-text-secondary);
  transition: all 0.2s ease;
  position: relative;
  width: 100%;
}

.nav-link:hover {
  background-color: var(--color-gray-100);
  color: var(--color-text-primary);
}

.nav-link-active {
  background-color: var(--color-primary-50);
  color: var(--color-primary);
  font-weight: var(--font-semibold);
}

.nav-link-active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 20px;
  background-color: var(--color-primary);
  border-radius: 0 2px 2px 0;
}

.nav-icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.nav-icon svg {
  width: 100%;
  height: 100%;
}

.nav-text {
  font-size: var(--text-sm);
  white-space: nowrap;
  flex: 1;
}

.nav-text-enter-active,
.nav-text-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.nav-text-enter-from,
.nav-text-leave-to {
  opacity: 0;
  transform: translateX(-10px);
}

/* ============================================
   COLLAPSED STATE - 折叠状态
   ============================================ */

.sidebar-collapsed .sidebar-header {
  padding: 0 var(--space-2);
  justify-content: center;
}

.sidebar-collapsed .sidebar-brand {
  justify-content: center;
}

.sidebar-collapsed .nav-link {
  justify-content: center;
  padding-left: 0;
  padding-right: 0;
}

.sidebar-collapsed .nav-link-active::before {
  display: none;
}

.sidebar-collapsed .nav-link-active {
  background-color: var(--color-primary);
  color: white;
}

.sidebar-collapsed .nav-link:hover {
  background-color: var(--color-gray-100);
  color: var(--color-text-primary);
}

.sidebar-collapsed .nav-link-active:hover {
  background-color: var(--color-primary);
  color: white;
  opacity: 0.9;
}

</style>
