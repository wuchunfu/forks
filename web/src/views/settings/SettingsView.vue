<template>
  <div class="settings-view">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">系统设置</h1>
      <p class="page-description">配置系统参数和偏好设置</p>
    </div>

    <div class="settings-layout">
      <!-- 左侧导航 -->
      <div class="settings-nav">
        <div class="nav-section">
          <div class="nav-section-title">设置</div>
          <nav class="nav-list">
            <button
              v-for="item in navItems"
              :key="item.key"
              class="nav-item"
              :class="{ active: activeTab === item.key }"
              @click="activeTab = item.key"
            >
              <component :is="item.icon" class="nav-icon" />
              <span class="nav-label">{{ item.label }}</span>
            </button>
          </nav>
        </div>
      </div>

      <!-- 右侧内容区 -->
      <div class="settings-content">
        <!-- 代理设置 -->
        <div v-if="activeTab === 'proxy'" class="settings-panel">
          <div class="panel-header">
            <h2 class="panel-title">代理设置</h2>
            <p class="panel-desc">配置 Git 操作的代理服务器</p>
          </div>

          <div class="settings-form">
            <!-- 代理开关 -->
            <div class="form-group">
              <label class="form-label">启用代理</label>
              <div class="radio-group">
                <label class="radio-item">
                  <input type="radio" v-model="settings.enabled" :value="false" />
                  <span>不使用代理</span>
                </label>
                <label class="radio-item">
                  <input type="radio" v-model="settings.enabled" :value="true" />
                  <span>使用代理</span>
                </label>
              </div>
            </div>

            <template v-if="settings.enabled">
              <!-- 代理类型 -->
              <div class="form-group">
                <label class="form-label">代理类型</label>
                <div class="radio-group">
                  <label class="radio-item">
                    <input type="radio" v-model="settings.type" value="http" />
                    <span>HTTP</span>
                  </label>
                  <label class="radio-item">
                    <input type="radio" v-model="settings.type" value="socks5" />
                    <span>SOCKS5</span>
                  </label>
                </div>
              </div>

              <!-- 代理地址 -->
              <div class="form-group">
                <label class="form-label">代理地址</label>
                <input
                  v-model="settings.host"
                  type="text"
                  class="form-input"
                  placeholder="127.0.0.1"
                />
              </div>

              <!-- 代理端口 -->
              <div class="form-group">
                <label class="form-label">代理端口</label>
                <input
                  v-model.number="settings.port"
                  type="number"
                  class="form-input"
                  placeholder="7890"
                />
              </div>

              <!-- 代理排除列表 -->
              <div class="form-group">
                <label class="form-label">不代理的主机（可选）</label>
                <input
                  v-model="settings.no_proxy"
                  type="text"
                  class="form-input"
                  placeholder="localhost,127.0.0.1,*.local"
                />
              </div>
            </template>

            <!-- 快捷代理预设 -->
            <div class="form-group" v-if="settings.enabled">
              <label class="form-label">快捷代理预设</label>
              <div class="preset-grid">
                <button
                  v-for="preset in proxyPresets"
                  :key="preset.name"
                  class="preset-btn"
                  @click="applyPreset(preset)"
                >
                  <div class="preset-icon">
                    <component :is="preset.icon" />
                  </div>
                  <div class="preset-info">
                    <div class="preset-name">{{ preset.label }}</div>
                    <div class="preset-desc">{{ preset.port }} 端口</div>
                  </div>
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- 平台代理 -->
        <div v-else-if="activeTab === 'platform-proxy'" class="settings-panel">
          <div class="panel-header">
            <h2 class="panel-title">平台代理</h2>
            <p class="panel-desc">单独控制每个平台是否走代理，未设置时跟随全局开关</p>
          </div>

          <div class="settings-form">
            <!-- 全局代理状态提示 -->
            <div class="proxy-status-hint" :class="{ disabled: !settings.enabled }">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="status-icon">
                <circle cx="12" cy="12" r="10"/>
                <path v-if="settings.enabled" d="M12 8v4l3 3"/>
                <path v-else d="M4.93 4.93l14.14 14.14"/>
              </svg>
              <span>全局代理：{{ settings.enabled ? '已启用' : '未启用' }}</span>
            </div>

            <!-- 平台列表 -->
            <div class="platform-list">
              <div
                v-for="p in platformList"
                :key="p.key"
                class="platform-item"
              >
                <div class="platform-info">
                  <span class="platform-name">{{ p.label }}</span>
                  <span class="platform-desc">{{ p.desc }}</span>
                </div>
                <label class="switch" :class="{ disabled: !settings.enabled }">
                  <input
                    type="checkbox"
                    :checked="getPlatformProxy(p.key)"
                    :disabled="!settings.enabled"
                    @change="togglePlatformProxy(p.key, $event)"
                  />
                  <span class="slider"></span>
                </label>
              </div>
            </div>
          </div>
        </div>

        <!-- 令牌管理 -->
        <div v-else-if="activeTab === 'token'" class="settings-panel">
          <div class="panel-header">
            <h2 class="panel-title">令牌管理</h2>
            <p class="panel-desc">管理 API 访问令牌，修改后当前会话自动更新</p>
          </div>

          <div class="settings-form">
            <!-- 当前 Token -->
            <div class="form-group">
              <label class="form-label">当前令牌</label>
              <div class="token-display">
                <code class="token-value">{{ currentToken || '加载中...' }}</code>
                <button
                  class="btn-icon"
                  title="复制令牌"
                  @click="copyToken"
                >
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

            <!-- 生成方式 -->
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

            <!-- 自动生成 -->
            <div v-if="tokenMode === 'auto'" class="form-group">
              <div class="token-actions">
                <button
                  class="btn btn-secondary"
                  :disabled="tokenLoading"
                  @click="regenerateToken"
                >
                  {{ tokenLoading ? '处理中...' : '重新生成' }}
                </button>
              </div>
            </div>

            <!-- 自定义输入 -->
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
        </div>

        <!-- 关于 -->
        <div v-else-if="activeTab === 'about'" class="settings-panel">
          <div class="panel-header">
            <h2 class="panel-title">关于</h2>
            <p class="panel-desc">Forks — Git 仓库管理工具</p>
          </div>

          <div class="settings-form">
            <div class="about-card">
              <div class="about-logo">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                  <circle cx="12" cy="12" r="3"/>
                  <path d="M12 1v2m0 18v2M4.22 4.22l1.42 1.42m12.72 12.72l1.42 1.42M1 12h2m18 0h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/>
                </svg>
              </div>
              <div class="about-info">
                <h3 class="about-name">Forks</h3>
                <p class="about-version">v{{ version }}</p>
              </div>
            </div>

            <div class="about-links">
              <a class="about-link-item" href="https://github.com/cicbyte/forks" target="_blank" rel="noopener noreferrer">
                <svg viewBox="0 0 24 24" fill="currentColor"><path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/></svg>
                <span>GitHub 仓库</span>
                <svg class="link-arrow" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M7 17L17 7M17 7H7M17 7v10"/></svg>
              </a>
            </div>

            <div class="about-desc">
              <p>Forks 是一个轻量级的 Git 仓库管理工具，提供 Web UI 来管理本地克隆的 GitHub / Gitee 仓库。支持批量克隆、拉取、文件浏览等功能。</p>
            </div>
          </div>
        </div>

        <!-- 其他设置占位 -->
        <div v-else class="settings-panel">
          <div class="panel-header">
            <h2 class="panel-title">{{ getCurrentNavLabel }}</h2>
            <p class="panel-desc">此功能开发中...</p>
          </div>
          <div class="empty-state">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <circle cx="12" cy="12" r="10"/>
              <path d="M12 8v4l3 3"/>
            </svg>
            <p>该设置项正在开发中，敬请期待</p>
          </div>
        </div>

        <!-- 底部操作按钮 -->
        <div v-if="activeTab !== 'about'" class="settings-footer">
          <div class="footer-info">
            <span class="info-text">设置自动保存到本地配置文件</span>
          </div>
          <div class="footer-actions">
            <button class="btn btn-secondary" @click="resetSettings">
              重置默认
            </button>
            <button class="btn btn-primary" @click="saveSettings">
              保存设置
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
/**
 * SettingsView - 系统设置页
 *
 * 功能特性:
 * - 左侧设置导航
 * - 代理设置
 * - 快捷代理预设
 * - 平台代理独立开关
 * - 保存和重置按钮
 */
