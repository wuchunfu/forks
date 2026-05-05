import request from '@/utils/request'

// 版本级缓存：语言数据随版本固定，同版本不重复请求
let cachedLanguagesPromise = null
let cachedVersion = null

async function getCachedLanguages() {
  try {
    const verRes = await request({ url: '/api/version', method: 'get' })
    const currentVersion = verRes.data?.data?.version || 'dev'

    if (cachedLanguagesPromise && cachedVersion === currentVersion) {
      return cachedLanguagesPromise
    }

    cachedVersion = currentVersion
    cachedLanguagesPromise = request({ url: '/api/trending/languages', method: 'get' })
    return cachedLanguagesPromise
  } catch {
    // version 接口失败时仍请求（降级）
    return request({ url: '/api/trending/languages', method: 'get' })
  }
}

export function getTrending(params = {}) {
  return request({
    url: '/api/trending',
    method: 'get',
    params,
    timeout: 30000
  })
}

export function getTrendingLanguages() {
  return getCachedLanguages()
}

export function getTrendingDates(params = {}) {
  return request({
    url: '/api/trending/dates',
    method: 'get',
    params
  })
}

export function getSyncConfig() {
  return request({
    url: '/api/trending/sync-config',
    method: 'get'
  })
}

export function updateSyncConfig(data) {
  return request({
    url: '/api/trending/sync-config',
    method: 'post',
    data
  })
}

export function syncNow() {
  return request({
    url: '/api/trending/sync-now',
    method: 'post',
    timeout: 10000
  })
}

export function getMCPTools() {
  return request({
    url: '/api/mcp/tools',
    method: 'get'
  })
}
