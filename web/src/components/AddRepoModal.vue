<template>
  <n-modal
    v-model:show="localShow"
    preset="dialog"
    title="添加仓库"
    style="width: 500px;"
    :mask-closable="false"
  >
    <div class="add-repo-modal">
      <n-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-placement="left"
        label-width="80px"
        require-mark-placement="right-hanging"
        size="medium"
      >
        <!-- 仓库URL -->
        <n-form-item label="仓库URL" path="url">
          <n-input
            v-model:value="formData.url"
            placeholder="https://github.com/author/repo 或 https://gitee.com/author/repo"
            clearable
            @keyup.enter="handleSubmit"
          >
            <template #prefix>
              <n-icon><LinkOutline /></n-icon>
            </template>
          </n-input>
        </n-form-item>

        <!-- 自动克隆 -->
        <n-form-item label="自动克隆">
          <n-switch v-model:value="formData.autoClone" />
          <n-text depth="3" style="margin-left: 12px;">
            添加后自动克隆到本地
          </n-text>
        </n-form-item>
      </n-form>

      <!-- 仓库预览 -->
      <div v-if="repoPreview" class="repo-preview">
        <n-divider style="margin: 12px 0;">仓库信息</n-divider>
        <div class="repo-preview-card">
          <div class="repo-preview-header">
            <n-text strong style="font-size: 15px;">
              <n-tag size="small" :type="repoPreview.platform === 'Gitee' ? 'success' : 'info'" style="margin-right: 6px;">{{ repoPreview.platform }}</n-tag>
              {{ repoPreview.author }}/<span class="repo-name">{{ repoPreview.repo }}</span>
            </n-text>
            <div class="repo-preview-stats">
              <n-tag size="small" round type="warning" v-if="repoPreview.stars > 0">
                <template #icon><n-icon><StarOutline /></n-icon></template>
                {{ repoPreview.stars }}
              </n-tag>
              <n-tag size="small" round type="info" v-if="repoPreview.forks > 0">
                <template #icon><n-icon><GitBranchOutline /></n-icon></template>
                {{ repoPreview.forks }}
              </n-tag>
            </div>
          </div>
          <div class="repo-preview-desc">
            {{ repoPreview.description || '暂无描述' }}
          </div>
        </div>
      </div>

      <!-- 克隆进度 -->
      <div v-if="cloning" class="clone-progress">
        <n-divider style="margin: 12px 0;">克隆进度</n-divider>
        <n-space align="center">
          <n-spin size="small" />
          <n-text>{{ cloneProgress || '正在克隆...' }}</n-text>
        </n-space>
      </div>
    </div>

    <template #action>
      <n-space justify="end">
        <n-button @click="handleCancel" :disabled="cloning">取消</n-button>
        <n-button
          type="primary"
          :loading="loading || cloning"
          :disabled="!formData.url.trim()"
          @click="handleSubmit"
        >
          {{ cloning ? '克隆中...' : '添加' }}
        </n-button>
      </n-space>
    </template>
  </n-modal>
</template>

<script setup>
import { ref, watch } from 'vue'
import { addRepo } from '@/api/repos'
import { useMessage } from 'naive-ui'
import { LinkOutline, StarOutline, GitBranchOutline } from '@vicons/ionicons5'
import { useClone } from '@/composables/useClone'

const props = defineProps({
  show: { type: Boolean, default: false }
})

const emit = defineEmits(['update:show', 'success'])

const message = useMessage()
const formRef = ref(null)
const loading = ref(false)
const localShow = ref(props.show)
const repoPreview = ref(null)

// 使用克隆 composable
const { cloning, cloneProgress, cloneRepo } = useClone()

const formData = ref({
  url: '',
  autoClone: true
})

const rules = {
  url: [
    { required: true, message: '请输入仓库URL', trigger: ['input', 'blur'] },
    { pattern: /^https:\/\/(github|gitee)\.com\/[^\/]+\/[^\/]+\/?$/, message: '请输入有效的仓库URL（支持 GitHub/Gitee）', trigger: ['blur'] }
  ]
}

watch(() => props.show, (val) => {
  localShow.value = val
  if (val) resetForm()
})

watch(localShow, (val) => emit('update:show', val))

watch(() => formData.value.url, (url) => {
  const match = url.match(/^https:\/\/(github|gitee)\.com\/([^\/]+)\/([^\/]+)\/?$/)
  if (match) {
    repoPreview.value = {
      platform: match[1] === 'gitee' ? 'Gitee' : 'GitHub',
      author: match[2],
      repo: match[3],
      description: '',
      stars: 0,
      forks: 0
    }
  } else {
    repoPreview.value = null
  }
})

const resetForm = () => {
  formData.value = { url: '', autoClone: true }
  repoPreview.value = null
  loading.value = false
}

const handleSubmit = async () => {
  try {
    await formRef.value?.validate()
    loading.value = true

    const response = await addRepo({
      url: formData.value.url.trim(),
      autoClone: formData.value.autoClone
    })

    if (response.data?.code === 0 || response.code === 0) {
      const data = response.data || response
      message.success('仓库添加成功')

      // 如果开启自动克隆，执行克隆
      if (formData.value.autoClone && data.data?.repoId) {
        loading.value = false
        await cloneRepo(data.data.repoId, {
          onProgress: (msg) => {
            cloneProgress.value = msg
          },
          onComplete: () => {
            message.success('克隆完成')
            emit('success', data)
            localShow.value = false
          },
          onError: (err) => {
            message.error('克隆失败: ' + err)
            emit('success', data) // 即使克隆失败，也触发成功事件（仓库已添加）
            localShow.value = false
          }
        })
      } else {
        emit('success', data)
        localShow.value = false
      }
    } else {
      throw new Error(response.message || '添加失败')
    }
  } catch (error) {
    message.error(error.message || '添加仓库失败')
  } finally {
    loading.value = false
  }
}

const handleCancel = () => {
  if (!cloning.value) {
    localShow.value = false
  }
}
</script>

<style scoped>
.add-repo-modal {
  max-height: 60vh;
  overflow-y: auto;
}

.repo-preview-card {
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  padding: var(--space-4);
}

.repo-preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.repo-name {
  color: var(--color-primary);
}

.repo-preview-stats {
  display: flex;
  gap: 8px;
}

.repo-preview-desc {
  margin-top: var(--space-2);
  color: var(--color-text-secondary);
  font-size: var(--text-sm);
}

.clone-progress {
  margin-top: var(--space-3);
}
</style>