import { ref, reactive, computed, onMounted } from 'vue'
import { useMessage } from 'naive-ui'
import {
  GitNetworkOutline,
  FlashOutline,
  GlobeOutline,
  RocketOutline,
  GitMergeOutline,
  KeyOutline,
  InformationCircleOutline
} from '@vicons/ionicons5'
import { getProxyConfig, updateProxyConfig, getTokenInfo, updateToken, getVersion } from '@/api/repos'
import { copyToClipboard } from '@/utils/clipboard'

const message = useMessage()

// 当前激活的 tab
const activeTab = ref('proxy')

// 导航项
const navItems = [
  {
    key: 'proxy',
    label: '代理',
    icon: GitNetworkOutline
  },
  {
    key: 'platform-proxy',
    label: '平台代理',
    icon: GitMergeOutline
  },
  {
    key: 'token',
    label: '令牌管理',
    icon: KeyOutline
  },
  {
    key: 'about',
    label: '关于',
    icon: InformationCircleOutline
  }
]

// 当前导航标签
const getCurrentNavLabel = computed(() => {
  const item = navItems.find(i => i.key === activeTab.value)
  return item?.label || '设置'
})

// 代理预设
const proxyPresets = [
  {
    name: 'clash',
    label: 'Clash',
    port: 7890,
    type: 'http',
    icon: FlashOutline
  },
  {
    name: 'shadowsocks',
    label: 'Shadowsocks',
    port: 1080,
    type: 'socks5',
    icon: GlobeOutline
  },
  {
    name: 'v2ray',
    label: 'V2Ray',
    port: 1080,
    type: 'http',
    icon: GitNetworkOutline
  },
  {
    name: 'ssrdog',
    label: 'SSRDOG',
    port: 7897,
    type: 'http',
    icon: RocketOutline
  }
]

