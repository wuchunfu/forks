<template>
  <div class="code-viewer">
    <!-- 操作区域 -->
    <div v-if="showOperations" class="operations-area">
      <n-tabs type="line" size="small" animated>
        <n-tab-pane name="git" tab="Git 操作">
          <div class="operations-content">
            <n-space size="medium">
              <n-card size="small" style="width: 300px;">
                <template #header>
                  <n-space align="center">
                    <n-icon><GitBranchOutline /></n-icon>
                    <span>版本控制</span>
                  </n-space>
                </template>
                <n-space vertical size="small">
                  <n-button v-if="!repoExists" @click="cloneRepository" :loading="operations.cloning" block>
                    <template #icon>
                      <n-icon><DownloadOutline /></n-icon>
                    </template>
                    克隆仓库到本地
                  </n-button>
                  <n-button @click="pullUpdates" :loading="operations.pulling" :disabled="!repoExists" block>
                    <template #icon>
                      <n-icon><RefreshOutline /></n-icon>
                    </template>
                    拉取最新更新
                  </n-button>
                  <n-button @click="checkStatus" :loading="operations.checking" block>
                    <template #icon>
                      <n-icon><InformationCircleOutline /></n-icon>
                    </template>
                    检查仓库状态
                  </n-button>
                </n-space>
              </n-card>

              <n-card size="small" style="width: 300px;">
                <template #header>
                  <n-space align="center">
                    <n-icon><FolderOutline /></n-icon>
                    <span>本地操作</span>
                  </n-space>
                </template>
                <n-space vertical size="small">
                  <n-button @click="refreshFileTree" block>
                    <template #icon>
                      <n-icon><RefreshOutline /></n-icon>
                    </template>
                    刷新文件树
                  </n-button>
                  <n-popconfirm @positive-click="deleteLocalRepo" negative-text="取消" positive-text="确认">
                    <template #trigger>
                      <n-button type="error" :disabled="!repoExists" block>
                        <template #icon>
                          <n-icon><TrashOutline /></n-icon>
                        </template>
                        删除本地仓库
                      </n-button>
                    </template>
                    确定删除本地仓库文件吗？此操作不可恢复。
                  </n-popconfirm>
                </n-space>
              </n-card>
            </n-space>
          </div>
        </n-tab-pane>

        <n-tab-pane name="repo" tab="仓库操作">
          <div class="operations-content">
            <n-card size="small" v-if="repoInfo">
              <template #header>
                <n-space align="center">
                  <n-icon><StarOutline /></n-icon>
                  <span>仓库管理</span>
                </n-space>
              </template>
              <RepoActions
                :repo="repoInfo"
                :update-loading="operations.updating"
                size="medium"
                :show-detail-button="false"
                :show-code-button="false"
                @open-repo="openRepo"
                @update-info="updateRepoInfo"
                @delete-repo="deleteRepository"
                @toggle-valid="handleToggleValid"
              />
            </n-card>
          </div>
        </n-tab-pane>

        <n-tab-pane name="status" tab="状态信息">
          <div class="operations-content">
            <n-card size="small">
              <template #header>
                <n-space align="center">
                  <n-icon><RadioButtonOnOutline /></n-icon>
                  <span>仓库状态</span>
                </n-space>
              </template>
              <div v-if="repoStatus">
                <n-descriptions :column="2" label-placement="left">
                  <n-descriptions-item label="本地状态">
                    <n-tag :type="repoExists ? 'success' : 'warning'" size="small">
                      {{ repoExists ? '已存在' : '未克隆' }}
                    </n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item label="当前分支" v-if="repoStatus.branch">
                    <n-tag type="info" size="small">{{ repoStatus.branch }}</n-tag>
                  </n-descriptions-item>
                  <n-descriptions-item label="最后更新" v-if="repoStatus.lastUpdate">
                    <n-text>{{ formatTime(repoStatus.lastUpdate) }}</n-text>
                  </n-descriptions-item>
                  <n-descriptions-item label="文件数量" v-if="repoStatus.fileCount">
                    <n-text>{{ repoStatus.fileCount }} 个文件</n-text>
                  </n-descriptions-item>
                </n-descriptions>
              </div>
              <div v-else>
                <n-empty description="暂无状态信息" size="small">
                  <template #extra>
                    <n-button size="small" @click="checkStatus">检查状态</n-button>
                  </template>
                </n-empty>
              </div>
            </n-card>
          </div>
        </n-tab-pane>
      </n-tabs>
    </div>

    <!-- 操作状态面板 -->
    <div class="status-panel" v-show="showStatusPanel">
      <n-card size="small">
        <template #header>
          <n-space justify="space-between" align="center">
            <n-space align="center">
              <n-icon><Terminal /></n-icon>
              <span>操作日志</span>
              <n-badge :value="operationLogs.length" :max="99" />
            </n-space>
            <n-space>
              <n-button size="small" @click="clearLogs" type="error" text>
                <template #icon>
                  <n-icon><TrashOutline /></n-icon>
                </template>
                清空
              </n-button>
              <n-button size="small" @click="showStatusPanel = false" text>
                <template #icon>
                  <n-icon><CloseOutline /></n-icon>
                </template>
              </n-button>
            </n-space>
          </n-space>
        </template>
        
        <div class="operation-logs">
          <n-scrollbar style="max-height: 300px;">
            <div v-if="operationLogs.length === 0" class="no-logs">
              <n-empty size="small" description="暂无操作日志" />
            </div>
            <div v-else>
              <div 
                v-for="(log, index) in operationLogs" 
                :key="index" 
                class="log-item"
                :class="`log-${log.type}`"
              >
                <div class="log-header">
                  <n-space align="center" size="small">
                    <n-icon v-if="log.type === 'info'" color="#18a058"><InformationCircleOutline /></n-icon>
                    <n-icon v-else-if="log.type === 'success'" color="#18a058"><CheckmarkCircleOutline /></n-icon>
                    <n-icon v-else-if="log.type === 'error'" color="#d03050"><CloseCircleOutline /></n-icon>
                    <n-icon v-else-if="log.type === 'warning'" color="#f0a020"><WarningOutline /></n-icon>
                    <n-text :depth="3" style="font-size: 12px;">{{ formatTime(log.timestamp) }}</n-text>
                  </n-space>
                </div>
                <div class="log-content">
                  <n-text :type="log.type === 'error' ? 'error' : log.type === 'success' ? 'success' : 'default'">
                    {{ log.message }}
                  </n-text>
                </div>
              </div>
            </div>
          </n-scrollbar>
        </div>

        <!-- 当前操作状态 -->
        <div v-if="currentOperation" class="current-operation">
          <n-divider />
          <n-space align="center">
            <n-spin size="small" :show="currentOperation.loading" />
            <n-text>{{ currentOperation.text }}</n-text>
            <n-progress 
              v-if="currentOperation.progress !== undefined" 
              :percentage="currentOperation.progress" 
              :show-indicator="false"
              style="width: 100px;"
            />
          </n-space>
        </div>

        <!-- 仓库状态概览 -->
        <div v-if="repoStatus && repoStatus.exists" class="repo-status-overview">
          <n-divider />
          <div class="status-header">
            <n-space align="center" size="small">
              <n-icon size="16" color="#18a058"><CheckmarkCircleOutline /></n-icon>
              <n-text strong>仓库状态概览</n-text>
            </n-space>
          </div>
          <div class="status-grid">
            <div class="status-item">
              <n-space align="center" size="small">
                <n-icon size="14"><GitBranchOutline /></n-icon>
                <n-text depth="3" style="font-size: 12px;">分支:</n-text>
                <n-tag size="small" type="primary">{{ repoStatus.branch }}</n-tag>
              </n-space>
            </div>
            <div class="status-item" v-if="repoStatus.hasChanges">
              <n-space align="center" size="small">
                <n-icon size="14" color="#f0a020"><WarningOutline /></n-icon>
                <n-text depth="3" style="font-size: 12px;">未提交:</n-text>
                <n-tag size="small" type="warning">{{ repoStatus.changes?.length || 0 }}</n-tag>
              </n-space>
            </div>
            <div class="status-item" v-if="repoStatus.hasUnpushed">
              <n-space align="center" size="small">
                <n-icon size="14" color="#f0a020"><ArrowUpOutline /></n-icon>
                <n-text depth="3" style="font-size: 12px;">未推送:</n-text>
                <n-tag size="small" type="warning">{{ repoStatus.unpushedCount || 0 }}</n-tag>
              </n-space>
            </div>
            <div class="status-item" v-if="repoStatus.hasBehind">
              <n-space align="center" size="small">
                <n-icon size="14" color="#2080f0"><ArrowDownOutline /></n-icon>
                <n-text depth="3" style="font-size: 12px;">待拉取:</n-text>
                <n-tag size="small" type="info">{{ repoStatus.behindCount || 0 }}</n-tag>
              </n-space>
            </div>
          </div>
        </div>
      </n-card>
    </div>

    <div class="viewer-main">
      <!-- 左侧：文件资源管理器 -->
      <div class="file-explorer">
        <div class="explorer-header">
          <span class="explorer-title">文件</span>
          <n-space>
            <n-button size="small" text @click="showStatusPanel = !showStatusPanel" :type="showStatusPanel ? 'primary' : 'default'">
              <template #icon>
                <n-icon><Terminal /></n-icon>
              </template>
            </n-button>
            <n-button size="small" text @click="toggleOperations">
              <template #icon>
                <n-icon><SettingsOutline /></n-icon>
              </template>
            </n-button>
          </n-space>
        </div>
        <div class="explorer-content">
          <div v-if="loading" class="loading-container">
            <n-spin size="small">
              <template #description>
                加载中...
              </template>
            </n-spin>
          </div>
          <div v-else-if="treeData.length === 0" class="empty-container">
            <div class="empty-text">暂无文件</div>
          </div>
          <div v-else class="file-tree">
            <NTree
              :data="treeDataComputed"
              :selected-keys="[currentFile]"
              :render-label="renderLabel"
              :render-switcher-icon="renderSwitcherIcon"
              @update:selected-keys="onSelectFile"
              @update:expanded-keys="updateExpandedKeys"
            />
          </div>
        </div>
      </div>
      
      <!-- 右侧：代码预览区域 -->
      <div class="code-container">
        <div class="code-content">
          <div v-if="!currentFile" class="no-file-selected">
            <div class="no-file-icon">
              <n-icon size="48" color="#ccc">
                <Document />
              </n-icon>
            </div>
            <div class="no-file-text">请选择左侧文件进行查看</div>
          </div>
          <div v-else class="file-viewer">
            <div class="file-header">
              <div class="file-info">
                <span class="file-name">{{ currentFileName }}</span>
              </div>
            </div>
            <div class="file-content">
              <!-- Markdown 渲染视图 -->
              <div v-if="isMarkdownFile" class="markdown-viewer">
                <div class="markdown-content" v-html="renderedMarkdown"></div>
              </div>
              <!-- 图片查看器视图 -->
              <div v-else-if="isImageFile" class="image-viewer">
                <div class="image-container">
                  <!-- 加载状态 -->
                  <div v-if="imageLoading" class="image-loading">
                    <n-spin size="large">
                      <template #description>
                        <n-text depth="3">正在加载图片...</n-text>
                      </template>
                    </n-spin>
                  </div>
                  
                  <!-- 图片内容 -->
                  <img 
                    v-else-if="imageUrl"
                    :src="imageUrl" 
                    :alt="currentFileName"
                    class="image-content"
                    @load="onImageLoad"
                    @error="onImageError"
                  />
                  
                  <!-- 加载失败状态 -->
                  <div v-else-if="!imageLoading" class="image-error">
                    <div class="error-content">
                      <p>图片加载失败</p>
                      <p class="error-filename">{{ currentFileName }}</p>
                      <n-button size="small" @click="loadImage" type="primary">重新加载</n-button>
                    </div>
                  </div>
                  
                  <div class="image-info">
                    <n-space align="center" size="small">
                      <n-text depth="3">{{ currentFileName }}</n-text>
                      <n-text depth="3" style="font-size: 12px;" id="imageDimensions"></n-text>
                    </n-space>
                  </div>
                </div>
              </div>
              <!-- 音频播放器视图 -->
              <div v-else-if="isAudioFile" class="media-viewer">
                <div class="media-container">
                  <div class="media-info">
                    <n-space align="center" size="small">
                      <n-icon size="24"><VolumeHighOutline /></n-icon>
                      <n-text strong>{{ currentFileName }}</n-text>
                    </n-space>
                  </div>
                  <audio 
                    :src="audioUrl" 
                    controls
                    class="audio-player"
                    @loadedmetadata="onAudioLoaded"
                    @error="onMediaError"
                  >
                    您的浏览器不支持音频播放
                  </audio>
                  <div class="media-details" id="audioDetails">
                    <n-text depth="3" style="font-size: 12px;">音频文件</n-text>
                  </div>
                </div>
              </div>
              <!-- 视频播放器视图 -->
              <div v-else-if="isVideoFile" class="media-viewer">
                <div class="media-container">
                  <div class="media-info">
                    <n-space align="center" size="small">
                      <n-icon size="24"><PlayCircleOutline /></n-icon>
                      <n-text strong>{{ currentFileName }}</n-text>
                    </n-space>
                  </div>
                  <video 
                    :src="videoUrl" 
                    controls
                    class="video-player"
                    @loadedmetadata="onVideoLoaded"
                    @error="onMediaError"
                  >
                    您的浏览器不支持视频播放
                  </video>
                  <div class="media-details" id="videoDetails">
                    <n-text depth="3" style="font-size: 12px;">视频文件</n-text>
                  </div>
                </div>
              </div>
              <!-- 二进制文件视图 -->
              <div v-else-if="isBinaryFile" class="unsupported-viewer">
                <div class="unsupported-container">
                  <div class="unsupported-icon">
                    <n-icon size="48" color="#8b949e"><DocumentOutline /></n-icon>
                  </div>
                  <div class="unsupported-info">
                    <n-text strong style="font-size: 16px;">{{ currentFileName }}</n-text>
                    <n-text depth="3" style="font-size: 14px; margin-top: 8px;">
                      {{ getBinaryFileDescription() }}
                    </n-text>
                  </div>
                  <div class="unsupported-message">
                    <n-space align="center" size="small">
                      <n-icon size="18" color="#f0a020"><AlertCircleOutline /></n-icon>
                      <n-text depth="3" style="font-size: 13px;">暂不支持预览此类型文件</n-text>
                    </n-space>
                  </div>
                  <div class="unsupported-actions">
                    <n-space size="medium">
                      <n-button size="small" @click="downloadFile" type="primary">
                        <template #icon>
                          <n-icon><DownloadOutline /></n-icon>
                        </template>
                        下载文件
                      </n-button>
                      <n-button size="small" @click="openWithCode" secondary>
                        <template #icon>
                          <n-icon><DocumentOutline /></n-icon>
                        </template>
                        强制以文本查看
                      </n-button>
                    </n-space>
                  </div>
                </div>
              </div>
              <!-- 代码编辑器视图 -->
              <div v-else class="code-editor">
                <div class="codemirror-container" ref="codeContainer"></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup>
