<template>
  <div class="settings-panel">
    <div class="panel-header">
      <h2 class="panel-title">令牌管理</h2>
      <p class="panel-desc">管理 API 访问令牌，修改后当前会话自动更新</p>
    </div>

    <div class="settings-form">
      <div class="form-group">
        <label class="form-label">当前令牌</label>
        <div class="token-display">
          <code class="token-value">{{ currentToken || '加载中...' }}</code>
          <button class="btn-icon" title="复制令牌" @click="copyToken">
            <svg v-if="!copied" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <rect x="9" y="9" width="13" height="13" rx="2"/>
              <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/>
            </svg>
            <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M20 6L9 17l-5-5"/>
            </svg>
          </button>
        </div>
        <span class="form-hint">服务启动后首次生成时可复制完整令牌，脱敏后仅页面内可见</span>
      </div>

      <div class="form-group">
        <label class="form-label">生成方式</label>
        <div class="radio-group">
          <label class="radio-item">
            <input type="radio" v-model="tokenMode" value="auto" />
            <span>自动生成</span>
          </label>
          <label class="radio-item">
            <input type="radio" v-model="tokenMode" value="custom" />
            <span>自定义</span>
          </label>
        </div>
      </div>

      <div v-if="tokenMode === 'auto'" class="form-group">
        <div class="token-actions">
          <button class="btn btn-secondary" :disabled="tokenLoading" @click="regenerateToken">
            {{ tokenLoading ? '处理中...' : '重新生成' }}
          </button>
        </div>
      </div>

      <div v-else class="form-group">
        <div class="token-custom">
          <input
            v-model="customTokenInput"
            type="text"
            class="form-input"
            placeholder="输入自定义令牌（至少8位）"
            @keyup.enter="applyCustomToken"
          />
          <button
            class="btn btn-primary"
            :disabled="tokenLoading || !customTokenInput || customTokenInput.length < 8"
            @click="applyCustomToken"
          >
            应用
          </button>
        </div>
        <span class="form-hint">修改后当前浏览器会话自动更新，其他浏览器需重新登录</span>
      </div>
    </div>

    <div class="settings-footer">
      <div class="footer-info">
        <span class="info-text">设置自动保存到本地配置文件</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useMessage } from 'naive-ui'
import { getTokenInfo, updateToken } from '@/api/repos'
import { copyToClipboard } from '@/utils/clipboard'

const message = useMessage()

const currentToken = ref('')
const plainToken = ref('')
const tokenMode = ref('auto')
const customTokenInput = ref('')
const tokenLoading = ref(false)
const copied = ref(false)

const loadTokenInfo = async () => {
  try {
    const response = await getTokenInfo()
    const res = response.data || response
    if (res.code === 0 && res.data) {
      currentToken.value = res.data.token
      plainToken.value = ''
    }
  } catch (error) {
    console.error('加载 Token 信息失败:', error)
  }
}

const copyToken = async () => {
  const text = plainToken.value || localStorage.getItem('token')
  if (!text) {
    message.warning('无法复制脱敏令牌，请重新生成后再复制')
    return
  }
  try {
    await copyToClipboard(text)
    copied.value = true
    message.success('令牌已复制到剪贴板')
    setTimeout(() => { copied.value = false }, 2000)
  } catch {
    message.error('复制失败')
  }
}

const onTokenUpdated = (newToken) => {
  localStorage.setItem('token', newToken)
  plainToken.value = newToken
  currentToken.value = newToken.substring(0, 4) + '****' + newToken.substring(newToken.length - 4)
  customTokenInput.value = ''
}

const regenerateToken = async () => {
  tokenLoading.value = true
  try {
    const response = await updateToken({ regenerate: true })
    const res = response.data || response
    if (res.code === 0 && res.data) {
      onTokenUpdated(res.data.token)
      message.success('令牌已重新生成并自动更新')
    } else {
      message.error(res.message || '重新生成失败')
    }
  } catch (error) {
    message.error('操作失败：' + error.message)
  } finally {
    tokenLoading.value = false
  }
}

const applyCustomToken = async () => {
  if (!customTokenInput.value || customTokenInput.value.length < 8) {
    message.warning('令牌长度不能少于8位')
    return
  }
  tokenLoading.value = true
  try {
    const response = await updateToken({ token: customTokenInput.value })
    const res = response.data || response
    if (res.code === 0 && res.data) {
      onTokenUpdated(res.data.token)
      message.success('令牌已更新并自动切换')
    } else {
      message.error(res.message || '更新失败')
    }
  } catch (error) {
    message.error('操作失败：' + error.message)
  } finally {
    tokenLoading.value = false
  }
}

onMounted(() => {
  loadTokenInfo()
})
</script>

<style scoped>
.settings-panel {
  padding: var(--space-6);
  max-height: calc(100vh - var(--navbar-height) - var(--space-6) * 2);
  overflow-y: auto;
}

.panel-header {
  margin-bottom: var(--space-6);
  padding-bottom: var(--space-4);
  border-bottom: 1px solid var(--color-border-light);
}

.panel-title {
  font-size: var(--text-xl);
  font-weight: var(--font-semibold);
  color: var(--color-text-primary);
  margin: 0 0 var(--space-1) 0;
}

.panel-desc {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  margin: 0;
}

.settings-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.form-label {
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  color: var(--color-text-primary);
}

.form-input {
  width: 100%;
  max-width: 400px;
  padding: var(--space-2_5) var(--space-3);
  font-size: var(--text-sm);
  color: var(--color-text-primary);
  background-color: var(--color-bg-page);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  outline: none;
  transition: all 0.2s ease;
}

.form-input:focus {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-50);
}

.radio-group {
  display: flex;
  gap: var(--space-6);
}

.radio-item {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  cursor: pointer;
  font-size: var(--text-sm);
  color: var(--color-text-primary);
}

.radio-item input[type="radio"] {
  width: 18px;
  height: 18px;
  cursor: pointer;
}

.form-hint {
  font-size: var(--text-xs);
  color: var(--color-text-tertiary);
}

.token-display {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.btn-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-bg-page);
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.btn-icon:hover:not(:disabled) {
  border-color: var(--color-primary);
  color: var(--color-primary);
  background-color: var(--color-primary-50);
}

.btn-icon:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.btn-icon svg {
  width: 16px;
  height: 16px;
}

.token-value {
  display: inline-block;
  min-width: 200px;
  max-width: 320px;
  padding: var(--space-2_5) var(--space-4);
  background-color: var(--color-bg-page);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  font-size: var(--text-sm);
  font-family: var(--font-mono, 'Courier New', Courier, monospace);
  color: var(--color-text-primary);
  letter-spacing: 1px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.token-actions {
  display: flex;
  gap: var(--space-3);
}

.token-custom {
  display: flex;
  gap: var(--space-3);
  align-items: center;
}

.token-custom .form-input {
  flex: 1;
  max-width: 360px;
}

.settings-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-4) var(--space-6);
  border-top: 1px solid var(--color-border-light);
  flex-wrap: wrap;
  gap: var(--space-4);
  flex-shrink: 0;
}

.footer-info {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.info-text {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
}

.btn {
  padding: var(--space-2_5) var(--space-4);
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  border-radius: var(--radius-md);
  border: 1px solid transparent;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-secondary {
  color: var(--color-text-primary);
  background-color: var(--color-gray-100);
  border-color: var(--color-border);
}

.btn-secondary:hover {
  background-color: var(--color-gray-200);
}

.btn-primary {
  color: white;
  background-color: var(--color-primary);
}

.btn-primary:hover {
  background-color: var(--color-primary-600);
}
</style>
