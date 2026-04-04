import request from '@/utils/request'

// 获取仓库列表
export function getRepos(params = {}) {
  return request({
    url: '/api/repos',
    method: 'get',
    params
  })
}

// 获取作者列表
export function getAuthors(params = {}) {
  return request({
    url: '/api/authors',
    method: 'get',
    params
  })
}

// 获取统计数据
export function getStats() {
  return request({
    url: '/api/stats',
    method: 'get'
  })
}

// 获取活动记录列表
export function getActivities(params = {}) {
  return request({
    url: '/api/activities',
    method: 'get',
    params
  })
}

// 添加活动记录
export function addActivity(data) {
  return request({
    url: '/api/activities',
    method: 'post',
    data
  })
}

// 清空活动记录
export function clearActivities() {
  return request({
    url: '/api/activities',
    method: 'delete'
  })
}

// 获取仓库时间轴数据（不分页，返回所有数据）
export function getReposTimeline(params = {}) {
  return request({
    url: '/api/repos/timeline',
    method: 'get',
    params
  })
}

// 添加仓库
export function addRepo(data) {
  return request({
    url: '/api/repos',
    method: 'post',
    data
  })
}

// 删除仓库
export function deleteRepo(id) {
  return request({
    url: `/api/repos/${id}`,
    method: 'delete'
  })
}

// 获取仓库状态
export function getRepoStatus(id) {
  return request({
    url: `/api/repos/${id}/status`,
    method: 'get'
  })
}

// 获取单个仓库
export function getRepo(id) {
  return request({
    url: `/api/repos/${id}`,
    method: 'get'
  })
}

// 获取系统信息
export function getSystemInfo() {
  return request({
    url: '/api/info',
    method: 'get'
  })
}

// 更新仓库信息
export function updateRepoInfo(id) {
  return request({
    url: `/api/repos/${id}/update`,
    method: 'put'
  })
}

// 获取时间轴维度信息（智能判断时间粒度）
export function getTimelineDimensions(params = {}) {
  return request({
    url: '/api/timeline/dimensions',
    method: 'get',
    params
  })
}

// 获取指定时间段的仓库数据（支持分页）
export function getTimelinePeriod(params = {}) {
  return request({
    url: '/api/timeline/period',
    method: 'get',
    params
  })
}

// 扫描本地Git仓库
export function scanRepos() {
  return request({
    url: '/api/repos/scan',
    method: 'post'
  })
}

// 获取 API 基础 URL
function getApiBaseUrl() {
  const apiUrl = import.meta.env.VITE_API_URL

  // 1. 只有当环境变量有效且不是 '/' 或空字符串时才使用
  if (apiUrl && apiUrl !== '/' && apiUrl !== '') {
    return apiUrl
  }

  // 2. 开发环境检测
  if (import.meta.env.DEV) {
    const backendPort = import.meta.env.VITE_API_PORT || '8080'
    return `http://localhost:${backendPort}`
  }

  // 3. 生产环境使用当前域名
  return window.location.origin
}

// 扫描进度SSE连接
export function scanReposSSE(taskId, tempToken, onProgress, onComplete, onError) {
  const baseUrl = getApiBaseUrl()
  const url = `${baseUrl}/api/repos/scan-status?taskId=${taskId}&tempToken=${tempToken}`

  console.log('[SSE] 连接扫描进度:', url)

  const eventSource = new EventSource(url)

  // 监听 start 事件
  eventSource.addEventListener('start', (event) => {
    console.log('[SSE] 扫描开始:', event.data)
  })

  // 监听 progress 事件
  eventSource.addEventListener('progress', (event) => {
    console.log('[SSE] 扫描进度:', event.data)
    try {
      const data = JSON.parse(event.data)
      if (onProgress) onProgress(data)
    } catch (e) {
      console.error('[SSE] 解析进度失败:', e)
    }
  })

  // 监听 complete 事件
  eventSource.addEventListener('complete', (event) => {
    console.log('[SSE] 扫描完成:', event.data)
    try {
      const data = JSON.parse(event.data)
      if (onComplete) onComplete(data)
    } catch (e) {
      console.error('[SSE] 解析完成数据失败:', e)
      if (onError) onError({ message: '解析结果失败' })
    }
    eventSource.close()
  })

  // 监听 error 事件
  eventSource.addEventListener('error', (event) => {
    console.error('[SSE] 错误事件:', event.data)
    try {
      const data = event.data ? JSON.parse(event.data) : { message: 'SSE连接错误' }
      if (onError) onError(data)
    } catch (e) {
      if (onError) onError({ message: '连接失败' })
    }
    eventSource.close()
  })

  // 处理连接错误
  eventSource.onerror = (err) => {
    console.error('[SSE] 连接错误:', err, 'readyState:', eventSource.readyState)
  }

  return eventSource
}

// 批量克隆仓库
export function batchCloneRepos(data = {}) {
  return request({
    url: '/api/repos/batch-clone',
    method: 'post',
    data
  })
}

// 批量克隆进度SSE连接
export function batchCloneSSE(tempToken, onProgress, onComplete, onError) {
  const baseUrl = getApiBaseUrl()
  const url = `${baseUrl}/api/repos/batch-clone-status?tempToken=${tempToken}`

  console.log('[SSE] 连接批量克隆进度:', url)

  const eventSource = new EventSource(url)

  eventSource.addEventListener('start', (event) => {
    const data = JSON.parse(event.data)
    if (onProgress) onProgress({ type: 'start', ...data })
  })

  eventSource.addEventListener('progress', (event) => {
    const data = JSON.parse(event.data)
    if (onProgress) onProgress({ type: 'progress', ...data })
  })

  eventSource.addEventListener('complete', (event) => {
    const data = JSON.parse(event.data)
    if (onComplete) onComplete(data)
    eventSource.close()
  })

  eventSource.addEventListener('error', (event) => {
    const data = event.data ? JSON.parse(event.data) : { message: '连接失败' }
    if (onError) onError(data)
    eventSource.close()
  })

  eventSource.onerror = () => {
    if (onError) onError({ message: '连接中断' })
  }

  return eventSource
}

// 获取代理配置
export function getProxyConfig() {
  return request({
    url: '/api/proxy',
    method: 'get'
  })
}

// 更新代理配置
export function updateProxyConfig(data) {
  return request({
    url: '/api/proxy',
    method: 'post',
    data
  })
}

// 获取当前 Token 信息（脱敏）
export function getTokenInfo() {
  return request({
    url: '/api/token',
    method: 'get'
  })
}

// 更新 Token
export function updateToken(data) {
  return request({
    url: '/api/token',
    method: 'post',
    data
  })
}