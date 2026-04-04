<template>
  <svg 
    :width="size" 
    :height="size" 
    viewBox="0 0 24 24" 
    fill="none" 
    xmlns="http://www.w3.org/2000/svg"
  >
    <!-- 垃圾桶盖子 -->
    <path 
      d="M8 6 L8 4 Q8 3 9 3 L15 3 Q16 3 16 4 L16 6" 
      :stroke="rimColor" 
      stroke-width="2" 
      stroke-linecap="round" 
      fill="none"
    />
    
    <!-- 垃圾桶顶部 -->
    <rect 
      x="6" y="6" 
      width="12" height="2" 
      rx="1" 
      :fill="rimColor"
    />
    
    <!-- 垃圾桶主体 -->
    <path 
      d="M7 8 L7 18 Q7 19 8 19 L16 19 Q17 19 17 18 L17 8" 
      :fill="bodyColor" 
      :stroke="borderColor" 
      stroke-width="1"
    />
    
    <!-- 删除线条 -->
    <g transform="translate(12, 13)">
      <line 
        x1="-2" y1="-2" 
        x2="2" y2="2" 
        :stroke="lineColor" 
        stroke-width="1.5" 
        stroke-linecap="round"
      />
      <line 
        x1="2" y1="-2" 
        x2="-2" y2="2" 
        :stroke="lineColor" 
        stroke-width="1.5" 
        stroke-linecap="round"
      />
    </g>
    
    <!-- 垃圾桶内容指示 -->
    <g transform="translate(10, 11)">
      <rect x="0" y="0" width="1" height="4" rx="0.5" :fill="contentColor" opacity="0.6"/>
      <rect x="2" y="1" width="1" height="3" rx="0.5" :fill="contentColor" opacity="0.6"/>
      <rect x="4" y="0" width="1" height="4" rx="0.5" :fill="contentColor" opacity="0.6"/>
    </g>
    
    <!-- 警告标识 -->
    <g transform="translate(18, 16)">
      <circle cx="0" cy="0" r="2.5" :fill="warningBg"/>
      <text 
        x="0" y="0.8" 
        text-anchor="middle" 
        :fill="warningColor" 
        font-size="2.5" 
        font-weight="bold"
      >!</text>
    </g>
  </svg>
</template>

<script setup>
const props = defineProps({
  size: {
    type: [String, Number],
    default: 24
  },
  rimColor: {
    type: String,
    default: '#757575'
  },
  bodyColor: {
    type: String,
    default: '#FFEBEE'
  },
  borderColor: {
    type: String,
    default: '#F44336'
  },
  lineColor: {
    type: String,
    default: '#F44336'
  },
  contentColor: {
    type: String,
    default: '#F44336'
  },
  warningBg: {
    type: String,
    default: '#FF9800'
  },
  warningColor: {
    type: String,
    default: 'white'
  }
})
</script>