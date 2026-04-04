<template>
  <div class="home-view">
    <n-space vertical size="large">
      <!-- 欢迎卡片 -->
      <n-card class="welcome-card">
        <div class="welcome-header">
          <div class="welcome-icon">
            <n-icon size="48" color="#66BB6A">
              <GitBranchOutline />
            </n-icon>
          </div>
          <div class="welcome-content">
            <h1 class="welcome-title">
              欢迎使用 <span class="brand-name">Git<span class="brand-accent">Plus</span></span>
            </h1>
            <p class="welcome-description">
              🚀 增强您的 Git 工作流程，提供智能仓库管理、快速代理克隆和便捷的开发工具
            </p>
            <div class="feature-tags">
              <n-tag type="info" size="small" round>
                <template #icon>
                  <n-icon><CodeOutline /></n-icon>
                </template>
                仓库管理
              </n-tag>
              <n-tag type="success" size="small" round>
                <template #icon>
                  <n-icon><StarOutline /></n-icon>
                </template>
                代理克隆
              </n-tag>
              <n-tag type="warning" size="small" round>
                <template #icon>
                  <n-icon><OpenOutline /></n-icon>
                </template>
                开发工具
              </n-tag>
            </div>
          </div>
        </div>

        <n-divider style="margin: 24px 0;" />

        <div class="action-section">
          <n-space size="large" justify="center">
            <n-button type="primary" size="large" @click="showAddRepoModal = true" class="action-button">
              <template #icon>
                <n-icon><Terminal /></n-icon>
              </template>
              添加仓库
            </n-button>
            <n-button size="large" @click="handleScanRepos" :loading="scanLoading" class="action-button">
              <template #icon>
                <n-icon><ScanOutline /></n-icon>
              </template>
              扫描仓库
            </n-button>
            <n-button size="large" @click="refreshRepos" class="action-button">
              <template #icon>
                <n-icon><GitBranchOutline /></n-icon>
              </template>
              刷新列表
            </n-button>
            <n-button size="large" @click="handleBatchClone" :loading="batchCloneLoading" class="action-button">
              <template #icon>
                <n-icon><DownloadOutline /></n-icon>
              </template>
              一键克隆
            </n-button>
          </n-space>
        </div>
      </n-card>

      <!-- 仓库统计栏 -->
      <n-card>
        <n-space justify="space-between" align="center">
          <n-text strong style="font-size: 16px;">仓库管理</n-text>
          <n-space>
            <n-text depth="3">共 {{ reposStore.total }} 个仓库</n-text>
            <n-divider vertical />
            <n-text type="success" depth="3">{{ reposStore.statistics.cloned }} 已克隆</n-text>
          </n-space>
        </n-space>
      </n-card>

      <!-- 筛选区域 -->
      <n-card>
        <n-space size="medium" wrap>
          <!-- 搜索框 -->
          <n-input
            v-model:value="searchQuery"
            placeholder="搜索仓库..."
            clearable
            style="width: 280px"
            @update:value="handleSearch"
          >
            <template #prefix>
              <n-icon><Search /></n-icon>
            </template>
          </n-input>

          <!-- 状态筛选 -->
          <n-select
            v-model:value="selectedStatus"
            :options="statusOptions"
            placeholder="仓库状态"
            clearable
            style="width: 140px"
            @update:value="handleStatusFilter"
          />

          <!-- 作者筛选 -->
          <n-select
            v-model:value="selectedAuthor"
            :options="authorOptions"
            placeholder="作者"
            clearable
            filterable
            style="width: 160px"
            @update:value="handleAuthorFilter"
          />

          <!-- 清除筛选 -->
          <n-button v-if="hasActiveFilters" quaternary @click="clearFilters">
            <template #icon>
              <n-icon><Close /></n-icon>
            </template>
            清除筛选
          </n-button>
        </n-space>
      </n-card>

      <!-- 仓库展示区域 - 卡片视图 -->
      <n-card>
        <n-spin :show="reposStore.loading">
          <CardView
            :repos="paginatedRepos"
            :repo-statuses="reposStore.repoStatuses"
            :loading="reposStore.loading"
            :total="reposStore.total"
            :current-page="currentPage"
            :page-size="pageSize"
            @repo-click="handleRepoClick"
            @view-code="handleViewCode"
            @open-folder="handleOpenFolder"
            @clone="handleClone"
            @pull="handlePull"
            @reset="handleReset"
            @delete="handleDelete"
            @update-info="handleUpdateInfo"
            @batch-clone="handleCardBatchClone"
            @batch-delete="handleBatchDelete"
            @page-change="handlePageChange"
            @page-size-change="handlePageSizeChange"
          />

          <!-- 空状态 -->
          <n-empty
            v-if="!reposStore.loading && paginatedRepos.length === 0"
            :description="emptyDescription"
            size="large"
          >
            <template #extra>
              <n-space>
                <n-button v-if="hasActiveFilters" @click="clearFilters">
                  清除筛选
                </n-button>
                <n-button type="primary" @click="showAddRepoModal = true">
                  添加第一个仓库
                </n-button>
              </n-space>
            </template>
          </n-empty>
        </n-spin>
      </n-card>
    </n-space>

    <!-- 添加仓库对话框 -->
    <AddRepoModal
      :show="showAddRepoModal"
      @update:show="showAddRepoModal = $event"
      @success="handleAddRepoSuccess"
    />

    <!-- 批量克隆选择弹窗 -->
    <BatchCloneModal
      :show="showBatchCloneModal"
      @update:show="showBatchCloneModal = $event"
      @confirm="handleBatchCloneConfirm"
    />

    <!-- 扫描结果对话框 -->
    <n-modal v-model:show="showScanResult" :style="{ width: '700px' }">
      <n-card title="扫描结果" :bordered="false" size="huge">
        <template v-if="scanResult">
          <!-- 摘要信息 -->
          <n-card :bordered="false" style="margin-bottom: 16px;">
            <n-statistic-group>
              <n-statistic label="本地仓库" :value="scanResult.summary.scanned_count" />
              <n-statistic label="新发现" :value="scanResult.summary.new_count" />
              <n-statistic label="未克隆" :value="scanResult.summary.missing_count" />
            </n-statistic-group>
          </n-card>

          <!-- 新发现的仓库 -->
          <div v-if="scanResult.new_repos && scanResult.new_repos.length > 0" style="margin-bottom: 16px;">
            <n-text strong style="font-size: 16px;">新发现的仓库 ({{ scanResult.new_repos.length }})</n-text>
            <n-list bordered style="margin-top: 12px; max-height: 200px; overflow-y: auto;">
              <n-list-item v-for="repo in scanResult.new_repos" :key="repo.author + '/' + repo.repo">
                <n-thing>
                  <template #header>
                    <n-space align="center">
                      <n-text strong>{{ repo.author }}/{{ repo.repo }}</n-text>
                      <n-tag size="small" type="success">{{ repo.source }}</n-tag>
                    </n-space>
                  </template>
                </n-thing>
              </n-list-item>
            </n-list>
          </div>

          <!-- 未克隆的仓库 -->
          <div v-if="scanResult.missing_repos && scanResult.missing_repos.length > 0" style="margin-bottom: 16px;">
            <n-text strong style="font-size: 16px;">未克隆的仓库 ({{ scanResult.missing_repos.length }})</n-text>
            <n-list bordered style="margin-top: 12px; max-height: 200px; overflow-y: auto;">
              <n-list-item v-for="repo in scanResult.missing_repos" :key="repo.id">
                <n-thing>
                  <template #header>
                    <n-space align="center">
                      <n-text strong>{{ repo.author }}/{{ repo.repo }}</n-text>
                      <n-tag size="small" type="info">{{ repo.source }}</n-tag>
                    </n-space>
                  </template>
                </n-thing>
              </n-list-item>
            </n-list>
          </div>

          <!-- 无变化 -->
          <n-empty v-if="(!scanResult.new_repos || scanResult.new_repos.length === 0) && (!scanResult.missing_repos || scanResult.missing_repos.length === 0)" description="扫描完成，数据已是最新" />
        </template>

        <template #footer>
          <n-space justify="end">
            <n-button @click="showScanResult = false">关闭</n-button>
            <n-button
              v-if="scanResult && scanResult.new_repos && scanResult.new_repos.length > 0"
              type="primary"
              @click="handleAddScannedRepos"
            >
              添加新仓库
            </n-button>
            <n-button
              v-if="scanResult && scanResult.missing_repos && scanResult.missing_repos.length > 0"
              type="warning"
              @click="handleCleanMissingRepos"
            >
              清理记录
            </n-button>
            <n-button
              v-if="scanResult && scanResult.summary && scanResult.summary.missing_count > 0"
              type="primary"
              @click="handleBatchCloneFromScan"
              :loading="batchCloneLoading"
            >
              一键克隆
            </n-button>
          </n-space>
        </template>
      </n-card>
    </n-modal>

    <!-- 批量克隆进度弹窗 -->
    <n-modal v-model:show="showBatchCloneProgress" :style="{ width: '500px' }" :closable="false" :maskClosable="false">
      <n-card title="一键克隆进度" :bordered="false">
        <div v-if="batchCloneProgress">
          <n-progress
            type="line"
            :percentage="batchCloneProgress.total > 0 ? Math.round((batchCloneProgress.current / batchCloneProgress.total) * 100) : 0"
            :status="batchCloneProgress.completed ? 'success' : 'default'"
          />

          <n-space vertical style="margin-top: 16px">
            <n-text>
              {{ batchCloneProgress.current || 0 }} / {{ batchCloneProgress.total || 0 }}
              <template v-if="batchCloneProgress.repo">
                - {{ batchCloneProgress.repo }}
              </template>
            </n-text>
            <n-text depth="3">{{ batchCloneProgress.status }}</n-text>

            <template v-if="batchCloneProgress.completed">
              <n-divider />
              <n-space>
                <n-tag type="success">成功: {{ batchCloneProgress.cloned }}</n-tag>
                <n-tag type="warning">失败: {{ batchCloneProgress.failed }}</n-tag>
                <n-tag type="error">失效: {{ batchCloneProgress.invalid }}</n-tag>
              </n-space>
            </template>
          </n-space>
        </div>

        <template #footer>
          <n-space justify="end">
            <n-button
              v-if="batchCloneProgress?.completed"
              type="primary"
              @click="showBatchCloneProgress = false"
            >
              完成
            </n-button>
          </n-space>
        </template>
      </n-card>
    </n-modal>

    <!-- 仓库详情抽屉 -->
    <RepoDetailDrawer
      :show="showDetailDrawer"
      :repo="selectedRepo"
      :update-loading="updateLoading"
      @update:show="showDetailDrawer = $event"
      @open-repo="handleOpenRepo"
      @view-code="handleViewCode"
      @update-info="handleUpdateInfo"
      @delete-repo="handleDelete"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import {
  GitBranchOutline,
  CodeOutline,
  StarOutline,
  OpenOutline,
  Terminal,
  Search,
  Close,
  ScanOutline,
  DownloadOutline
} from '@vicons/ionicons5'

