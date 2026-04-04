<template>
  <!-- 悬浮添加按钮 -->
  <div class="floating-add-btn" @click="handleClick">
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
      <line x1="12" y1="5" x2="12" y2="19"/>
      <line x1="5" y1="12" x2="19" y2="12"/>
    </svg>
  </div>
</template>

<script setup>
const emit = defineEmits(['click'])

const handleClick = () => {
  emit('click')
}
</script>

<style scoped>
.floating-add-btn {
  position: fixed;
  right: 24px;
  bottom: 24px;
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-primary-600) 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.4);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  z-index: 1000;
}

.floating-add-btn:hover {
  transform: scale(1.1);
  box-shadow: 0 6px 20px rgba(37, 99, 235, 0.5);
}

.floating-add-btn:active {
  transform: scale(0.95);
}

.floating-add-btn svg {
  width: 24px;
  height: 24px;
}

/* 响应式 */
@media (max-width: 768px) {
  .floating-add-btn {
    right: 16px;
    bottom: 16px;
    width: 48px;
    height: 48px;
  }

  .floating-add-btn svg {
    width: 20px;
    height: 20px;
  }
}
</style>
