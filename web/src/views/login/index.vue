<template>
  <div class="login-bg">
    <div class="login-card">
      <div class="login-brand">Forks</div>
      <n-form ref="loginFormRef" :model="loginForm" :rules="loginRules" class="login-form">
        <n-form-item path="token" label="访问令牌">
          <n-input v-model:value="loginForm.token" placeholder="请输入访问令牌" size="large" />
        </n-form-item>
        <n-form-item>
          <n-button type="primary" size="large" block @click="onLogin" :loading="loading">登录</n-button>
        </n-form-item>
      </n-form>
      <div class="login-extra">
        请使用启动时生成的访问令牌进行登录
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { NForm, NFormItem, NInput, NButton, useMessage } from 'naive-ui'
import { getRepos } from '@/api/repos'

const router = useRouter()
const message = useMessage()

// 登录表单
const loginFormRef = ref(null)
const loginForm = ref({ token: '' })
const loading = ref(false)
const loginRules = {
  token: [
    { required: true, message: '请输入访问令牌', trigger: 'blur' }
  ]
}

async function onLogin() {
  try {
    await loginFormRef.value?.validate()

    if (!loginForm.value.token) {
      message.error('请输入访问令牌')
      return
    }

    loading.value = true

    // 临时保存token进行验证
    const tempToken = loginForm.value.token

    // 先临时设置token到localStorage供request拦截器使用
    localStorage.setItem('token', tempToken)

    // 验证token是否有效
    try {
      console.log('🔐 验证token有效性...')
      const response = await getRepos({ page: 1, page_size: 1 })

      if (response && response.data && response.data.code === 0) {
        console.log('✅ Token验证成功')

        // 保存用户信息（token已经设置）
        localStorage.setItem('userInfo', JSON.stringify({
          id: 1,
          username: 'user',
          role: 'user'
        }))

        // 触发自定义事件通知导航栏更新
        window.dispatchEvent(new CustomEvent('userLoggedIn'))

        message.success('登录成功！')
        router.push('/')
      } else {
        console.log('❌ Token验证失败:', response)
        message.error('访问令牌无效，请检查后重试')
        // 验证失败，清除临时token
        localStorage.removeItem('token')
      }
    } catch (apiError) {
      console.error('❌ API验证失败:', apiError)
      message.error('访问令牌验证失败，请检查网络连接和令牌是否正确')
      // 验证失败，清除临时token
      localStorage.removeItem('token')
    }
  } catch (e) {
    if (e.message) {
      message.error(e.message)
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-bg {
  min-height: 100vh;
  width: 100vw;
  background: linear-gradient(120deg, #4CAF50 0%, #2196F3 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
}
.login-card {
  background: #fff;
  border-radius: 18px;
  box-shadow: 0 8px 32px 0 rgba(60, 60, 60, 0.18);
  padding: 48px 40px 32px 40px;
  min-width: 340px;
  max-width: 90vw;
  display: flex;
  flex-direction: column;
  align-items: center;
}
.login-brand {
  font-size: 2.2rem;
  font-weight: 900;
  letter-spacing: 2px;
  margin-bottom: 32px;
  background: linear-gradient(90deg, #4CAF50 0%, #2196F3 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  font-family: 'Fira Code', 'Lato', 'Segoe UI', 'Arial', sans-serif;
}
.brand-accent {
  color: #4CAF50;
  -webkit-text-fill-color: #4CAF50;
  background: none;
  font-weight: 900;
}
.login-form {
  width: 320px;
  max-width: 100%;
}
.login-extra {
  margin-top: 18px;
  text-align: center;
  color: #888;
  font-size: 15px;
}
</style>