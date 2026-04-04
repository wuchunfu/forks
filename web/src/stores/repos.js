import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getRepos, deleteRepo, addRepo, getStats, getRepoStatus } from '@/api/repos'

export const useReposStore = defineStore('repos', () => {
  // 状态
  const repos = ref([])
  const loading = ref(false)
  const error = ref(null)

  // 统计数据 (API)
  const stats = ref({
    total_repos: 0,
    cloned_count: 0,
    not_cloned_count: 0,
    author_count: 0
  })

  // 分页
  const total = ref(0)
  const currentPage = ref(1)
  const pageSize = ref(20)
  const totalPages = ref(0)

  // 筛选
  const searchQuery = ref('')
  const activeFilters = ref({
    languages: [],
    dateRange: null,
    status: 'all',
    author: null,
    starsRange: null,
    hasClone: null,
    source: null
  })

  // 仓库状态
  const repoStatuses = ref({})
  const selectedRepos = ref([])
  const batchMode = ref(false)

  // 横向时间轴状态
  const timelineDimensions = ref(null) // 时间维度结构
  const currentDimension = ref(null) // 当前维度 (year/month/week/day)
  const currentPeriod = ref(null) // 当前时间段
  const dimensionRepos = ref([]) // 当前时间段的仓库数据
  const dimensionLoading = ref(false) // 维度加载状态

  // 计算属性 - 直接使用后端返回的数据
  const filteredRepos = computed(() => {
    return repos.value
  })

  const paginatedRepos = computed(() => {
    return repos.value
  })

  const uniqueAuthors = computed(() => {
    const authors = [...new Set(repos.value.map(repo => repo.author))]
    return authors.sort()
  })

  const uniqueLanguages = computed(() => {
    const languages = new Set()
    repos.value.forEach(repo => {
      if (repo.languages) {
        try {
          JSON.parse(repo.languages).forEach(lang => languages.add(lang))
        } catch {
          // 忽略解析错误
        }
      }
    })
    return Array.from(languages).sort()
  })

  // 方法
  const fetchRepos = async (params = {}) => {
    loading.value = true
    error.value = null

    // 构建筛选参数
    const filterParams = {
      page: currentPage.value,
      page_size: pageSize.value,
      search: searchQuery.value || undefined,
      author: activeFilters.value.author || undefined,
      status: activeFilters.value.status !== 'all' ? activeFilters.value.status : undefined,
      source: activeFilters.value.source || undefined,
      ...params
    }

    // 移除 undefined 值
    Object.keys(filterParams).forEach(key => {
      if (filterParams[key] === undefined) {
        delete filterParams[key]
      }
    })

    try {
      const response = await getRepos(filterParams)

      // 检查响应结构
      const responseData = response.data || response
      if (response && responseData && responseData.code === 0) {
        // 处理数据
        repos.value = responseData.data?.list || responseData.list || []
        total.value = responseData.data?.total || responseData.total || 0
        totalPages.value = responseData.data?.total_pages || responseData.total_pages || 0
        currentPage.value = responseData.data?.page || responseData.page || 1
        pageSize.value = responseData.data?.page_size || responseData.page_size || 10
      } else {
        const errorMsg = responseData?.message || 'API返回错误'
        throw new Error(errorMsg)
      }
    } catch (err) {
      error.value = err.message
    } finally {
      loading.value = false
    }
  }

  // 时间轴专用方法 - 获取所有数据不分页
  const fetchTimelineRepos = async (params = {}) => {
    loading.value = true
    error.value = null

    try {
      // 导入时间轴专用API
      const { getReposTimeline } = await import('@/api/repos')

      const response = await getReposTimeline({
        search: searchQuery.value,
        ...params
      })

      const responseData = response.data || response
      if (response && responseData && responseData.code === 0) {
        // 时间轴返回所有数据,不分页
        repos.value = responseData.data || []
        total.value = responseData.total || repos.value.length

        // 获取仓库状态
        await fetchRepoStatuses()
      } else {
        throw new Error(responseData?.message || 'API返回错误')
      }
    } catch (err) {
      error.value = err.message
    } finally {
      loading.value = false
    }
  }

  // 横向时间轴 - 获取时间维度结构
  const fetchTimelineDimensions = async (params = {}) => {
    dimensionLoading.value = true
    error.value = null

    try {
      const { getTimelineDimensions } = await import('@/api/repos')

      const response = await getTimelineDimensions({
        search: searchQuery.value,
        ...params
      })

      const responseData = response.data || response
      if (response && responseData && responseData.code === 0) {
        timelineDimensions.value = responseData.data
        // 设置默认推荐维度
        currentDimension.value = responseData.data?.recommended_granularity || 'month'
      } else {
        throw new Error(responseData?.message || 'API返回错误')
      }
    } catch (err) {
      error.value = err.message
    } finally {
      dimensionLoading.value = false
    }
  }

  // 横向时间轴 - 获取指定时间段的仓库数据
  const fetchTimelinePeriod = async (periodParams, page = 1, pageSize = 12) => {
    dimensionLoading.value = true
    error.value = null

    try {
      const { getTimelinePeriod } = await import('@/api/repos')

      const response = await getTimelinePeriod({
        year: periodParams.year,
        month: periodParams.month,
        week: periodParams.week,
        day: periodParams.day,
        page,
        page_size: pageSize,
        search: searchQuery.value
      })

      const responseData = response.data || response
      if (response && responseData && responseData.code === 0) {
        const list = responseData.data?.list || []
        const total = responseData.data?.total || 0

        // 如果是第一页,直接替换;否则追加
        if (page === 1) {
          dimensionRepos.value = list
        } else {
          dimensionRepos.value = [...dimensionRepos.value, ...list]
        }

        // 保存当前时间段信息
        currentPeriod.value = periodParams

        return {
          list,
          total,
          hasMore: dimensionRepos.value.length < total
        }
      } else {
        throw new Error(responseData?.message || 'API返回错误')
      }
    } catch (err) {
      error.value = err.message
      throw err
    } finally {
      dimensionLoading.value = false
    }
  }

  // 切换时间维度
  const switchDimension = (dimension) => {
    currentDimension.value = dimension
    // 重置当前时间段和数据
    currentPeriod.value = null
    dimensionRepos.value = []
  }

  const addNewRepo = async (repoData) => {
    try {
      const response = await addRepo(repoData)
      if (response.code === 0) {
        // 刷新列表
        await fetchRepos()
        return response.data
      } else {
        throw new Error(response.message)
      }
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  const removeRepo = async (repoId) => {
    try {
      await deleteRepo(repoId)

      // 从列表中移除
      repos.value = repos.value.filter(repo => repo.id !== repoId)
      selectedRepos.value = selectedRepos.value.filter(repo => repo.id !== repoId)

      // 移除状态缓存
      delete repoStatuses.value[repoId]

      // 更新总数
      total.value = Math.max(0, total.value - 1)
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  const batchDeleteRepos = async (repoIds) => {
    const results = []
    const errors = []

    for (const repoId of repoIds) {
      try {
        await removeRepo(repoId)
        results.push(repoId)
      } catch (err) {
        errors.push({ repoId, error: err.message })
      }
    }

    return { results, errors }
  }

  const fetchRepoStatuses = async () => {
    const promises = repos.value.map(async (repo) => {
      try {
        const response = await getRepoStatus(repo.id)
        const responseData = response.data || response
        if (responseData && responseData.code === 0) {
          repoStatuses.value[repo.id] = responseData.data
        }
      } catch {
        // 单个仓库状态获取失败不影响其他仓库
      }
    })
    await Promise.all(promises)
  }

  const updateRepoStatus = (repoId, status) => {
    repoStatuses.value[repoId] = { ...repoStatuses.value[repoId], ...status }
  }

  const setSearchQuery = (query) => {
    searchQuery.value = query
    currentPage.value = 1
  }

  const setActiveFilter = (filterType, value) => {
    activeFilters.value[filterType] = value
    currentPage.value = 1
  }

  const clearFilters = () => {
    activeFilters.value = {
      languages: [],
      dateRange: null,
      status: 'all',
      author: null,
      starsRange: null,
      hasClone: null,
      source: null
    }
    searchQuery.value = ''
    currentPage.value = 1
  }

  const setCurrentPage = (page) => {
    currentPage.value = page
  }

  const setPageSize = (size) => {
    pageSize.value = size
    currentPage.value = 1
  }

  // 选择相关方法
  const selectRepo = (repo, selected) => {
    if (selected) {
      if (!selectedRepos.value.some(r => r.id === repo.id)) {
        selectedRepos.value.push(repo)
      }
    } else {
      selectedRepos.value = selectedRepos.value.filter(r => r.id !== repo.id)
    }
  }

  const selectAllRepos = () => {
    selectedRepos.value = [...filteredRepos.value]
  }

  const clearSelection = () => {
    selectedRepos.value = []
  }

  const toggleBatchMode = () => {
    batchMode.value = !batchMode.value
    if (!batchMode.value) {
      clearSelection()
    }
  }

  const isRepoSelected = (repo) => {
    return selectedRepos.value.some(r => r.id === repo.id)
  }

  const isAllSelected = computed(() => {
    return filteredRepos.value.length > 0 &&
           selectedRepos.value.length === filteredRepos.value.length
  })

  const isIndeterminate = computed(() => {
    return selectedRepos.value.length > 0 &&
           selectedRepos.value.length < filteredRepos.value.length
  })

  // 统计信息 - 来自 API
  const statistics = computed(() => {
    return {
      total: stats.value.total_repos,
      cloned: stats.value.cloned_count,
      notCloned: stats.value.not_cloned_count,
      authorCount: stats.value.author_count
    }
  })

  // 获取统计数据
  const fetchStats = async () => {
    try {
      const response = await getStats()
      const responseData = response.data || response
      if (response && responseData && responseData.code === 0) {
        stats.value = responseData.data
      }
    } catch (err) {
      console.error('获取统计数据失败:', err)
    }
  }

  return {
    // 状态
    repos,
    loading,
    error,
    total,
    stats,
    currentPage,
    pageSize,
    totalPages,
    searchQuery,
    activeFilters,
    repoStatuses,
    selectedRepos,
    batchMode,
    timelineDimensions,
    currentDimension,
    currentPeriod,
    dimensionRepos,
    dimensionLoading,

    // 计算属性
    filteredRepos,
    paginatedRepos,
    uniqueAuthors,
    uniqueLanguages,
    isAllSelected,
    isIndeterminate,
    statistics,

    // 方法
    fetchRepos,
    fetchStats,
    fetchTimelineRepos, // 时间轴专用方法
    fetchTimelineDimensions, // 横向时间轴维度结构
    fetchTimelinePeriod, // 横向时间轴时间段数据
    switchDimension, // 切换时间维度
    addNewRepo,
    removeRepo,
    batchDeleteRepos,
    fetchRepoStatuses,
    updateRepoStatus,
    setSearchQuery,
    setActiveFilter,
    clearFilters,
    setCurrentPage,
    setPageSize,
    selectRepo,
    selectAllRepos,
    clearSelection,
    toggleBatchMode,
    isRepoSelected
  }
})