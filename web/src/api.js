// API client: all requests go through here, token attached automatically.
const TOKEN_KEY = 'emby_token'
const USER_KEY = 'emby_user'

export function getToken() {
  return localStorage.getItem(TOKEN_KEY) || ''
}

export function getUser() {
  try {
    return JSON.parse(localStorage.getItem(USER_KEY) || 'null')
  } catch {
    return null
  }
}

export function setAuth(token, user) {
  localStorage.setItem(TOKEN_KEY, token)
  localStorage.setItem(USER_KEY, JSON.stringify(user))
  emitAuth()
}

export function clearAuth() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
  emitAuth()
}

// Reactive auth change notification (localStorage itself is not reactive).
const authListeners = new Set()
export function onAuthChange(fn) {
  authListeners.add(fn)
  return () => authListeners.delete(fn)
}
function emitAuth() {
  authListeners.forEach((fn) => fn())
}

export function isAdmin() {
  const u = getUser()
  return !!(u && u.Policy && u.Policy.IsAdministrator)
}

export async function api(path, { method = 'GET', body, query } = {}) {
  const headers = {}
  const token = getToken()
  if (token) headers['X-Emby-Token'] = token
  if (body !== undefined) headers['Content-Type'] = 'application/json'
  let url = path
  if (query) url += '?' + new URLSearchParams(query).toString()
  const res = await fetch(url, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined
  })
  if (res.status === 401) {
    clearAuth()
    if (location.hash !== '#/login') location.hash = '#/login'
    throw new Error('登录已过期')
  }
  if (!res.ok) {
    const t = await res.text()
    let msg = t
    try {
      const j = JSON.parse(t)
      if (j.error) msg = j.error
    } catch {}
    throw new Error(msg || ('请求失败 ' + res.status))
  }
  const ct = res.headers.get('content-type') || ''
  if (ct.includes('application/json')) return res.json()
  return res.text()
}

// URL builders for <img> / <video> / <track> / EventSource (can't set headers)
export function imgUrl(itemId, kind = 'Primary') {
  return `/emby/Items/${itemId}/Images/${kind}?api_key=${encodeURIComponent(getToken())}`
}

export function streamUrl(itemId) {
  return `/emby/Videos/${itemId}/stream?api_key=${encodeURIComponent(getToken())}`
}

export function subtitleUrl(itemId, idx = 0) {
  return `/emby/Videos/${itemId}/Subtitles/${idx}/Stream?api_key=${encodeURIComponent(getToken())}`
}

export function logsUrl() {
  return `/api/admin/logs?api_key=${encodeURIComponent(getToken())}`
}

export function fmtRuntime(ticks) {
  if (!ticks) return ''
  const m = Math.round(ticks / 600000000)
  const h = Math.floor(m / 60)
  return h > 0 ? `${h}时${m % 60}分` : `${m}分`
}

export function fmtSize(bytes) {
  if (bytes == null) return '-'
  const u = ['B', 'KB', 'MB', 'GB', 'TB']
  let i = 0
  let v = bytes
  while (v >= 1024 && i < u.length - 1) { v /= 1024; i++ }
  return v.toFixed(1) + ' ' + u[i]
}

export function fmtTime(ts) {
  if (!ts) return '-'
  return new Date(ts * 1000).toLocaleString()
}
