<template>
  <n-space>
    <n-button @click="$emit('open-repo', repo.url)" type="primary" :size="size">
      <template #icon>
        <n-icon><OpenOutline /></n-icon>
      </template>
      打开仓库
    </n-button>
    <n-button @click="$emit('view-code', repo)" type="info" :size="size" v-if="showCodeButton">
      <template #icon>
        <n-icon><CodeOutline /></n-icon>
      </template>
      查看代码
    </n-button>
    <n-button @click="$emit('update-info', repo)" type="warning" :loading="updateLoading" :size="size" v-if="repo.valid !== 0">
      <template #icon>
        <n-icon><RefreshOutline /></n-icon>
      </template>
      更新信息
    </n-button>
    <n-button @click="$emit('show-detail', repo)" type="default" :size="size" v-if="showDetailButton">
      <template #icon>
        <n-icon><InformationCircleOutline /></n-icon>
      </template>
      详情
    </n-button>
    <n-button v-if="repo.valid === 0" @click="$emit('toggle-valid', repo)" type="success" :size="size">
      <template #icon>
        <n-icon><CheckmarkCircleOutline /></n-icon>
      </template>
      取消失效标记
    </n-button>
    <n-button v-else @click="$emit('toggle-valid', repo)" type="default" :size="size">
      <template #icon>
        <n-icon><CloseCircleOutline /></n-icon>
      </template>
      标记为失效
    </n-button>
    <n-popconfirm @positive-click="$emit('delete-repo', repo)" negative-text="取消" positive-text="确认">
      <template #trigger>
        <n-button type="error" :size="size">
          <template #icon>
            <n-icon><TrashOutline /></n-icon>
          </template>
          删除
        </n-button>
      </template>
      确定删除这个仓库吗？
    </n-popconfirm>
  </n-space>
</template>

<script setup>
import { NSpace, NButton, NIcon, NPopconfirm } from 'naive-ui'
import {
  OpenOutline, CodeOutline, RefreshOutline, InformationCircleOutline, TrashOutline, CheckmarkCircleOutline, CloseCircleOutline
} from '@vicons/ionicons5'

// Props
defineProps({
  repo: {
    type: Object,
    required: true
  },
  updateLoading: {
    type: Boolean,
    default: false
  },
  size: {
    type: String,
    default: 'medium'
  },
  showDetailButton: {
    type: Boolean,
    default: true
  },
  showCodeButton: {
    type: Boolean,
    default: true
  }
})

// Events
defineEmits(['open-repo', 'view-code', 'update-info', 'show-detail', 'delete-repo', 'toggle-valid'])
</script>