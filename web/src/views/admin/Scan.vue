<template>
  <h2 style="margin-top:0">媒体库扫描</h2>
  <div class="err" v-if="err">{{ err }}</div>
  <div class="card">
    <p class="muted" style="margin-bottom:14px">
      全量扫描：重新索引所有文件（含 NFO、海报、ffprobe 媒体信息）<br />
      增量刷新：仅处理新增或修改过的文件
    </p>
    <div style="display:flex; gap:10px; flex-wrap:wrap">
      <button class="btn primary" :disabled="running" @click="scan('full')">全量扫描</button>
      <button class="btn" :disabled="running" @click="scan('incremental')">增量刷新</button>
    </div>
    <p style="margin-top:14px">
      状态：<span :class="running ? 'ok' : 'muted'">{{ running ? '● 扫描中…' : '空闲' }}</span>
    </p>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { api } from '../../api.js'

const running = ref(false)
const err = ref('')
let timer = null

async function status() {
  try {
    const r = await api('/api/admin/scan/status')
    running.value = !!r.running
  } catch {}
}

async function scan(mode) {
  err.value = ''
  try {
    await api('/api/admin/scan', { method: 'POST', query: { mode } })
    running.value = true
  } catch (e) { err.value = e.message }
}

onMounted(() => {
  status()
  timer = setInterval(status, 3000)
})
onBeforeUnmount(() => clearInterval(timer))
</script>
