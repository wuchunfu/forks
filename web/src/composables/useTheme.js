/**
 * Forks Naive UI Theme Configuration
 * Grid Design System - Chromatic Team
 *
 * Naive UI 主题覆盖配置
 * 版本: 1.0.0
 */

import { computed, watch } from 'vue'
import { darkTheme } from 'naive-ui'

/**
 * 创建主题配置
 * @param {Ref<string>} themeMode - 主题模式 ('light' | 'dark' | 'auto')
 * @returns {Object} 主题配置对象
 */
export function createThemeConfig(themeMode) {
  // 获取实际主题
  const actualTheme = computed(() => {
    if (themeMode.value === 'auto') {
      return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
    }
    return themeMode.value
  })

  // Naive UI 主题对象
  const naiveTheme = computed(() => {
    return actualTheme.value === 'dark' ? darkTheme : null
  })

  // 更新 HTML data-theme 属性
  const updateDocumentTheme = (theme) => {
    if (typeof document !== 'undefined') {
      document.documentElement.setAttribute('data-theme', theme)
    }
  }

  // 监听主题变化
  watch(actualTheme, (newTheme) => {
    updateDocumentTheme(newTheme)
  }, { immediate: true })

  // 监听系统主题变化
  if (themeMode.value === 'auto' && typeof window !== 'undefined') {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    const handleThemeChange = () => {
      updateDocumentTheme(mediaQuery.matches ? 'dark' : 'light')
    }
    mediaQuery.addEventListener('change', handleThemeChange)

    // 返回清理函数
    return () => {
      mediaQuery.removeEventListener('change', handleThemeChange)
    }
  }

  return {
    naiveTheme,
    actualTheme
  }
}

/**
 * 亮色主题配置
 */