// 组件导入
import CardView from '@/components/view/CardView.vue'
import AddRepoModal from '@/components/AddRepoModal.vue'
import RepoDetailDrawer from '@/components/RepoDetailDrawer.vue'
import BatchCloneModal from '@/components/BatchCloneModal.vue'

// Store导入
import { useReposStore } from '@/stores/repos'

const router = useRouter()
const message = useMessage()

// Store
const reposStore = useReposStore()

// 响应式数据
const showAddRepoModal = ref(false)
const showBatchCloneModal = ref(false)
const showDetailDrawer = ref(false)
const selectedRepo = ref(null)
const updateLoading = ref(false)
const scanLoading = ref(false)
const showScanResult = ref(false)
const scanResult = ref(null)
const batchCloneLoading = ref(false)
const showBatchCloneProgress = ref(false)
const batchCloneProgress = ref(null)

// 搜索和筛选
const searchQuery = ref('')
const selectedStatus = ref(null)
const selectedAuthor = ref(null)

// 分页
const currentPage = ref(1)
const pageSize = ref(12)

// 计算属性
const paginatedRepos = computed(() => reposStore.filteredRepos)

const hasActiveFilters = computed(() => {
  return searchQuery.value || selectedStatus.value || selectedAuthor.value
})

const emptyDescription = computed(() => {
  return hasActiveFilters.value ? '没有找到符合条件的仓库' : '还没有添加任何仓库'
})

