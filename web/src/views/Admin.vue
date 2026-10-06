<template>
  <div class="page">
    <h1>管理后台</h1>
    <div class="admin-layout">
      <div class="admin-nav">
        <router-link to="/admin/libraries">媒体库</router-link>
        <router-link to="/admin/scan">扫描</router-link>
        <router-link to="/admin/users">用户</router-link>
        <router-link to="/admin/files">文件</router-link>
        <router-link to="/admin/settings">设置</router-link>
        <router-link to="/admin/logs">实时日志</router-link>
      </div>
      <div class="admin-body">
        <router-view />
      </div>
    </div>
  </div>
</template>
