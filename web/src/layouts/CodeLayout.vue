<template>
  <div class="code-layout">
    <div class="code-header">
      <div class="header-left">
        <div class="logo">
          <span class="logo-text">Forks</span>
        </div>
        <div class="breadcrumb">
          <n-breadcrumb>
            <n-breadcrumb-item @click="$router.push('/')">
              <n-icon><HomeOutline /></n-icon>
              首页
            </n-breadcrumb-item>
            <n-breadcrumb-item>
              <n-icon><CodeOutline /></n-icon>
              代码查看
            </n-breadcrumb-item>
          </n-breadcrumb>
        </div>
      </div>
      <div class="header-center">
        <div v-if="repoInfo" class="repo-info">
          <n-icon class="repo-icon"><GitBranchOutline /></n-icon>
          <span class="repo-name">{{ repoInfo.author }}/{{ repoInfo.repo }}</span>
          <n-tag v-if="repoInfo.license" size="small" type="info">{{ repoInfo.license }}</n-tag>
        </div>
      </div>
      <div class="header-right">
        <n-space>
          <n-button size="small" @click="showRepoDetail" type="default">
            <template #icon>
              <n-icon><InformationCircleOutline /></n-icon>
            </template>
            详细信息
          </n-button>
          <n-button size="small" @click="openRepo()" type="primary" :disabled="!repoInfo">
            <template #icon>
              <n-icon><OpenOutline /></n-icon>
            </template>
            查看仓库
          </n-button>
          <n-button size="small" @click="$router.back()">
            <template #icon>
              <n-icon><ArrowBack /></n-icon>
            </template>
            返回
          </n-button>
        </n-space>
      </div>
    </div>
    
    <div class="code-content">
      <router-view @repo-info-loaded="handleRepoInfoLoaded" />
    </div>

    <!-- 仓库详细信息抽屉 -->
    <RepoDetailDrawer
      v-model:show="showDetailDrawer"
      :repo="repoInfo"
      :update-loading="updateLoading"
      :show-code-button="false"
      @open-repo="openRepo"
      @update-info="updateRepoInfo"
      @delete-repo="deleteRepository"
      @toggle-valid="handleToggleValid"
    />
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { 
  NSpace, NButton, NIcon, NBreadcrumb, NBreadcrumbItem, NTag, useMessage
} from 'naive-ui'
import { 
  HomeOutline, CodeOutline, GitBranchOutline, InformationCircleOutline,
  OpenOutline, ArrowBack
} from '@vicons/ionicons5'
import RepoDetailDrawer from '@/components/RepoDetailDrawer.vue'
import request from '@/utils/request'

const router = useRouter()
const message = useMessage()

// 仓库信息
const repoInfo = ref(null)
const showDetailDrawer = ref(false)
const updateLoading = ref(false)

// 处理仓库信息加载
const handleRepoInfoLoaded = (info) => {
  repoInfo.value = info
}

// 显示仓库详细信息
const showRepoDetail = () => {
  showDetailDrawer.value = true
}

// 打开仓库
const openRepo = (url) => {
  const repoUrl = url || repoInfo.value?.url
  if (repoUrl) {
    window.open(repoUrl, '_blank')
  }
}

// 更新仓库信息
const updateRepoInfo = async (repo) => {
  const id = repo?.id ?? repo
  try {
    updateLoading.value = true
    
    const response = await request.put(`/api/repos/${id}/update`)
    
    if (response.data.code === 0) {
      // 更新本地仓库信息
      if (repoInfo.value && repoInfo.value.id === id) {
        const updatedRepo = response.data.data
        repoInfo.value = { ...repoInfo.value, ...updatedRepo }
      }
      
      message.success('仓库信息更新成功')
    } else {
      throw new Error(response.data.message || '更新失败')
    }
  } catch (error) {
    console.error('更新仓库信息失败:', error)
    
    let errorMessage = '更新仓库信息失败'
    if (error.response) {
      const status = error.response.status
      const responseMessage = error.response.data?.message || ''
      
      if (status === 404) {
        errorMessage = '仓库不存在或无法访问'
      } else if (status === 500) {
        errorMessage = responseMessage || '服务器处理失败，请稍后重试'
      } else {
        errorMessage = responseMessage || `请求失败 (${status})`
      }
    } else if (error.request) {
      errorMessage = '网络连接失败，请检查网络连接'
    } else {
      errorMessage = error.message || '未知错误'
    }
    
    message.error(errorMessage)
  } finally {
    updateLoading.value = false
  }
}

// 切换仓库有效状态
const handleToggleValid = async (repo) => {
  try {
    const res = await request.post(`/api/repos/${repo.id}/toggle-valid`)
    const apiData = res.data
    if (apiData && apiData.code === 0) {
      message.success(apiData.message)
      if (repoInfo.value && repoInfo.value.id === repo.id) {
        repoInfo.value = { ...repoInfo.value, valid: apiData.data.valid }
      }
    } else {
      throw new Error(apiData?.message || '操作失败')
    }
  } catch (error) {
    message.error('操作失败：' + error.message)
  }
}

// 删除仓库
const deleteRepository = async (repo) => {
  const id = repo?.id ?? repo
  try {
    await request.delete(`/api/repos/${id}`)
    message.success('仓库删除成功')
    showDetailDrawer.value = false
    
    // 删除成功后返回首页
    router.push('/')
  } catch (error) {
    message.error('删除仓库失败: ' + (error.message || '未知错误'))
  }
}
</script>

<style scoped>
.code-layout {
  height: 100vh;
  display: flex;
  flex-direction: column;
}

.code-header {
  height: 64px;
  background: #fff;
  border-bottom: 1px solid #e0e0e0;
  display: flex;
  align-items: center;
  padding: 0 24px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.03);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 24px;
  flex: 0 0 auto;
}

.logo {
  display: flex;
  align-items: center;
}

.logo-text {
  font-size: 20px;
  font-weight: bold;
  color: #18a058;
}

.breadcrumb {
  margin-left: 8px;
}

.header-center {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: center;
}

.repo-info {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 16px;
  background: #f8f9fa;
  border-radius: 6px;
  border: 1px solid #e0e0e0;
}

.repo-icon {
  color: #666;
}

.repo-name {
  font-size: 16px;
  font-weight: 600;
  color: #333;
}

.header-right {
  flex: 0 0 auto;
}

.code-content {
  flex: 1;
  overflow: hidden;
}


/* 面包屑样式 */
:deep(.n-breadcrumb-item) {
  cursor: pointer;
}

:deep(.n-breadcrumb-item:hover) {
  color: #18a058;
}
</style>