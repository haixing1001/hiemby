<template>
  <div class="page">
    <button class="btn small" @click="$router.push('/')" style="margin-bottom:14px">← 媒体库</button>
    <h1>{{ libName }}</h1>
    <div class="err" v-if="err">{{ err }}</div>
    <div class="grid">
      <div class="poster-card" v-for="it in items" :key="it.Id" @click="$router.push('/item/' + it.Id)">
        <div class="thumb">
          <img v-if="it.ImageTags && it.ImageTags.Primary" :src="imgUrl(it.Id)" :alt="it.Name" loading="lazy" />
          <span v-else>{{ it.Type }}</span>
        </div>
        <div class="title">{{ it.Name }}</div>
        <div class="sub">{{ it.ProductionYear || '' }}</div>
      </div>
    </div>
    <div class="pager" v-if="total > limit">
      <button class="btn small" :disabled="page === 0" @click="gotoPage(page - 1)">上一页</button>
      <span class="muted">第 {{ page + 1 }} / {{ pages }} 页（共 {{ total }}）</span>
      <button class="btn small" :disabled="page + 1 >= pages" @click="gotoPage(page + 1)">下一页</button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, getUser, imgUrl } from '../api.js'

const route = useRoute()
const router = useRouter()
const items = ref([])
const total = ref(0)
const err = ref('')
const libName = ref('')
const limit = 60
const page = ref(0)
const pages = computed(() => Math.ceil(total.value / limit))

async function load() {
  err.value = ''
  try {
    const u = getUser()
    const libId = route.params.id
    const r = await api(`/emby/Users/${u.Id}/Items`, {
      query: { ParentId: libId, Limit: limit, StartIndex: page.value * limit }
    })
    items.value = r.Items || []
    total.value = r.TotalRecordCount || 0
    // 取媒体库名
    if (!libName.value) {
      const v = await api(`/emby/Users/${u.Id}/Views`)
      const lib = (v.Items || []).find(x => x.Id === libId)
      libName.value = lib ? lib.Name : '媒体库'
    }
  } catch (e) {
    err.value = e.message
  }
}

function gotoPage(p) {
  page.value = p
  router.replace({ query: { p } })
  load()
  window.scrollTo(0, 0)
}

onMounted(() => {
  page.value = parseInt(route.query.p || '0', 10) || 0
  load()
})
watch(() => route.params.id, () => { page.value = 0; libName.value = ''; load() })
</script>
