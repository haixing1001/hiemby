<template>
  <h2 style="margin-top:0">设置</h2>
  <div class="err" v-if="err">{{ err }}</div>
  <div class="ok" v-if="ok">{{ ok }}</div>
  <div class="card">
    <div class="form-row" style="max-width:420px">
      <label>TMDB API Key（用于元数据刮削）</label>
      <input v-model="tmdbKey" type="password" placeholder="留空则不修改" />
    </div>
    <button class="btn primary" @click="save">保存设置</button>
  </div>
  <h2>元数据刮削</h2>
  <div class="card">
    <p class="muted" style="margin-bottom:12px">对缺少简介/评分的影片调用 TMDB 搜索并回填元数据。</p>
    <button class="btn" @click="scrape">开始刮削</button>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { api } from '../../api.js'

const tmdbKey = ref('')
const err = ref('')
const ok = ref('')

onMounted(async () => {
  try {
    const s = await api('/api/admin/settings')
    if (s.tmdb_key && s.tmdb_key !== '***') tmdbKey.value = s.tmdb_key
  } catch (e) { err.value = e.message }
})

async function save() {
  err.value = ''; ok.value = ''
  try {
    const body = {}
    if (tmdbKey.value && tmdbKey.value !== '***') body.tmdb_key = tmdbKey.value
    await api('/api/admin/settings', { method: 'POST', body })
    ok.value = '已保存'
  } catch (e) { err.value = e.message }
}

async function scrape() {
  err.value = ''; ok.value = ''
  try {
    await api('/api/admin/scrape', { method: 'POST' })
    ok.value = '刮削任务已启动，可在实时日志查看进度'
  } catch (e) { err.value = e.message }
}
</script>
