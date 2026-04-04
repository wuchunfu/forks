/**
 * 日期格式化工具函数
 */

/**
 * 格式化日期
 * @param {string|Date} date - 日期字符串或Date对象
 * @param {string} format - 格式化模式
 * @returns {string} 格式化后的日期字符串
 */
export function formatDate(date, format = 'YYYY-MM-DD') {
  if (!date) return ''

  const d = new Date(date)
  if (isNaN(d.getTime())) return ''

  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  const hours = String(d.getHours()).padStart(2, '0')
  const minutes = String(d.getMinutes()).padStart(2, '0')
  const seconds = String(d.getSeconds()).padStart(2, '0')

  return format
    .replace('YYYY', year)
    .replace('MM', month)
    .replace('DD', day)
    .replace('HH', hours)
    .replace('mm', minutes)
    .replace('ss', seconds)
}

/**
 * 格式化为相对时间
 * @param {string|Date} date - 日期
 * @returns {string} 相对时间字符串
 */
export function formatRelativeTime(date) {
  if (!date) return ''

  const now = new Date()
  const target = new Date(date)
  const diff = now - target

  const minutes = Math.floor(diff / (1000 * 60))
  const hours = Math.floor(diff / (1000 * 60 * 60))
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))

  if (minutes < 1) {
    return '刚刚'
  } else if (minutes < 60) {
    return `${minutes}分钟前`
  } else if (hours < 24) {
    return `${hours}小时前`
  } else if (days < 7) {
    return `${days}天前`
  } else if (days < 30) {
    return `${Math.floor(days / 7)}周前`
  } else if (days < 365) {
    return `${Math.floor(days / 30)}个月前`
  } else {
    return `${Math.floor(days / 365)}年前`
  }
}

/**
 * 格式化为本地日期
 * @param {string|Date} date - 日期
 * @returns {string} 本地日期字符串
 */
export function toLocaleDate(date) {
  if (!date) return ''

  const d = new Date(date)
  if (isNaN(d.getTime())) return ''

  return d.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
}

/**
 * 格式化为本地时间
 * @param {string|Date} date - 日期
 * @returns {string} 本地时间字符串
 */
export function toLocaleTime(date) {
  if (!date) return ''

  const d = new Date(date)
  if (isNaN(d.getTime())) return ''

  return d.toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit'
  })
}

/**
 * 格式化为本地日期时间
 * @param {string|Date} date - 日期
 * @returns {string} 本地日期时间字符串
 */
export function toLocaleDateTime(date) {
  if (!date) return ''

  const d = new Date(date)
  if (isNaN(d.getTime())) return ''

  return d.toLocaleString('zh-CN', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

/**
 * 获取日期范围的开始和结束
 * @param {string} range - 范围类型：'today', 'week', 'month', 'year'
 * @returns {{ start: Date, end: Date }} 日期范围
 */
export function getDateRange(range) {
  const now = new Date()
  const start = new Date()
  const end = new Date()

  switch (range) {
    case 'today':
      start.setHours(0, 0, 0, 0)
      end.setHours(23, 59, 59, 999)
      break
    case 'week':
      start.setDate(now.getDate() - now.getDay())
      start.setHours(0, 0, 0, 0)
      end.setDate(start.getDate() + 6)
      end.setHours(23, 59, 59, 999)
      break
    case 'month':
      start.setDate(1)
      start.setHours(0, 0, 0, 0)
      end.setMonth(now.getMonth() + 1, 0)
      end.setHours(23, 59, 59, 999)
      break
    case 'year':
      start.setMonth(0, 1)
      start.setHours(0, 0, 0, 0)
      end.setMonth(11, 31)
      end.setHours(23, 59, 59, 999)
      break
  }

  return { start, end }
}

/**
 * 判断是否为今天
 * @param {string|Date} date - 日期
 * @returns {boolean} 是否为今天
 */
export function isToday(date) {
  const today = new Date()
  const target = new Date(date)

  return today.getDate() === target.getDate() &&
         today.getMonth() === target.getMonth() &&
         today.getFullYear() === target.getFullYear()
}

/**
 * 判断是否为昨天
 * @param {string|Date} date - 日期
 * @returns {boolean} 是否为昨天
 */
export function isYesterday(date) {
  const yesterday = new Date()
  yesterday.setDate(yesterday.getDate() - 1)
  const target = new Date(date)

  return yesterday.getDate() === target.getDate() &&
         yesterday.getMonth() === target.getMonth() &&
         yesterday.getFullYear() === target.getFullYear()
}