// 设置数据
const settings = reactive({
  enabled: false,
  type: 'none',
  host: '127.0.0.1',
  port: 7890,
  no_proxy: 'localhost,127.0.0.1,0.0.0.0,::1,*.local',
  platforms: {} // {"github": true, "gitee": false}
})

// 平台列表
const platformList = [
  { key: 'github', label: 'GitHub', desc: 'github.com' },
  { key: 'gitee', label: 'Gitee', desc: 'gitee.com' }
]

// 加载代理配置
const loadProxySettings = async () => {
  try {
    const response = await getProxyConfig()
    const res = response.data || response
    if (res.code === 0 && res.data) {
      const config = res.data
      settings.enabled = config.enabled || false
      settings.type = config.type || 'none'
      settings.host = config.host || '127.0.0.1'
      settings.port = config.port || 7890
      settings.no_proxy = config.no_proxy || 'localhost,127.0.0.1,0.0.0.0,::1,*.local'
      settings.platforms = config.platforms || {}
    }
  } catch (error) {
    console.error('加载代理配置失败:', error)
  }
}

// 方法
const applyPreset = (preset) => {
  settings.enabled = true
  settings.type = preset.type
  settings.host = '127.0.0.1'
  settings.port = preset.port
  message.success(`已应用 ${preset.label} 代理配置，点击保存生效`)
}

const saveSettings = async () => {
  try {
    const config = {
      enabled: settings.enabled && settings.type !== 'none',
      type: settings.type,
      host: settings.host,
      port: settings.port,
      no_proxy: settings.no_proxy,
      platforms: { ...settings.platforms }
    }
    const response = await updateProxyConfig(config)
    const res = response.data || response
    if (res.code === 0) {
      message.success('代理配置已保存并生效')
    } else {
      message.error(res.message || '保存失败')
    }
  } catch (error) {
    message.error('保存失败：' + error.message)
  }
}

const resetSettings = () => {
  settings.enabled = false
  settings.type = 'none'
  settings.host = '127.0.0.1'
  settings.port = 7890
  settings.no_proxy = 'localhost,127.0.0.1,0.0.0.0,::1,*.local'
  settings.platforms = {}
  message.success('设置已重置，点击保存生效')
}

