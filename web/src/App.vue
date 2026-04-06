<script setup>
import { computed, onMounted, toRef } from 'vue'
import {
  NMessageProvider,
  NNotificationProvider,
  NConfigProvider,
  NModalProvider,
  NDialogProvider
} from 'naive-ui'
import { useThemeStore } from '@/stores/theme'
import { createThemeConfig, getThemeOverrides } from './composables/useTheme'

// 导入 Design Tokens 样式
import './styles/index.css'

const themeStore = useThemeStore()
const { naiveTheme, actualTheme } = createThemeConfig(toRef(themeStore, 'mode'))

// Naive UI 主题 + 组件覆盖
const themeOverrides = computed(() => getThemeOverrides(actualTheme.value))

onMounted(() => {
  themeStore.initTheme()
})
</script>

<template>
  <NConfigProvider :theme="naiveTheme" :theme-overrides="themeOverrides">
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
