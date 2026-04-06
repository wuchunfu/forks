<template>
  <div class="view-switcher">
    <n-radio-group :value="currentView" size="small" @update:value="handleViewChange">
      <n-radio-button
        v-for="view in views"
        :key="view.key"
        :value="view.key"
        :disabled="view.disabled"
      >
        <template #icon>
          <n-icon :size="16">
            <component :is="view.icon" />
          </n-icon>
        </template>
        {{ view.label }}
      </n-radio-button>
    </n-radio-group>

    <!-- 视图选项 -->
    <div class="view-options" v-if="showOptions">
      <n-space size="small">
        <!-- 卡片视图选项 -->
        <template v-if="currentView === 'card'">
          <n-select
            :value="cardSize"
            :options="cardSizeOptions"
            size="small"
            style="width: 120px"
            @update:value="handleCardSizeChange"
          />
        </template>

        <!-- 列表视图选项 -->
        <template v-if="currentView === 'list'">
          <n-button
            size="small"
            quaternary
            @click="showColumnSelector = true"
          >
            <template #icon>
              <n-icon>
                <SettingsOutline />
              </n-icon>
            </template>
            列设置
          </n-button>
        </template>

        <!-- 时间轴视图选项 -->
        <template v-if="currentView === 'timeline'">
          <n-select
            :value="timelineGroupBy"
            :options="timelineGroupOptions"
            size="small"
            style="width: 120px"
            @update:value="handleTimelineGroupChange"
          />
        </template>
      </n-space>
    </div>

    <!-- 列选择器对话框 -->
    <n-modal :show="showColumnSelector" preset="dialog" title="列显示设置" @update:show="showColumnSelector = $event">
      <div class="column-selector">
        <n-space vertical>
          <n-text depth="3">选择要在列表视图中显示的列：</n-text>
          <n-checkbox-group :value="selectedColumns" @update:value="selectedColumns = $event">
            <n-space vertical>
              <n-checkbox
                v-for="column in columnOptions"
                :key="column.key"
                :value="column.key"
                :label="column.label"
                :disabled="column.required"
              />
            </n-space>
          </n-checkbox-group>
        </n-space>
      </div>
      <template #action>
        <n-space>
          <n-button @click="resetColumns">重置</n-button>
          <n-button type="primary" @click="applyColumnSettings">应用</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup>
import { ref, computed, watch, h } from 'vue'
import {
  ListOutline,
  GridOutline,
  CalendarOutline,
  SettingsOutline
} from '@vicons/ionicons5'

const props = defineProps({
  modelValue: {
    type: String,
    default: 'list'
  },
  showOptions: {
    type: Boolean,
    default: true
  }
})

const emit = defineEmits([
  'update:modelValue',
  'view-change',
  'card-size-change',
  'timeline-group-change',
  'columns-change'
])

// 响应式数据
const currentView = ref(props.modelValue)
const cardSize = ref('medium')
const timelineGroupBy = ref('month')
const showColumnSelector = ref(false)
const selectedColumns = ref(['author', 'repo', 'stars', 'forks', 'created_at'])

// 视图配置
const views = [
  {
    key: 'list',
    label: '列表',
    icon: ListOutline,
    description: '简洁的列表展示'
  },
  {
    key: 'card',
    label: '卡片',
    icon: GridOutline,
    description: '丰富的卡片展示'
  },
  {
    key: 'timeline',
    label: '时间轴',
    icon: CalendarOutline,
    description: '按时间顺序展示'
  }
]

// 卡片大小选项
const cardSizeOptions = [
  { label: '紧凑', value: 'small' },
  { label: '标准', value: 'medium' },
  { label: '大', value: 'large' }
]

// 时间轴分组选项
const timelineGroupOptions = [
  { label: '按年', value: 'year' },
  { label: '按月', value: 'month' },
  { label: '按周', value: 'week' },
  { label: '按日', value: 'day' }
]

// 列选项
const columnOptions = [
  { key: 'author', label: '作者', required: true },
  { key: 'repo', label: '仓库名', required: true },
  { key: 'description', label: '描述' },
  { key: 'stars', label: '星标' },
  { key: 'forks', label: 'Fork' },
  { key: 'language', label: '语言' },
  { key: 'license', label: '许可证' },
  { key: 'created_at', label: '创建时间' },
  { key: 'status', label: '状态' },
  { key: 'size', label: '大小' }
]

// 计算属性
const viewConfig = computed(() => {
  return views.find(view => view.key === currentView.value) || views[0]
})

// 监听器
watch(() => props.modelValue, (newValue) => {
  currentView.value = newValue
})

// 方法
const handleViewChange = (value) => {
  currentView.value = value
  emit('update:modelValue', value)
  emit('view-change', value, viewConfig.value)
}

const handleCardSizeChange = (value) => {
  cardSize.value = value
  emit('card-size-change', value)
}

const handleTimelineGroupChange = (value) => {
  timelineGroupBy.value = value
  emit('timeline-group-change', value)
}

const resetColumns = () => {
  selectedColumns.value = ['author', 'repo', 'stars', 'forks', 'created_at']
}

const applyColumnSettings = () => {
  emit('columns-change', [...selectedColumns.value])
  showColumnSelector.value = false
}

// 暴露配置给父组件
defineExpose({
  cardSize,
  timelineGroupBy,
  selectedColumns
})
</script>

<style scoped>
.view-switcher {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 0;
  border-bottom: 1px solid var(--color-border);
  margin-bottom: 16px;
}

.view-options {
  margin-left: auto;
}

.column-selector {
  min-width: 300px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .view-switcher {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }

  .view-options {
    margin-left: 0;
    align-self: flex-end;
  }
}
</style>