<template>
  <div class="settings-view">
    <nav class="settings-nav">
      <router-link
        v-for="item in navItems"
        :key="item.path"
        :to="item.path"
        class="nav-item"
        :class="{ active: isActive(item.path) }"
      >
        <svg v-if="item.icon === 'proxy'" class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <circle cx="12" cy="12" r="10"/><path d="M2 12h20M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z"/>
        </svg>
        <svg v-else-if="item.icon === 'key'" class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 11-7.778 7.778 5.5 5.5 0 017.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4"/>
        </svg>
        <svg v-else-if="item.icon === 'trending'" class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <polyline points="23 6 13.5 15.5 8.5 10.5 1 18"/><polyline points="17 6 23 6 23 12"/>
        </svg>
        <svg v-else-if="item.icon === 'mcp'" class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <rect x="2" y="3" width="20" height="14" rx="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/>
        </svg>
        <svg v-else-if="item.icon === 'info'" class="nav-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/>
        </svg>
        <span class="nav-label">{{ item.label }}</span>
      </router-link>
    </nav>
    <div class="settings-content">
      <router-view />
    </div>
  </div>
</template>

<script setup>
import { useRoute } from 'vue-router'

const route = useRoute()

const navItems = [
  { path: '/settings/proxy', label: '代理', icon: 'proxy' },
  { path: '/settings/token', label: '令牌管理', icon: 'key' },
  { path: '/settings/trending', label: '趋势同步', icon: 'trending' },
  { path: '/settings/mcp', label: 'MCP', icon: 'mcp' },
  { path: '/settings/about', label: '关于', icon: 'info' },
]

function isActive(path) {
  return route.path === path
}
</script>

<style scoped>
.settings-view {
  display: flex;
  gap: var(--space-6);
  animation: fadeIn 0.3s ease-out;
  height: calc(100vh - var(--navbar-height) - var(--space-6) * 2);
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.settings-nav {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex-shrink: 0;
  width: 160px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: var(--space-2_5);
  padding: var(--space-2_5) var(--space-3);
  border-radius: var(--radius-md);
  text-decoration: none;
  color: var(--color-text-secondary);
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  transition: all 0.15s ease;
  cursor: pointer;
}

.nav-item:hover {
  color: var(--color-text-primary);
  background-color: var(--color-gray-50);
}

.nav-item.active {
  color: var(--color-primary);
  background-color: var(--color-primary-50);
}

.nav-icon {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
}

.nav-label {
  white-space: nowrap;
}

.settings-content {
  flex: 1;
  min-width: 0;
  background: var(--color-bg-card);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg);
  overflow: hidden;
}
</style>