import { ref, onMounted, computed, h, nextTick, onBeforeUnmount, watch } from 'vue'
import { useRoute } from 'vue-router'
import { 
  useMessage, NIcon, NTree, NSpin, NButton, NCard, NSpace, 
  NPopconfirm, NText, NTabs, NTabPane, NDescriptions, 
  NDescriptionsItem, NTag, NEmpty, NBadge, NScrollbar,
  NDivider, NProgress
} from 'naive-ui'
import { 
  Document, 
  Folder,
  FolderOpenOutline,
  FileTrayFullOutline,
  ChevronForward,
  SettingsOutline,
  GitBranchOutline,
  DownloadOutline,
  RefreshOutline,
  InformationCircleOutline,
  FolderOutline,
  TrashOutline,
  RadioButtonOnOutline,
  Terminal,
  CloseOutline,
  CheckmarkCircleOutline,
  CloseCircleOutline,
  WarningOutline,
  ArrowUpOutline,
  ArrowDownOutline,
  ReloadOutline,
  VolumeHighOutline,
  PlayCircleOutline,
  DocumentOutline,
  AlertCircleOutline
} from '@vicons/ionicons5'
import request from '@/utils/request'
import { marked } from 'marked'

// CodeMirror 核心模块
import { EditorView, lineNumbers, highlightActiveLineGutter } from '@codemirror/view'
import { EditorState } from '@codemirror/state'
import { defaultHighlightStyle, syntaxHighlighting } from '@codemirror/language'

