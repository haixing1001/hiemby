<template>
  <h2 style="margin-top:0">媒体库管理</h2>
  <div class="err" v-if="err">{{ err }}</div>
  <table class="tbl">
    <thead><tr><th>名称</th><th>路径</th><th>类型</th><th>条目数</th><th></th></tr></thead>
    <tbody>
      <tr v-for="l in libs" :key="l.Id">
        <td>{{ l.Name }}</td>
        <td class="muted">{{ l.Path }}</td>
        <td>{{ l.Kind }}</td>
        <td>{{ l.Count }}</td>
        <td><button class="btn small danger" @click="remove(l)">删除</button></td>
      </tr>
    </tbody>
  </table>
  <h2>添加媒体库</h2>
  <div class="card">
    <div class="form-inline">
      <div class="form-row"><label>名称</label><input v-model="f.Name" placeholder="电影" /></div>
      <div class="form-row"><label>路径</label><input v-model="f.Path" placeholder="/media/movies" /></div>
      <div class="form-row" style="max-width:160px"><label>类型</label>
        <select v-model="f.Kind">
          <option value="movies">movies</option>
          <option value="tvshows">tvshows</option>
          <option value="music">music</option>
        </select>
      </div>
      <button class="btn primary" @click="add">添加</button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { api } from '../../api.js'

const libs = ref([])
const err = ref('')
const f = ref({ Name: '', Path: '', Kind: 'movies' })

async function load() {
  err.value = ''
  try { libs.value = await api('/api/admin/libraries') }
  catch (e) { err.value = e.message }
}

async function add() {
  if (!f.value.Path) { err.value = '请填写路径'; return }
  try {
    await api('/api/admin/libraries', { method: 'POST', body: f.value })
    f.value = { Name: '', Path: '', Kind: 'movies' }
    load()
  } catch (e) { err.value = e.message }
}

async function remove(l) {
  if (!confirm(`删除媒体库「${l.Name}」？（仅删除索引，不删除文件）`)) return
  try {
    await api(`/api/admin/libraries/${l.Id}`, { method: 'DELETE' })
    load()
  } catch (e) { err.value = e.message }
}

onMounted(load)
</script>
