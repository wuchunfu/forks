import request from '@/utils/request'

export function getTrending(params = {}) {
  return request({
    url: '/api/trending',
    method: 'get',
    params,
    timeout: 30000
  })
}

export function getTrendingLanguages() {
  return request({
    url: '/api/trending/languages',
    method: 'get'
  })
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
