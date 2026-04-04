# Forks Design Tokens - 快速参考

> Grid Design System - Chromatic Team

## 快速查找

### 最常用的变量

```css
/* 颜色 */
var(--color-primary)              /* #2563eb - 主色 */
var(--color-text-primary)         /* 主要文本 */
var(--color-text-secondary)       /* 次要文本 */
var(--color-bg-surface)           /* 表面背景 */
var(--color-border)               /* 边框 */

/* 间距 */
var(--space-2)                    /* 8px */
var(--space-4)                    /* 16px */
var(--space-6)                    /* 24px */
var(--space-8)                    /* 32px */

/* 圆角 */
var(--radius-md)                  /* 8px - 默认 */
var(--radius-full)                /* 圆形 */

/* 阴影 */
var(--shadow-sm)                  /* 小阴影 */
var(--shadow-md)                  /* 中阴影 */
var(--shadow-lg)                  /* 大阴影 */

/* 字体 */
var(--font-sans)                  /* 无衬线字体 */
var(--font-mono)                  /* 等宽字体 */
var(--text-sm)                    /* 14px */
var(--text-base)                  /* 16px */
var(--text-lg)                    /* 18px */
```

---

## 颜色速查表

### 品牌色
```
Primary: #2563EB (主要操作)
Success: #10B981 (成功状态)
Warning: #F59E0B (警告状态)
Error:   #EF4444 (错误状态)
Info:    #3B82F6 (信息提示)
```

### 文本色
```
Primary:   #111827 (主要文本)
Secondary: #6B7280 (次要文本)
Tertiary:  #9CA3AF (辅助文本)
Disabled:  #D1D5DB (禁用文本)
```

### 背景色
```
Body:    #F9FAFB (页面背景)
Surface: #FFFFFF (卡片/容器)
Modal:   rgba(0,0,0,0.5) (遮罩)
```

---

## 间距速查表

```
0  = 0px
1  = 4px   (微小)
2  = 8px   (小)
3  = 12px
4  = 16px  (中)
6  = 24px  (大)
8  = 32px  (超大)
12 = 48px  (特大)
```

**使用建议**:
- 卡片内边距: `var(--space-6)` (24px)
- 按钮内边距: `var(--space-4)` (16px)
- 元素间距: `var(--space-4)` (16px)
- 区块间距: `var(--space-8)` (32px)

---

## 圆角速查表

```
sm   = 4px  (标签、小按钮)
md   = 8px  (默认圆角)
lg   = 12px (大卡片)
xl   = 16px (模态框)
full = 圆形 (头像、徽章)
```

---

## 阴影速查表

```
xs  = 微小阴影 (工具栏)
sm  = 小阴影 (卡片)
md  = 中阴影 (焦点状态)
lg  = 大阴影 (下拉菜单)
xl  = 超大阴影 (弹出层)
2xl = 特大阴影 (模态框)
```

---

## 字体速查表

### 字号
```
xs   = 12px (标签)
sm   = 14px (按钮、辅助文本)
base = 16px (正文)
lg   = 18px (小标题)
xl   = 20px (中等标题)
2xl  = 24px (大标题)
```

### 字重
```
normal   = 400 (正文)
medium   = 500 (中等强调)
semibold = 600 (标题)
bold     = 700 (重点标题)
```

---

## 组件默认样式

### 按钮
```css
height: 40px
padding: 0 16px
border-radius: 8px
font-size: 14px
font-weight: 500
```

### 输入框
```css
height: 40px
padding: 0 12px
border-radius: 8px
border: 1px solid #d1d5db
font-size: 16px
```

### 卡片
```css
border-radius: 12px
padding: 24px
background: #ffffff
border: 1px solid #e5e7eb
box-shadow: 0 1px 3px 0 rgb(0 0 0 / 0.1)
```

---

## 常用模式

### 居中容器
```css
.container {
  max-width: 1280px;
  margin: 0 auto;
  padding: 0 var(--space-4);
}
```

### Flex 布局
```css
.flex-center {
  display: flex;
  align-items: center;
  justify-content: center;
}

.flex-between {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
```

### 过渡效果
```css
.transition {
  transition: all 200ms cubic-bezier(0.4, 0, 0.2, 1);
}

.transition-colors {
  transition: color 150ms ease-out,
              background-color 150ms ease-out,
              border-color 150ms ease-out;
}
```

---

## 响应式断点

```
xs  = 480px  (超小屏幕)
sm  = 640px  (小屏幕)
md  = 768px  (平板)
lg  = 1024px (桌面)
xl  = 1280px (大桌面)
2xl = 1536px (超大桌面)
```

---

## Git 状态颜色

```css
/* 新增文件 */
--color-git-added: #10b981

/* 修改文件 */
--color-git-modified: #f59e0b

/* 删除文件 */
--color-git-deleted: #ef4444

/* 重命名 */
--color-git-renamed: #0ea5e9

/* 复制 */
--color-git-copied: #3b82f6
```

---

## 主题切换

### JavaScript
```javascript
import { useThemeStore } from '@/stores/theme'

const themeStore = useThemeStore()

// 切换主题
themeStore.toggleTheme()

// 设置主题
themeStore.setTheme('dark')  // 'light' | 'dark' | 'auto'

// 获取当前主题
const currentTheme = themeStore.getCurrentTheme()
```

### CSS
```css
/* 暗色模式特定样式 */
:root[data-theme="dark"] {
  /* 自定义样式 */
}
```

---

## Naive UI 组件样式覆盖

```javascript
import { getThemeOverrides } from '@/composables/useTheme'

const themeOverrides = getThemeOverrides('light')

// 在 NConfigProvider 中使用
<NConfigProvider :theme-overrides="themeOverrides">
  <App />
</NConfigProvider>
```

---

## 常见问题

### Q: 如何自定义颜色？
A: 修改 `variables.css` 中的 CSS 变量值。

### Q: 如何添加新的组件样式？
A: 在 `index.css` 中添加，确保使用 Design Tokens。

### Q: 如何支持暗色模式？
A: 在 `theme-dark.css` 中覆盖相应的变量。

### Q: 为什么使用 CSS 变量？
A: 方便主题切换、提高可维护性、确保一致性。

---

## 文件结构

```
web/src/styles/
├── variables.css    # Design Tokens 定义
├── theme-light.css  # 亮色主题
├── theme-dark.css   # 暗色主题
└── index.css        # 主入口 + 基础样式

web/src/composables/
└── useTheme.js      # Naive UI 主题配置

web/src/stores/
└── theme.js         # 主题状态管理
```

---

**提示**: 将此文件加入书签，快速查找 Design Tokens！
