<template>
  <svg 
    :width="size" 
    :height="size" 
    viewBox="0 0 24 24" 
    fill="none" 
    xmlns="http://www.w3.org/2000/svg"
  >
    <!-- 刷新箭头圆环 -->
    <path 
      d="M12 2 A10 10 0 0 1 22 12" 
      :stroke="primaryColor" 
      stroke-width="2.5" 
      stroke-linecap="round" 
      fill="none"
    />
    <path 
      d="M22 12 A10 10 0 0 1 12 22" 
      :stroke="accentColor" 
      stroke-width="2.5" 
      stroke-linecap="round" 
      fill="none"
    />
    <path 
      d="M12 22 A10 10 0 0 1 2 12" 
      :stroke="primaryColor" 
      stroke-width="2.5" 
      stroke-linecap="round" 
      fill="none"
    />
    <path 
      d="M2 12 A10 10 0 0 1 12 2" 
      :stroke="accentColor" 
      stroke-width="2.5" 
      stroke-linecap="round" 
      fill="none"
      stroke-dasharray="8 4"
    />
    
    <!-- 箭头指示器 -->
    <g transform="translate(20, 8)">
      <path 
        d="M-2 -2 L0 0 L-2 2" 
        :stroke="arrowColor" 
        stroke-width="2" 
        stroke-linecap="round" 
        stroke-linejoin="round" 
        fill="none"
      />
    </g>
    
    <!-- 中心数据指示器 -->
    <g transform="translate(12, 12)">
      <circle cx="0" cy="0" r="4" :fill="centerBg" opacity="0.9"/>
      
      <!-- 数据条 -->
      <rect x="-2" y="-1" width="1" height="2" :fill="dataColor" opacity="0.8"/>
      <rect x="-0.5" y="-1.5" width="1" height="3" :fill="dataColor"/>
      <rect x="1" y="-0.5" width="1" height="1" :fill="dataColor" opacity="0.6"/>
    </g>
    
    <!-- 装饰性光点 -->
    <circle cx="6" cy="6" r="1" :fill="glowColor" opacity="0.6"/>
    <circle cx="18" cy="18" r="1" :fill="glowColor" opacity="0.6"/>
  </svg>
</template>

<script setup>
const props = defineProps({
  size: {
    type: [String, Number],
    default: 24
  },
  primaryColor: {
    type: String,
    default: '#FF9800'
  },
  accentColor: {
    type: String,
    default: '#FFB74D'
  },
  arrowColor: {
    type: String,
    default: '#F57C00'
  },
  centerBg: {
    type: String,
    default: '#FFF3E0'
  },
  dataColor: {
    type: String,
    default: '#FF9800'
  },
  glowColor: {
    type: String,
    default: '#FFE0B2'
  }
})
</script>