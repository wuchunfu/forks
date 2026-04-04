<template>
  <div
    class="repo-card"
    :class="{
      'repo-card--selected': selected,
      'repo-card--clone-success': repoStatus?.exists && repoStatus?.status === '已克隆',
      'repo-card--has-update': repoStatus?.hasBehind
    }"
    @click="$emit('click', repo)"
    @contextmenu="handleContextMenu"
  >
    <!-- 卡片头部 -->
    <div class="repo-card__header">
      <div class="repo-card__title">
        <n-ellipsis style="max-width: 200px">
          <strong>{{ repo.author }}</strong
          ><span class="repo-card__separator">/</span>{{ repo.repo }}
        </n-ellipsis>
        <n-button
          text
          size="tiny"
          @click.stop="toggleFavorite"
          class="repo-card__favorite"
        >
          <n-icon
            :color="isFavorite ? '#f5a623' : '#c8c8c8'"
            size="16"
          >
            <Star v-if="isFavorite" />
            <StarOutline v-else />
          </n-icon>
        </n-button>
      </div>

      <!-- 状态指示器 -->
      <div class="repo-card__status">
        <n-tag
          v-if="repoStatus?.exists"
          type="success"
          size="small"
          round
        >
          已克隆
        </n-tag>
        <n-tag
          v-if="repoStatus?.hasBehind"
          type="warning"
          size="small"
          round
        >
          有更新
        </n-tag>
        <n-tag
          v-if="repoStatus?.hasChanges"
          type="info"
          size="small"
          round
        >
          有变更
        </n-tag>
      </div>
    </div>

    <!-- 描述 -->
    <div class="repo-card__description">
      <n-ellipsis :line-clamp="2" style="line-height: 20px;">
        {{ repo.description || '暂无描述' }}
      </n-ellipsis>
    </div>

    <!-- 标签 -->
    <div class="repo-card__tags" v-if="hasTags">
      <n-tag
        v-for="tag in mainTags"
        :key="tag.name"
        :type="tag.type"
        size="small"
        round
      >
        {{ tag.name }}
      </n-tag>
      <n-dropdown
        v-if="remainingTags.length > 0"
        trigger="hover"
        :options="tagDropdownOptions"
      >
        <n-tag size="small" round>
          +{{ remainingTags.length }}
        </n-tag>
      </n-dropdown>
    </div>

    <!-- 统计信息 -->
    <div class="repo-card__stats">
      <div class="repo-card__stat">
        <n-icon size="14">
          <StarOutline />
        </n-icon>
        <span>{{ formatNumber(repo.stars) }}</span>
      </div>
      <div class="repo-card__stat">
        <n-icon size="14">
          <GitBranchOutline />
        </n-icon>
        <span>{{ formatNumber(repo.forks) }}</span>
      </div>
      <div class="repo-card__stat" v-if="repoStatus?.fileCount">
        <n-icon size="14">
          <DocumentTextOutline />
        </n-icon>
        <span>{{ formatNumber(repoStatus.fileCount) }}</span>
      </div>
      <div class="repo-card__stat" v-if="repoStatus?.repoSize">
        <n-icon size="14">
          <ServerOutline />
        </n-icon>
        <span>{{ repoStatus.repoSize }}</span>
      </div>
      <div class="repo-card__stat">
        <n-icon size="14">
          <TimeOutline />
        </n-icon>
        <span>{{ formatTime(repo.created_at) }}</span>
      </div>
    </div>

    <!-- 编程语言 -->
    <div class="repo-card__language" v-if="mainLanguage">
      <div
        class="repo-card__language-dot"
        :style="{ backgroundColor: getLanguageColor(mainLanguage) }"
      ></div>
      <span>{{ mainLanguage }}</span>
    </div>

    <!-- 快速操作按钮 -->
    <div class="repo-card__actions" @click.stop>
      <n-dropdown
        trigger="click"
        :options="actionMenuOptions"
        @select="handleAction"
      >
        <n-button circle size="small" quaternary>
          <template #icon>
            <n-icon>
              <EllipsisVerticalOutline />
            </n-icon>
          </template>
        </n-button>
      </n-dropdown>

      <n-button
        circle
        size="small"
        quaternary
        @click.stop="$emit('view-code', repo)"
      >
        <template #icon>
          <n-icon>
            <CodeSlashOutline />
          </n-icon>
        </template>
      </n-button>

      <n-button
        v-if="repoStatus?.exists"
        circle
        size="small"
        quaternary
        @click.stop="$emit('open-folder', repo)"
      >
        <template #icon>
          <n-icon>
            <FolderOutline />
          </n-icon>
        </template>
      </n-button>
    </div>

    <!-- 复选框（批量选择模式） -->
    <div v-if="batchMode" class="repo-card__checkbox" @click.stop>
      <n-checkbox
        :checked="selected"
        @update:checked="$emit('select', repo, $event)"
      />
    </div>
  </div>
