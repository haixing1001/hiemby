<template>
  <h2 style="margin-top:0">实时日志</h2>
  <div style="margin-bottom:12px; display:flex; gap:10px; align-items:center">
    <span :class="connected ? 'ok' : 'muted'" style="font-size:13px">{{ connected ? '● 已连接' : '○ 未连接' }}</span>
    <button class="btn small" @click="clear">清空</button>
    <label class="muted" style="font-size:13px"><input type="checkbox" v-model="autoscroll" style="width:auto" /> 自动滚动</label>
  </div>
  <div class="logbox" ref="box">
    <div v-for="(l, i) in lines" :key="i" :class="'lv-' + l.level">{{ l.time }} [{{ l.level }}] {{ l.msg }}</div>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { logsUrl } from '../../api.js'

const lines = ref([])
const box = ref(null)
const connected = ref(false)
const autoscroll = ref(true)
let es = null

function scroll() {
  if (autoscroll.value && box.value) {
    nextTick(() => { box.value.scrollTop = box.value.scrollHeight })
  }
}

function clear() { lines.value = [] }

onMounted(() => {
  es = new EventSource(logsUrl())
  es.onopen = () => { connected.value = true }
  es.onerror = () => { connected.value = false }
  es.onmessage = (ev) => {
    // "HH:MM:SS [LEVEL] message"
    const m = ev.data.match(/^(\S+) \[(\w+)\] ([\s\S]*)$/)
    if (m) lines.value.push({ time: m[1], level: m[2], msg: m[3] })
    else lines.value.push({ time: '', level: 'INFO', msg: ev.data })
    if (lines.value.length > 1000) lines.value.splice(0, lines.value.length - 1000)
    scroll()
  }
})

onBeforeUnmount(() => { if (es) es.close() })
</script>
