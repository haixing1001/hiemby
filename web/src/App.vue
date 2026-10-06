<template>
  <div class="navbar" v-if="showNav">
    <span class="brand" @click="$router.push('/')">emby-server</span>
    <nav>
      <router-link to="/">媒体库</router-link>
      <router-link to="/admin" v-if="admin">管理后台</router-link>
    </nav>
    <span class="spacer"></span>
    <span class="user">{{ userName }}</span>
    <button class="btn small" @click="logout">退出</button>
  </div>
  <router-view />
</template>

<script setup>
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getUser, clearAuth, isAdmin, onAuthChange } from './api.js'

const route = useRoute()
const router = useRouter()
// bumped on login/logout so auth-dependent computeds re-evaluate
const authTick = ref(0)
onAuthChange(() => authTick.value++)
const showNav = computed(() => route.path !== '/login')
const admin = computed(() => { authTick.value; return isAdmin() })
const userName = computed(() => { authTick.value; return getUser()?.Name || '' })

function logout() {
  clearAuth()
  router.push('/login')
}
</script>