</template>

<script setup>
import { computed, ref, h } from 'vue'
import {
  Star,
  StarOutline,
  GitBranchOutline,
  DocumentTextOutline,
  ServerOutline,
  TimeOutline,
  CodeSlashOutline,
  FolderOutline,
  EllipsisVerticalOutline,
  PlayOutline,
  DownloadOutline,
  RefreshOutline,
  TrashOutline,
  EyeOutline,
  ShareSocialOutline
} from '@vicons/ionicons5'
import { copyToClipboard as clipboardWrite } from '../../utils/clipboard'

const props = defineProps({
  repo: {
    type: Object,
    required: true
  },
  repoStatus: {
    type: Object,
    default: null
  },
  selected: {
    type: Boolean,
    default: false
  },
  batchMode: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits([
  'click',
  'select',
  'view-code',
  'open-folder',
  'clone',
  'pull',
  'reset',
  'delete',
  'update-info',
  'share'
])

// 响应式数据
const isFavorite = ref(false)

// 计算属性
const hasTags = computed(() => {
  return props.repo.topics || mainLanguage.value
})

const mainTags = computed(() => {
  const tags = []

  // 添加主题标签
  if (props.repo.topics) {
    try {
      const topics = JSON.parse(props.repo.topics)
      tags.push(...topics.slice(0, 2).map(topic => ({
        name: topic,
        type: 'info'
      })))
    } catch (e) {
      // 忽略解析错误
    }
  }

  return tags.slice(0, 2)
})

const remainingTags = computed(() => {
  if (!props.repo.topics) return []

  try {
    const topics = JSON.parse(props.repo.topics)
    return topics.slice(2)
  } catch (e) {
    return []
  }
})

const mainLanguage = computed(() => {
  if (props.repo.languages) {
    try {
      const languages = JSON.parse(props.repo.languages)
      return languages[0] // 取第一个主要语言
    } catch (e) {
      return null
    }
  }
  return null
})

const tagDropdownOptions = computed(() => {
  return remainingTags.value.map(tag => ({
    label: tag,
    key: tag,
    disabled: true
  }))
})

const actionMenuOptions = [
  {
    label: '克隆仓库',
    key: 'clone',
    icon: () => h('span', '⬇️'),
    show: !props.repoStatus?.exists
  },
  {
    label: '拉取更新',
    key: 'pull',
    icon: () => h('span', '🔄'),
    show: props.repoStatus?.exists
  },
  {
    label: '重置仓库',
    key: 'reset',
    icon: () => h('span', '🔁'),
    show: props.repoStatus?.exists
  },
  {
    type: 'divider',
    show: props.repoStatus?.exists
  },
  {
    label: '打开文件夹',
    key: 'open-folder',
    icon: () => h('span', '📁'),
    show: props.repoStatus?.exists
  },
  {
    label: '查看代码',
    key: 'view-code',
    icon: () => h('span', '💻')
  },
  {
    label: '更新信息',
    key: 'update-info',
    icon: () => h('span', '🔄')
  },
  {
    type: 'divider'
  },
  {
    label: '复制链接',
    key: 'copy-link',
    icon: () => h('span', '📋')
  },
  {
    label: '分享',
    key: 'share',
    icon: () => h('span', '🔗')
  },
  {
    type: 'divider'
  },
  {
    label: '删除仓库',
    key: 'delete',
    icon: () => h('span', '🗑️'),
    props: {
      style: 'color: #e74c3c;'
    }
  }
].filter(item => item.show !== false)

// 方法
const toggleFavorite = () => {
  isFavorite.value = !isFavorite.value
  // TODO: 保存收藏状态到本地存储或后端
}

const handleContextMenu = (event) => {
  event.preventDefault()
  // TODO: 显示上下文菜单
}

const handleAction = (key) => {
  switch (key) {
    case 'clone':
      emit('clone', props.repo)
      break
    case 'pull':
      emit('pull', props.repo)
      break
    case 'reset':
      emit('reset', props.repo)
      break
    case 'open-folder':
      emit('open-folder', props.repo)
      break
    case 'view-code':
      emit('view-code', props.repo)
      break
    case 'update-info':
      emit('update-info', props.repo)
      break
    case 'copy-link':
      copyToClipboard(props.repo.url)
      break
    case 'share':
      emit('share', props.repo)
      break
    case 'delete':
      emit('delete', props.repo)
      break
  }
}

const copyToClipboard = async (text) => {
  try {
    await clipboardWrite(text)
    console.log('链接已复制到剪贴板')
  } catch (err) {
    console.error('复制失败:', err)
  }
}

const formatNumber = (num) => {
  if (!num) return '0'
  if (num >= 1000) {
    return (num / 1000).toFixed(1) + 'k'
  }
  return num.toString()
}

const formatTime = (dateString) => {
  if (!dateString) return ''

  const date = new Date(dateString)
  const now = new Date()
  const diff = now - date

  const minutes = Math.floor(diff / (1000 * 60))
  const hours = Math.floor(diff / (1000 * 60 * 60))
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))

  if (minutes < 60) {
    return `${minutes}分钟前`
  } else if (hours < 24) {
    return `${hours}小时前`
  } else if (days < 30) {
    return `${days}天前`
  } else {
    return date.toLocaleDateString('zh-CN')
  }
}

