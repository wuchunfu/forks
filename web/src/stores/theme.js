/**
 * Forks Theme Store
 * Grid Design System - Chromatic Team
 *
 * 主题状态管理
 * 版本: 1.0.0
 */

import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export const useThemeStore = defineStore('theme', () => {
  // 主题模式: 'light' | 'dark' | 'auto'
  const mode = ref(localStorage.getItem('forks-theme') || 'light')

  // 监听主题变化并持久化
  watch(mode, (newMode) => {
    localStorage.setItem('forks-theme', newMode)
  })

  /**
   * 设置主题模式
   * @param {string} newMode - 新的主题模式
   */
  function setTheme(newMode) {
    if (['light', 'dark', 'auto'].includes(newMode)) {
      mode.value = newMode
    }
  }

  /**
   * 切换主题
   */
  function toggleTheme() {
    const currentTheme = getCurrentTheme()
    mode.value = currentTheme === 'dark' ? 'light' : 'dark'
  }

  /**
   * 获取当前实际主题
   * @returns {string} 'light' | 'dark'
   */
  function getCurrentTheme() {
    if (mode.value === 'auto') {
      return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
    }
    return mode.value
  }

  /**
   * 初始化主题
   */
  function initTheme() {
    const actualTheme = getCurrentTheme()
    if (typeof document !== 'undefined') {
      document.documentElement.setAttribute('data-theme', actualTheme)
    }
  }

  return {
    mode,
    setTheme,
    toggleTheme,
    getCurrentTheme,
    initTheme,
  }
})