// 筛选选项
const authorOptions = computed(() => {
  return reposStore.uniqueAuthors.map(author => ({
    label: author,
    value: author
  }))
})

const statusOptions = [
  { label: '全部', value: 'all' },
  { label: '已克隆', value: 'cloned' },
  { label: '未克隆', value: 'not-cloned' }
]

// 方法
const refreshRepos = async () => {
  try {
    await reposStore.fetchRepos()
    message.success('仓库列表已刷新')
  } catch (error) {
    message.error('刷新失败：' + error.message)
  }
}

// 一键克隆 - 打开选择弹窗
const handleBatchClone = () => {
  showBatchCloneModal.value = true
}

// 一键克隆 - 用户确认选择后执行
const handleBatchCloneConfirm = async (ids) => {
  try {
    batchCloneLoading.value = true
    const { batchCloneRepos, batchCloneSSE } = await import('@/api/repos')
    const response = await batchCloneRepos({ ids })
    const apiData = response.data

    if (apiData && apiData.code === 0 && apiData.data?.useSSE) {
      const { tempToken, total } = apiData.data

      if (total === 0) {
        message.info('没有需要克隆的仓库')
        batchCloneLoading.value = false
        return
      }

      showBatchCloneProgress.value = true
      batchCloneProgress.value = { total, current: 0, status: '准备中...' }

      batchCloneSSE(
        tempToken,
        (progress) => {
          if (progress.type === 'start') {
            batchCloneProgress.value = { ...batchCloneProgress.value, status: '开始克隆...', total: progress.total }
          } else if (progress.type === 'progress') {
            batchCloneProgress.value = {
              ...batchCloneProgress.value,
              current: progress.current,
              repo: progress.repo,
              status: progress.message
            }
          }
        },
        (result) => {
          batchCloneLoading.value = false
          batchCloneProgress.value = {
            ...batchCloneProgress.value,
            status: '完成',
            cloned: result.cloned,
            failed: result.failed,
            invalid: result.invalid,
            completed: true
          }
          message.success(`克隆完成: 成功 ${result.cloned}, 失败 ${result.failed}, 失效 ${result.invalid}`)
          refreshRepos()
        },
        (error) => {
          batchCloneLoading.value = false
          message.error('批量克隆失败：' + (error.message || '未知错误'))
        }
      )
    } else {
      batchCloneLoading.value = false
      message.error('启动批量克隆失败')
    }
  } catch (error) {
    batchCloneLoading.value = false
    message.error('批量克隆失败：' + error.message)
  }
}

