<template>
  <div class="settings-panel">
    <div class="panel-header">
      <h2 class="panel-title">MCP 工具</h2>
      <p class="panel-desc">Forks 提供的 Model Context Protocol 工具，供 AI 助手调用</p>
    </div>

    <n-tabs v-model:value="mcpSubTab" type="line" size="small" class="mcp-sub-tabs">
      <n-tab-pane name="tools" tab="工具列表">
        <div class="mcp-tools">
          <div
            v-for="tool in mcpTools"
            :key="tool.name"
            class="mcp-tool-card"
            :class="{ 'is-expanded': tool._expanded }"
            @click="tool._expanded = !tool._expanded"
          >
            <div class="mcp-tool-header">
              <div class="mcp-tool-title-row">
                <span class="mcp-tool-name">{{ tool.name }}</span>
              </div>
              <div class="mcp-tool-desc">{{ tool.description }}</div>
            </div>
            <div v-if="tool.params.length > 0" class="mcp-tool-params" v-show="tool._expanded">
              <table class="mcp-params-table">
                <thead>
                  <tr>
                    <th style="width: 1%">参数</th>
                    <th style="width: 1%">类型</th>
                    <th style="width: 1%">必填</th>
                    <th>说明</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="p in tool.params" :key="p.name">
                    <td><code class="mcp-param-name">{{ p.name }}</code></td>
                    <td><span class="mcp-type-badge" :class="'type-' + p.type">{{ p.type }}</span></td>
                    <td>
                      <span v-if="p.required" class="mcp-required">*</span>
                      <span v-else class="mcp-optional">-</span>
                    </td>
                    <td>{{ p.description }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div v-if="tool.params.length === 0" class="mcp-tool-params mcp-no-params" v-show="tool._expanded">
              <span>无需参数</span>
            </div>
          </div>
        </div>
      </n-tab-pane>

      <n-tab-pane name="config" tab="配置">
        <div class="codemirror-wrapper">
          <button class="codemirror-copy-btn" @click="copyMcpConfig" title="复制">复制</button>
          <div class="codemirror-container" ref="mcpCodeRef"></div>
        </div>
      </n-tab-pane>
    </n-tabs>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useMessage, NTabs, NTabPane } from 'naive-ui'
import { getMCPTools } from '@/api/trending'
import { copyToClipboard } from '@/utils/clipboard'
import { EditorView, lineNumbers, highlightActiveLineGutter } from '@codemirror/view'
import { EditorState } from '@codemirror/state'
import { defaultHighlightStyle, syntaxHighlighting } from '@codemirror/language'
import { json } from '@codemirror/lang-json'
import { dracula } from '@uiw/codemirror-theme-dracula'

const message = useMessage()

const mcpSubTab = ref('tools')
const mcpCodeRef = ref(null)
let mcpEditorView = null

const mcpTools = ref([])

const loadMCPTools = async () => {
  try {
    const res = await getMCPTools()
    const data = res.data?.data || []
    mcpTools.value = data.map(t => ({ ...t, _expanded: false, params: t.params || [] }))
  } catch (e) {
    console.error('加载 MCP 工具列表失败:', e)
  }
}

const mcpConfigDisplay = computed(() => {
  const hasToken = !!localStorage.getItem('token')
  const config = {
    servers: {
      forks: {
        type: 'http',
        url: `${window.location.origin}/mcp`
      }
    }
  }
  if (hasToken) {
    config.servers.forks.headers = {
      Authorization: 'Bearer <your-token>'
    }
  }
  return JSON.stringify(config, null, 2)
})

function copyMcpConfig() {
  const token = localStorage.getItem('token') || ''
  const config = {
    servers: {
      forks: {
        type: 'http',
        url: `${window.location.origin}/mcp`
      }
    }
  }
  if (token) {
    config.servers.forks.headers = {
      Authorization: `Bearer ${token}`
    }
  }
  copyToClipboard(JSON.stringify(config, null, 2))
  message.success('已复制完整配置')
}

