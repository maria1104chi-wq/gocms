// Pinia 状态管理 - 用户 Store
// 文件名: user.js
// 路径: /workspace/frontend/src/store/user.js

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useUserStore = defineStore('user', () => {
  // 状态
  const token = ref(localStorage.getItem('token') || '')
  const userInfo = ref(JSON.parse(localStorage.getItem('user') || '{}'))

  // 计算属性
  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => userInfo.value.role === 3)
  const isEditor = computed(() => userInfo.value.role >= 2)

  // 方法
  function setToken(newToken) {
    token.value = newToken
    localStorage.setItem('token', newToken)
  }

  function setUserInfo(info) {
    userInfo.value = info
    localStorage.setItem('user', JSON.stringify(info))
  }

  function login(newToken, info) {
    setToken(newToken)
    setUserInfo(info)
  }

  function logout() {
    token.value = ''
    userInfo.value = {}
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  return {
    token,
    userInfo,
    isLoggedIn,
    isAdmin,
    isEditor,
    setToken,
    setUserInfo,
    login,
    logout,
  }
})
