<template>
  <div class="dashboard-layout">
    <!-- 左侧侧边栏 -->
    <AppSidebar
      v-model:collapsed="sidebarCollapsed"
      v-model:mobile-open="sidebarMobileOpen"
      class="dashboard-sidebar"
    />

    <!-- 主内容区域 -->
    <div
      class="dashboard-main"
      :style="{
        marginLeft: isMobile ? '0' : sidebarCollapsed ? 'var(--sidebar-width-collapsed)' : 'var(--sidebar-width-expanded)'
      }"
    >
      <!-- 顶部标题栏 -->
      <AppHeader
        :collapsed="sidebarCollapsed"
        @toggle-sidebar="toggleSidebar"
        @toggle-mobile-sidebar="toggleMobileSidebar"
      />

      <!-- 页面内容区 -->
      <main class="dashboard-content">
        <router-view v-slot="{ Component, route }">
          <transition
            name="page-fade"
            mode="out-in"
            @before-enter="onBeforeEnter"
            @enter="onEnter"
            @leave="onLeave"
          >
            <component
              :is="Component"
              :key="route.path + '-' + refreshKey"
              :refresh-key="refreshKey"
              @add-repo="showAddRepoModal = true"
            />
          </transition>
        </router-view>
      </main>
    </div>

    <!-- 移动端遮罩层 -->
    <div
      v-if="isMobile && sidebarMobileOpen"
      class="mobile-overlay"
      @click="closeMobileSidebar"
    ></div>

    <!-- 悬浮添加按钮 -->
    <FloatingAddButton v-if="showFloatingButton" @click="showAddRepoModal = true" />

    <!-- 添加仓库弹窗 -->
    <AddRepoModal
      v-model:show="showAddRepoModal"
      @success="handleAddRepoSuccess"
    />
  </div>
</template>

<script setup>
/**
 * DashboardLayout - 后台管理系统主布局
 *
 * Grid Design System - Flow Layout
 * Chromatic Team - Pixel Implementation
 *
 * 功能特性:
 * - 左侧固定侧边栏 (240px expanded, 64px collapsed)
 * - 顶部固定标题栏 (64px height)
 * - 主内容区域自适应
 * - 响应式断点处理
 * - 移动端抽屉模式
 * - 悬浮添加按钮
 */
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useMessage } from 'naive-ui'
import { useReposStore } from '@/stores/repos'
import AppSidebar from './AppSidebar.vue'
import AppHeader from './AppHeader.vue'
import FloatingAddButton from '@/components/FloatingAddButton.vue'
import AddRepoModal from '@/components/AddRepoModal.vue'

const message = useMessage()
const reposStore = useReposStore()
const route = useRoute()

// 设置页、任务页等不显示悬浮按钮
const showFloatingButton = computed(() => {
  return !route.path.startsWith('/settings')
})

// 侧边栏状态
const sidebarCollapsed = ref(false) // 桌面端折叠状态
const sidebarMobileOpen = ref(false) // 移动端打开状态

// 添加仓库弹窗
const showAddRepoModal = ref(false)

// 刷新标记，用于通知子组件刷新
const refreshKey = ref(0)

// 响应式断点检测
const isMobile = ref(false)
const checkMobile = () => {
  isMobile.value = window.innerWidth < 1024 // lg breakpoint
}

// 侧边栏切换
const toggleSidebar = () => {
  sidebarCollapsed.value = !sidebarCollapsed.value
}

const toggleMobileSidebar = () => {
  sidebarMobileOpen.value = !sidebarMobileOpen.value
}

const closeMobileSidebar = () => {
  sidebarMobileOpen.value = false
}

// 添加仓库成功回调
const handleAddRepoSuccess = async () => {
  // 刷新仓库列表
  await reposStore.fetchRepos()
  // 触发子组件刷新（包括活动记录）
  refreshKey.value++
}

// 页面过渡动画
const onBeforeEnter = (el) => {
  el.style.opacity = '0'
  el.style.transform = 'translateY(10px)'
}

const onEnter = (el, done) => {
  el.style.transition = 'opacity 0.3s ease-out, transform 0.3s ease-out'
  // 强制重绘
  el.offsetHeight
  el.style.opacity = '1'
  el.style.transform = 'translateY(0)'
  setTimeout(done, 300)
}

const onLeave = (el, done) => {
  el.style.transition = 'opacity 0.2s ease-in, transform 0.2s ease-in'
  el.style.opacity = '0'
  el.style.transform = 'translateY(-10px)'
  setTimeout(done, 200)
}

// 生命周期
onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})
</script>

<style scoped>
/* ============================================
   LAYOUT STRUCTURE - 布局结构
   ============================================ */

.dashboard-layout {
  display: flex;
  min-height: 100vh;
  background-color: var(--color-bg-page);
  position: relative;
}

/* 侧边栏定位 */
.dashboard-sidebar {
  position: fixed;
  left: 0;
  top: 0;
  height: 100vh;
  z-index: var(--z-sticky);
  transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

/* 桌面端: 侧边栏始终可见 */
@media (min-width: 1024px) {
  .dashboard-sidebar {
    transform: translateX(0);
  }
}

/* 移动端: 侧边栏默认隐藏 */
@media (max-width: 1023px) {
  .dashboard-sidebar {
    transform: translateX(-100%);
  }

  .dashboard-sidebar.mobile-open {
    transform: translateX(0);
  }
}

/* 主内容区域 */
.dashboard-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  transition: margin-left 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  will-change: margin-left;
}

/* 移动端主内容区域 */
@media (max-width: 1023px) {
  .dashboard-main {
    margin-left: 0 !important;
  }
}

/* ============================================
   HEADER - 顶部标题栏
   ============================================ */

/* 顶部标题栏由 AppHeader 组件处理 */
/* 预留 64px 高度空间 */

/* ============================================
   CONTENT - 内容区域
   ============================================ */

.dashboard-content {
  flex: 1;
  padding: var(--space-6);
  margin-top: var(--navbar-height);
  min-height: calc(100vh - var(--navbar-height));
}

/* 内容区子元素不限制最大宽度，充分利用屏幕 */

/* 响应式内边距 */
@media (max-width: 768px) {
  .dashboard-content {
    padding: var(--space-4);
  }
}

@media (max-width: 480px) {
  .dashboard-content {
    padding: var(--space-3);
  }
}

/* ============================================
   MOBILE OVERLAY - 移动端遮罩层
   ============================================ */

.mobile-overlay {
  position: fixed;
  inset: 0;
  background-color: rgba(0, 0, 0, 0.5);
  z-index: calc(var(--z-sticky) - 1);
  opacity: 0;
  animation: fadeIn 0.3s ease-out forwards;
}

@keyframes fadeIn {
  to {
    opacity: 1;
  }
}

/* ============================================
   PAGE TRANSITIONS - 页面过渡动画
   ============================================ */

.page-fade-enter-active,
.page-fade-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.page-fade-enter-from {
  opacity: 0;
  transform: translateY(10px);
}

.page-fade-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
</style>
