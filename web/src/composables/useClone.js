/**
 * useClone - 克隆仓库的通用 composable
 *
 * 封装了 SSE 实时克隆进度的逻辑，可在多个组件中复用
 */
import { ref, onUnmounted } from 'vue'
import { useMessage } from 'naive-ui'
import request from '@/utils/request'

// 获取 API 基础 URL
function getApiBaseUrl() {
  const apiUrl = import.meta.env.VITE_API_URL

  // 只有当环境变量有效且不是 '/' 或空字符串时才使用
  if (apiUrl && apiUrl !== '/' && apiUrl !== '') {
    return apiUrl
  }

  if (import.meta.env.DEV) {
    const backendPort = import.meta.env.VITE_API_PORT || '8080'
    return `http://localhost:${backendPort}`
  }
  return window.location.origin
}

export function useClone() {
  const message = useMessage()
  const cloning = ref(false)
  const cloneProgress = ref('')
  const eventSource = ref(null)

  /**
   * 克隆仓库
   * @param {number|string} repoId - 仓库ID
   * @param {object} options - 配置选项
   * @param {function} options.onProgress - 进度回调
   * @param {function} options.onComplete - 完成回调
   * @param {function} options.onError - 错误回调
   * @returns {Promise<boolean>}
   */
  const cloneRepo = async (repoId, options = {}) => {
    const { onProgress, onComplete, onError } = options
    cloning.value = true
    cloneProgress.value = ''

    try {
      const response = await request.post(`/api/repos/${repoId}/clone`)

      if (response.data.data?.useSSE) {
        // 使用 SSE 获取实时进度
        const baseUrl = getApiBaseUrl()
        const tempToken = response.data.data.tempToken

        return new Promise((resolve) => {
          eventSource.value = new EventSource(
            `${baseUrl}/api/repos/${repoId}/clone-status?tempToken=${tempToken}`
          )

          eventSource.value.onopen = () => {
            cloneProgress.value = '开始克隆仓库...'
            onProgress?.('开始克隆仓库...')
          }

          eventSource.value.addEventListener('start', (event) => {
            try {
              const data = JSON.parse(event.data)
              cloneProgress.value = data.message
              onProgress?.(data.message)
            } catch (e) {
              console.error('Parse start event error:', e)
            }
          })

          eventSource.value.addEventListener('progress', (event) => {
            try {
              const data = JSON.parse(event.data)
              cloneProgress.value = data.message
              onProgress?.(data.message)
            } catch (e) {
              console.error('Parse progress event error:', e)
            }
          })

          eventSource.value.addEventListener('complete', (event) => {
            try {
              const data = JSON.parse(event.data)
              cloneProgress.value = data.message
            } catch (e) {
              cloneProgress.value = '克隆完成'
            }
            closeEventSource()
            cloning.value = false
            onComplete?.()
            resolve(true)
          })

          eventSource.value.addEventListener('error', (event) => {
            let errorMsg = '克隆过程中发生错误'
            try {
              const data = JSON.parse(event.data)
              errorMsg = data.message
            } catch (e) {
              console.error('Parse error event error:', e)
            }
            cloneProgress.value = errorMsg
            closeEventSource()
            cloning.value = false
            onError?.(errorMsg)
            resolve(false)
          })

          eventSource.value.onerror = () => {
            cloneProgress.value = '连接中断'
            closeEventSource()
            cloning.value = false
            onError?.('连接中断')
            resolve(false)
          }
        })
      } else {
        // 不使用 SSE，直接返回结果
        cloning.value = false
        if (response.data.message === '本地仓库已存在') {
          onComplete?.()
          return true
        }
        message.success(response.data.message)
        onComplete?.()
        return true
      }
    } catch (error) {
      console.error('克隆仓库失败:', error)
      cloning.value = false
      const errorMsg = error.message || '克隆仓库失败'
      message.error(errorMsg)
      onError?.(errorMsg)
      return false
    }
  }

  /**
   * 关闭 SSE 连接
   */
  const closeEventSource = () => {
    if (eventSource.value) {
      eventSource.value.close()
      eventSource.value = null
    }
  }

  /**
   * 取消克隆
   */
  const cancelClone = () => {
    closeEventSource()
    cloning.value = false
    cloneProgress.value = ''
  }

  // 组件卸载时自动关闭 SSE 连接
  onUnmounted(() => {
    closeEventSource()
  })

  return {
    cloning,
    cloneProgress,
    cloneRepo,
    cancelClone
  }
}
