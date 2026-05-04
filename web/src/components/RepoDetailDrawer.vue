<template>
  <n-drawer v-model:show="visible" :width="680" placement="right">
    <n-drawer-content title="仓库详细信息" closable>
      <div v-if="repo" class="repo-detail">
        <n-space vertical size="large">
          <!-- 失效提示 -->
          <n-alert v-if="repo.valid === 0" title="仓库已失效" type="error" :bordered="false">
            该仓库可能在远端已被删除、归档或不可访问。拉取和更新操作已隐藏，您可以右键选择「取消失效标记」恢复。
          </n-alert>

          <!-- 基本信息 -->
          <n-card title="基本信息" size="small">
            <n-descriptions :column="1" label-placement="left">
              <n-descriptions-item label="仓库名称">
                <n-text strong>{{ repo.author }}/{{ repo.repo }}</n-text>
              </n-descriptions-item>
              <n-descriptions-item label="描述">
                <n-text>{{ repo.description || '暂无描述' }}</n-text>
              </n-descriptions-item>
              <n-descriptions-item label="仓库地址">
                <n-button text @click="$emit('open-repo', repo.url)">
                  {{ repo.url }}
                </n-button>
              </n-descriptions-item>
              <n-descriptions-item v-if="isGithubRepo" label="快捷链接">
                <n-space size="small" :wrap="false">
                  <n-button
                    text
                    type="primary"
                    tag="a"
                    :href="'https://deepwiki.com/' + repo.author + '/' + repo.repo"
                    target="_blank"
                    rel="noopener"
                    size="small"
                  >
                    DeepWiki
                  </n-button>
                  <n-button
                    text
                    type="primary"
                    tag="a"
                    :href="'https://zread.ai/' + repo.author + '/' + repo.repo"
                    target="_blank"
                    rel="noopener"
                    size="small"
                  >
                    ZRead
                  </n-button>
                </n-space>
              </n-descriptions-item>
              <n-descriptions-item v-if="gitMirrorUrl" label="镜像克隆">
                <n-tooltip trigger="hover">
                  <template #trigger>
                    <n-space align="center" size="small" :wrap="false">
                      <n-text code class="git-mirror-url">{{ gitMirrorUrl }}</n-text>
                      <n-button size="tiny" quaternary @click="copyMirrorUrl">
                        <template #icon>
                          <n-icon size="14"><CopyOutline /></n-icon>
                        </template>
                      </n-button>
                    </n-space>
                  </template>
                  通过本服务加速克隆（需仓库已缓存到本地）
                </n-tooltip>
              </n-descriptions-item>
              <n-descriptions-item label="许可证">
                <n-tag v-if="repo.license" size="small">{{ repo.license }}</n-tag>
                <n-text v-else depth="3">未设置</n-text>
              </n-descriptions-item>
              <n-descriptions-item label="添加时间">
                <n-text>{{ formatDate(repo.created_at) }}</n-text>
              </n-descriptions-item>
              <n-descriptions-item v-if="repo.updated_at" label="更新时间">
                <n-text>{{ formatDate(repo.updated_at) }}</n-text>
              </n-descriptions-item>
            </n-descriptions>
          </n-card>

          <!-- 统计信息 -->
          <n-card title="统计信息" size="small">
            <n-space size="large">
              <n-statistic label="Stars" :value="repo.stars">
                <template #prefix>
                  <n-icon><StarOutline /></n-icon>
                </template>
              </n-statistic>
              <n-statistic label="Forks" :value="repo.forks">
                <template #prefix>
                  <n-icon><GitBranchOutline /></n-icon>
                </template>
              </n-statistic>
            </n-space>
          </n-card>

          <!-- 主题标签 -->
          <n-card v-if="repo.topics && parseTopics(repo.topics).length > 0" title="主题标签" size="small">
            <n-space size="small">
              <n-tag
                v-for="topic in parseTopics(repo.topics)"
                :key="topic"
                size="medium"
                round
                type="primary"
              >
                {{ topic }}
              </n-tag>
            </n-space>
          </n-card>

          <!-- 编程语言 -->
          <n-card v-if="repo.languages && parseLanguages(repo.languages).length > 0" title="编程语言" size="small">
            <n-space size="small">
              <n-tag
                v-for="language in parseLanguages(repo.languages)"
                :key="language"
                size="medium"
                type="info"
              >
                {{ language }}
              </n-tag>
            </n-space>
          </n-card>

          <!-- 操作按钮 -->
          <n-card title="操作" size="small">
            <RepoActions
              :repo="repo"
              :update-loading="updateLoading"
              size="medium"
              :show-detail-button="false"
              :show-code-button="showCodeButton"
              @open-repo="$emit('open-repo', $event)"
              @view-code="$emit('view-code', $event)"
              @update-info="$emit('update-info', $event)"
              @delete-repo="handleDeleteRepo"
              @toggle-valid="$emit('toggle-valid', $event)"
            />
          </n-card>
        </n-space>
      </div>
    </n-drawer-content>
  </n-drawer>
