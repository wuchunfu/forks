<template>
  <n-modal
    v-model:show="localShow"
    preset="dialog"
    title="选择要克隆的仓库"
    style="width: 600px;"
    :mask-closable="false"
    :show-icon="false"
  >
    <div class="batch-clone-modal">
      <!-- 加载中 -->
      <div v-if="loading" class="batch-clone-loading">
        <n-spin size="medium" />
        <n-text depth="3" style="margin-top: 12px;">加载未克隆仓库列表...</n-text>
      </div>

      <!-- 无数据 -->
      <div v-else-if="repos.length === 0" class="batch-clone-empty">
        <n-text depth="3">没有需要克隆的仓库</n-text>
      </div>

      <!-- 仓库列表 -->
      <template v-else>
        <div class="batch-clone-header">
          <n-checkbox
            :checked="isAllSelected"
            :indeterminate="isPartialSelected"
            @update:checked="toggleSelectAll"
          >
            全选
          </n-checkbox>
          <n-text depth="3" style="font-size: 13px;">
            共 {{ repos.length }} 个未克隆仓库，已选 {{ selectedIds.length }} 个
          </n-text>
        </div>

        <div class="batch-clone-list">
          <n-checkbox-group v-model:value="selectedIds">
            <div
              v-for="repo in repos"
              :key="repo.id"
              class="batch-clone-item"
            >
              <n-checkbox :value="repo.id" :label="undefined">
                <div class="batch-clone-item-content">
                  <n-text>{{ repo.author }}/{{ repo.repo }}</n-text>
                  <n-tag
                    size="small"
                    :type="repo.source === 'gitee' ? 'success' : 'info'"
                    style="margin-left: 8px;"
                  >
                    {{ repo.source }}
                  </n-tag>
                </div>
              </n-checkbox>
            </div>
          </n-checkbox-group>
        </div>
      </template>
    </div>

    <template #action>
      <n-space>
        <n-button @click="localShow = false">取消</n-button>
        <n-button
          type="primary"
          :disabled="selectedIds.length === 0"
          @click="handleConfirm"
        >
          开始克隆 ({{ selectedIds.length }})
        </n-button>
      </n-space>
    </template>
  </n-modal>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { getRepos } from '@/api/repos'

const props = defineProps({
  show: { type: Boolean, default: false }
})

const emit = defineEmits(['update:show', 'confirm'])

const localShow = computed({
  get: () => props.show,
  set: (val) => emit('update:show', val)
})

const loading = ref(false)
const repos = ref([])
const selectedIds = ref([])

const isAllSelected = computed(() => repos.value.length > 0 && selectedIds.value.length === repos.value.length)
const isPartialSelected = computed(() => selectedIds.value.length > 0 && selectedIds.value.length < repos.value.length)

const toggleSelectAll = (checked) => {
  selectedIds.value = checked ? repos.value.map((r) => r.id) : []
}

// 弹窗打开时加载未克隆仓库
watch(
  () => props.show,
  async (val) => {
    if (val) {
      loading.value = true
      selectedIds.value = []
      try {
        // 获取所有未克隆仓库（不分页）
        const allRepos = []
        let page = 1
        const pageSize = 100
        let hasMore = true

        while (hasMore) {
          const response = await getRepos({ status: 'not-cloned', page, pageSize })
          const apiData = response.data
          if (apiData && apiData.code === 0 && apiData.data) {
            const items = apiData.data.list || apiData.data.repos || []
            allRepos.push(...items)
            const total = apiData.data.total || 0
            hasMore = allRepos.length < total
            page++
          } else {
            hasMore = false
          }
        }

        repos.value = allRepos
        // 默认全选
        selectedIds.value = allRepos.map((r) => r.id)
      } catch (e) {
        console.error('获取未克隆仓库失败:', e)
        repos.value = []
      } finally {
        loading.value = false
      }
    }
  }
)

const handleConfirm = () => {
  emit('confirm', selectedIds.value)
  localShow.value = false
}
</script>

<style scoped>
.batch-clone-modal {
  min-height: 100px;
}

.batch-clone-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 40px 0;
}

.batch-clone-empty {
  text-align: center;
  padding: 40px 0;
}

.batch-clone-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px solid var(--n-border-color, rgba(255, 255, 255, 0.1));
  margin-bottom: 8px;
}

.batch-clone-list {
  max-height: 400px;
  overflow-y: auto;
}

.batch-clone-item {
  padding: 6px 4px;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.batch-clone-item:hover {
  background-color: var(--n-color-hover, rgba(255, 255, 255, 0.06));
}

.batch-clone-item-content {
  display: flex;
  align-items: center;
}
</style>
