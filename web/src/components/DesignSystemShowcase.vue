/**
 * Forks Design System Showcase
 * Grid Design System - Chromatic Team
 *
 * 设计系统展示组件
 * 用于演示所有 Design Tokens 的使用
 */

<template>
  <div class="design-showcase">
    <section class="showcase-section">
      <h2 class="section-title">颜色系统 (Color System)</h2>

      <!-- 品牌色 -->
      <div class="color-group">
        <h3 class="group-title">品牌色 (Brand Colors)</h3>
        <div class="color-grid">
          <div
            v-for="i in 9"
            :key="`primary-${i * 100}`"
            class="color-swatch"
            :style="{ backgroundColor: `var(--color-primary-${i * 100})` }"
          >
            <span class="color-label">{{ i * 100 }}</span>
          </div>
        </div>
      </div>

      <!-- 语义色 -->
      <div class="color-group">
        <h3 class="group-title">语义色 (Semantic Colors)</h3>
        <div class="semantic-colors">
          <div class="semantic-item">
            <div class="color-circle success"></div>
            <span>Success: var(--color-success)</span>
          </div>
          <div class="semantic-item">
            <div class="color-circle warning"></div>
            <span>Warning: var(--color-warning)</span>
          </div>
          <div class="semantic-item">
            <div class="color-circle error"></div>
            <span>Error: var(--color-error)</span>
          </div>
          <div class="semantic-item">
            <div class="color-circle info"></div>
            <span>Info: var(--color-info)</span>
          </div>
        </div>
      </div>

      <!-- Git 状态色 -->
      <div class="color-group">
        <h3 class="group-title">Git 状态色</h3>
        <div class="git-status">
          <n-tag type="success">Added: var(--color-git-added)</n-tag>
          <n-tag type="warning">Modified: var(--color-git-modified)</n-tag>
          <n-tag type="error">Deleted: var(--color-git-deleted)</n-tag>
          <n-tag type="info">Renamed: var(--color-git-renamed)</n-tag>
        </div>
      </div>
    </section>

    <section class="showcase-section">
      <h2 class="section-title">间距系统 (Spacing)</h2>
      <div class="spacing-demo">
        <div class="spacing-box" style="padding: var(--space-1)">
          --space-1 (4px)
        </div>
        <div class="spacing-box" style="padding: var(--space-2)">
          --space-2 (8px)
        </div>
        <div class="spacing-box" style="padding: var(--space-4)">
          --space-4 (16px)
        </div>
        <div class="spacing-box" style="padding: var(--space-6)">
          --space-6 (24px)
        </div>
        <div class="spacing-box" style="padding: var(--space-8)">
          --space-8 (32px)
        </div>
      </div>
    </section>

    <section class="showcase-section">
      <h2 class="section-title">字体系统 (Typography)</h2>
      <div class="typography-demo">
        <p class="text-xs">text-xs - 12px: 辅助文本</p>
        <p class="text-sm">text-sm - 14px: 按钮文本</p>
        <p class="text-base">text-base - 16px: 正文内容</p>
        <p class="text-lg">text-lg - 18px: 小标题</p>
        <p class="text-xl">text-xl - 20px: 中等标题</p>
        <p class="text-2xl">text-2xl - 24px: 大标题</p>

        <div class="font-weights">
          <p class="font-normal">font-normal (400): 正常字重</p>
          <p class="font-medium">font-medium (500): 中等字重</p>
          <p class="font-semibold">font-semibold (600): 半粗体</p>
          <p class="font-bold">font-bold (700): 粗体</p>
        </div>
      </div>
    </section>

    <section class="showcase-section">
      <h2 class="section-title">圆角系统 (Border Radius)</h2>
      <div class="radius-demo">
        <div class="radius-box" style="border-radius: var(--radius-none)">
          None (0px)
        </div>
        <div class="radius-box" style="border-radius: var(--radius-sm)">
          Small (4px)
        </div>
        <div class="radius-box" style="border-radius: var(--radius-md)">
          Medium (8px) - 默认
        </div>
        <div class="radius-box" style="border-radius: var(--radius-lg)">
          Large (12px)
        </div>
        <div class="radius-box" style="border-radius: var(--radius-xl)">
          XLarge (16px)
        </div>
        <div class="radius-box" style="border-radius: var(--radius-full)">
          Full (圆形)
        </div>
      </div>
    </section>

    <section class="showcase-section">
      <h2 class="section-title">阴影系统 (Shadows)</h2>
      <div class="shadow-demo">
        <div class="shadow-box" style="box-shadow: var(--shadow-xs)">
          XS Shadow
        </div>
        <div class="shadow-box" style="box-shadow: var(--shadow-sm)">
          SM Shadow (卡片默认)
        </div>
        <div class="shadow-box" style="box-shadow: var(--shadow-md)">
          MD Shadow
        </div>
        <div class="shadow-box" style="box-shadow: var(--shadow-lg)">
          LG Shadow (下拉菜单)
        </div>
        <div class="shadow-box" style="box-shadow: var(--shadow-xl)">
          XL Shadow
        </div>
        <div class="shadow-box" style="box-shadow: var(--shadow-2xl)">
          2XL Shadow (模态框)
        </div>
      </div>
    </section>

    <section class="showcase-section">
      <h2 class="section-title">组件示例 (Components)</h2>

      <!-- 按钮 -->
      <div class="component-group">
        <h3 class="group-title">按钮 (Buttons)</h3>
        <div class="button-demo">
          <n-button type="primary">Primary Button</n-button>
          <n-button>Default Button</n-button>
          <n-button type="info">Info Button</n-button>
          <n-button type="success">Success Button</n-button>
          <n-button type="warning">Warning Button</n-button>
          <n-button type="error">Error Button</n-button>
        </div>

        <div class="button-demo mt-4">
          <n-button size="small">Small</n-button>
          <n-button size="medium">Medium</n-button>
          <n-button size="large">Large</n-button>
        </div>
      </div>

      <!-- 输入框 -->
      <div class="component-group">
        <h3 class="group-title">输入框 (Inputs)</h3>
        <div class="input-demo">
          <n-input
            placeholder="Default input"
            style="width: 240px"
          />
          <n-input
            placeholder="Disabled input"
            disabled
            style="width: 240px"
          />
        </div>
      </div>

      <!-- 卡片 -->
      <div class="component-group">
        <h3 class="group-title">卡片 (Cards)</h3>
        <n-card class="demo-card">
          <template #header>
            <span class="card-title">卡片标题</span>
          </template>
          <p>这是卡片内容，使用了 var(--radius-lg) 圆角和 var(--shadow-sm) 阴影。</p>
        </n-card>
      </div>

      <!-- 徽章 -->
      <div class="component-group">
        <h3 class="group-title">徽章 (Badges)</h3>
        <div class="badge-demo">
          <n-badge type="primary" :value="10">
            <n-button>Primary</n-button>
          </n-badge>
          <n-badge type="success" :value="20">
            <n-button>Success</n-button>
          </n-badge>
          <n-badge type="warning" :value="30">
            <n-button>Warning</n-button>
          </n-badge>
          <n-badge type="error" :value="40">
            <n-button>Error</n-button>
          </n-badge>
        </div>
      </div>
    </section>

    <section class="showcase-section">
      <h2 class="section-title">动画示例 (Animations)</h2>
      <div class="animation-demo">
        <n-button @click="triggerAnimation('fadeIn')">
          Fade In
        </n-button>
        <n-button @click="triggerAnimation('slideUp')">
          Slide Up
        </n-button>
        <n-button @click="triggerAnimation('scaleIn')">
          Scale In
        </n-button>
        <div
          v-if="showAnimation"
          class="animated-box"
          :class="animationClass"
        >
          Animated Element
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const showAnimation = ref(false)
const animationClass = ref('')

