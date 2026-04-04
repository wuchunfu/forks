<template>
  <svg 
    :width="size" 
    :height="size" 
    viewBox="0 0 24 24" 
    fill="none" 
    xmlns="http://www.w3.org/2000/svg"
  >
    <!-- 信息卡片背景 -->
    <rect 
      x="3" y="4" 
      width="18" height="16" 
      rx="3" 
      :fill="cardBg" 
      :stroke="cardBorder" 
      stroke-width="1"
    />
    
    <!-- 标题区域 -->
    <rect 
      x="3" y="4" 
      width="18" height="4" 
      rx="3" 
      :fill="headerBg"
    />
    
    <!-- 信息行 -->
    <g transform="translate(6, 11)">
      <!-- 第一行 -->
      <circle cx="1" cy="1" r="1" :fill="dotColor"/>
      <rect x="3" y="0.5" width="8" height="1" rx="0.5" :fill="textColor" opacity="0.8"/>
      
      <!-- 第二行 -->
      <circle cx="1" cy="4" r="1" :fill="dotColor"/>
      <rect x="3" y="3.5" width="6" height="1" rx="0.5" :fill="textColor" opacity="0.6"/>
      
      <!-- 第三行 -->
      <circle cx="1" cy="7" r="1" :fill="dotColor"/>
      <rect x="3" y="6.5" width="10" height="1" rx="0.5" :fill="textColor" opacity="0.4"/>
    </g>
    
    <!-- 信息图标 -->
    <g transform="translate(17, 6)">
      <circle cx="0" cy="0" r="2" :fill="iconBg"/>
      <text 
        x="0" y="0.5" 
        text-anchor="middle" 
        :fill="iconColor" 
        font-size="2.5" 
        font-weight="bold"
      >i</text>
    </g>
    
    <!-- 装饰性元素 -->
    <g transform="translate(19, 15)">
      <circle cx="0" cy="0" r="1.5" :fill="accentBg" opacity="0.3"/>
      <path 
        d="M-0.5 0 L0.5 0 M0 -0.5 L0 0.5" 
        :stroke="accentColor" 
        stroke-width="0.5" 
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
  cardBg: {
    type: String,
    default: '#F8F9FA'
  },
  cardBorder: {
    type: String,
    default: '#E0E0E0'
  },
  headerBg: {
    type: String,
    default: '#2196F3'
  },
  dotColor: {
    type: String,
    default: '#2196F3'
  },
  textColor: {
    type: String,
    default: '#666666'
  },
  iconBg: {
    type: String,
    default: '#2196F3'
  },
  iconColor: {
    type: String,
    default: 'white'
  },
  accentBg: {
    type: String,
    default: '#4CAF50'
  },
  accentColor: {
    type: String,
    default: '#4CAF50'
  }
})
</script>