// API 请求封装
// 文件名: api.js
// 路径: /workspace/frontend/src/api/api.js

import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

// 创建 axios 实例
const request = axios.create({
  baseURL: '/api',
  timeout: 10000,
})

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    // 从 localStorage 获取 token
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response) => {
    const res = response.data
    
    // 如果返回的状态码不是 0，说明有错误
    if (res.code !== 0) {
      ElMessage.error(res.message || '请求失败')
      
      // 401: 未授权，跳转到登录页
      if (res.code === 401) {
        localStorage.removeItem('token')
        localStorage.removeItem('user')
        router.push('/login')
      }
      
      return Promise.reject(new Error(res.message || '请求失败'))
    }
    
    return res
  },
  (error) => {
    ElMessage.error(error.message || '网络错误')
    return Promise.reject(error)
  }
)

export default request

// 文章相关 API
export const articleApi = {
  // 获取文章列表
  getList(params) {
    return request.get('/articles', { params })
  },
  
  // 获取文章详情
  getDetail(slug) {
    return request.get(`/articles/${slug}`)
  },
  
  // 获取热门文章
  getTop(limit = 10) {
    return request.get('/articles/top', { params: { limit } })
  },
  
  // 点赞文章
  like(id) {
    return request.post(`/articles/${id}/like`)
  },
  
  // 分享文章
  share(id) {
    return request.post(`/articles/${id}/share`)
  },
  
  // 创建文章
  create(data) {
    return request.post('/articles', data)
  },
  
  // 更新文章
  update(id, data) {
    return request.put(`/articles/${id}`, data)
  },
  
  // 删除文章
  delete(id) {
    return request.delete(`/articles/${id}`)
  },
}

// 评论相关 API
export const commentApi = {
  // 获取文章评论
  getByArticle(articleId, params) {
    return request.get(`/comments/article/${articleId}`, { params })
  },
  
  // 发表评论
  create(articleId, data) {
    return request.post(`/comments/article/${articleId}`, data)
  },
  
  // 删除评论 (管理员)
  delete(id) {
    return request.delete(`/comments/${id}`)
  },
}

// 分类相关 API
export const categoryApi = {
  // 获取所有分类
  getList() {
    return request.get('/categories')
  },
}

// 用户相关 API
export const userApi = {
  // 注册
  register(data) {
    return request.post('/users/register', data)
  },
  
  // 登录
  login(data) {
    return request.post('/users/login', data)
  },
  
  // 发送短信验证码
  sendSMSCode(data) {
    return request.post('/users/sms/send', data)
  },
  
  // 验证短信验证码
  verifySMSCode(data) {
    return request.post('/users/sms/verify', data)
  },
  
  // 获取个人信息
  getProfile() {
    return request.get('/users/profile')
  },
}

// 管理后台 API
export const adminApi = {
  // 获取用户列表
  getUsers(params) {
    return request.get('/admin/users', { params })
  },
  
  // 更新用户角色
  updateUserRole(id, data) {
    return request.put(`/admin/users/${id}/role`, data)
  },
  
  // 获取敏感词列表
  getSensitiveWords() {
    return request.get('/admin/sensitive')
  },
  
  // 添加敏感词
  addSensitiveWord(data) {
    return request.post('/admin/sensitive', data)
  },
  
  // 删除敏感词
  deleteSensitiveWord(word) {
    return request.delete(`/admin/sensitive/${word}`)
  },
  
  // 获取统计数据
  getStats() {
    return request.get('/admin/stats')
  },
}