// 获取平台代理状态（未设置时跟随全局开关）
const getPlatformProxy = (platformKey) => {
  if (platformKey in settings.platforms) {
    return settings.platforms[platformKey]
  }
  return settings.enabled
}

// 切换平台代理开关
const togglePlatformProxy = (platformKey, event) => {
  const checked = event.target.checked
  settings.platforms[platformKey] = checked
}

// ============ Token 管理 ============
const currentToken = ref('')
const plainToken = ref('') // 完整 token，用于复制
const tokenMode = ref('auto')
const customTokenInput = ref('')
const tokenLoading = ref(false)
const copied = ref(false)

// 加载 Token 信息
const loadTokenInfo = async () => {
  try {
    const response = await getTokenInfo()
    const res = response.data || response
    if (res.code === 0 && res.data) {
      currentToken.value = res.data.token
      plainToken.value = '' // 脱敏接口无法获取完整 token
    }
  } catch (error) {
    console.error('加载 Token 信息失败:', error)
  }
}

// 复制 Token
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

// 更新 token 后同步本地状态
const onTokenUpdated = (newToken) => {
  localStorage.setItem('token', newToken)
  plainToken.value = newToken
  currentToken.value = newToken.substring(0, 4) + '****' + newToken.substring(newToken.length - 4)
  customTokenInput.value = ''
}

// 重新生成 Token
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

// 应用自定义 Token
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

// ============ 关于 ============
const version = ref('...')

const loadVersion = async () => {
  try {
    const response = await getVersion()
    const res = response.data || response
    if (res.code === 0 && res.data) {
      version.value = res.data.version
    }
  } catch (error) {
    console.error('加载版本号失败:', error)
  }
}

// 页面加载时获取配置
onMounted(() => {
  loadProxySettings()
  loadTokenInfo()
  loadVersion()
})
</script>

<style scoped>
/* ============================================
   PAGE HEADER
   ============================================ */

.settings-view {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
  animation: fadeIn 0.3s ease-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.page-header {
  margin-bottom: var(--space-2);
}

.page-title {
  font-size: 28px;
  font-weight: var(--font-bold);
  color: var(--color-text-primary);
  margin: 0 0 var(--space-1) 0;
}

.page-description {
  font-size: var(--text-base);
  color: var(--color-text-secondary);
  margin: 0;
}

/* ============================================
   SETTINGS LAYOUT
   ============================================ */

.settings-layout {
  display: grid;
  grid-template-columns: 200px 1fr;
  gap: var(--space-6);
  align-items: start;
}

@media (max-width: 768px) {
  .settings-layout {
    grid-template-columns: 1fr;
  }

  .settings-nav {
    position: static;
    width: 100%;
  }

  .nav-list {
    flex-direction: row;
    flex-wrap: wrap;
  }
}

/* ============================================
   SETTINGS NAV
   ============================================ */

.settings-nav {
  position: sticky;
  top: calc(var(--navbar-height) + var(--space-6));
  background-color: var(--color-bg-card);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg);
  padding: var(--space-4);
  align-self: start;
}

.nav-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.nav-section-title {
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
  color: var(--color-text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  padding: 0 var(--space-2);
}

.nav-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.nav-item {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-2_5) var(--space-3);
  border-radius: var(--radius-md);
  border: none;
  background: none;
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: all 0.2s ease;
  text-align: left;
  width: 100%;
}

.nav-item:hover {
  background-color: var(--color-gray-100);
  color: var(--color-text-primary);
}

.nav-item.active {
  background-color: var(--color-primary-50);
  color: var(--color-primary);
  font-weight: var(--font-semibold);
}

.nav-icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.nav-label {
  font-size: var(--text-sm);
}

/* ============================================
   SETTINGS CONTENT
   ============================================ */

.settings-content {
  background-color: var(--color-bg-card);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg);
  min-height: 400px;
}

.settings-panel {
  padding: var(--space-6);
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

/* ============================================
   EMPTY STATE
   ============================================ */

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-12);
  color: var(--color-text-tertiary);
}

