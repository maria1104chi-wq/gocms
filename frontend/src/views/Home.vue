// 首页视图 - 包含轮播、分类显示、无限滚动
// 文件名: Home.vue
// 路径: /workspace/frontend/src/views/Home.vue

<template>
  <div class="home">
    <!-- 顶部导航 -->
    <header class="header">
      <div class="container">
        <div class="logo">
          <router-link to="/">中文个人博客</router-link>
        </div>
        <nav class="nav">
          <router-link to="/">首页</router-link>
          <a v-for="cat in categories" :key="cat.id" :href="`/category/${cat.id}`">{{ cat.name }}</a>
        </nav>
        <div class="search-box">
          <input v-model="searchKeyword" @keyup.enter="doSearch" placeholder="搜索文章..." />
          <button @click="doSearch"><el-icon><Search /></el-icon></button>
        </div>
        <div class="user-actions">
          <template v-if="userStore.isLoggedIn">
            <router-link to="/admin">管理后台</router-link>
            <span>{{ userStore.userInfo.username }}</span>
            <button @click="handleLogout">退出</button>
          </template>
          <template v-else>
            <router-link to="/login">登录</router-link>
            <router-link to="/register">注册</router-link>
          </template>
        </div>
      </div>
    </header>

    <!-- 轮播图 -->
    <section class="carousel">
      <el-carousel :interval="4000" height="300px">
        <el-carousel-item v-for="item in topArticles.slice(0, 5)" :key="item.id">
          <div class="carousel-item" @click="$router.push(`/article/${item.slug}`)">
            <img v-if="item.cover_image" :src="item.cover_image" :alt="item.title" />
            <div class="carousel-info">
              <h3>{{ item.title }}</h3>
              <p>{{ item.summary }}</p>
            </div>
          </div>
        </el-carousel-item>
      </el-carousel>
    </section>

    <!-- 主要内容区 -->
    <main class="main container">
      <!-- 左侧文章列表 -->
      <div class="content-left">
        <div class="section-title">
          <h2>最新文章</h2>
        </div>
        
        <div class="article-list">
          <div 
            v-for="article in articles" 
            :key="article.id" 
            class="article-item"
            @click="$router.push(`/article/${article.slug}`)"
          >
            <div class="article-cover">
              <img v-if="article.cover_image" :src="article.cover_image" :alt="article.title" />
            </div>
            <div class="article-info">
              <h3>{{ article.title }}</h3>
              <p class="summary">{{ article.summary }}</p>
              <div class="meta">
                <span><el-icon><User /></el-icon> {{ article.author?.username || '匿名' }}</span>
                <span><el-icon><Clock /></el-icon> {{ formatDate(article.created_at) }}</span>
                <span><el-icon><View /></el-icon> {{ article.view_count }}</span>
                <span><el-icon><Star /></el-icon> {{ article.like_count }}</span>
                <span><el-icon><ChatDotRound /></el-icon> {{ article.comment_count }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 加载更多 -->
        <div class="load-more" v-intersection-observer="loadMore">
          <el-button v-if="loading" loading>加载中...</el-button>
          <el-button v-else-if="hasMore" type="primary">加载更多</el-button>
          <p v-else>没有更多文章了</p>
        </div>
      </div>

      <!-- 右侧边栏 -->
      <aside class="sidebar">
        <!-- 热门文章排行 -->
        <div class="widget">
          <h3>🔥 热门排行</h3>
          <ul class="hot-list">
            <li v-for="(article, index) in hotArticles" :key="article.id" @click="$router.push(`/article/${article.slug}`)">
              <span class="rank" :class="'rank-' + (index + 1)">{{ index + 1 }}</span>
              <span class="title">{{ article.title }}</span>
              <span class="views">{{ article.view_count }}</span>
            </li>
          </ul>
        </div>

        <!-- 最新评论 -->
        <div class="widget">
          <h3>💬 最新评论</h3>
          <ul class="comment-list">
            <li v-for="comment in recentComments" :key="comment.id">
              <p class="comment-content">{{ comment.content }}</p>
              <p class="comment-meta">
                <span>{{ comment.user?.username || '匿名访客' }}</span>
                <span>{{ comment.ip_location }}</span>
              </p>
            </li>
          </ul>
        </div>
      </aside>
    </main>

    <!-- 页脚 -->
    <footer class="footer">
      <div class="container">
        <p>&copy; 2024 中文个人博客 CMS. All rights reserved.</p>
      </div>
    </footer>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/store/user'
import { articleApi, categoryApi, commentApi } from '@/api/api'
import { Search, User, Clock, View, Star, ChatDotRound } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const router = useRouter()
const userStore = useUserStore()

// 数据
const categories = ref([])
const articles = ref([])
const topArticles = ref([])
const hotArticles = ref([])
const recentComments = ref([])
const searchKeyword = ref('')

// 分页
const page = ref(1)
const pageSize = 10
const loading = ref(false)
const hasMore = ref(true)

// 格式化日期
const formatDate = (dateStr) => {
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN')
}

// 加载分类
const loadCategories = async () => {
  try {
    const res = await categoryApi.getList()
    categories.value = res.data
  } catch (error) {
    console.error('加载分类失败', error)
  }
}

// 加载文章列表
const loadArticles = async () => {
  if (loading.value) return
  
  loading.value = true
  try {
    const res = await articleApi.getList({
      page: page.value,
      pageSize,
    })
    
    const newArticles = res.data.list
    if (newArticles.length < pageSize) {
      hasMore.value = false
    }
    
    articles.value = [...articles.value, ...newArticles]
    page.value++
  } catch (error) {
    ElMessage.error('加载文章失败')
  } finally {
    loading.value = false
  }
}

// 加载更多 (无限滚动)
const loadMore = () => {
  if (hasMore.value && !loading.value) {
    loadArticles()
  }
}

// 加载热门文章
const loadHotArticles = async () => {
  try {
    const res = await articleApi.getTop(10)
    hotArticles.value = res.data
  } catch (error) {
    console.error('加载热门文章失败', error)
  }
}

// 加载轮播文章
const loadTopArticles = async () => {
  try {
    const res = await articleApi.getTop(5)
    topArticles.value = res.data
  } catch (error) {
    console.error('加载轮播文章失败', error)
  }
}

// 加载最新评论
const loadRecentComments = async () => {
  // TODO: 实现获取最新评论的 API
  recentComments.value = []
}

// 搜索
const doSearch = () => {
  if (searchKeyword.value.trim()) {
    router.push({ path: '/search', query: { q: searchKeyword.value } })
  }
}

// 退出登录
const handleLogout = () => {
  userStore.logout()
  ElMessage.success('已退出登录')
  router.push('/')
}

// 自定义指令：元素可见时触发
const intersectionObserver = {
  mounted(el, binding) {
    const observer = new IntersectionObserver((entries) => {
      if (entries[0].isIntersecting) {
        binding.value()
      }
    })
    observer.observe(el)
  },
}

onMounted(() => {
  loadCategories()
  loadArticles()
  loadHotArticles()
  loadTopArticles()
  loadRecentComments()
})
</script>

<style scoped>
.home {
  min-height: 100vh;
}

.header {
  background: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  position: sticky;
  top: 0;
  z-index: 100;
}

.header .container {
  display: flex;
  align-items: center;
  padding: 15px 20px;
  gap: 20px;
}

.logo a {
  font-size: 24px;
  font-weight: bold;
  color: #409eff;
}

.nav {
  display: flex;
  gap: 15px;
}

.nav a {
  color: #666;
  transition: color 0.3s;
}

.nav a:hover {
  color: #409eff;
}

.search-box {
  display: flex;
  margin-left: auto;
}

.search-box input {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px 0 0 4px;
  outline: none;
}

.search-box button {
  padding: 8px 12px;
  background: #409eff;
  color: #fff;
  border: none;
  border-radius: 0 4px 4px 0;
  cursor: pointer;
}

.user-actions {
  display: flex;
  gap: 10px;
}

.carousel {
  margin: 20px auto;
  max-width: 1200px;
}

.carousel-item {
  cursor: pointer;
  background: #f5f7fa;
  height: 100%;
  display: flex;
  overflow: hidden;
}

.carousel-item img {
  width: 60%;
  height: 100%;
  object-fit: cover;
}

.carousel-info {
  width: 40%;
  padding: 30px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.carousel-info h3 {
  font-size: 20px;
  margin-bottom: 10px;
}

.main {
  display: grid;
  grid-template-columns: 1fr 300px;
  gap: 20px;
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

.section-title h2 {
  font-size: 20px;
  margin-bottom: 15px;
  padding-left: 10px;
  border-left: 4px solid #409eff;
}

.article-item {
  display: flex;
  gap: 15px;
  background: #fff;
  padding: 15px;
  margin-bottom: 15px;
  border-radius: 8px;
  cursor: pointer;
  transition: transform 0.3s, box-shadow 0.3s;
}

.article-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.article-cover {
  width: 200px;
  flex-shrink: 0;
}

.article-cover img {
  width: 100%;
  height: 120px;
  object-fit: cover;
  border-radius: 4px;
}

.article-info {
  flex: 1;
}

.article-info h3 {
  font-size: 18px;
  margin-bottom: 10px;
  color: #333;
}

.summary {
  color: #666;
  font-size: 14px;
  margin-bottom: 10px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.meta {
  display: flex;
  gap: 15px;
  font-size: 13px;
  color: #999;
}

.meta span {
  display: flex;
  align-items: center;
  gap: 4px;
}

.load-more {
  text-align: center;
  padding: 20px;
}

.sidebar {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.widget {
  background: #fff;
  padding: 15px;
  border-radius: 8px;
}

.widget h3 {
  font-size: 16px;
  margin-bottom: 15px;
  padding-bottom: 10px;
  border-bottom: 1px solid #eee;
}

.hot-list {
  list-style: none;
}

.hot-list li {
  display: flex;
  align-items: center;
  padding: 8px 0;
  cursor: pointer;
  transition: background 0.3s;
}

.hot-list li:hover {
  background: #f5f7fa;
}

.rank {
  width: 24px;
  height: 24px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 12px;
  margin-right: 10px;
  flex-shrink: 0;
}

.rank-1 { background: #f56c6c; }
.rank-2 { background: #e6a23c; }
.rank-3 { background: #f9ae3d; }
.rank-4, .rank-5 { background: #909399; }

.title {
  flex: 1;
  font-size: 14px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.views {
  font-size: 12px;
  color: #999;
}

.comment-list {
  list-style: none;
}

.comment-list li {
  padding: 10px 0;
  border-bottom: 1px solid #eee;
}

.comment-content {
  font-size: 14px;
  color: #666;
  margin-bottom: 5px;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.comment-meta {
  font-size: 12px;
  color: #999;
  display: flex;
  justify-content: space-between;
}

.footer {
  background: #333;
  color: #fff;
  text-align: center;
  padding: 20px;
  margin-top: 40px;
}

@media (max-width: 768px) {
  .main {
    grid-template-columns: 1fr;
  }
  
  .sidebar {
    order: -1;
  }
  
  .article-item {
    flex-direction: column;
  }
  
  .article-cover {
    width: 100%;
  }
}
</style>
