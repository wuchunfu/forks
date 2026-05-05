<template>
  <div class="settings-panel">
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
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getVersion } from '@/api/repos'

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

onMounted(() => {
  loadVersion()
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