.empty-state svg {
  width: 48px;
  height: 48px;
  margin-bottom: var(--space-4);
  opacity: 0.5;
}

.empty-state p {
  font-size: var(--text-sm);
}

/* ============================================
   FORM ELEMENTS
   ============================================ */

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

/* ============================================
   PLATFORM PROXY
   ============================================ */

.proxy-status-hint {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-3) var(--space-4);
  border-radius: var(--radius-md);
  font-size: var(--text-sm);
  color: var(--color-primary);
  background-color: var(--color-primary-50);
}

.proxy-status-hint.disabled {
  color: var(--color-text-tertiary);
  background-color: var(--color-gray-100);
}

.proxy-status-hint .status-icon {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
}

.platform-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  max-width: 400px;
}

.platform-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-3) var(--space-4);
  background: var(--color-bg-page);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
}

.platform-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.platform-name {
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  color: var(--color-text-primary);
}

.platform-desc {
  font-size: var(--text-xs);
  color: var(--color-text-tertiary);
}

.switch {
  position: relative;
  display: inline-block;
  width: 44px;
  height: 24px;
  flex-shrink: 0;
}

.switch.disabled {
  opacity: 0.5;
  pointer-events: none;
}

.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: var(--color-gray-300);
  transition: 0.2s;
  border-radius: 24px;
}

.slider:before {
  position: absolute;
  content: '';
  height: 18px;
  width: 18px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: 0.2s;
  border-radius: 50%;
}

.switch input:checked + .slider {
  background-color: var(--color-primary);
}

.switch input:checked + .slider:before {
  transform: translateX(20px);
}

/* ============================================
   TOKEN MANAGEMENT
   ============================================ */

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

.form-hint {
  font-size: var(--text-xs);
  color: var(--color-text-tertiary);
}

/* ============================================
   PRESET GRID
   ============================================ */

.preset-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--space-3);
  max-width: 400px;
}

@media (max-width: 640px) {
  .preset-grid {
    grid-template-columns: 1fr;
  }
}

.preset-btn {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all 0.2s ease;
  background: var(--color-bg-page);
}

.preset-btn:hover {
  border-color: var(--color-primary);
  background-color: var(--color-primary-50);
}

.preset-icon {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: var(--color-primary-50);
  color: var(--color-primary);
  border-radius: var(--radius-md);
  flex-shrink: 0;
}

.preset-icon svg {
  width: 18px;
  height: 18px;
}

.preset-info {
  flex: 1;
  text-align: left;
}

.preset-name {
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  color: var(--color-text-primary);
}

.preset-desc {
  font-size: var(--text-xs);
  color: var(--color-text-secondary);
}

/* ============================================
   SETTINGS FOOTER
   ============================================ */

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

.footer-actions {
  display: flex;
  gap: var(--space-3);
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

/* ============================================
   ABOUT
   ============================================ */

.about-card {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  padding: var(--space-5);
  background: var(--color-bg-page);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
}

.about-logo {
  width: 56px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: var(--color-primary-50);
  color: var(--color-primary);
  border-radius: var(--radius-lg);
  flex-shrink: 0;
}

.about-logo svg {
  width: 28px;
  height: 28px;
}

.about-name {
  font-size: var(--text-lg);
  font-weight: var(--font-bold);
  color: var(--color-text-primary);
  margin: 0 0 var(--space-1) 0;
}

.about-version {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  margin: 0;
  font-family: var(--font-mono, 'Courier New', Courier, monospace);
}

.about-links {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.about-link-item {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-4);
  background: var(--color-bg-page);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  text-decoration: none;
  color: var(--color-text-primary);
  font-size: var(--text-sm);
  transition: all 0.2s ease;
}

.about-link-item:hover {
  border-color: var(--color-primary);
  background-color: var(--color-primary-50);
}

.about-link-item svg:first-child {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
  color: var(--color-text-primary);
}

.about-link-item span {
  flex: 1;
}

.link-arrow {
  width: 16px !important;
  height: 16px !important;
  color: var(--color-text-tertiary) !important;
}

.about-desc p {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  line-height: 1.7;
  margin: 0;
}
</style>
