<template>
  <h2 style="margin-top:0">文件管理</h2>
  <div class="err" v-if="err">{{ err }}</div>
  <div class="card">
    <div style="margin-bottom:12px; display:flex; gap:8px; align-items:center; flex-wrap:wrap">
      <button class="btn small" @click="go('..')" :disabled="cur === '/'">↑ 上级</button>
      <span class="muted" style="font-size:13px; word-break:break-all">{{ cur }}</span>
    </div>
    <table class="tbl">
      <thead><tr><th>名称</th><th>大小</th><th>修改时间</th><th></th></tr></thead>
      <tbody>
        <tr v-for="it in items" :key="it.Path">
          <td>
            <a v-if="it.IsDir" href="#" @click.prevent="go(it.Path)">📁 {{ it.Name }}</a>
            <span v-else>📄 {{ it.Name }}</span>
          </td>
          <td class="muted">{{ it.IsDir ? '-' : fmtSize(it.Size) }}</td>
          <td class="muted">{{ fmtTime(it.ModTime) }}</td>
          <td>
            <button v-if="!it.IsDir" class="btn small danger" @click="remove(it)">删除</button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { api, fmtSize, fmtTime } from '../../api.js'

const cur = ref('/')
const items = ref([])
const err = ref('')

async function load(path) {
  err.value = ''
  try {
    const r = await api('/api/admin/files', { query: { path } })
    cur.value = r.Path
    items.value = r.Items || []
  } catch (e) { err.value = e.message }
}

function go(p) {
  if (p === '..') {
    const parts = cur.value.split('/').filter(Boolean)
    parts.pop()
    load('/' + parts.join('/'))
  } else {
    load(p)
  }
}

async function remove(it) {
  if (!confirm(`删除文件「${it.Name}」？不可恢复！`)) return
  try {
    await api('/api/admin/files', { method: 'DELETE', query: { path: it.Path } })
    load(cur.value)
  } catch (e) { err.value = e.message }
}

onMounted(() => load('/'))
</script>
