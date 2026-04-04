<script setup>
import { computed, onMounted } from 'vue'
import {
  NMessageProvider,
  NNotificationProvider,
  NConfigProvider,
  NModalProvider,
  NDialogProvider
} from 'naive-ui'
import { getThemeOverrides } from './composables/useTheme'

// 导入 Design Tokens 样式
import './styles/index.css'

// 固定使用浅色主题
const themeOverrides = computed(() => {
  return getThemeOverrides('light')
})
</script>

<template>
  <NConfigProvider :theme-overrides="themeOverrides">
    <NNotificationProvider>
      <NMessageProvider>
        <NDialogProvider>
          <NModalProvider>
            <router-view />
          </NModalProvider>
        </NDialogProvider>
      </NMessageProvider>
    </NNotificationProvider>
  </NConfigProvider>
</template>

<style>
/* 全局样式覆盖 */
#app {
  min-height: 100vh;
  background-color: var(--color-bg-body);
  color: var(--color-text-primary);
}
</style>