function initMCPCodeMirror() {
  if (!mcpCodeRef.value) return
  if (mcpEditorView) {
    mcpEditorView.destroy()
    mcpEditorView = null
  }
  mcpEditorView = new EditorView({
    state: EditorState.create({
      doc: mcpConfigDisplay.value,
      extensions: [
        json(),
        dracula,
        lineNumbers(),
        highlightActiveLineGutter(),
        syntaxHighlighting(defaultHighlightStyle),
        EditorView.editable.of(false),
        EditorView.theme({
          '&': { height: 'auto', fontSize: '13px' },
          '.cm-scroller': { maxHeight: '300px', overflow: 'auto' }
        })
      ]
    }),
    parent: mcpCodeRef.value
  })
}

watch(mcpSubTab, (sub) => {
  if (sub === 'config') {
    nextTick(() => initMCPCodeMirror())
  }
})

onMounted(() => {
  loadMCPTools()
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

.mcp-tools {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: var(--space-6);
}

.mcp-tool-card {
  border-radius: var(--radius-md);
  background-color: var(--color-bg-page);
  border: 1px solid var(--color-border-light);
  cursor: pointer;
  transition: border-color 0.2s, box-shadow 0.2s;
  overflow: hidden;
}

.mcp-tool-card:hover {
  border-color: var(--color-primary-200);
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
}

.mcp-tool-card.is-expanded {
  border-color: var(--color-primary);
}

.mcp-tool-header {
  padding: 12px 16px;
}

.mcp-tool-title-row {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 4px;
}

.mcp-tool-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--color-primary);
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
}

.mcp-tool-desc {
  font-size: 13px;
  color: var(--color-text-tertiary);
  line-height: 1.4;
}

.mcp-tool-card.is-expanded .mcp-tool-desc {
  color: var(--color-text-secondary);
}

.mcp-tool-params {
  border-top: 1px solid var(--color-border-light);
  padding: 8px 16px 12px;
  animation: slideDown 0.15s ease-out;
}

@keyframes slideDown {
  from { opacity: 0; transform: translateY(-4px); }
  to { opacity: 1; transform: translateY(0); }
}

.mcp-no-params {
  text-align: center;
  padding: 12px 16px;
  color: var(--color-text-quaternary);
  font-size: 12px;
}

.mcp-params-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}

.mcp-params-table th {
  text-align: left;
  padding: 6px 10px;
  background-color: var(--color-gray-50);
  color: var(--color-text-tertiary);
  font-weight: 500;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.3px;
}

.mcp-params-table td {
  padding: 7px 10px;
  border-bottom: 1px solid var(--color-border-light);
  color: var(--color-text-secondary);
}

.mcp-params-table tr:last-child td {
  border-bottom: none;
}

.mcp-param-name {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
  color: var(--color-primary);
  font-weight: 500;
}

.mcp-type-badge {
  display: inline-block;
  padding: 1px 6px;
  border-radius: 3px;
  font-size: 11px;
  font-weight: 500;
}

.mcp-type-badge.type-string { background: #dbeafe; color: #2563eb; }
.mcp-type-badge.type-number { background: #dcfce7; color: #16a34a; }

.mcp-required {
  color: #ef4444;
  font-weight: 600;
}

.mcp-optional {
  color: var(--color-text-quaternary);
}

.codemirror-wrapper {
  position: relative;
  border-radius: var(--radius-md);
  overflow: hidden;
  border: 1px solid var(--color-border-light);
}

.codemirror-container {
  border-radius: 0;
  border: none;
}

.codemirror-copy-btn {
  position: absolute;
  top: 8px;
  right: 8px;
  z-index: 10;
  padding: 4px 10px;
  font-size: 12px;
  border-radius: 4px;
  border: 1px solid var(--color-border-light);
  background: var(--color-bg-card);
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: all 0.15s;
  opacity: 0;
}

.codemirror-wrapper:hover .codemirror-copy-btn {
  opacity: 1;
}

.codemirror-copy-btn:hover {
  background: var(--color-primary-50);
  color: var(--color-primary);
  border-color: var(--color-primary);
}
</style>
