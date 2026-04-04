<template>
  <div class="not-found-view">
    <n-result
      status="404"
      title="404"
      description="抱歉，您访问的页面不存在"
      size="large"
    >
      <template #footer>
        <n-space justify="center">
          <n-button @click="goBack">
            <template #icon>
              <n-icon><ArrowBackOutline /></n-icon>
            </template>
            返回上一页
          </n-button>
          <n-button type="primary" @click="goHome">
            <template #icon>
              <n-icon><HomeOutline /></n-icon>
            </template>
            回到首页
          </n-button>
        </n-space>
      </template>
    </n-result>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { NResult, NSpace, NButton, NIcon } from 'naive-ui'
import { ArrowBackOutline, HomeOutline } from '@vicons/ionicons5'

const router = useRouter()

function goBack() {
  router.back()
}

function goHome() {
  router.push('/')
}
</script>

<style scoped>
.not-found-view {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 50vh;
  padding: 24px;
}
</style>