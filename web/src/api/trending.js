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
