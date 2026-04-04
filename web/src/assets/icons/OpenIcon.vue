<template>
  <svg 
    :width="size" 
    :height="size" 
    viewBox="0 0 24 24" 
    fill="none" 
    xmlns="http://www.w3.org/2000/svg"
  >
    <!-- 外部链接箭头背景 -->
    <rect 
      x="2" y="6" 
      width="14" height="14" 
      rx="2" 
      :fill="bgColor" 
      :stroke="borderColor" 
      stroke-width="1"
    />
    
    <!-- 内容区域 -->
    <rect 
      x="4" y="8" 
      width="10" height="4" 
      rx="1" 
      :fill="contentColor" 
      opacity="0.6"
    />
    <rect 
      x="4" y="13" 
      width="8" height="2" 
      rx="0.5" 
      :fill="contentColor" 
      opacity="0.4"
    />
    
    <!-- 外部链接箭头 -->
    <g transform="translate(17, 7)">
      <rect 
        x="0" y="0" 
        width="5" height="5" 
        rx="1" 
        :fill="arrowBg"
      />
      <!-- 箭头 -->
      <path 
        d="M1.5 3.5 L3.5 1.5 M3.5 1.5 L3.5 3.5 M3.5 1.5 L1.5 1.5" 
        :stroke="arrowColor" 
        stroke-width="1" 
        stroke-linecap="round" 
        stroke-linejoin="round" 
        fill="none"
      />
    </g>
    
    <!-- Git标识 -->
    <g transform="translate(5, 17)">
      <circle cx="0" cy="0" r="1.5" :fill="gitBg"/>
      <path 
        d="M-0.8 0 Q0 -0.8 0.8 0 Q0 0.8 -0.8 0" 
        :fill="gitColor"
      />
    </g>
    
    <!-- 装饰性连接线 -->
    <path 
      d="M16 13 L17 12" 
      :stroke="connectionColor" 
      stroke-width="1.5" 
      stroke-linecap="round"
      stroke-dasharray="2 2"
    />
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
    default: '#E3F2FD'
  },
  borderColor: {
    type: String,
    default: '#2196F3'
  },
  contentColor: {
    type: String,
    default: '#2196F3'
  },
  arrowBg: {
    type: String,
    default: '#2196F3'
  },
  arrowColor: {
    type: String,
    default: 'white'
  },
  gitBg: {
    type: String,
    default: '#4CAF50'
  },
  gitColor: {
    type: String,
    default: 'white'
  },
  connectionColor: {
    type: String,
    default: '#FF9800'
  }
})
</script>