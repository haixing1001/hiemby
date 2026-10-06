<template>
  <h2 style="margin-top:0">用户管理</h2>
  <div class="err" v-if="err">{{ err }}</div>
  <table class="tbl">
    <thead><tr><th>用户名</th><th>管理员</th><th>设备（已用/上限）</th><th></th></tr></thead>
    <tbody>
      <tr v-for="u in users" :key="u.Id">
        <td>{{ u.Name }}</td>
        <td>{{ u.IsAdmin ? '是' : '否' }}</td>
        <td>{{ u.Devices }} / {{ u.MaxDevices }}</td>
        <td style="white-space:nowrap">
          <button class="btn small" @click="edit(u)">编辑</button>
          <button class="btn small danger" @click="remove(u)" style="margin-left:6px">删除</button>
        </td>
      </tr>
    </tbody>
  </table>

  <h2>新增用户</h2>
  <div class="card">
    <div class="form-inline">
      <div class="form-row"><label>用户名</label><input v-model="f.Name" /></div>
      <div class="form-row"><label>密码（≥3位）</label><input v-model="f.Password" type="password" /></div>
      <div class="form-row" style="max-width:120px"><label>设备上限</label><input v-model.number="f.MaxDevices" type="number" min="1" /></div>
      <div class="form-row" style="max-width:140px; flex:0"><label>管理员</label>
        <input v-model="f.IsAdmin" type="checkbox" style="width:auto; margin-top:10px" />
      </div>
      <button class="btn primary" @click="add">新增</button>
    </div>
  </div>

  <div class="card" v-if="editing" style="margin-top:18px">
    <h2 style="margin-top:0">编辑 {{ editing.Name }}</h2>
    <div class="form-inline">
      <div class="form-row"><label>新密码（留空不改）</label><input v-model="e.Password" type="password" /></div>
      <div class="form-row" style="max-width:120px"><label>设备上限</label><input v-model.number="e.MaxDevices" type="number" min="1" /></div>
      <button class="btn primary" @click="save">保存</button>
      <button class="btn" @click="editing = null">取消</button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { api } from '../../api.js'

const users = ref([])
const err = ref('')
const f = ref({ Name: '', Password: '', IsAdmin: false, MaxDevices: 2 })
const editing = ref(null)
const e = ref({ Password: '', MaxDevices: 2 })

async function load() {
  err.value = ''
  try { users.value = await api('/api/admin/users') }
  catch (e2) { err.value = e2.message }
}

async function add() {
  if (!f.value.Name || !f.value.Password) { err.value = '用户名和密码必填'; return }
  try {
    await api('/api/admin/users', { method: 'POST', body: f.value })
    f.value = { Name: '', Password: '', IsAdmin: false, MaxDevices: 2 }
    load()
  } catch (e2) { err.value = e2.message }
}

function edit(u) {
  editing.value = u
  e.value = { Password: '', MaxDevices: u.MaxDevices }
}

async function save() {
  const body = {}
  if (e.value.Password) body.Password = e.value.Password
  body.MaxDevices = e.value.MaxDevices
  try {
    await api(`/api/admin/users/${editing.value.Id}`, { method: 'PUT', body })
    editing.value = null
    load()
  } catch (e2) { err.value = e2.message }
}

async function remove(u) {
  if (!confirm(`删除用户「${u.Name}」？`)) return
  try {
    await api(`/api/admin/users/${u.Id}`, { method: 'DELETE' })
    load()
  } catch (e2) { err.value = e2.message }
}

onMounted(load)
</script>
