<template>
  <svg 
    :width="size" 
    :height="size" 
    viewBox="0 0 24 24" 
    fill="none" 
    xmlns="http://www.w3.org/2000/svg"
  >
    <!-- 仓库文件夹 -->
    <path 
      d="M3 7V19C3 19.5523 3.44772 20 4 20H20C20.5523 20 21 19.5523 21 19V9C21 8.44772 20.5523 8 20 8H12L10 6H4C3.44772 6 3 6.44772 3 7Z" 
      :fill="primaryColor" 
      opacity="0.9"
    />
    
    <!-- Git分支标识 -->
    <g transform="translate(6, 10)">
      <!-- 主分支 -->
      <circle cx="2" cy="3" r="1.5" :fill="accentColor"/>
      <circle cx="6" cy="3" r="1.5" :fill="accentColor"/>
      <circle cx="10" cy="3" r="1.5" :fill="accentColor"/>
      
      <!-- 连接线 -->
      <path 
        d="M3.5 3 L8.5 3" 
        :stroke="accentColor" 
        stroke-width="1" 
        stroke-linecap="round"
      />
      
      <!-- 分支线 -->
      <path 
        d="M6 3 Q7 1 8 0" 
        :stroke="highlightColor" 
        stroke-width="1" 
        stroke-linecap="round" 
        fill="none"
      />
      <circle cx="8" cy="0" r="1" :fill="highlightColor"/>
    </g>
    
    <!-- 星标 -->
    <g transform="translate(16, 3)" v-if="showStar">
      <path 
        d="M2 1 L2.6 2.2 L4 2.2 L3 3 L3.4 4.2 L2 3.6 L0.6 4.2 L1 3 L0 2.2 L1.4 2.2 Z" 
        :fill="starColor"
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
  primaryColor: {
    type: String,
    default: '#4CAF50'
  },
  accentColor: {
    type: String,
    default: '#2E7D32'
  },
  highlightColor: {
    type: String,
    default: '#81C784'
  },
  starColor: {
    type: String,
    default: '#FFB300'
  },
  showStar: {
    type: Boolean,
    default: false
  }
})
</script>