// Dracula 主题
import { dracula } from '@uiw/codemirror-theme-dracula'

// 语言支持
import { javascript } from '@codemirror/lang-javascript'
import { html } from '@codemirror/lang-html'
import { css } from '@codemirror/lang-css'
import { json } from '@codemirror/lang-json'
import { markdown } from '@codemirror/lang-markdown'
import { python } from '@codemirror/lang-python'
import { java } from '@codemirror/lang-java'
import { cpp } from '@codemirror/lang-cpp'
import { rust } from '@codemirror/lang-rust'
import { go } from '@codemirror/lang-go'
import { sql } from '@codemirror/lang-sql'
import { xml } from '@codemirror/lang-xml'
import { yaml } from '@codemirror/lang-yaml'
import { vue } from '@codemirror/lang-vue'
import RepoActions from '@/components/RepoActions.vue'
import RepoDetailDrawer from '@/components/RepoDetailDrawer.vue'

// 获取 API 基础 URL
function getApiBaseUrl() {
  const apiUrl = import.meta.env.VITE_API_URL

  // 只有当环境变量有效且不是 '/' 或空字符串时才使用
  if (apiUrl && apiUrl !== '/' && apiUrl !== '') {
    return apiUrl
  }

  if (import.meta.env.DEV) {
    const backendPort = import.meta.env.VITE_API_PORT || '8080'
    return `http://localhost:${backendPort}`
  }
  return window.location.origin
}

const route = useRoute()
const message = useMessage()
const emit = defineEmits(['repo-info-loaded'])

// 基础数据
const repoId = ref(route.params.id)
const repoInfo = ref(null)
const treeData = ref([])
const loading = ref(false)
const currentFile = ref('')
const currentFileName = ref('')
const currentFileContent = ref('')
const forceTextMode = ref(false) // 强制文本模式标志

// 操作面板相关
const showOperations = ref(false)
const repoExists = ref(false)
const repoStatus = ref(null)

const operations = ref({
  cloning: false,
  pulling: false,
  checking: false,
  updating: false
})

// 状态面板相关
const showStatusPanel = ref(false)
const operationLogs = ref([])
const currentOperation = ref(null)

// 日志方法
const addLog = (message, type = 'info') => {
  operationLogs.value.push({
    message,
    type,
    timestamp: new Date()
  })
  
  // 保持最多100条日志
  if (operationLogs.value.length > 100) {
    operationLogs.value.shift()
  }
}

const clearLogs = () => {
  operationLogs.value = []
}

// formatTime函数已在下面定义，这里删除重复定义

const setCurrentOperation = (text, loading = true, progress = undefined) => {
  currentOperation.value = { text, loading, progress }
}

const clearCurrentOperation = () => {
  currentOperation.value = null
}

// CodeMirror 相关
const codeContainer = ref(null)
let editorView = null

// 语言映射
const languageMap = {
  'js': javascript(),
  'javascript': javascript(),
  'ts': javascript({typescript: true}),
  'typescript': javascript({typescript: true}),
  'jsx': javascript({jsx: true}),
  'tsx': javascript({typescript: true, jsx: true}),
  'vue': vue(),
  'html': html(),
  'htm': html(),
  'css': css(),
  'scss': css(),
  'sass': css(),
  'less': css(),
  'json': json(),
  'md': markdown(),
  'markdown': markdown(),
  'py': python(),
  'python': python(),
  'java': java(),
  'cpp': cpp(),
  'cc': cpp(),
  'cxx': cpp(),
  'c': cpp(),
  'rs': rust(),
  'rust': rust(),
  'go': go(),
  'sql': sql(),
  'xml': xml(),
  'yaml': yaml(),
  'yml': yaml()
}

function getLanguageExtension(filename) {
  const ext = filename.split('.').pop()?.toLowerCase()
  return languageMap[ext] || null
}

// 展开状态
const expandedKeys = ref(new Set())

// 转换树数据为NTree格式
const treeDataComputed = computed(() => {
  return treeToNaive(treeData.value)
})

// Markdown相关计算属性
const isMarkdownFile = computed(() => {
  if (!currentFileName.value) return false
  const lowerName = currentFileName.value.toLowerCase()
  return lowerName.endsWith('.md') || lowerName.endsWith('.markdown')
})

const renderedMarkdown = computed(() => {
  if (!isMarkdownFile.value || !currentFileContent.value) return ''
  try {
    return marked(currentFileContent.value)
  } catch (error) {
    console.error('Markdown rendering error:', error)
    return '<p>Markdown 渲染失败</p>'
  }
})

// 图片相关计算属性
const isImageFile = computed(() => {
  if (!currentFileName.value) return false
  const lowerName = currentFileName.value.toLowerCase()
  const imageExtensions = ['.jpg', '.jpeg', '.png', '.gif', '.svg', '.bmp', '.webp', '.ico']
  return imageExtensions.some(ext => lowerName.endsWith(ext))
})

const imageUrl = ref('')
const imageLoading = ref(false)