export const lightThemeOverrides = {
  common: {
    primaryColor: '#2563eb',
    primaryColorHover: '#1d4ed8',
    primaryColorPressed: '#1e40af',
    primaryColorSuppl: '#3b82f6',

    successColor: '#10b981',
    successColorHover: '#059669',
    successColorPressed: '#047857',
    successColorSuppl: '#34d399',

    warningColor: '#f59e0b',
    warningColorHover: '#d97706',
    warningColorPressed: '#b45309',
    warningColorSuppl: '#fbbf24',

    errorColor: '#ef4444',
    errorColorHover: '#dc2626',
    errorColorPressed: '#b91c1c',
    errorColorSuppl: '#f87171',

    infoColor: '#3b82f6',
    infoColorHover: '#2563eb',
    infoColorPressed: '#1d4ed8',
    infoColorSuppl: '#60a5fa',

    textColorBase: '#111827',
    textColor1: '#111827',
    textColor2: '#6b7280',
    textColor3: '#9ca3af',
    textColorDisabled: '#d1d5db',

    textColorFocus: '#2563eb',

    opacity1Disabled: 0.5,
    opacity2Disabled: 0.5,

    dividerColor: '#e5e7eb',
    borderColor: '#e5e7eb',

    borderRadius: '8px',
    borderRadiusSmall: '4px',

    fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
    fontFamilyMono: '"JetBrains Mono", "Fira Code", Consolas, monospace',

    boxShadow1: '0 1px 2px 0 rgb(0 0 0 / 0.05)',
    boxShadow2: '0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)',
    boxShadow3: '0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)',
  },

  // Layout
  Layout: {
    headerColor: '#ffffff',
    footerColor: '#ffffff',
    siderColor: '#ffffff',
    color: '#f9fafb',
  },

  // Card
  Card: {
    color: '#ffffff',
    colorModal: '#ffffff',
    colorTarget: '#ffffff',
    colorEmbeddedModal: '#ffffff',
    borderColor: '#e5e7eb',
  },

  // Button
  Button: {
    textColor: '#111827',
    textColorHover: '#111827',
    textColorPressed: '#111827',
    textColorFocus: '#111827',
    textColorDisabled: '#9ca3af',

    color: '#ffffff',
    colorHover: '#f3f4f6',
    colorPressed: '#e5e7eb',
    colorFocus: '#f3f4f6',
    colorDisabled: '#f3f4f6',

    borderColor: '#d1d5db',
    borderColorHover: '#9ca3af',
    borderColorPressed: '#6b7280',
    borderColorFocus: '#2563eb',
    borderColorDisabled: '#e5e7eb',

    colorPrimary: '#2563eb',
    colorPrimaryHover: '#1d4ed8',
    colorPrimaryPressed: '#1e40af',
    colorPrimaryFocus: '#1d4ed8',

    textColorPrimary: '#ffffff',
    textColorPrimaryHover: '#ffffff',
    textColorPrimaryPressed: '#ffffff',
    textColorPrimaryFocus: '#ffffff',

    textColorText: '#111827',
    textColorTextHover: '#111827',
    textColorTextPressed: '#111827',
    textColorTextFocus: '#111827',

    textColorGhost: '#111827',
    textColorGhostHover: '#111827',
    textColorGhostPressed: '#111827',
    textColorGhostFocus: '#111827',

    colorGhost: '#ffffff',
    colorGhostHover: '#f3f4f6',
    colorGhostPressed: '#e5e7eb',
    colorGhostFocus: '#f3f4f6',

    borderRadius: '8px',
    paddingSmall: '0 12px',
    paddingMedium: '0 16px',
    paddingLarge: '0 24px',

    heightSmall: '32px',
    heightMedium: '40px',
    heightLarge: '48px',

    fontSizeSmall: '14px',
    fontSizeMedium: '14px',
    fontSizeLarge: '16px',
  },

  // Input
  Input: {
    color: '#ffffff',
    colorFocus: '#ffffff',
    textColor: '#111827',
    caretColor: '#2563eb',
    border: '1px solid #d1d5db',
    borderHover: '1px solid #9ca3af',
    borderFocus: '1px solid #2563eb',
    placeholderColor: '#9ca3af',
    colorDisabled: '#f3f4f6',
    textColorDisabled: '#d1d5db',
    borderDisabled: '1px solid #e5e7eb',
    boxShadowFocus: '0 0 0 2px rgba(37, 99, 235, 0.1)',
    borderRadius: '8px',
    heightSmall: '32px',
    heightMedium: '40px',
    heightLarge: '48px',
  },

  // Select
  Select: {
    peers: {
      InternalSelection: {
        color: '#ffffff',
        colorActive: '#f3f4f6',
        textColor: '#111827',
        placeholderColor: '#9ca3af',
        caretColor: '#2563eb',
        border: '1px solid #d1d5db',
        borderHover: '1px solid #9ca3af',
        borderActive: '1px solid #2563eb',
        borderFocus: '1px solid #2563eb',
        boxShadowFocus: '0 0 0 2px rgba(37, 99, 235, 0.1)',
        borderRadius: '8px',
        heightSmall: '32px',
        heightMedium: '40px',
        heightLarge: '48px',
      }
    }
  },

  // Modal
  Modal: {
    color: '#ffffff',
    textColor: '#111827',
    boxShadow: '0 25px 50px -12px rgb(0 0 0 / 0.25)',
    borderRadius: '12px',
  },

  // Dropdown
  Dropdown: {
    color: '#ffffff',
    optionColorHover: '#f3f4f6',
    optionTextColorActive: '#2563eb',
    optionColorActive: '#eff6ff',
    dividerColor: '#e5e7eb',
    borderRadius: '8px',
    boxShadow: '0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)',
  },

  // Tooltip
  Tooltip: {
    color: '#1f2937',
    textColor: '#ffffff',
    borderRadius: '6px',
    padding: '8px 12px',
    fontSize: '14px',
  },

  // Table
  DataTable: {
    thColor: '#f9fafb',
    thTextColor: '#111827',
    tdColor: '#ffffff',
    tdTextColor: '#111827',
    tdColorHover: '#f3f4f6',
    tdColorStriped: '#f9fafb',
    borderColor: '#e5e7eb',
    borderRadius: '8px',
    fontSizeSmall: '12px',
    fontSizeMedium: '14px',
    fontSizeLarge: '16px',
  },

  // Tabs
  Tabs: {
    tabTextColorBar: '#6b7280',
    tabTextColorActiveBar: '#2563eb',
    tabTextColorHoverBar: '#111827',
    barColor: 'transparent',
    barColorHover: 'transparent',
    panePaddingSmall: '12px 0',
    panePaddingMedium: '16px 0',
    panePaddingLarge: '24px 0',
    gapSmall: '24px',
    gapMedium: '28px',
    gapLarge: '32px',
  },

  // Menu
  Menu: {
    itemTextColor: '#374151',
    itemTextColorHover: '#111827',
    itemTextColorActive: '#2563eb',
    itemIconColor: '#6b7280',
    itemIconColorHover: '#111827',
    itemIconColorActive: '#2563eb',
    itemColorActive: '#eff6ff',
    itemColorHover: '#f3f4f6',
    arrowColor: '#9ca3af',
    arrowColorHover: '#111827',
    arrowColorActive: '#2563eb',
    borderRadius: '6px',
  },

  // Tag
  Tag: {
    borderRadius: '4px',
    padding: '0 8px',
    heightSmall: '20px',
    heightMedium: '24px',
    heightLarge: '28px',
    fontSizeSmall: '12px',
    fontSizeMedium: '12px',
    fontSizeLarge: '14px',
  },

  // Message
  Message: {
    borderRadius: '8px',
    padding: '12px 16px',
    fontSize: '14px',
  },

  // Notification
  Notification: {
    borderRadius: '8px',
    padding: '16px 20px',
    fontSize: '14px',
    boxShadow: '0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)',
  },

  // Switch
  Switch: {
    borderRadius: '12px',
    railColor: '#e5e7eb',
    railColorActive: '#2563eb',
    buttonColor: '#ffffff',
    boxShadowFocus: '0 0 0 2px rgba(37, 99, 235, 0.1)',
  },

  // Checkbox
  Checkbox: {
    borderRadius: '4px',
    borderColor: '#d1d5db',
    borderColorChecked: '#2563eb',
    checkMarkColor: '#ffffff',
    boxShadowFocus: '0 0 0 2px rgba(37, 99, 235, 0.1)',
  },

  // Radio
  Radio: {
    boxShadowFocus: '0 0 0 2px rgba(37, 99, 235, 0.1)',
    dotColorActive: '#ffffff',
    colorActive: '#2563eb',
    boxShadowActive: '0 0 0 1px #2563eb',
  },

  // Slider
  Slider: {
    railColor: '#e5e7eb',
    railColorHover: '#d1d5db',
    fillColor: '#2563eb',
    fillColorHover: '#1d4ed8',
    handleColor: '#ffffff',
    dotColor: '#ffffff',
    dotBorderColor: '#2563eb',
  },

  // Progress
  Progress: {
    railColor: '#e5e7eb',
    fillColor: '#2563eb',
    textColor: '#111827',
    fontSizeCircle: '24px',
  },

  // Alert
  Alert: {
    borderRadius: '8px',
    padding: '12px 16px',
    fontSize: '14px',
  },

  // Popover
  Popover: {
    color: '#ffffff',
    textColor: '#111827',
    borderRadius: '8px',
    boxShadow: '0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)',
    padding: '12px 16px',
  },

  // Drawer
  Drawer: {
    color: '#ffffff',
    textColor: '#111827',
    boxShadow: '0 25px 50px -12px rgb(0 0 0 / 0.25)',
  },

  // Pagination
  Pagination: {
    itemColor: '#ffffff',
    itemColorHover: '#f3f4f6',
    itemColorPressed: '#e5e7eb',
    itemColorActive: '#2563eb',
    itemTextColor: '#111827',
    itemTextColorHover: '#111827',
    itemTextColorPressed: '#111827',
    itemTextColorActive: '#ffffff',
    itemBorderColor: '#e5e7eb',
    itemBorderRadius: '6px',
  },

  // Breadcrumb
  Breadcrumb: {
    itemTextColor: '#6b7280',
    itemTextColorHover: '#2563eb',
    itemTextColorActive: '#111827',
    fontSize: '14px',
  },

  // Steps
  Steps: {
    stepTextColor: '#6b7280',
    stepTextColorActive: '#2563eb',
    stepTextColorFinished: '#111827',
  },

  // Timeline
  Timeline: {
    titleTextColor: '#111827',
    contentTextColor: '#6b7280',
    lineColor: '#e5e7eb',
  },

  // Skeleton
  Skeleton: {
    color: '#f3f4f6',
    colorEnd: '#e5e7eb',
    borderRadius: '6px',
  },

  // Spin
  Spin: {
    sizeSmall: '18px',
    sizeMedium: '24px',
    sizeLarge: '36px',
  },
}