const getLanguageColor = (language) => {
  const colors = {
    'JavaScript': '#f1e05a',
    'TypeScript': '#2b7489',
    'Python': '#3572A5',
    'Java': '#b07219',
    'Go': '#00ADD8',
    'Rust': '#dea584',
    'C++': '#f34b7d',
    'C#': '#239120',
    'PHP': '#4F5D95',
    'Ruby': '#701516',
    'Swift': '#ffac45',
    'Kotlin': '#F18E33',
    'Vue': '#4FC08D',
    'React': '#61DAFB',
    'HTML': '#e34c26',
    'CSS': '#1572b6',
    'Shell': '#89e051'
  }
  return colors[language] || '#666'
}
</script>

<style scoped>
.repo-card {
  position: relative;
  background: white;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  padding: 16px;
  cursor: pointer;
  transition: all 0.2s ease;
  height: 160px;
  display: flex;
  flex-direction: column;
  min-height: 160px;
  max-height: 160px;
}

.repo-card:hover {
  border-color: #18a058;
  box-shadow: 0 4px 12px rgba(24, 160, 88, 0.15);
  transform: translateY(-2px);
}

.repo-card--selected {
  border-color: #18a058;
  box-shadow: 0 0 0 2px rgba(24, 160, 88, 0.2);
}

.repo-card__header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 8px;
}

.repo-card__title {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 16px;
  font-weight: 600;
  flex: 1;
  min-width: 0;
}

.repo-card__separator {
  color: #999;
  margin: 0 2px;
}

.repo-card__favorite {
  opacity: 0.7;
  transition: opacity 0.2s;
}

.repo-card__favorite:hover {
  opacity: 1;
}

.repo-card__status {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}

.repo-card__description {
  color: #666;
  font-size: 14px;
  line-height: 20px;
  margin-bottom: 12px;
  flex: 1;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  max-height: 40px;
}

.repo-card__tags {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  margin-bottom: 12px;
}

.repo-card__stats {
  display: flex;
  gap: 16px;
  margin-bottom: 12px;
  font-size: 12px;
  color: #666;
}

.repo-card__stat {
  display: flex;
  align-items: center;
  gap: 4px;
}

.repo-card__language {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #666;
  margin-top: auto;
}

.repo-card__language-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.repo-card__actions {
  display: flex;
  justify-content: flex-end;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.2s;
  flex-shrink: 0;
  margin-top: auto;
}

.repo-card:hover .repo-card__actions {
  opacity: 1;
}

.repo-card__checkbox {
  position: absolute;
  top: 8px;
  left: 8px;
}

/* 状态样式 */
.repo-card--clone-success {
  border-left: 3px solid #52c41a;
}

.repo-card--has-update {
  border-left: 3px solid #faad14;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .repo-card {
    height: auto;
    min-height: 160px;
  }

  .repo-card__stats {
    flex-wrap: wrap;
    gap: 8px;
  }

  .repo-card__actions {
    opacity: 1;
    justify-content: flex-end;
  }
}
</style>