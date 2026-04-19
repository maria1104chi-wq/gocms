// Vue Router 路由配置
// 文件名: index.js
// 路径: /workspace/frontend/src/router/index.js

import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    name: 'Home',
    component: () => import('@/views/Home.vue'),
    meta: { title: '首页' },
  },
  {
    path: '/article/:slug',
    name: 'ArticleDetail',
    component: () => import('@/views/ArticleDetail.vue'),
    meta: { title: '文章详情' },
  },
  {
    path: '/category/:id',
    name: 'Category',
    component: () => import('@/views/Category.vue'),
    meta: { title: '分类' },
  },
  {
    path: '/search',
    name: 'Search',
    component: () => import('@/views/Search.vue'),
    meta: { title: '搜索' },
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { title: '登录' },
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/Register.vue'),
    meta: { title: '注册' },
  },
  {
    path: '/admin',
    name: 'Admin',
    component: () => import('@/views/admin/Dashboard.vue'),
    meta: { title: '管理后台', requiresAuth: true, requiresRole: 3 },
    children: [
      {
        path: '',
        name: 'AdminDashboard',
        component: () => import('@/views/admin/Dashboard.vue'),
      },
      {
        path: 'articles',
        name: 'AdminArticles',
        component: () => import('@/views/admin/ArticleList.vue'),
        meta: { requiresAuth: true, requiresRole: 2 },
      },
      {
        path: 'article/edit/:id?',
        name: 'AdminArticleEdit',
        component: () => import('@/views/admin/ArticleEdit.vue'),
        meta: { requiresAuth: true, requiresRole: 2 },
      },
      {
        path: 'users',
        name: 'AdminUsers',
        component: () => import('@/views/admin/UserList.vue'),
        meta: { requiresAuth: true, requiresRole: 3 },
      },
      {
        path: 'sensitive',
        name: 'AdminSensitive',
        component: () => import('@/views/admin/SensitiveWords.vue'),
        meta: { requiresAuth: true, requiresRole: 3 },
      },
      {
        path: 'stats',
        name: 'AdminStats',
        component: () => import('@/views/admin/Stats.vue'),
        meta: { requiresAuth: true, requiresRole: 3 },
      },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/NotFound.vue'),
    meta: { title: '页面不存在' },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) {
      return savedPosition
    } else {
      return { top: 0 }
    }
  },
})

// 路由守卫
router.beforeEach((to, from, next) => {
  // 设置页面标题
  document.title = to.meta.title ? `${to.meta.title} - 中文个人博客` : '中文个人博客'
  
  // 检查是否需要登录
  if (to.meta.requiresAuth) {
    const token = localStorage.getItem('token')
    const user = JSON.parse(localStorage.getItem('user') || '{}')
    
    if (!token) {
      next('/login')
      return
    }
    
    // 检查角色权限
    if (to.meta.requiresRole && user.role < to.meta.requiresRole) {
      next('/admin')
      return
    }
  }
  
  next()
})

export default router
