<template>
  <div class="settings-panel">
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

    <!-- 底部操作按钮 -->
    <div class="settings-footer">
      <div class="footer-info">
        <span class="info-text">设置自动保存到本地配置文件</span>
      </div>
      <div class="footer-actions">
        <button class="btn btn-secondary" @click="resetSettings">重置默认</button>
        <button class="btn btn-primary" @click="saveSettings">保存设置</button>
      </div>
    </div>

    <!-- 平台代理 -->
    <div class="panel-header" style="margin-top: var(--space-8)">
      <h2 class="panel-title">平台代理</h2>
      <p class="panel-desc">单独控制每个平台是否走代理，未设置时跟随全局开关</p>
    </div>

    <div class="settings-form">
      <div class="proxy-status-hint" :class="{ disabled: !settings.enabled }">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="status-icon">
          <circle cx="12" cy="12" r="10"/>
          <path v-if="settings.enabled" d="M12 8v4l3 3"/>
          <path v-else d="M4.93 4.93l14.14 14.14"/>
        </svg>
        <span>全局代理：{{ settings.enabled ? '已启用' : '未启用' }}</span>
      </div>

      <div class="platform-list">
        <div v-for="p in platformList" :key="p.key" class="platform-item">
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
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useMessage } from 'naive-ui'
import {
  GitNetworkOutline,
  FlashOutline,
  GlobeOutline,
  RocketOutline,
} from '@vicons/ionicons5'
import { getProxyConfig, updateProxyConfig } from '@/api/repos'

const message = useMessage()

const proxyPresets = [
  { name: 'clash', label: 'Clash', port: 7890, type: 'http', icon: FlashOutline },
  { name: 'shadowsocks', label: 'Shadowsocks', port: 1080, type: 'socks5', icon: GlobeOutline },
  { name: 'v2ray', label: 'V2Ray', port: 1080, type: 'http', icon: GitNetworkOutline },
  { name: 'ssrdog', label: 'SSRDOG', port: 7897, type: 'http', icon: RocketOutline }
]

const settings = reactive({
  enabled: false,
  type: 'none',
  host: '127.0.0.1',
  port: 7890,
  no_proxy: 'localhost,127.0.0.1,0.0.0.0,::1,*.local',
  platforms: {}
})

const platformList = [
  { key: 'github', label: 'GitHub', desc: 'github.com' },
  { key: 'gitee', label: 'Gitee', desc: 'gitee.com' }
]

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

const getPlatformProxy = (platformKey) => {
  if (platformKey in settings.platforms) {
    return settings.platforms[platformKey]
  }
  return settings.enabled
}

const togglePlatformProxy = (platformKey, event) => {
  settings.platforms[platformKey] = event.target.checked
}

onMounted(() => {
  loadProxySettings()
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
  background-color: var(--color-bg-card);
  transition: 0.2s;
  border-radius: 50%;
}

.switch input:checked + .slider {
  background-color: var(--color-primary);
}

.switch input:checked + .slider:before {
  transform: translateX(20px);
}
</style>
