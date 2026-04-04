<template>
  <svg 
    :width="size" 
    :height="size" 
    viewBox="0 0 64 64" 
    fill="none" 
    xmlns="http://www.w3.org/2000/svg"
  >
    <!-- Git分支图标背景 -->
    <circle cx="32" cy="32" r="30" :fill="bgColor" stroke="none"/>
    
    <!-- 主分支线 -->
    <path 
      d="M16 32 L48 32" 
      :stroke="primaryColor" 
      stroke-width="3" 
      stroke-linecap="round"
    />
    
    <!-- 分支节点 -->
    <circle cx="20" cy="32" r="4" :fill="primaryColor"/>
    <circle cx="32" cy="32" r="4" :fill="primaryColor"/>
    <circle cx="44" cy="32" r="4" :fill="primaryColor"/>
    
    <!-- 向上分支 -->
    <path 
      d="M32 32 Q36 24 40 20" 
      :stroke="accentColor" 
      stroke-width="2.5" 
      stroke-linecap="round" 
      fill="none"
    />
    <circle cx="40" cy="20" r="3" :fill="accentColor"/>
    
    <!-- 向下分支 -->
    <path 
      d="M20 32 Q24 40 28 44" 
      :stroke="accentColor" 
      stroke-width="2.5" 
      stroke-linecap="round" 
      fill="none"
    />
    <circle cx="28" cy="44" r="3" :fill="accentColor"/>
    
    <!-- Plus符号 -->
    <g transform="translate(48, 16)">
      <circle cx="0" cy="0" r="8" :fill="plusBg" stroke="white" stroke-width="1"/>
      <path 
        d="M-4 0 L4 0 M0 -4 L0 4" 
        stroke="white" 
        stroke-width="2" 
        stroke-linecap="round"
      />
    </g>
    
    <!-- 装饰性光点 -->
    <circle cx="24" cy="24" r="1.5" :fill="accentColor" opacity="0.6"/>
    <circle cx="40" cy="40" r="1.5" :fill="accentColor" opacity="0.6"/>
    <circle cx="16" cy="40" r="1" :fill="primaryColor" opacity="0.4"/>
  </svg>
</template>

<script setup>
const props = defineProps({
  size: {
    type: [String, Number],
    default: 64
  },
  primaryColor: {
    type: String,
    default: '#4CAF50'
  },
  accentColor: {
    type: String,
    default: '#81C784'
  },
  bgColor: {
    type: String,
    default: '#F1F8E9'
  },
  plusBg: {
    type: String,
    default: '#FF9800'
  }
})
</script>