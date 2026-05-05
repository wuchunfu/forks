<template>
  <div class="settings-panel">
    <div class="panel-header">
      <h2 class="panel-title">趋势自动同步</h2>
      <p class="panel-desc">配置 GitHub 趋势数据的自动采集任务</p>
    </div>

    <div class="settings-form">
      <!-- 全局开关 -->
      <div class="form-row">
        <div class="form-label">
          <span class="label-text">启用自动同步</span>
          <span class="label-desc">每天在指定时间自动采集趋势数据</span>
        </div>
        <n-switch v-model:value="syncConfig.enabled" />
      </div>

      <!-- 同步时间 -->
      <div class="form-row">
        <div class="form-label">
          <span class="label-text">同步时间</span>
          <span class="label-desc">每天执行采集的时间</span>
        </div>
        <n-input
          v-model:value="syncConfig.sync_time"
          placeholder="01:00"
          size="small"
          style="width: 100px"
        />
      </div>

      <!-- 任务列表 -->
      <div class="form-section">
        <div class="section-header">
          <span class="section-title">同步任务</span>
          <span class="section-desc">配置需要自动采集的语言组合</span>
        </div>

        <div class="sync-tasks-list" v-if="syncConfig.tasks && syncConfig.tasks.length > 0">
          <div v-for="(task, index) in syncConfig.tasks" :key="index" class="sync-task-item">
            <span class="task-lang">{{ task.language || '全部语言' }}</span>
            <span class="task-sep">/</span>
            <span class="task-spoken">{{ task.spoken_language_code || '全部' }}</span>
            <span class="task-sep">/</span>
            <span class="task-since">{{ { daily: '每天', weekly: '每周', monthly: '每月' }[task.since] }}</span>
            <button class="task-remove" @click="removeSyncTask(index)">✕</button>
          </div>
        </div>
        <div v-else class="sync-tasks-empty">暂无同步任务，请添加</div>

        <!-- 添加任务 -->
        <div class="sync-task-add">
          <n-select
            v-model:value="newTask.language"
            :options="syncLanguageOptions"
            placeholder="编程语言"
            clearable
            filterable
            size="small"
            style="width: 180px"
          />
          <n-select
            v-model:value="newTask.spoken_language_code"
            :options="syncSpokenOptions"
            placeholder="自然语言"
            clearable
            filterable
            size="small"
            style="width: 160px"
          />
          <n-select
            v-model:value="newTask.since"
            :options="[{ label: '每天', value: 'daily' }, { label: '每周', value: 'weekly' }, { label: '每月', value: 'monthly' }]"
            size="small"
            style="width: 90px"
          />
          <n-button size="small" type="primary" @click="addSyncTask">添加</n-button>
        </div>
      </div>

      <!-- 操作按钮 -->
      <div class="settings-footer">
        <div class="footer-actions">
          <button class="btn btn-secondary" @click="loadSyncConfig">重置</button>
          <button class="btn btn-primary" @click="saveSyncConfig">保存配置</button>
        </div>
        <div class="footer-actions">
          <button class="btn btn-secondary" @click="handleSyncNow" :disabled="syncRunning">
            {{ syncRunning ? '同步中...' : '立即同步' }}
          </button>
          <span v-if="syncConfig.last_sync_date" class="last-sync-info">
            上次同步：{{ syncConfig.last_sync_date }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useMessage } from 'naive-ui'
import { getSyncConfig, updateSyncConfig, syncNow, getTrendingLanguages } from '@/api/trending'

const message = useMessage()

const syncConfig = ref({
  enabled: false,
  sync_time: '01:00',
  last_sync_date: '',
  tasks: []
})
const syncRunning = ref(false)
const syncLanguageOptions = ref([{ label: '全部语言', value: '' }])
const syncSpokenOptions = ref([{ label: '全部', value: '' }])
const newTask = ref({ language: '', spoken_language_code: '', since: 'daily' })

async function loadSyncConfig() {
  try {
    const res = await getSyncConfig()
    const data = res.data.data
    syncConfig.value = {
      enabled: data.enabled || false,
      sync_time: data.sync_time || '01:00',
      last_sync_date: data.last_sync_date || '',
      tasks: data.tasks || []
    }
  } catch (e) {
    console.error('加载同步配置失败:', e)
  }
}