const triggerAnimation = (type) => {
  showAnimation.value = false
  setTimeout(() => {
    animationClass.value = `animate-${type}`
    showAnimation.value = true
  }, 50)
}
</script>

<style scoped>
/* 布局 */
.design-showcase {
  max-width: 1200px;
  margin: 0 auto;
  padding: var(--space-8);
  background-color: var(--color-bg-page);
  min-height: 100vh;
}

.showcase-section {
  margin-bottom: var(--space-12);
}

.section-title {
  font-size: var(--text-2xl);
  font-weight: var(--font-bold);
  color: var(--color-text-primary);
  margin-bottom: var(--space-6);
  padding-bottom: var(--space-3);
  border-bottom: 2px solid var(--color-border);
}

.group-title {
  font-size: var(--text-lg);
  font-weight: var(--font-semibold);
  color: var(--color-text-primary);
  margin-bottom: var(--space-4);
}

/* 颜色系统 */
.color-grid {
  display: grid;
  grid-template-columns: repeat(9, 1fr);
  gap: var(--space-2);
  margin-bottom: var(--space-6);
}

.color-swatch {
  height: 60px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--color-border);
  box-shadow: var(--shadow-sm);
}

.color-label {
  font-size: var(--text-xs);
  font-weight: var(--font-medium);
  color: var(--color-text-primary);
}