// 从扫描结果一键克隆
const handleBatchCloneFromScan = () => {
  showScanResult.value = false // 关闭扫描结果弹窗
  showBatchCloneModal.value = true
}

// 卡片视图批量克隆（用户已在卡片中选择仓库，直接执行）
const handleCardBatchClone = (selectedRepos) => {
  const ids = selectedRepos.map((r) => r.id)
  handleBatchCloneConfirm(ids)
}

// 扫描本地Git仓库
const handleScanRepos = async () => {
  try {
    scanLoading.value = true
    const { scanRepos, scanReposSSE } = await import('@/api/repos')
    const response = await scanRepos()
    const apiData = response.data

    if (apiData && apiData.code === 0 && apiData.data && apiData.data.useSSE) {
      const { taskId, tempToken } = apiData.data

      scanReposSSE(
        taskId,
        tempToken,
        () => {}, // 进度更新
        (result) => {
          scanLoading.value = false
          scanResult.value = result
          showScanResult.value = true

          const new_repos = result.new_repos || []
          const missing_repos = result.missing_repos || []

          if (new_repos.length > 0) {
            message.success(`发现 ${new_repos.length} 个新仓库`)
          } else if (missing_repos.length > 0) {
            message.warning(`发现 ${missing_repos.length} 个未克隆仓库`)
          } else {
            message.success('扫描完成，数据已是最新')
          }
        },
        (error) => {
          scanLoading.value = false
          message.error('扫描失败：' + (error.message || '未知错误'))
        }
      )
    } else {
      scanLoading.value = false
      message.error('启动扫描失败')
    }
  } catch (error) {
    scanLoading.value = false
    message.error('扫描失败：' + error.message)
  }
}

const handleSearch = (value) => {
  reposStore.setSearchQuery(value)
  currentPage.value = 1
  // 触发后端请求
  reposStore.fetchRepos()
}

const handleStatusFilter = (value) => {
  selectedStatus.value = value
  reposStore.setActiveFilter('status', value)
  currentPage.value = 1
  // 触发后端请求
  reposStore.fetchRepos()
}

const handleAuthorFilter = (value) => {
  selectedAuthor.value = value
  reposStore.setActiveFilter('author', value)
  currentPage.value = 1
  // 触发后端请求
  reposStore.fetchRepos()
}

const clearFilters = () => {
  searchQuery.value = ''
  selectedStatus.value = null
  selectedAuthor.value = null
  reposStore.clearFilters()
  currentPage.value = 1
  // 触发后端请求
  reposStore.fetchRepos()
}

// 仓库操作
const handleRepoClick = (repo) => {
  selectedRepo.value = repo
  showDetailDrawer.value = true
}

const handleViewCode = (repo) => {
  if (!repo || !repo.id) {
    message.error('仓库信息无效')
    return
  }
  router.push(`/code/${repo.id}`)
}

const handleOpenFolder = (repo) => {
  console.log('Open folder:', repo)
}

const handleOpenRepo = (url) => {
  window.open(url, '_blank')
}

const handleClone = async (repo) => {
  if (!repo || !repo.id) {
    message.error('仓库信息无效')
    return
  }
  message.success(`开始克隆 ${repo.repo}...`)
}