</template>

<script setup>
import { computed } from 'vue'
import {
  NDrawer, NDrawerContent, NSpace, NCard, NDescriptions, NDescriptionsItem,
  NText, NButton, NTag, NStatistic, NIcon, NTooltip, NAlert, useMessage
} from 'naive-ui'
import { StarOutline, GitBranchOutline, CopyOutline } from '@vicons/ionicons5'
import { copyToClipboard } from '../utils/clipboard'
import RepoActions from './RepoActions.vue'

// Props
const props = defineProps({
  show: {
    type: Boolean,
    default: false
  },
  repo: {
    type: Object,
    default: null
  },
  updateLoading: {
    type: Boolean,
    default: false
  },
  showCodeButton: {
    type: Boolean,
    default: true
  }
})

// Events
const emit = defineEmits(['update:show', 'open-repo', 'view-code', 'update-info', 'delete-repo', 'toggle-valid'])

const message = useMessage()

// Git 镜像克隆地址（从当前浏览器 URL 拼接）
const gitMirrorUrl = computed(() => {
  if (!props.repo || !props.repo.source || !props.repo.author || !props.repo.repo) return ''
  const origin = window.location.origin
  return `${origin}/git/${props.repo.source}/${props.repo.author}/${props.repo.repo}.git`
})

const isGithubRepo = computed(() => {
  if (!props.repo || !props.repo.url) return false
  return props.repo.url.includes('github.com')
})

// 复制镜像地址
const copyMirrorUrl = async () => {
  if (!gitMirrorUrl.value) return
  try {
    await copyToClipboard(gitMirrorUrl.value)
    message.success('已复制镜像克隆地址')
  } catch {
    message.error('复制失败')
  }
}

// Computed
const visible = computed({
  get: () => props.show,
  set: (value) => emit('update:show', value)
})

// Methods
const handleDeleteRepo = (id) => {
  emit('delete-repo', id)
  visible.value = false
}

const parseTopics = (topicsStr) => {
  if (!topicsStr) return []
  try {
    return JSON.parse(topicsStr)
  } catch {
    return []
  }
}

const parseLanguages = (languagesStr) => {
  if (!languagesStr) return []
  try {
    return JSON.parse(languagesStr)
  } catch {
    return []
  }
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  try {
    return new Date(dateStr).toLocaleDateString('zh-CN')
  } catch {
    return dateStr
  }
}
</script>

<style scoped>
.repo-detail {
  padding: 0;
}

.repo-detail .n-card {
  margin-bottom: 0;
}

.repo-detail .n-descriptions-item {
  padding: 8px 0;
}

.repo-detail .n-statistic {
  text-align: center;
}

.git-mirror-url {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  word-break: break-all;
}
</style>