/**
 * 暗色主题配置
 */
export const darkThemeOverrides = {
  common: {
    primaryColor: '#3b82f6',
    primaryColorHover: '#60a5fa',
    primaryColorPressed: '#2563eb',
    primaryColorSuppl: '#1d4ed8',

    successColor: '#10b981',
    successColorHover: '#34d399',
    successColorPressed: '#059669',
    successColorSuppl: '#6ee7b7',

    warningColor: '#f59e0b',
    warningColorHover: '#fbbf24',
    warningColorPressed: '#d97706',
    warningColorSuppl: '#fcd34d',

    errorColor: '#ef4444',
    errorColorHover: '#f87171',
    errorColorPressed: '#dc2626',
    errorColorSuppl: '#fca5a5',

    infoColor: '#3b82f6',
    infoColorHover: '#60a5fa',
    infoColorPressed: '#2563eb',
    infoColorSuppl: '#93c5fd',

    textColorBase: '#f9fafb',
    textColor1: '#f9fafb',
    textColor2: '#d1d5db',
    textColor3: '#9ca3af',
    textColorDisabled: '#4b5563',

    textColorFocus: '#3b82f6',

    dividerColor: '#374151',
    borderColor: '#374151',

    fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
    fontFamilyMono: '"JetBrains Mono", "Fira Code", Consolas, monospace',

    boxShadow1: '0 1px 2px 0 rgb(0 0 0 / 0.5)',
    boxShadow2: '0 4px 6px -1px rgb(0 0 0 / 0.5), 0 2px 4px -2px rgb(0 0 0 / 0.5)',
    boxShadow3: '0 10px 15px -3px rgb(0 0 0 / 0.5), 0 4px 6px -4px rgb(0 0 0 / 0.5)',
  },

  // Layout
  Layout: {
    headerColor: '#111827',
    footerColor: '#111827',
    siderColor: '#111827',
    color: '#030712',
  },

  // Card
  Card: {
    color: '#1f2937',
    colorModal: '#1f2937',
    colorTarget: '#1f2937',
    colorEmbeddedModal: '#111827',
    borderColor: '#374151',
  },

  // Button
  Button: {
    textColor: '#f9fafb',
    textColorHover: '#f9fafb',
    textColorPressed: '#f9fafb',
    textColorFocus: '#f9fafb',
    textColorDisabled: '#4b5563',

    color: '#1f2937',
    colorHover: '#374151',
    colorPressed: '#4b5563',
    colorFocus: '#374151',
    colorDisabled: '#111827',

    borderColor: '#4b5563',
    borderColorHover: '#6b7280',
    borderColorPressed: '#9ca3af',
    borderColorFocus: '#3b82f6',
    borderColorDisabled: '#374151',

    colorPrimary: '#3b82f6',
    colorPrimaryHover: '#60a5fa',
    colorPrimaryPressed: '#2563eb',
    colorPrimaryFocus: '#60a5fa',

    textColorPrimary: '#ffffff',
    textColorPrimaryHover: '#ffffff',
    textColorPrimaryPressed: '#ffffff',
    textColorPrimaryFocus: '#ffffff',

    textColorText: '#f9fafb',
    textColorTextHover: '#f9fafb',
    textColorTextPressed: '#f9fafb',
    textColorTextFocus: '#f9fafb',

    textColorGhost: '#f9fafb',
    textColorGhostHover: '#f9fafb',
    textColorGhostPressed: '#f9fafb',
    textColorGhostFocus: '#f9fafb',

    colorGhost: '#1f2937',
    colorGhostHover: '#374151',
    colorGhostPressed: '#4b5563',
    colorGhostFocus: '#374151',
  },

  // Input
  Input: {
    color: '#1f2937',
    colorFocus: '#1f2937',
    textColor: '#f9fafb',
    caretColor: '#3b82f6',
    border: '1px solid #4b5563',
    borderHover: '1px solid #6b7280',
    borderFocus: '1px solid #3b82f6',
    placeholderColor: '#6b7280',
    colorDisabled: '#111827',
    textColorDisabled: '#4b5563',
    borderDisabled: '1px solid #374151',
    boxShadowFocus: '0 0 0 2px rgba(59, 130, 246, 0.2)',
  },

  // Select
  Select: {
    peers: {
      InternalSelection: {
        color: '#1f2937',
        colorActive: '#374151',
        textColor: '#f9fafb',
        placeholderColor: '#6b7280',
        caretColor: '#3b82f6',
        border: '1px solid #4b5563',
        borderHover: '1px solid #6b7280',
        borderActive: '1px solid #3b82f6',
        borderFocus: '1px solid #3b82f6',
        boxShadowFocus: '0 0 0 2px rgba(59, 130, 246, 0.2)',
      }
    }
  },

  // Modal
  Modal: {
    color: '#1f2937',
    textColor: '#f9fafb',
    boxShadow: '0 25px 50px -12px rgb(0 0 0 / 0.75)',
  },

  // Dropdown
  Dropdown: {
    color: '#1f2937',
    optionColorHover: '#374151',
    optionTextColorActive: '#3b82f6',
    optionColorActive: '#1e3a8a',
    dividerColor: '#374151',
    boxShadow: '0 10px 15px -3px rgb(0 0 0 / 0.5), 0 4px 6px -4px rgb(0 0 0 / 0.5)',
  },

  // Tooltip
  Tooltip: {
    color: '#f3f4f6',
    textColor: '#111827',
  },

  // Table
  DataTable: {
    thColor: '#1f2937',
    thTextColor: '#f9fafb',
    tdColor: '#1f2937',
    tdTextColor: '#f9fafb',
    tdColorHover: '#374151',
    tdColorStriped: '#111827',
    borderColor: '#374151',
  },

  // Menu
  Menu: {
    itemTextColor: '#d1d5db',
    itemTextColorHover: '#f9fafb',
    itemTextColorActive: '#3b82f6',
    itemIconColor: '#9ca3af',
    itemIconColorHover: '#f9fafb',
    itemIconColorActive: '#3b82f6',
    itemColorActive: '#1e3a8a',
    itemColorHover: '#374151',
    arrowColor: '#6b7280',
    arrowColorHover: '#f9fafb',
    arrowColorActive: '#3b82f6',
  },

  // Tag
  Tag: {
    borderRadius: '4px',
    padding: '0 8px',
    heightSmall: '20px',
    heightMedium: '24px',
    heightLarge: '28px',
    fontSizeSmall: '12px',
    fontSizeMedium: '12px',
    fontSizeLarge: '14px',
  },

  // Message
  Message: {
    borderRadius: '8px',
    padding: '12px 16px',
    fontSize: '14px',
  },

  // Notification
  Notification: {
    borderRadius: '8px',
    padding: '16px 20px',
    fontSize: '14px',
    boxShadow: '0 10px 15px -3px rgb(0 0 0 / 0.5), 0 4px 6px -4px rgb(0 0 0 / 0.5)',
  },

  // Switch
  Switch: {
    borderRadius: '12px',
    railColor: '#4b5563',
    railColorActive: '#3b82f6',
    buttonColor: '#1f2937',
    boxShadowFocus: '0 0 0 2px rgba(59, 130, 246, 0.2)',
  },

  // Checkbox
  Checkbox: {
    borderRadius: '4px',
    borderColor: '#4b5563',
    borderColorChecked: '#3b82f6',
    checkMarkColor: '#ffffff',
    boxShadowFocus: '0 0 0 2px rgba(59, 130, 246, 0.2)',
  },

  // Radio
  Radio: {
    boxShadowFocus: '0 0 0 2px rgba(59, 130, 246, 0.2)',
    dotColorActive: '#1f2937',
    colorActive: '#3b82f6',
    boxShadowActive: '0 0 0 1px #3b82f6',
  },

  // Slider
  Slider: {
    railColor: '#4b5563',
    railColorHover: '#6b7280',
    fillColor: '#3b82f6',
    fillColorHover: '#60a5fa',
    handleColor: '#1f2937',
    dotColor: '#1f2937',
    dotBorderColor: '#3b82f6',
  },

  // Progress
  Progress: {
    railColor: '#4b5563',
    fillColor: '#3b82f6',
    textColor: '#f9fafb',
    fontSizeCircle: '24px',
  },

  // Alert
  Alert: {
    borderRadius: '8px',
    padding: '12px 16px',
    fontSize: '14px',
  },

  // Popover
  Popover: {
    color: '#1f2937',
    textColor: '#f9fafb',
    borderRadius: '8px',
    boxShadow: '0 10px 15px -3px rgb(0 0 0 / 0.5), 0 4px 6px -4px rgb(0 0 0 / 0.5)',
    padding: '12px 16px',
  },

  // Drawer
  Drawer: {
    color: '#1f2937',
    textColor: '#f9fafb',
    boxShadow: '0 25px 50px -12px rgb(0 0 0 / 0.75)',
  },

  // Pagination
  Pagination: {
    itemColor: '#1f2937',
    itemColorHover: '#374151',
    itemColorPressed: '#4b5563',
    itemColorActive: '#3b82f6',
    itemTextColor: '#f9fafb',
    itemTextColorHover: '#f9fafb',
    itemTextColorPressed: '#f9fafb',
    itemTextColorActive: '#ffffff',
    itemBorderColor: '#374151',
    itemBorderRadius: '6px',
  },

  // Breadcrumb
  Breadcrumb: {
    itemTextColor: '#9ca3af',
    itemTextColorHover: '#3b82f6',
    itemTextColorActive: '#f9fafb',
    fontSize: '14px',
  },

  // Steps
  Steps: {
    stepTextColor: '#9ca3af',
    stepTextColorActive: '#3b82f6',
    stepTextColorFinished: '#f9fafb',
  },

  // Timeline
  Timeline: {
    titleTextColor: '#f9fafb',
    contentTextColor: '#9ca3af',
    lineColor: '#374151',
  },

  // Skeleton
  Skeleton: {
    color: '#374151',
    colorEnd: '#4b5563',
    borderRadius: '6px',
  },

  // Spin
  Spin: {
    sizeSmall: '18px',
    sizeMedium: '24px',
    sizeLarge: '36px',
  },

  // Tabs
  Tabs: {
    tabTextColorBar: '#9ca3af',
    tabTextColorActiveBar: '#3b82f6',
    tabTextColorHoverBar: '#f9fafb',
    barColor: 'transparent',
    barColorHover: 'transparent',
    panePaddingSmall: '12px 0',
    panePaddingMedium: '16px 0',
    panePaddingLarge: '24px 0',
    gapSmall: '24px',
    gapMedium: '28px',
    gapLarge: '32px',
  },
}

/**
 * 获取主题覆盖配置
 * @param {string} theme - 主题模式 ('light' | 'dark')
 * @returns {Object} Naive UI 主题覆盖配置
 */
export function getThemeOverrides(theme) {
  return theme === 'dark' ? darkThemeOverrides : lightThemeOverrides
}