const handlePull = async (repo) => {
  if (!repo || !repo.id) {
    message.error('仓库信息无效')
    return
  }
  message.success(`开始拉取 ${repo.repo}...`)
}

const handleReset = async (repo) => {
  if (!repo || !repo.id) {
    message.error('仓库信息无效')
    return
  }
  message.success(`开始重置 ${repo.repo}...`)
}

const handleDelete = async (repo) => {
  if (!repo || !repo.id) {
    message.error('仓库信息无效')
    return
  }
  try {
    await reposStore.removeRepo(repo.id)
    message.success('仓库删除成功')
  } catch (error) {
    message.error('删除失败：' + error.message)
  }
}

const handleUpdateInfo = async (repo) => {
  try {
    updateLoading.value = true
    const { updateRepoInfo } = await import('@/api/repos')
    const response = await updateRepoInfo(repo.id)
    const apiData = response.data
    if (apiData && apiData.code === 0) {
      message.success(`更新 ${repo.repo} 信息成功`)
      // 即时更新详情抽屉数据
      if (selectedRepo.value && selectedRepo.value.id === repo.id) {
        selectedRepo.value = { ...selectedRepo.value, ...apiData.data }
      }
      refreshRepos()
    } else {
      throw new Error(apiData?.message || '更新失败')
    }
  } catch (error) {
    message.error('更新失败：' + error.message)
  } finally {
    updateLoading.value = false
  }
}

const handleBatchDelete = async (repos) => {
  try {
    const repoIds = repos.map(repo => repo.id)
    await reposStore.batchDeleteRepos(repoIds)
    message.success('批量删除成功')
  } catch (error) {
    message.error('批量删除失败：' + error.message)
  }
}

const handleAddRepoSuccess = () => {
  showAddRepoModal.value = false
  refreshRepos()
}

const handleAddScannedRepos = async () => {
  try {
    const { addRepo } = await import('@/api/repos')
    let successCount = 0

    for (const repo of scanResult.value.new_repos) {
      try {
        const response = await addRepo({ url: repo.url, autoClone: false })
        if (response.code === 0) successCount++
      } catch (error) {
        console.error('添加失败:', repo.author + '/' + repo.repo)
      }
    }

    message.success(`成功添加 ${successCount} 个仓库`)
    showScanResult.value = false
    await refreshRepos()
  } catch (error) {
    message.error('添加仓库失败：' + error.message)
  }
}

const handleCleanMissingRepos = async () => {
  try {
    const { deleteRepo } = await import('@/api/repos')
    let deletedCount = 0

    for (const repo of scanResult.value.missing_repos) {
      try {
        await deleteRepo(repo.id)
        deletedCount++
      } catch (error) {
        console.error('删除失败:', repo.id)
      }
    }

    message.success(`已清理 ${deletedCount} 条记录`)
    showScanResult.value = false
    await refreshRepos()
  } catch (error) {
    message.error('清理失败：' + error.message)
  }
}

// 分页处理
const handlePageChange = async (page) => {
  currentPage.value = page
  reposStore.setCurrentPage(page)
  await reposStore.fetchRepos()
}

const handlePageSizeChange = async (size) => {
  pageSize.value = size
  currentPage.value = 1
  reposStore.setPageSize(size)
  await reposStore.fetchRepos()
}

// 生命周期
onMounted(async () => {
  await refreshRepos()
})

watch([searchQuery, selectedStatus, selectedAuthor], () => {
  currentPage.value = 1
})
</script>

<style scoped>
.home-view {
  padding: 24px;
  max-width: 1400px;
  margin: 0 auto;
}

.welcome-card {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
}

.welcome-header {
  display: flex;
  align-items: center;
  gap: 24px;
}

.welcome-content {
  flex: 1;
}

.welcome-title {
  font-size: 32px;
  font-weight: 700;
  margin: 0 0 12px 0;
  color: white;
}

.brand-name {
  color: #ffd700;
}

.brand-accent {
  color: #ff6b6b;
}

.welcome-description {
  font-size: 18px;
  margin: 0 0 20px 0;
  opacity: 0.9;
  line-height: 1.6;
}

.feature-tags {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.action-section {
  margin-top: 16px;
}

.action-button {
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: white;
  backdrop-filter: blur(10px);
  transition: all 0.3s ease;
}

.action-button:hover {
  background: rgba(255, 255, 255, 0.3);
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.2);
}

@media (max-width: 768px) {
  .home-view {
    padding: 16px;
  }

  .welcome-header {
    flex-direction: column;
    text-align: center;
  }
}
</style>
