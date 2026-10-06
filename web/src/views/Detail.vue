<template>
  <div class="page">
    <button class="btn small" @click="$router.back()" style="margin-bottom:14px">← 返回</button>
    <div class="err" v-if="err">{{ err }}</div>
    <div v-if="item">
      <img v-if="hasBackdrop" :src="imgUrl(item.Id, 'Backdrop')" class="backdrop" alt="" />
      <div class="detail-head">
        <div class="detail-poster">
          <img v-if="hasPoster" :src="imgUrl(item.Id, 'Primary')" :alt="item.Name" />
        </div>
        <div class="detail-info">
          <h1>{{ item.Name }}</h1>
          <div class="meta-line">
            <span v-if="item.ProductionYear">{{ item.ProductionYear }}</span>
            <span v-if="item.CommunityRating">★ {{ item.CommunityRating.toFixed(1) }}</span>
            <span>{{ fmtRuntime(item.RunTimeTicks) }}</span>
            <span v-if="item.Type">{{ item.Type }}</span>
          </div>
          <p class="overview">{{ item.Overview || '暂无简介' }}</p>
          <button class="btn primary" @click="playing = true" v-if="!playing">
            ▶ {{ resumeText }}
          </button>
        </div>
      </div>
      <div v-if="playing" style="margin-top:20px">
        <div class="video-wrap">
          <video ref="videoEl" :src="streamUrl(item.Id)" controls autoplay playsinline
                 @timeupdate="onTime" @pause="onStop" @ended="onStop">
            <track :src="subtitleUrl(item.Id, 0)" kind="subtitles" srclang="zh" label="中文" default />
          </video>
        </div>
        <p class="muted" style="margin-top:8px; font-size:13px">直连播放，无转码</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api, imgUrl, streamUrl, subtitleUrl, fmtRuntime } from '../api.js'

const route = useRoute()
const item = ref(null)
const err = ref('')
const playing = ref(false)
const videoEl = ref(null)
let lastReport = 0
let resumeTicks = 0

const hasPoster = computed(() => !!(item.value && item.value.ImageTags && item.value.ImageTags.Primary))
const hasBackdrop = computed(() => !!(item.value && item.value.BackdropImageTags))
const resumeText = computed(() => {
  if (resumeTicks > 0) {
    const m = Math.floor(resumeTicks / 600000000)
    return `继续播放（${m} 分钟处）`
  }
  return '播放'
})

const posTicks = () => Math.round((videoEl.value?.currentTime || 0) * 10000000)

async function reportPlaying() {
  try {
    await api('/emby/Sessions/Playing', {
      method: 'POST',
      body: { ItemId: item.value.Id, PositionTicks: posTicks() }
    })
  } catch {}
}

async function reportStopped() {
  try {
    await api('/emby/Sessions/Playing/Stopped', {
      method: 'POST',
      body: { ItemId: item.value.Id, PositionTicks: posTicks(), Failed: false }
    })
  } catch {}
}

function onTime() {
  const now = Date.now()
  if (now - lastReport > 10000) {
    lastReport = now
    reportPlaying()
  }
}

function onStop() {
  reportStopped()
}

onMounted(async () => {
  try {
    const r = await api(`/emby/Items/${route.params.id}`)
    item.value = r
    resumeTicks = r.UserData?.PlaybackPositionTicks || 0
  } catch (e) {
    err.value = e.message
  }
})

watch(playing, async (v) => {
  if (v) {
    await nextTick()
    if (videoEl.value && resumeTicks > 0) {
      videoEl.value.currentTime = resumeTicks / 10000000
    }
    lastReport = Date.now()
    reportPlaying()
  }
})

onBeforeUnmount(() => {
  if (playing.value) reportStopped()
})
window.addEventListener('beforeunload', () => {
  if (playing.value && item.value) {
    navigator.sendBeacon('/emby/Sessions/Playing/Stopped', JSON.stringify({
      ItemId: item.value.Id, PositionTicks: posTicks(), Failed: false
    }))
  }
})
</script>