.semantic-colors {
  display: flex;
  gap: var(--space-6);
  flex-wrap: wrap;
}

.semantic-item {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
}

.color-circle {
  width: 24px;
  height: 24px;
  border-radius: var(--radius-full);
  box-shadow: var(--shadow-sm);
}

.color-circle.success {
  background-color: var(--color-success);
}

.color-circle.warning {
  background-color: var(--color-warning);
}

.color-circle.error {
  background-color: var(--color-error);
}

.color-circle.info {
  background-color: var(--color-info);
}

.git-status {
  display: flex;
  gap: var(--space-3);
  flex-wrap: wrap;
}

/* 间距系统 */
.spacing-demo {
  display: flex;
  gap: var(--space-4);
  flex-wrap: wrap;
}

.spacing-box {
  background-color: var(--color-primary);
  color: white;
  border-radius: var(--radius-md);
  font-size: var(--text-sm);
  box-shadow: var(--shadow-sm);
}

/* 字体系统 */
.typography-demo {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.font-weights {
  margin-top: var(--space-6);
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.text-xs { font-size: var(--text-xs); }
.text-sm { font-size: var(--text-sm); }
.text-base { font-size: var(--text-base); }
.text-lg { font-size: var(--text-lg); }
.text-xl { font-size: var(--text-xl); }
.text-2xl { font-size: var(--text-2xl); }

.font-normal { font-weight: var(--font-normal); }
.font-medium { font-weight: var(--font-medium); }
.font-semibold { font-weight: var(--font-semibold); }
.font-bold { font-weight: var(--font-bold); }

/* 圆角系统 */
.radius-demo {
  display: flex;
  gap: var(--space-4);
  flex-wrap: wrap;
}

.radius-box {
  width: 80px;
  height: 80px;
  background-color: var(--color-primary);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: var(--text-xs);
  text-align: center;
  padding: var(--space-2);
  box-shadow: var(--shadow-sm);
}

/* 阴影系统 */
.shadow-demo {
  display: flex;
  gap: var(--space-6);
  flex-wrap: wrap;
}

.shadow-box {
  width: 120px;
  height: 120px;
  background-color: var(--color-card-bg);
  border: 1px solid var(--color-card-border);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  color: var(--color-text-primary);
}

/* 组件 */
.component-group {
  margin-bottom: var(--space-8);
}

.button-demo,
.input-demo,
.badge-demo {
  display: flex;
  gap: var(--space-3);
  flex-wrap: wrap;
}

.demo-card {
  max-width: 400px;
}

.card-title {
  font-size: var(--text-lg);
  font-weight: var(--font-semibold);
  color: var(--color-text-primary);
}

.mt-4 {
  margin-top: var(--space-4);
}

/* 动画 */
.animation-demo {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.animated-box {
  width: 200px;
  height: 100px;
  background-color: var(--color-primary);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-lg);
  font-size: var(--text-base);
  font-weight: var(--font-medium);
}

/* 响应式 */
@media (max-width: 768px) {
  .design-showcase {
    padding: var(--space-4);
  }

  .color-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}
</style>
