import { createRouter, createWebHistory } from 'vue-router'
import DefaultLayout from '@/layouts/DefaultLayout.vue'
import FullScreenLayout from '@/layouts/FullScreenLayout.vue'
import CodeLayout from '@/layouts/CodeLayout.vue'
import DashboardLayout from '@/layouts/DashboardLayout.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    // 新版后台管理系统路由
    {
      path: '/',
      component: DashboardLayout,
      children: [
        {
          path: '',
          name: 'dashboard',
          component: () => import('@/views/dashboard/DashboardView.vue'),
          meta: { title: '首页', breadcrumb: [{ label: '首页' }] }
        },
        {
          path: 'authors',
          name: 'authors',
          component: () => import('@/views/authors/AuthorsView.vue'),
          meta: { title: '作者管理', breadcrumb: [{ label: '作者管理' }] }
        },
        {
          path: 'repos',
          name: 'repos',
          component: () => import('@/views/repos/ReposView.vue'),
          meta: { title: '仓库管理', breadcrumb: [{ label: '仓库管理' }] }
        },
        {
          path: 'settings',
          component: () => import('@/views/settings/SettingsView.vue'),
          meta: { title: '系统设置' },
          children: [
            { path: '', redirect: { name: 'settings-proxy' } },
            { path: 'proxy', name: 'settings-proxy', component: () => import('@/views/settings/SettingsProxy.vue'), meta: { title: '代理设置' } },
            { path: 'token', name: 'settings-token', component: () => import('@/views/settings/SettingsToken.vue'), meta: { title: '令牌管理' } },
            { path: 'trending', name: 'settings-trending', component: () => import('@/views/settings/SettingsTrending.vue'), meta: { title: '趋势同步' } },
            { path: 'mcp', name: 'settings-mcp', component: () => import('@/views/settings/SettingsMCP.vue'), meta: { title: 'MCP' } },
            { path: 'about', name: 'settings-about', component: () => import('@/views/settings/SettingsAbout.vue'), meta: { title: '关于' } },
          ]
        },
        {
          path: 'tasks',
          name: 'tasks',
          component: () => import('@/views/tasks/TasksView.vue'),
          meta: { title: '任务管理', breadcrumb: [{ label: '任务管理' }] }
        },
        {
          path: 'activities',
          name: 'activities',
          component: () => import('@/views/activities/ActivitiesView.vue'),
          meta: { title: '活动记录', breadcrumb: [{ label: '活动记录' }] }
        },
        {
          path: 'trending',
          name: 'trending',
          component: () => import('@/views/trending/TrendingView.vue'),
          meta: { title: 'GitHub Trending', breadcrumb: [{ label: 'GitHub Trending' }] }
        },
        {
          path: ':pathMatch(.*)*',
          name: 'DashboardNotFound',
          component: () => import('@/views/NotFoundView.vue'),
          meta: { title: '页面未找到' }
        }
      ]
    },
    // 代码查看页面（保持原有布局）
    {
      path: '/code',
      component: CodeLayout,
      children: [
        {
          path: ':id',
          name: 'code',
          component: () => import('@/views/CodeView.vue'),
          meta: { title: '代码查看' }
        }
      ]
    },
    // 登录页面
    {
      path: '/login',
      component: FullScreenLayout,
      children: [
        {
          path: '',
          name: 'login',
          component: () => import('@/views/login/index.vue'),
          meta: { title: '登录', requiresAuth: false }
        }
      ]
    },
    // 旧版路由重定向（向后兼容）
    {
      path: '/home',
      redirect: '/'
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'NotFound',
      component: () => import('@/views/NotFoundView.vue'),
      meta: { title: '页面未找到' }
    }
  ],
})

// 路由守卫
router.beforeEach((to, from, next) => {
  // 设置页面标题
  if (to.meta.title) {
    document.title = `${to.meta.title} - forks`
  }
  
  // 检查URL中的token参数
  const urlParams = new URLSearchParams(window.location.search)
  const urlToken = urlParams.get('token')
  
  if (urlToken) {
    // 自动设置token并移除URL参数
    localStorage.setItem('token', urlToken)
    localStorage.setItem('userInfo', JSON.stringify({
      id: 1,
      username: 'user',
      role: 'user'
    }))
    
    // 清除URL中的token参数
    const newUrl = new URL(window.location)
    newUrl.searchParams.delete('token')
    window.history.replaceState({}, '', newUrl.pathname + newUrl.search + newUrl.hash)
    
    // 触发登录事件
    window.dispatchEvent(new CustomEvent('userLoggedIn'))
  }
  
  // 获取认证信息
  const token = localStorage.getItem('token')
  const userInfoStr = localStorage.getItem('userInfo')
  let userInfo = null
  
  if (userInfoStr) {
    try {
      userInfo = JSON.parse(userInfoStr)
    } catch (e) {
      console.error('解析用户信息失败:', e)
    }
  }
  
  // 检查是否需要登录
  if (to.meta.requiresAuth !== false && !token) {
    next({ name: 'login' })
    return
  }
  
  
  // 如果已登录访问登录页，重定向到首页
  if (to.name === 'login' && token) {
    next({ name: 'dashboard' })
    return
  }
  
  next()
})

export default router