async function loadSyncLanguageOptions() {
  try {
    const res = await getTrendingLanguages()
    const data = res.data.data
    syncLanguageOptions.value = [
      { label: '全部语言', value: '' },
      ...data.programming_languages.map(l => ({ label: l.name, value: l.slug }))
    ]
    syncSpokenOptions.value = [
      { label: '全部', value: '' },
      ...data.spoken_languages.map(l => ({ label: `${l.name} (${l.code})`, value: l.code }))
    ]
  } catch (e) {
    console.error('加载语言选项失败:', e)
  }
}

async function saveSyncConfig() {
  try {
    await updateSyncConfig(syncConfig.value)
    message.success('同步配置已保存')
  } catch (e) {
    message.error('保存失败')
  }
}

function addSyncTask() {
  const task = { ...newTask.value }
  if (!task.since) task.since = 'daily'
  syncConfig.value.tasks.push(task)
  newTask.value = { language: '', spoken_language_code: '', since: 'daily' }
}

function removeSyncTask(index) {
  syncConfig.value.tasks.splice(index, 1)
}

async function handleSyncNow() {
  syncRunning.value = true
  try {
    await syncNow()
    message.success('同步已启动，后台执行中')
    setTimeout(() => loadSyncConfig(), 3000)
  } catch (e) {
    message.error(e.response?.data?.message || '同步失败')
  } finally {
    syncRunning.value = false
  }
}

onMounted(() => {
  loadSyncConfig()
  loadSyncLanguageOptions()
})
</script>

<style scoped>
.settings-panel {
  padding: var(--space-6);
  max-height: calc(100vh - var(--navbar-height) - var(--space-6) * 2);
  overflow-y: auto;
}

.panel-header {
  margin-bottom: var(--space-6);
  padding-bottom: var(--space-4);
  border-bottom: 1px solid var(--color-border-light);
}

.panel-title {
  font-size: var(--text-xl);
  font-weight: var(--font-semibold);
  color: var(--color-text-primary);
  margin: 0 0 var(--space-1) 0;
}

.panel-desc {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  margin: 0;
}

.settings-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
}

.form-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.form-label {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.label-text {
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  color: var(--color-text-primary);
}

.label-desc {
  font-size: var(--text-xs);
  color: var(--color-text-tertiary);
}

.form-section {
  margin-top: var(--space-4);
}

.section-header {
  margin-bottom: var(--space-3);
}

.section-title {
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--color-text-primary);
  display: block;
}

.section-desc {
  font-size: var(--text-xs);
  color: var(--color-text-tertiary);
}

.sync-tasks-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  margin-bottom: var(--space-3);
}

.sync-task-item {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-md);
  background-color: var(--color-gray-50);
  border: 1px solid var(--color-border-light);
  font-size: var(--text-sm);
}

.task-lang {
  color: var(--color-text-primary);
  font-weight: var(--font-medium);
  min-width: 100px;
}

.task-sep {
  color: var(--color-text-quaternary);
}

.task-spoken {
  color: var(--color-text-secondary);
  min-width: 80px;
}

.task-since {
  color: var(--color-primary);
  font-weight: var(--font-medium);
}

.task-remove {
  margin-left: auto;
  background: none;
  border: none;
  color: var(--color-text-quaternary);
  cursor: pointer;
  font-size: var(--text-sm);
  padding: 2px 6px;
  border-radius: var(--radius-sm);
}

.task-remove:hover {
  background-color: var(--color-error-50, #fef2f2);
  color: var(--color-error);
}

.sync-tasks-empty {
  text-align: center;
  padding: var(--space-4);
  color: var(--color-text-tertiary);
  font-size: var(--text-sm);
  border: 1px dashed var(--color-border-light);
  border-radius: var(--radius-md);
  margin-bottom: var(--space-3);
}

.sync-task-add {
  display: flex;
  gap: var(--space-2);
  align-items: center;
  flex-wrap: wrap;
}

.last-sync-info {
  font-size: var(--text-xs);
  color: var(--color-text-tertiary);
}

.settings-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-4) var(--space-6);
  border-top: 1px solid var(--color-border-light);
  flex-wrap: wrap;
  gap: var(--space-4);
}

.footer-actions {
  display: flex;
  gap: var(--space-3);
  align-items: center;
}

.btn {
  padding: var(--space-2_5) var(--space-4);
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  border-radius: var(--radius-md);
  border: 1px solid transparent;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-secondary {
  color: var(--color-text-primary);
  background-color: var(--color-gray-100);
  border-color: var(--color-border);
}

.btn-secondary:hover {
  background-color: var(--color-gray-200);
}

.btn-primary {
  color: white;
  background-color: var(--color-primary);
}

.btn-primary:hover {
  background-color: var(--color-primary-600);
}
</style>
