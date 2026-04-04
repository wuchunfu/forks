<template>
  <svg 
    :width="size" 
    :height="size" 
    viewBox="0 0 24 24" 
    fill="none" 
    xmlns="http://www.w3.org/2000/svg"
  >
    <!-- 代码窗口背景 -->
    <rect 
      x="2" y="4" 
      width="20" height="16" 
      rx="2" 
      :fill="bgColor" 
      :stroke="borderColor" 
      stroke-width="1"
    />
    
    <!-- 标题栏 -->
    <rect 
      x="2" y="4" 
      width="20" height="4" 
      rx="2" 
      :fill="headerColor"
    />
    
    <!-- 标题栏按钮 -->
    <circle cx="5" cy="6" r="0.8" fill="#FF5F56"/>
    <circle cx="7" cy="6" r="0.8" fill="#FFBD2E"/>
    <circle cx="9" cy="6" r="0.8" fill="#27CA3F"/>
    
    <!-- 代码标签符号 -->
    <g transform="translate(4, 10)">
      <!-- 左标签 -->
      <path 
        d="M2 4 L0 6 L2 8" 
        :stroke="codeColor" 
        stroke-width="1.5" 
        stroke-linecap="round" 
        stroke-linejoin="round" 
        fill="none"
      />
      
      <!-- 右标签 -->
      <path 
        d="M14 4 L16 6 L14 8" 
        :stroke="codeColor" 
        stroke-width="1.5" 
        stroke-linecap="round" 
        stroke-linejoin="round" 
        fill="none"
      />
      
      <!-- 斜杠 -->
      <path 
        d="M10 2 L6 10" 
        :stroke="slashColor" 
        stroke-width="1.5" 
        stroke-linecap="round"
      />
    </g>
    
    <!-- Git分支指示器 -->
    <g transform="translate(17, 16)">
      <circle cx="1" cy="1" r="2" :fill="branchBg" opacity="0.8"/>
      <path 
        d="M0 1 L2 1 M1 0 L1 2" 
        :stroke="branchColor" 
        stroke-width="0.8" 
        stroke-linecap="round"
      />
    </g>
  </svg>
</template>

<script setup>
const props = defineProps({
  size: {
    type: [String, Number],
    default: 24
  },
  bgColor: {
    type: String,
    default: '#F8F9FA'
  },
  borderColor: {
    type: String,
    default: '#E0E0E0'
  },
  headerColor: {
    type: String,
    default: '#4CAF50'
  },
  codeColor: {
    type: String,
    default: '#2196F3'
  },
  slashColor: {
    type: String,
    default: '#FF9800'
  },
  branchBg: {
    type: String,
    default: '#4CAF50'
  },
  branchColor: {
    type: String,
    default: 'white'
  }
})
</script>