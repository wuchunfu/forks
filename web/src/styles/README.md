# Forks Design System

> **Grid Design System** by Chromatic Team
> 版本: 1.0.0 | 更新: 2025-02-20

---

## 快速开始

### 1. 导入样式

在 `main.js` 中已自动导入：

```javascript
import App from './App.vue'
```

在 `App.vue` 中导入样式：

```vue
<script setup>
import './styles/index.css'
</script>
```

### 2. 使用 Design Tokens

```vue
<template>
  <div class="my-component">
    <h1 class="title">Hello Forks</h1>
    <p class="description">使用 Design Tokens 构建一致的界面</p>
    <n-button type="primary">主要按钮</n-button>
  </div>
</template>

<style scoped>
.my-component {
  padding: var(--space-6);
  background-color: var(--color-card-bg);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-sm);
}

.title {
  font-size: var(--text-2xl);
  font-weight: var(--font-bold);
  color: var(--color-text-primary);
}

.description {
  font-size: var(--text-base);
  color: var(--color-text-secondary);
  margin-bottom: var(--space-4);
}
</style>
```

---

## 文件说明

### 核心文件

| 文件 | 说明 |
|-----|------|
| `variables.css` | Design Tokens 变量定义 |
| `theme-light.css` | 亮色主题样式 |
| `theme-dark.css` | 暗色主题样式 |
| `index.css` | 主入口文件，导入所有样式 |
| `useTheme.js` | Naive UI 主题配置 |
| `theme.js` | 主题状态管理 Store |

### 文档文件

| 文件 | 说明 |
|-----|------|
| `DESIGN_TOKENS.md` | 完整的设计系统文档 |
| `CHEATSHEET.md` | 快速参考指南 |
| `README.md` | 本文件 |

### 示例文件

| 文件 | 说明 |
|-----|------|
| `DesignSystemShowcase.vue` | 设计系统展示组件 |

---

## 使用指南

### 主题切换

#### 方式 1: 使用 Store

```vue
<script setup>
import { useThemeStore } from '@/stores/theme'

const themeStore = useThemeStore()
</script>

<template>
  <div>
    <n-button @click="themeStore.toggleTheme()">
      切换主题
    </n-button>
    <n-button @click="themeStore.setTheme('dark')">
      暗色模式
    </n-button>
    <n-button @click="themeStore.setTheme('light')">
      亮色模式
    </n-button>
  </div>
</template>
```

#### 方式 2: 直接操作 HTML

```javascript
// 设置为暗色主题
document.documentElement.setAttribute('data-theme', 'dark')

// 设置为亮色主题
document.documentElement.setAttribute('data-theme', 'light')

// 移除主题（使用默认）
document.documentElement.removeAttribute('data-theme')
```

### 读取 Token 值

```javascript
// 获取主色
const primaryColor = getComputedStyle(document.documentElement)
  .getPropertyValue('--color-primary')
  .trim()

console.log(primaryColor) // '#2563eb'
```

---

## Design Tokens 分类

### 颜色 (Colors)

```css
/* 品牌色 */
--color-primary-50 ~ --color-primary-900
--color-primary (默认: #2563eb)

/* 语义色 */
--color-success
--color-warning
--color-error
--color-info

/* 中性色 */
--color-gray-50 ~ --color-gray-900
--color-text-primary
--color-text-secondary
--color-border
```

### 间距 (Spacing)

```css
--space-0 ~ --space-64
--space-xs (4px)
--space-sm (8px)
--space-md (16px)
--space-lg (24px)
--space-xl (32px)
```

### 字体 (Typography)

```css
--font-sans (无衬线字体)
--font-mono (等宽字体)
--text-xs ~ --text-9xl
--font-normal ~ --font-black
```

### 圆角 (Border Radius)

```css
--radius-none
--radius-sm (4px)
--radius-md (8px) - 默认
--radius-lg (12px)
--radius-full (圆形)
```

### 阴影 (Shadows)

```css
--shadow-xs
--shadow-sm (卡片默认)
--shadow-md
--shadow-lg (下拉菜单)
--shadow-xl
--shadow-2xl (模态框)
```

### 布局 (Layout)

```css
--sidebar-width-expanded (240px)
--sidebar-width-collapsed (64px)
--navbar-height (64px)
--breakpoint-sm ~ --breakpoint-2xl
```

---

## 组件样式规范

### 按钮