// 加载图片数据并生成blob URL
const loadImage = async () => {
  if (!isImageFile.value || !repoId.value || !currentFile.value) return
  
  imageLoading.value = true
  try {
    const baseUrl = getApiBaseUrl().replace(/\/$/, '')
    const apiUrl = `${baseUrl}/api/repos/${repoId.value}/file-content?path=${encodeURIComponent(currentFile.value)}&type=blob`
    console.log('请求图片URL:', apiUrl)
    
    const token = localStorage.getItem('token')
    const headers = {}
    if (token) {
      headers.Authorization = `Bearer ${token}`
    }
    
    const response = await fetch(apiUrl, { headers })
    
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`)
    }
    
    const blob = await response.blob()
    const blobUrl = URL.createObjectURL(blob)
    imageUrl.value = blobUrl
    
    console.log('图片加载成功，blob URL:', blobUrl)
  } catch (error) {
    console.error('图片加载失败:', error)
    imageUrl.value = '' // 清空URL，触发错误显示
  } finally {
    imageLoading.value = false
  }
}

// 监听图片文件变化，自动加载图片
watch([isImageFile, currentFile], ([isImage, file]) => {
  // 清理之前的blob URL
  if (imageUrl.value && imageUrl.value.startsWith('blob:')) {
    URL.revokeObjectURL(imageUrl.value)
    imageUrl.value = ''
  }
  
  if (isImage && file) {
    loadImage()
  }
}, { immediate: true })

// 组件卸载时清理blob URL
onBeforeUnmount(() => {
  if (imageUrl.value && imageUrl.value.startsWith('blob:')) {
    URL.revokeObjectURL(imageUrl.value)
  }
})

// 音频相关计算属性
const isAudioFile = computed(() => {
  if (!currentFileName.value) return false
  const lowerName = currentFileName.value.toLowerCase()
  const audioExtensions = ['.mp3', '.wav', '.ogg', '.aac', '.flac', '.m4a', '.wma']
  return audioExtensions.some(ext => lowerName.endsWith(ext))
})

const audioUrl = computed(() => {
  if (!isAudioFile.value || !repoId.value || !currentFile.value) return ''
  const baseUrl = getApiBaseUrl().replace(/\/$/, '')
  return `${baseUrl}/api/repos/${repoId.value}/file-content?path=${encodeURIComponent(currentFile.value)}&type=blob`
})

// 视频相关计算属性
const isVideoFile = computed(() => {
  if (!currentFileName.value) return false
  const lowerName = currentFileName.value.toLowerCase()
  const videoExtensions = ['.mp4', '.webm', '.avi', '.mov', '.mkv', '.wmv', '.flv', '.m4v']
  return videoExtensions.some(ext => lowerName.endsWith(ext))
})

const videoUrl = computed(() => {
  if (!isVideoFile.value || !repoId.value || !currentFile.value) return ''
  const baseUrl = getApiBaseUrl().replace(/\/$/, '')
  return `${baseUrl}/api/repos/${repoId.value}/file-content?path=${encodeURIComponent(currentFile.value)}&type=blob`
})

// 二进制文件相关计算属性
const isBinaryFile = computed(() => {
  if (!currentFileName.value) return false
  
  // 如果强制文本模式，不显示为二进制文件
  if (forceTextMode.value) return false
  
  const lowerName = currentFileName.value.toLowerCase()
  
  // 已经支持的文件类型不需要显示为二进制
  if (isMarkdownFile.value || isImageFile.value || isAudioFile.value || isVideoFile.value) {
    return false
  }
  
  // 常见的文本文件扩展名
  const textExtensions = [
    '.txt', '.md', '.json', '.xml', '.yaml', '.yml', '.ini', '.cfg', '.conf',
    '.js', '.ts', '.jsx', '.tsx', '.vue', '.html', '.htm', '.css', '.scss', '.less',
    '.py', '.java', '.cpp', '.c', '.h', '.cs', '.php', '.rb', '.go', '.rs', '.swift',
    '.sh', '.bat', '.ps1', '.sql', '.r', '.m', '.pl', '.lua', '.dart', '.kt',
    '.dockerfile', '.gitignore', '.gitattributes', '.editorconfig',
    '.log', '.csv', '.env', '.properties', '.toml'
  ]
  
  // 如果是已知的文本文件扩展名，不是二进制文件
  if (textExtensions.some(ext => lowerName.endsWith(ext))) {
    return false
  }
  
  // 常见的二进制文件扩展名
  const binaryExtensions = [
    '.exe', '.dll', '.so', '.dylib', '.app', '.deb', '.rpm', '.msi',
    '.zip', '.rar', '.7z', '.tar', '.gz', '.bz2', '.xz',
    '.pdf', '.doc', '.docx', '.xls', '.xlsx', '.ppt', '.pptx',
    '.bin', '.dat', '.db', '.sqlite', '.sqlite3',
    '.ttf', '.otf', '.woff', '.woff2', '.eot',
    '.psd', '.ai', '.sketch', '.fig',
    '.class', '.jar', '.war', '.ear',
    '.o', '.obj', '.lib', '.a'
  ]
  
  return binaryExtensions.some(ext => lowerName.endsWith(ext))
})

const getFileType = computed(() => {
  if (!currentFileName.value) return 'unknown'
  
  if (isMarkdownFile.value) return 'markdown'
  if (isImageFile.value) return 'image'
  if (isAudioFile.value) return 'audio'
  if (isVideoFile.value) return 'video'
  if (isBinaryFile.value) return 'binary'
  return 'text'
})

function treeToNaive(tree) {
  if (!Array.isArray(tree)) return []
  
  // 排序逻辑：目录在前，文件在后，同类型按名称排序
  const customSort = (a, b) => {
    if ((b.isDirectory || 0) - (a.isDirectory || 0) !== 0) {
      return (b.isDirectory || 0) - (a.isDirectory || 0)
    }
    const nameA = (a.fileName || a.label || '').toLowerCase()
    const nameB = (b.fileName || b.label || '').toLowerCase()
    return nameA.localeCompare(nameB)
  }
  
  const sorted = [...tree].sort(customSort)
  
  return sorted.map(node => {
    const nodeKey = node.key || node.id
    const isExpanded = expandedKeys.value.has(String(nodeKey))
    
    const getDisplayName = (node) => {
      return node.fileName || node.name || ''
    }
    
    return {
      label: getDisplayName(node),
      key: nodeKey,
      isLeaf: node.isDirectory !== 1,
      filePath: node.filePath,
      fileName: node.fileName,
      prefix: () => h(NIcon, null, { 
        default: () => h(node.isDirectory === 1 ? (isExpanded ? FolderOpenOutline : Folder) : FileTrayFullOutline)
      }),
      children: node.children ? treeToNaive(node.children) : []
    }
  })
}

// 加载仓库信息
const loadRepoInfo = async () => {
  try {
    const response = await request.get(`/api/repos/${repoId.value}`)
    repoInfo.value = response.data.data
    // 用 is_cloned 初始化 repoExists，避免页面闪烁
    if (repoInfo.value && repoInfo.value.is_cloned === 1) {
      repoExists.value = true
    }
    // 向父组件传递仓库信息
    emit('repo-info-loaded', repoInfo.value)
  } catch (error) {
    console.error('加载仓库信息失败:', error)
    message.error('加载仓库信息失败')
  }
}

// 加载文件树
const loadFileTree = async () => {
  loading.value = true
  try {
    const response = await request.get(`/api/repos/${repoId.value}/files`)
    treeData.value = response.data.data.tree || []
  } catch (error) {
    console.error('加载文件树失败:', error)
    message.error('加载文件树失败')
    treeData.value = []
  } finally {
    loading.value = false
  }
}

// 选择文件
const onSelectFile = async (keys) => {
  if (!keys || keys.length === 0) return
  
  // 重置强制文本模式标志
  forceTextMode.value = false
  
  const selectedKey = keys[0]
  
  // 从树数据中查找对应的节点
  const findNodeByKey = (nodes, key) => {
    for (const node of nodes) {
      if (node.key === key || node.id === key) {
        return node
      }
      if (node.children) {
        const found = findNodeByKey(node.children, key)
        if (found) return found
      }
    }
    return null
  }
  
  const selectedNode = findNodeByKey(treeData.value, selectedKey)
  if (!selectedNode || selectedNode.isDirectory === 1) {
    return
  }
  
  // 加载文件内容
  try {
    const response = await request.get(`/api/repos/${repoId.value}/file-content`, {
      params: { path: selectedNode.filePath }
    })
    
    currentFile.value = selectedKey
    currentFileName.value = selectedNode.fileName
    currentFileContent.value = response.data.data.content
    
    // 等待DOM更新后创建或更新CodeMirror编辑器
    await nextTick()
    createOrUpdateEditor()
  } catch (error) {
    console.error('加载文件内容失败:', error)
    message.error('加载文件内容失败')
  }
}

// 创建或更新CodeMirror编辑器
const createOrUpdateEditor = async () => {
  if (!codeContainer.value) return
  
  // 销毁现有编辑器
  if (editorView) {
    editorView.destroy()
    editorView = null
  }
  
  // 获取语言扩展
  const languageExt = currentFileName.value ? getLanguageExtension(currentFileName.value) : null
  
  // 创建编辑器扩展
  const extensions = [
    dracula,
    EditorView.editable.of(false),
    lineNumbers(),
    highlightActiveLineGutter(),
    syntaxHighlighting(defaultHighlightStyle),
    EditorView.scrollMargins.of(() => ({ top: 10, bottom: 10 })),
    EditorView.theme({
      "&": { height: "100%" },
      ".cm-scroller": { 
        overflow: "auto !important",
        height: "100% !important"
      }
    })
  ]
  
  // 添加语言支持
  if (languageExt) {
    extensions.push(languageExt)
  }
  
  // 创建编辑器状态
  const state = EditorState.create({
    doc: currentFileContent.value,
    extensions
  })
  
  // 创建编辑器视图
  editorView = new EditorView({
    state,
    parent: codeContainer.value
  })
  
  // 确保编辑器可以滚动
  setTimeout(() => {
    if (editorView && editorView.scrollDOM) {
      editorView.requestMeasure()
    }
  }, 100)
}

// 更新展开状态
const updateExpandedKeys = (keys) => {
  expandedKeys.value = new Set(keys)
}

// 渲染标签
const renderLabel = ({ option }) => {
  return option.label
}

// 渲染切换图标
const renderSwitcherIcon = () => h(NIcon, null, { default: () => h(ChevronForward) })

// 操作面板相关函数
const toggleOperations = () => {
  showOperations.value = !showOperations.value
}

// Git 操作函数
const cloneRepository = async () => {
  operations.value.cloning = true
  try {
    const response = await request.post(`/api/repos/${repoId.value}/clone`)
    
    if (response.data.data?.useSSE) {
      // 使用SSE获取实时状态，使用临时token
      const baseUrl = getApiBaseUrl()
      const tempToken = response.data.data.tempToken
      const eventSource = new EventSource(`${baseUrl}/api/repos/${repoId.value}/clone-status?tempToken=${tempToken}`)
      
      eventSource.onopen = () => {
        addLog('开始克隆仓库，正在连接...', 'info')
        setCurrentOperation('连接服务器中...', true)
        showStatusPanel.value = true
      }
      
      eventSource.addEventListener('start', (event) => {
        try {
          const data = JSON.parse(event.data)
          addLog(data.message, 'info')
          setCurrentOperation(data.message, true)
        } catch (error) {
          console.error('Failed to parse event data:', event.data)
          addLog('解析服务器数据失败', 'warning')
        }
      })
      
      eventSource.addEventListener('progress', (event) => {
        try {
          const data = JSON.parse(event.data)
          addLog(data.message, 'info')
          setCurrentOperation(data.message, true)
        } catch (error) {
          console.error('Failed to parse event data:', event.data)
          addLog('解析进度数据失败', 'warning')
        }
      })
      
      eventSource.addEventListener('complete', (event) => {
        try {
          const data = JSON.parse(event.data)
          addLog(data.message, 'success')
        } catch (error) {
          console.error('Failed to parse event data:', event.data)
          addLog('克隆操作完成', 'success')
        }
        eventSource.close()
        operations.value.cloning = false
        clearCurrentOperation()
        repoExists.value = true
        checkRepoStatus()
        loadFileTree()
      })
      
      eventSource.addEventListener('error', (event) => {
        try {
          const data = JSON.parse(event.data)
          addLog(data.message, 'error')
        } catch (error) {
          addLog('克隆过程中发生错误', 'error')
        }
        eventSource.close()
        operations.value.cloning = false
        clearCurrentOperation()
      })
      
      eventSource.onerror = () => {
        addLog('连接中断', 'error')
        eventSource.close()
        operations.value.cloning = false
        clearCurrentOperation()
      }
    } else {
      message.success(response.data.message)
      operations.value.cloning = false
      if (response.data.message === '本地仓库已存在') {
        repoExists.value = true
      }
    }
  } catch (error) {
    console.error('克隆仓库失败:', error)
    message.error('克隆仓库失败: ' + (error.message || '未知错误'))
    operations.value.cloning = false
  }
}

const pullUpdates = async () => {
  operations.value.pulling = true
  try {
    const response = await request.post(`/api/repos/${repoId.value}/pull`)
    
    if (response.data.data?.useSSE) {
      // 使用SSE获取实时状态，使用临时token
      const baseUrl = getApiBaseUrl()
      const tempToken = response.data.data.tempToken
      const eventSource = new EventSource(`${baseUrl}/api/repos/${repoId.value}/pull-status?tempToken=${tempToken}`)
      
      eventSource.onopen = () => {
        addLog('开始拉取更新，正在连接...', 'info')
        setCurrentOperation('连接服务器中...', true)
        showStatusPanel.value = true
      }
      
      eventSource.addEventListener('start', (event) => {
        try {
          const data = JSON.parse(event.data)
          addLog(data.message, 'info')
          setCurrentOperation(data.message, true)
        } catch (error) {
          console.error('Failed to parse event data:', event.data)
          addLog('解析服务器数据失败', 'warning')
        }
      })
      
      eventSource.addEventListener('progress', (event) => {
        try {
          const data = JSON.parse(event.data)
          addLog(data.message, 'info')
          setCurrentOperation(data.message, true)
        } catch (error) {
          console.error('Failed to parse event data:', event.data)
          addLog('解析进度数据失败', 'warning')
        }
      })
      
      eventSource.addEventListener('complete', (event) => {
        try {
          const data = JSON.parse(event.data)
          addLog(data.message, 'success')
        } catch (error) {
          console.error('Failed to parse event data:', event.data)
          addLog('拉取操作完成', 'success')
        }
        eventSource.close()
        operations.value.pulling = false
        clearCurrentOperation()
        checkRepoStatus()
        loadFileTree()
      })
      
      eventSource.addEventListener('error', (event) => {
        try {
          const data = JSON.parse(event.data)
          addLog(data.message, 'error')
        } catch (error) {
          addLog('拉取过程中发生错误', 'error')
        }
        eventSource.close()
        operations.value.pulling = false
        clearCurrentOperation()
      })
      
      eventSource.onerror = () => {
        addLog('连接中断', 'error')
        eventSource.close()
        operations.value.pulling = false
        clearCurrentOperation()
      }
    } else {
      addLog(response.data.message, 'success')
      operations.value.pulling = false
      showStatusPanel.value = true
    }
  } catch (error) {
    console.error('拉取更新失败:', error)
    addLog('拉取更新失败: ' + (error.message || '未知错误'), 'error')
    operations.value.pulling = false
    showStatusPanel.value = true
  }
}

const checkStatus = async () => {
  operations.value.checking = true
  setCurrentOperation('正在检查仓库状态...', true)
  try {
    const response = await request.get(`/api/repos/${repoId.value}/status`)
    repoStatus.value = response.data.data
    
    // 详细的状态日志
    if (response.data.data.exists) {
      addLog(`✅ 仓库已克隆 - ${response.data.data.branch} 分支`, 'success')
      
      if (response.data.data.lastCommitMessage) {
        addLog(`📝 最新提交: ${response.data.data.lastCommitHash} - ${response.data.data.lastCommitMessage}`, 'info')
      }
      
      if (response.data.data.hasChanges) {
        addLog(`⚠️  有 ${response.data.data.changes.length} 个未提交的更改`, 'warning')
      }
      
      if (response.data.data.hasUnpushed) {
        addLog(`⬆️  有 ${response.data.data.unpushedCount} 个未推送的提交`, 'warning')
      }
      
      if (response.data.data.hasBehind) {
        addLog(`⬇️  远程有 ${response.data.data.behindCount} 个新提交可拉取`, 'info')
      }
      
      if (response.data.data.fileCount) {
        addLog(`📁 仓库包含 ${response.data.data.fileCount} 个文件`, 'info')
      }
      
      if (response.data.data.repoSize) {
        addLog(`💾 仓库大小: ${response.data.data.repoSize}`, 'info')
      }
      
      if (!response.data.data.hasChanges && !response.data.data.hasUnpushed && !response.data.data.hasBehind) {
        addLog('✨ 仓库状态良好，与远程同步', 'success')
      }
    } else {
      addLog('❌ 仓库尚未克隆到本地', 'warning')
    }
    
    showStatusPanel.value = true
    clearCurrentOperation()
  } catch (error) {
    console.error('检查状态失败:', error)
    addLog('检查状态失败: ' + (error.message || '未知错误'), 'error')
    showStatusPanel.value = true
    clearCurrentOperation()
  } finally {
    operations.value.checking = false
  }
}

// 图片处理函数
const onImageLoad = (event) => {
  const img = event.target
  const dimensionsEl = document.getElementById('imageDimensions')
  if (dimensionsEl && img.naturalWidth && img.naturalHeight) {
    const fileSize = img.src.length // 这是一个简化的大小估算
    dimensionsEl.textContent = `${img.naturalWidth} × ${img.naturalHeight}`
  }
}

const onImageError = async (event) => {
  const img = event.target
  const imageUrl = img.src
  console.error('图片加载失败:', imageUrl, event)
  
  // 尝试直接请求API获取更详细的错误信息
  try {
    const response = await fetch(imageUrl)
    const responseText = await response.text()
    console.error('API响应:', response.status, responseText)
  } catch (fetchError) {
    console.error('获取API响应失败:', fetchError)
  }
  
  img.style.display = 'none'
  const container = img.closest('.image-container')
  if (container) {
    // 清除可能存在的旧错误消息
    const existingError = container.querySelector('.image-error')
    if (existingError) {
      existingError.remove()
    }
    
    const errorDiv = document.createElement('div')
    errorDiv.className = 'image-error'
    errorDiv.innerHTML = `
      <div class="error-content">
        <p>图片加载失败</p>
        <p class="error-filename">${currentFileName.value}</p>
        <div class="error-url" style="font-size: 10px; color: #999; word-break: break-all; margin-top: 8px;">
          URL: ${imageUrl}
        </div>
      </div>
    `
    container.appendChild(errorDiv)
  }
}

// 音频处理函数
const onAudioLoaded = (event) => {
  const audio = event.target
  const detailsEl = document.getElementById('audioDetails')
  if (detailsEl && audio.duration) {
    const duration = formatDuration(audio.duration)
    detailsEl.innerHTML = `
      <span style="font-size: 12px; color: #656d76;">
        时长: ${duration}
      </span>
    `
  }
}

// 视频处理函数
const onVideoLoaded = (event) => {
  const video = event.target
  const detailsEl = document.getElementById('videoDetails')
  if (detailsEl && video.duration && video.videoWidth && video.videoHeight) {
    const duration = formatDuration(video.duration)
    detailsEl.innerHTML = `
      <span style="font-size: 12px; color: #656d76;">
        ${video.videoWidth} × ${video.videoHeight} • 时长: ${duration}
      </span>
    `
  }
}

// 媒体错误处理
const onMediaError = (event) => {
  console.error('媒体文件加载失败:', event)
  const media = event.target
  const container = media.closest('.media-container')
  if (container) {
    const errorDiv = document.createElement('div')
    errorDiv.className = 'media-error'
    errorDiv.innerHTML = `
      <div class="error-content">
        <p>媒体文件加载失败</p>
        <p class="error-filename">${currentFileName.value}</p>
        <p class="error-hint">请检查文件格式是否受支持</p>
      </div>
    `
    media.style.display = 'none'
    container.appendChild(errorDiv)
  }
}

// 格式化时长
const formatDuration = (seconds) => {
  if (isNaN(seconds) || seconds < 0) return '00:00'
  
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const secs = Math.floor(seconds % 60)
  
  if (hours > 0) {
    return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
  } else {
    return `${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
  }
}

// 二进制文件处理函数
const getBinaryFileDescription = () => {
  if (!currentFileName.value) return '二进制文件'
  
  const lowerName = currentFileName.value.toLowerCase()
  
  // 根据文件扩展名返回描述
  if (lowerName.endsWith('.exe') || lowerName.endsWith('.msi')) return '可执行文件'
  if (lowerName.endsWith('.dll') || lowerName.endsWith('.so') || lowerName.endsWith('.dylib')) return '动态链接库'
  if (lowerName.match(/\.(zip|rar|7z|tar|gz|bz2|xz)$/)) return '压缩文件'
  if (lowerName.match(/\.(pdf|doc|docx|xls|xlsx|ppt|pptx)$/)) return '文档文件'
  if (lowerName.match(/\.(ttf|otf|woff|woff2|eot)$/)) return '字体文件'
  if (lowerName.match(/\.(psd|ai|sketch|fig)$/)) return '设计文件'
  if (lowerName.match(/\.(class|jar|war|ear)$/)) return 'Java字节码文件'
  if (lowerName.match(/\.(o|obj|lib|a)$/)) return '编译对象文件'
  if (lowerName.match(/\.(db|sqlite|sqlite3)$/)) return '数据库文件'
  if (lowerName.match(/\.(bin|dat)$/)) return '数据文件'
  
  return '二进制文件'
}

const downloadFile = () => {
  if (!repoId.value || !currentFile.value) return
  
  const downloadUrl = `/api/repos/${repoId.value}/file-content?path=${encodeURIComponent(currentFile.value)}&type=blob`
  
  // 创建隐藏的下载链接
  const link = document.createElement('a')
  link.href = downloadUrl
  link.download = currentFileName.value
  link.style.display = 'none'
  
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  
  addLog(`开始下载文件: ${currentFileName.value}`, 'info')
  showStatusPanel.value = true
}

const openWithCode = async () => {
  // 强制以文本方式查看二进制文件
  try {
    const response = await request.get(`/api/repos/${repoId.value}/file-content`, {
      params: { path: currentFile.value }
    })
    
    currentFileContent.value = response.data.data.content
    
    // 设置强制文本模式标志
    forceTextMode.value = true
    
    // 等待DOM更新后设置CodeMirror
    await nextTick()
    await createOrUpdateEditor()
    
    addLog(`已强制以文本方式打开: ${currentFileName.value}`, 'warning')
    showStatusPanel.value = true
  } catch (error) {
    console.error('强制打开文件失败:', error)
    message.error('无法以文本方式打开此文件')
  }
}

// 本地操作函数
const refreshFileTree = async () => {
  await loadFileTree()
  addLog('文件树已刷新', 'success')
  showStatusPanel.value = true
}

const deleteLocalRepo = async () => {
  try {
    await request.delete(`/api/repos/${repoId.value}/local`)
    addLog('本地仓库已删除', 'success')
    repoExists.value = false
    repoStatus.value = null
    await loadFileTree() // 重新加载文件树
    showStatusPanel.value = true
  } catch (error) {
    addLog('删除本地仓库失败', 'error')
    showStatusPanel.value = true
  }
}

// 检查仓库是否存在
const checkRepoStatus = async () => {
  try {
    const response = await request.get(`/api/repos/${repoId.value}/status`)
    repoExists.value = response.data.data.exists
    repoStatus.value = response.data.data
  } catch (error) {
    repoExists.value = false
    repoStatus.value = null
  }
}

// 仓库操作函数
const openRepo = (url) => {
  window.open(url, '_blank')
}

const updateRepoInfo = async (id) => {
  try {
    operations.value.updating = true
    addLog('开始更新仓库信息...', 'info')
    showStatusPanel.value = true
    
    const response = await request.put(`/api/repos/${id}/update`)
    
    if (response.data.code === 200) {
      // 更新本地仓库信息
      if (repoInfo.value && repoInfo.value.id === id) {
        const updatedRepo = response.data.data
        repoInfo.value = { ...repoInfo.value, ...updatedRepo }
      }
      
      addLog('✅ 仓库信息更新成功', 'success')
      message.success('仓库信息更新成功')
    } else {
      throw new Error(response.data.message || '更新失败')
    }
  } catch (error) {
    console.error('更新仓库信息失败:', error)
    
    let errorMessage = '更新仓库信息失败'
    if (error.response) {
      const status = error.response.status
      const responseMessage = error.response.data?.message || ''
      
      if (status === 404) {
        errorMessage = '仓库不存在或无法访问'
      } else if (status === 500) {
        errorMessage = responseMessage || '服务器处理失败，请稍后重试'
      } else {
        errorMessage = responseMessage || `请求失败 (${status})`
      }
    } else if (error.request) {
      errorMessage = '网络连接失败，请检查网络连接'
    } else {
      errorMessage = error.message || '未知错误'
    }
    
    addLog(`❌ ${errorMessage}`, 'error')
    message.error(errorMessage)
  } finally {
    operations.value.updating = false
  }
}

const deleteRepository = async (id) => {
  try {
    await request.delete(`/api/repos/${id}`)
    addLog('仓库已从数据库中删除', 'success')
    message.success('仓库删除成功')
    showStatusPanel.value = true

    // 可以选择跳转回首页或显示删除成功的消息
    window.location.href = '/'
  } catch (error) {
    addLog('删除仓库失败', 'error')
    message.error('删除仓库失败: ' + (error.message || '未知错误'))
    showStatusPanel.value = true
  }
}

const handleToggleValid = async (repo) => {
  try {
    const res = await request.post(`/api/repos/${repo.id}/toggle-valid`)
    const apiData = res.data
    if (apiData && apiData.code === 0) {
      message.success(apiData.message)
      // 更新 repoInfo 中的 valid 状态
      if (repoInfo.value && repoInfo.value.id === repo.id) {
        repoInfo.value = { ...repoInfo.value, valid: apiData.data.valid }
      }
    } else {
      throw new Error(apiData?.message || '操作失败')
    }
  } catch (error) {
    message.error('操作失败：' + error.message)
  }
}


// 格式化时间
const formatTime = (time) => {
  if (!time) return ''
  try {
    // 如果已经是Date对象，直接格式化
    if (time instanceof Date) {
      return time.toLocaleTimeString('zh-CN')
    }
    // 如果是字符串，先转换为Date对象
    return new Date(time).toLocaleString('zh-CN')
  } catch {
    return time.toString()
  }
}

onMounted(async () => {
  // 先加载仓库信息（获取 is_cloned 状态），再并行加载文件树和详细状态
  await loadRepoInfo()
  await Promise.all([loadFileTree(), checkRepoStatus()])
})

onBeforeUnmount(() => {
  if (editorView) {
    editorView.destroy()
  }
})
</script>

<style scoped>
.code-viewer {
  height: 100%;
  background: #f5f5f5;
  display: flex;
  flex-direction: column;
}

/* 操作区域样式 */
.operations-area {
  background: #fff;
  border-bottom: 1px solid #e0e0e0;
  box-shadow: 0 2px 8px rgba(0,0,0,0.03);
}

.operations-content {
  padding: 16px 24px;
  min-height: 120px;
}

.operations-area :deep(.n-tabs-nav) {
  padding: 0 24px;
}

.operations-area :deep(.n-card) {
  transition: all 0.3s ease;
}

.operations-area :deep(.n-card:hover) {
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
}

.viewer-main {
  flex: 1;
  display: flex;
  min-height: 0;
}

/* 文件资源管理器 */
.file-explorer {
  width: 280px;
  background: #fff;
  border-right: 1px solid #e0e0e0;
  display: flex;
  flex-direction: column;
}

.explorer-header {
  height: 48px;
  background: #f8f9fa;
  border-bottom: 1px solid #e0e0e0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
}

.explorer-title {
  font-size: 14px;
  font-weight: bold;
  color: #333;
}

.explorer-content {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.loading-container {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.empty-container {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.empty-text {
  color: #999;
  font-size: 14px;
}

.file-tree {
  height: 100%;
}

/* 代码容器 */
.code-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  background: #fff;
}

.code-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.no-file-selected {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #999;
}

.no-file-icon {
  margin-bottom: 16px;
}

.no-file-text {
  font-size: 16px;
}

.file-viewer {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.file-header {
  height: 48px;
  background: #f8f9fa;
  border-bottom: 1px solid #e0e0e0;
  display: flex;
  align-items: center;
  padding: 0 16px;
}

.file-info {
  display: flex;
  align-items: center;
}

.file-name {
  font-size: 14px;
  font-weight: bold;
  color: #333;
}

.file-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.code-editor {
  flex: 1;
  overflow: auto;
  background: #1e1e1e;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.03);
}

.codemirror-container {
  height: 100%;
  min-height: 400px;
}

/* CodeMirror 样式覆盖 */
:deep(.cm-editor) {
  height: 100% !important;
  font-size: 15px;
  outline: none !important;
}

:deep(.cm-editor .cm-scroller) {
  font-family: 'Fira Code', 'Consolas', 'Monaco', monospace;
  overflow: auto !important;
  height: 100% !important;
  max-height: none !important;
}

:deep(.cm-editor .cm-line) {
  padding: 0;
}

/* 只读模式下的光标隐藏 */
:deep(.cm-editor .cm-cursor) {
  display: none !important;
}

:deep(.cm-editor .cm-cursor-primary) {
  display: none !important;
}

/* 状态面板样式 */
.status-panel {
  position: fixed;
  bottom: 20px;
  right: 20px;
  width: 400px;
  max-width: 90vw;
  z-index: 1000;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
  border-radius: 8px;
}

.operation-logs {
  min-height: 100px;
}

.log-item {
  padding: 8px 0;
  border-bottom: 1px solid var(--border-color);
}

.log-item:last-child {
  border-bottom: none;
}

.log-header {
  margin-bottom: 4px;
}

.log-content {
  margin-left: 20px;
  font-size: 13px;
  line-height: 1.4;
}

.current-operation {
  padding-top: 12px;
}

.no-logs {
  padding: 20px 0;
  text-align: center;
}

.repo-status-overview {
  padding-top: 12px;
}

.status-header {
  margin-bottom: 12px;
}

.status-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 8px;
}

.status-item {
  padding: 6px 8px;
  background: var(--color-fill-2);
  border-radius: 6px;
  display: flex;
  align-items: center;
}

/* Markdown 渲染样式 */
.markdown-viewer {
  flex: 1;
  overflow: auto;
  background: #fff;
  border-radius: 8px;
}

.markdown-content {
  padding: 20px;
  max-width: none;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Helvetica Neue', Arial, sans-serif;
  line-height: 1.6;
  color: #333;
}

/* Markdown 样式 */
.markdown-content h1,
.markdown-content h2,
.markdown-content h3,
.markdown-content h4,
.markdown-content h5,
.markdown-content h6 {
  margin-top: 24px;
  margin-bottom: 16px;
  font-weight: 600;
  line-height: 1.25;
  color: #1f2328;
}

.markdown-content h1 {
  font-size: 2rem;
  border-bottom: 1px solid #d1d9e0;
  padding-bottom: 10px;
}

.markdown-content h2 {
  font-size: 1.5rem;
  border-bottom: 1px solid #d1d9e0;
  padding-bottom: 8px;
}

.markdown-content h3 {
  font-size: 1.25rem;
}

.markdown-content h4 {
  font-size: 1rem;
}

.markdown-content h5 {
  font-size: 0.875rem;
}

.markdown-content h6 {
  font-size: 0.75rem;
  color: #656d76;
}

.markdown-content p {
  margin-bottom: 16px;
}

.markdown-content ul,
.markdown-content ol {
  margin-bottom: 16px;
  padding-left: 2rem;
}

.markdown-content li {
  margin-bottom: 0.25rem;
}

.markdown-content blockquote {
  margin: 0 0 16px;
  padding: 0 1rem;
  color: #656d76;
  border-left: 0.25rem solid #d1d9e0;
}

.markdown-content code {
  background: #f6f8fa;
  border-radius: 6px;
  font-size: 85%;
  margin: 0;
  padding: 0.2rem 0.4rem;
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
}

.markdown-content pre {
  background: #f6f8fa;
  border-radius: 6px;
  font-size: 85%;
  line-height: 1.45;
  margin-bottom: 16px;
  overflow: auto;
  padding: 16px;
}

.markdown-content pre code {
  background: transparent;
  border: 0;
  display: inline;
  font-size: 100%;
  margin: 0;
  max-width: auto;
  padding: 0;
  word-wrap: normal;
}

.markdown-content table {
  border-collapse: collapse;
  border-spacing: 0;
  display: block;
  margin-bottom: 16px;
  overflow: auto;
  width: 100%;
  width: max-content;
  max-width: 100%;
}

.markdown-content table th,
.markdown-content table td {
  border: 1px solid #d1d9e0;
  padding: 6px 13px;
}

.markdown-content table th {
  background-color: #f6f8fa;
  font-weight: 600;
}

.markdown-content table tr {
  background-color: #fff;
  border-top: 1px solid #d1d9e0;
}

.markdown-content table tr:nth-child(2n) {
  background-color: #f6f8fa;
}

.markdown-content a {
  color: #0969da;
  text-decoration: none;
}

.markdown-content a:hover {
  text-decoration: underline;
}

.markdown-content img {
  max-width: 100%;
  height: auto;
  border-radius: 6px;
}

.markdown-content hr {
  background-color: #d1d9e0;
  border: 0;
  height: 2px;
  margin: 24px 0;
}

/* 图片查看器样式 */
.image-viewer {
  flex: 1;
  overflow: auto;
  background: #f8f9fa;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
}

.image-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  max-width: 100%;
  max-height: 100%;
  padding: 20px;
  min-height: 200px; /* 确保加载状态有足够高度 */
}

.image-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
}

.image-content {
  max-width: 100%;
  max-height: calc(100vh - 200px);
  object-fit: contain;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  background: white;
  border: 1px solid #e0e0e0;
}

.image-info {
  margin-top: 16px;
  text-align: center;
  padding: 8px 16px;
  background: white;
  border-radius: 6px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.image-error {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 300px;
  height: 200px;
  background: white;
  border: 2px dashed #d1d9e0;
  border-radius: 8px;
}

.image-error .error-content {
  text-align: center;
  color: #656d76;
}

.image-error .error-content p {
  margin: 8px 0;
}

.image-error .error-filename {
  font-family: monospace;
  font-size: 12px;
  color: #8b949e;
}

/* 媒体播放器样式 */
.media-viewer {
  flex: 1;
  overflow: auto;
  background: #f8f9fa;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
}

.media-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  max-width: 100%;
  max-height: 100%;
  padding: 20px;
  gap: 16px;
}

.media-info {
  text-align: center;
  padding: 8px 16px;
  background: white;
  border-radius: 6px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.audio-player {
  width: 100%;
  max-width: 500px;
  height: 60px;
  outline: none;
  border-radius: 8px;
}

.video-player {
  width: 100%;
  max-width: 800px;
  max-height: calc(100vh - 250px);
  outline: none;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  background: #000;
}

.media-details {
  text-align: center;
  padding: 6px 12px;
  background: white;
  border-radius: 4px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  min-height: 20px;
}

.media-error {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 400px;
  height: 200px;
  background: white;
  border: 2px dashed #d1d9e0;
  border-radius: 8px;
  margin-top: 16px;
}

.media-error .error-content {
  text-align: center;
  color: #656d76;
}

.media-error .error-content p {
  margin: 8px 0;
}

.media-error .error-filename {
  font-family: monospace;
  font-size: 12px;
  color: #8b949e;
}

.media-error .error-hint {
  font-size: 11px;
  color: #8b949e;
  font-style: italic;
}

/* 媒体控件样式优化 */
.audio-player::-webkit-media-controls-panel {
  background: #f8f9fa;
  border-radius: 8px;
}

.video-player::-webkit-media-controls-panel {
  background: rgba(0, 0, 0, 0.8);
}

/* 不支持文件类型查看器样式 */
.unsupported-viewer {
  flex: 1;
  overflow: auto;
  background: #f8f9fa;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
}

.unsupported-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  max-width: 500px;
  padding: 40px 20px;
  text-align: center;
  gap: 20px;
}

.unsupported-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 80px;
  height: 80px;
  background: white;
  border-radius: 50%;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  border: 2px solid #e0e0e0;
}

.unsupported-info {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  max-width: 400px;
}

.unsupported-message {
  padding: 12px 20px;
  background: #fff3cd;
  border: 1px solid #f0ad4e;
  border-radius: 6px;
  color: #856404;
}

.unsupported-actions {
  margin-top: 8px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .unsupported-container {
    max-width: 90%;
    padding: 20px 10px;
    gap: 16px;
  }
  
  .unsupported-icon {
    width: 60px;
    height: 60px;
  }
  
  .unsupported-actions .n-space {
    flex-direction: column;
  }
}

</style>