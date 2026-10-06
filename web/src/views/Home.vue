<template>
  <div class="page">
    <h1>媒体库</h1>
    <div class="err" v-if="err">{{ err }}</div>
    <div class="grid" v-if="libs.length">
      <div class="poster-card" v-for="lib in libs" :key="lib.Id" @click="$router.push('/library/' + lib.Id)">
        <div class="thumb">{{ lib.CollectionType || '媒体库' }}</div>
        <div class="title">{{ lib.Name }}</div>
        <div class="sub">{{ lib.CollectionType }}</div>
      </div>
    </div>
    <p class="muted" v-else-if="!err">暂无媒体库，请在管理后台添加。</p>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { api, getUser } from '../api.js'

const libs = ref([])
const err = ref('')

onMounted(async () => {
  try {
    const u = getUser()
    const r = await api(`/emby/Users/${u.Id}/Views`)
    libs.value = r.Items || []
  } catch (e) {
    err.value = e.message
  }
})
</script>