```css
height: var(--btn-height-md)        /* 40px */
padding: var(--btn-padding-x-md)    /* 0 16px */
border-radius: var(--btn-radius)    /* 8px */
font-size: var(--text-sm)           /* 14px */
font-weight: var(--btn-font-weight) /* 500 */
```

### 输入框

```css
height: var(--input-height-md)      /* 40px */
padding: 0 var(--input-padding-x)   /* 0 12px */
border-radius: var(--input-radius)  /* 8px */
border-color: var(--color-input-border)
```

### 卡片

```css
background: var(--color-card-bg)
border: 1px solid var(--color-card-border)
border-radius: var(--radius-lg)     /* 12px */
padding: var(--card-padding)        /* 24px */
box-shadow: var(--shadow-sm)
```

---

## Naive UI 集成

### 主题覆盖

```vue
<script setup>
import { computed } from 'vue'
import { useThemeStore } from '@/stores/theme'
import { darkTheme } from 'naive-ui'
import { getThemeOverrides } from '@/composables/useTheme'

const themeStore = useThemeStore()

const naiveTheme = computed(() => {
  return themeStore.getCurrentTheme() === 'dark' ? darkTheme : null
})

const themeOverrides = computed(() => {
  return getThemeOverrides(themeStore.getCurrentTheme())
})
</script>

<template>
  <n-config-provider
    :theme="naiveTheme"
    :theme-overrides="themeOverrides"
  >
    <app />
  </n-config-provider>
</template>
```

---

## 最佳实践

### 1. 始终使用 Design Tokens

❌ **不推荐**:
```css
.my-component {
  padding: 16px;
  background: #ffffff;
  border-radius: 8px;
}
```

✅ **推荐**:
```css
.my-component {
  padding: var(--space-4);
  background: var(--color-card-bg);
  border-radius: var(--radius-md);
}
```

### 2. 语义化命名

❌ **不推荐**:
```css
color: var(--color-blue-600);
```

✅ **推荐**:
```css
color: var(--color-primary);
```

### 3. 保持一致性

所有组件使用相同的间距、圆角、阴影值：

```css
/* 统一使用 8px 网格 */
padding: var(--space-2);  /* 8px */
padding: var(--space-4);  /* 16px */
padding: var(--space-6);  /* 24px */

/* 统一使用默认圆角 */
border-radius: var(--radius-md);  /* 8px */
```

### 4. 支持暗色模式

在需要特定暗色样式时：

```css
:root[data-theme="dark"] .my-component {
  /* 暗色模式特定样式 */
  background: var(--color-bg-surface);
}
```

---

## 调试工具

### 查看所有变量

在浏览器控制台中运行：

```javascript
// 获取所有 CSS 变量
const root = document.documentElement
const variables = getComputedStyle(root)
const colorVariables = {}

for (let i = 0; i < variables.length; i++) {
  const name = variables[i]
  if (name.startsWith('--color-')) {
    colorVariables[name] = variables.getPropertyValue(name)
  }
}

console.table(colorVariables)
```

### 主题切换调试

```javascript
// 手动切换主题
import { useThemeStore } from '@/stores/theme'
const themeStore = useThemeStore()

// 在控制台使用
window.themeStore = themeStore
// 然后可以调用
// themeStore.toggleTheme()
// themeStore.setTheme('dark')
```

---

## 更新日志

### v1.0.0 (2025-02-20)

**新增**:
- 完整的 Design Tokens 系统
- 亮色/暗色主题支持
- Naive UI 主题集成
- 主题切换功能
- 组件基础样式
- 文档和示例

---

## 贡献指南

### 修改 Design Tokens

1. 在 `variables.css` 中修改
2. 更新 `DESIGN_TOKENS.md` 文档
3. 测试所有受影响的组件
4. 提交 PR

### 添加新组件样式

1. 在 `index.css` 中添加
2. 确保使用 Design Tokens
3. 在 `DESIGN_TOKENS.md` 中添加文档
4. 在 `DesignSystemShowcase.vue` 中添加示例

---

## 参考资源

- [完整文档](../DESIGN_TOKENS.md)
- [快速参考](./CHEATSHEET.md)
- [展示组件](../components/DesignSystemShowcase.vue)
- [Naive UI 文档](https://www.naiveui.com/)

---

**维护**: Grid (Chromatic Team)
**座右铭**: "秩序产生美。"
