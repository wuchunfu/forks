import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useViewStore = defineStore('view', () => {
  // 状态
  const currentView = ref('list') // list, card, timeline
  const cardSize = ref('medium') // small, medium, large
  const gridColumns = ref(4)
  const timelineGroupBy = ref('month') // year, month, week, day

  // 列表视图配置
  const listColumns = ref([
    { key: 'author', label: '作者', visible: true, required: true, width: 120 },
    { key: 'repo', label: '仓库名', visible: true, required: true, width: 200 },
    { key: 'description', label: '描述', visible: true, required: false, width: 300 },
    { key: 'stars', label: '星标', visible: true, required: false, width: 80 },
    { key: 'forks', label: 'Fork', visible: true, required: false, width: 80 },
    { key: 'language', label: '语言', visible: false, required: false, width: 100 },
    { key: 'license', label: '许可证', visible: false, required: false, width: 100 },
    { key: 'created_at', label: '创建时间', visible: true, required: false, width: 120 },
    { key: 'status', label: '状态', visible: true, required: false, width: 100 },
    { key: 'size', label: '大小', visible: false, required: false, width: 80 }
  ])

  // 排序配置
  const sortBy = ref('created_at')
  const sortOrder = ref('desc')

  // 视图配置预设
  const viewPresets = {
    compact: {
      cardSize: 'small',
      gridColumns: 6,
      listColumns: ['author', 'repo', 'stars', 'forks', 'created_at']
    },
    comfortable: {
      cardSize: 'medium',
      gridColumns: 4,
      listColumns: ['author', 'repo', 'description', 'stars', 'forks', 'created_at', 'status']
    },
    detailed: {
      cardSize: 'large',
      gridColumns: 3,
      listColumns: ['author', 'repo', 'description', 'stars', 'forks', 'language', 'license', 'created_at', 'status', 'size']
    }
  }

  // 计算属性
  const visibleListColumns = computed(() => {
    return listColumns.value.filter(col => col.visible)
  })

  const listColumnKeys = computed(() => {
    return visibleListColumns.value.map(col => col.key)
  })

  const currentViewConfig = computed(() => {
    return {
      view: currentView.value,
      cardSize: cardSize.value,
      gridColumns: gridColumns.value,
      timelineGroupBy: timelineGroupBy.value,
      sortBy: sortBy.value,
      sortOrder: sortOrder.value,
      listColumns: listColumnKeys.value
    }
  })

  // 方法
  const setView = (view) => {
    currentView.value = view
    saveViewConfig()
  }

  const setCardSize = (size) => {
    cardSize.value = size
    saveViewConfig()
  }

  const setGridColumns = (columns) => {
    gridColumns.value = Math.min(6, Math.max(2, columns))
    saveViewConfig()
  }

  const setTimelineGroupBy = (groupBy) => {
    timelineGroupBy.value = groupBy
    saveViewConfig()
  }

  const setSort = (by, order = 'desc') => {
    sortBy.value = by
    sortOrder.value = order
    saveViewConfig()
  }

  const setListColumns = (columns) => {
    listColumns.value.forEach(col => {
      col.visible = columns.includes(col.key) || col.required
    })
    saveViewConfig()
  }

  const toggleListColumn = (columnKey) => {
    const column = listColumns.value.find(col => col.key === columnKey)
    if (column && !column.required) {
      column.visible = !column.visible
      saveViewConfig()
    }
  }

  const applyViewPreset = (presetName) => {
    const preset = viewPresets[presetName]
    if (preset) {
      cardSize.value = preset.cardSize
      gridColumns.value = preset.gridColumns
      setListColumns(preset.listColumns)
      saveViewConfig()
    }
  }

  const resetViewConfig = () => {
    currentView.value = 'list'
    cardSize.value = 'medium'
    gridColumns.value = 4
    timelineGroupBy.value = 'month'
    sortBy.value = 'created_at'
    sortOrder.value = 'desc'
    setListColumns(['author', 'repo', 'description', 'stars', 'forks', 'created_at', 'status'])
    saveViewConfig()
  }

  // 本地存储
  const saveViewConfig = () => {
    try {
      localStorage.setItem('forks-view-config', JSON.stringify(currentViewConfig.value))
    } catch (e) {
      console.warn('Failed to save view config:', e)
    }
  }

  const loadViewConfig = () => {
    try {
      const saved = localStorage.getItem('forks-view-config')
      if (saved) {
        const config = JSON.parse(saved)
        currentView.value = config.view || 'list'
        cardSize.value = config.cardSize || 'medium'
        gridColumns.value = config.gridColumns || 4
        timelineGroupBy.value = config.timelineGroupBy || 'month'
        sortBy.value = config.sortBy || 'created_at'
        sortOrder.value = config.sortOrder || 'desc'
        if (config.listColumns) {
          setListColumns(config.listColumns)
        }
      }
    } catch (e) {
      console.warn('Failed to load view config:', e)
    }
  }

  // 初始化时加载配置
  loadViewConfig()

  return {
    // 状态
    currentView,
    cardSize,
    gridColumns,
    timelineGroupBy,
    listColumns,
    sortBy,
    sortOrder,

    // 计算属性
    visibleListColumns,
    listColumnKeys,
    currentViewConfig,

    // 方法
    setView,
    setCardSize,
    setGridColumns,
    setTimelineGroupBy,
    setSort,
    setListColumns,
    toggleListColumn,
    applyViewPreset,
    resetViewConfig,
    saveViewConfig,
    loadViewConfig
  }
})