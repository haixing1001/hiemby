<template>
  <div class="login-wrap">
    <div class="card login-card">
      <h1>emby-server</h1>
      <div class="form-row">
        <label>用户名</label>
        <input v-model="username" @keyup.enter="doLogin" placeholder="用户名" autofocus />
      </div>
      <div class="form-row">
        <label>密码</label>
        <input v-model="password" type="password" @keyup.enter="doLogin" placeholder="密码" />
      </div>
      <button class="btn primary" :disabled="loading" @click="doLogin" style="width:100%">
        {{ loading ? '登录中…' : '登录' }}
      </button>
      <div class="err" v-if="err">{{ err }}</div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { api, setAuth } from '../api.js'

const router = useRouter()
const username = ref('')
const password = ref('')
const loading = ref(false)
const err = ref('')

async function doLogin() {
  err.value = ''
  if (!username.value) { err.value = '请输入用户名'; return }
  loading.value = true
  try {
    const r = await api('/emby/Users/AuthenticateByName', {
      method: 'POST',
      body: { Username: username.value, Pw: password.value }
    })
    setAuth(r.AccessToken, r.User)
    router.push('/')
  } catch (e) {
    err.value = e.message
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-wrap { min-height: 100vh; display: flex; align-items: center; justify-content: center; padding: 20px; }
.login-card { width: 340px; }
.login-card h1 { text-align: center; color: var(--accent); margin-bottom: 24px; font-size: 24px; }
</style>
