import { createRouter, createWebHashHistory } from 'vue-router'
import { getToken, isAdmin } from './api.js'
import Login from './views/Login.vue'
import Home from './views/Home.vue'
import Library from './views/Library.vue'
import Detail from './views/Detail.vue'
import Admin from './views/Admin.vue'
import AdminLibraries from './views/admin/Libraries.vue'
import AdminScan from './views/admin/Scan.vue'
import AdminUsers from './views/admin/Users.vue'
import AdminFiles from './views/admin/Files.vue'
import AdminSettings from './views/admin/Settings.vue'
import AdminLogs from './views/admin/Logs.vue'

const routes = [
  { path: '/login', component: Login },
  { path: '/', component: Home, meta: { auth: true } },
  { path: '/library/:id', component: Library, meta: { auth: true } },
  { path: '/item/:id', component: Detail, meta: { auth: true } },
  {
    path: '/admin', component: Admin, meta: { auth: true, admin: true },
    children: [
      { path: '', redirect: '/admin/libraries' },
      { path: 'libraries', component: AdminLibraries },
      { path: 'scan', component: AdminScan },
      { path: 'users', component: AdminUsers },
      { path: 'files', component: AdminFiles },
      { path: 'settings', component: AdminSettings },
      { path: 'logs', component: AdminLogs }
    ]
  }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

router.beforeEach((to) => {
  if (to.meta.auth && !getToken()) return '/login'
  if (to.meta.admin && !isAdmin()) return '/'
})

export default router
