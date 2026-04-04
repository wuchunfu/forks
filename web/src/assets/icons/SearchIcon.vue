<template>
  <svg 
    :width="size" 
    :height="size" 
    viewBox="0 0 24 24" 
    fill="none" 
    xmlns="http://www.w3.org/2000/svg"
  >
    <!-- 搜索放大镜 -->
    <circle 
      cx="10" cy="10" 
      r="6" 
      :stroke="glassColor" 
      stroke-width="2.5" 
      fill="none"
    />
    
    <!-- 放大镜柄 -->
    <path 
      d="M15.5 15.5 L20 20" 
      :stroke="handleColor" 
      stroke-width="2.5" 
      stroke-linecap="round"
    />
    
    <!-- 放大镜内容 - 搜索指示 -->
    <g transform="translate(10, 10)">
      <!-- 数据点 -->
      <circle cx="-2" cy="-1" r="0.8" :fill="dataColor" opacity="0.8"/>
      <circle cx="1" cy="-2" r="0.8" :fill="dataColor" opacity="0.6"/>
      <circle cx="0" cy="1" r="0.8" :fill="dataColor" opacity="0.9"/>
      <circle cx="-1" cy="2" r="0.8" :fill="dataColor" opacity="0.7"/>
    </g>
    
    <!-- 搜索光效 -->
    <g transform="translate(10, 10)">
      <circle 
        cx="0" cy="0" 
        r="3" 
        :stroke="glowColor" 
        stroke-width="0.5" 
        fill="none" 
        opacity="0.4"
      />
      <circle 
        cx="0" cy="0" 
        r="4.5" 
        :stroke="glowColor" 
        stroke-width="0.3" 
        fill="none" 
        opacity="0.2"
      />
    </g>
    
    <!-- Git相关标识 -->
    <g transform="translate(6, 6)">
      <path 
        d="M0 0 Q1 -1 2 0 Q1 1 0 0" 
        :fill="gitColor"
        opacity="0.6"
      />
    </g>
    
    <!-- 装饰性搜索线条 -->
    <path 
      d="M18 18 L21 21" 
      :stroke="accentColor" 
      stroke-width="1" 
      stroke-linecap="round"
      opacity="0.5"
    />
  </svg>
</template>

<script setup>
const props = defineProps({
  size: {
    type: [String, Number],
    default: 24
  },
  glassColor: {
    type: String,
    default: '#2196F3'
  },
  handleColor: {
    type: String,
    default: '#757575'
  },
  dataColor: {
    type: String,
    default: '#4CAF50'
  },
  glowColor: {
    type: String,
    default: '#2196F3'
  },
  gitColor: {
    type: String,
    default: '#4CAF50'
  },
  accentColor: {
    type: String,
    default: '#FF9800'
  }
})
</script>