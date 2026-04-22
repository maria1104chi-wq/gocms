# 中文个人博客 CMS 系统

## 项目简介

这是一个功能完善的中文个人博客及内容管理系统（CMS），采用 Golang + Gin 框架开发后端，Vue 3 开发前端，MySQL 作为主数据库，Redis 作为缓存数据库。支持 Docker + Caddy 一键部署到 Debian 云服务器。
# 中文个人 CMS 系统

## 项目简介

这是一个功能完善的中文内容管理系统（CMS），采用 Golang + Gin 框架开发后端，Vue 3 开发前端，MySQL 作为主数据库，Redis 作为缓存数据库。支持 Docker + Caddy 一键部署到 Debian 云服务器。

## 技术栈

### 后端
- **语言**: Golang 1.21+
- **Web框架**: Gin
- **ORM**: GORM
- **数据库**: MySQL 8.0
- **缓存**: Redis
- **认证**: JWT
- **密码加密**: bcrypt

### 前端
- **框架**: Vue 3 (Composition API)
- **构建工具**: Vite
- **UI组件**: Element Plus
- **Markdown编辑器**: v-md-editor
- **HTTP客户端**: Axios

### 部署
- **容器化**: Docker + Docker Compose
- **反向代理**: Caddy (自动HTTPS)
- **操作系统**: Debian 11/12

## 功能特性

### 前台功能
- ✅ 首页轮播图展示
- ✅ 文章分类显示
- ✅ 无限滚动加载（鼠标下拉）
- ✅ 站内搜索
- ✅ 侧栏热门文章排行（点击率TOP10）
- ✅ 文章详情页
  - 点赞、分享功能
  - 评论、跟评（支持匿名）
  - 显示点击数、点赞数、分享数、评论数
  - 评论显示访客IP归属地
  - 敏感词自动过滤
- ✅ 伪静态URL

### 后台管理
- ✅ 多账号权限管理
  - 系统管理员：账号管理、版块管理、备份管理、日志管理
  - 版块管理员：负责特定版块内容发布和管理
- ✅ 账号绑定手机号，登录需短信验证
- ✅ Markdown富文本编辑器
  - 支持图片、PDF、视频上传（限定大小）
- ✅ 自动SEO优化
- ✅ 敏感词自动屏蔽
- ✅ 流量统计分析
  - IP来源统计
  - 内容和版块点击量统计
  - 其他网站运营数据

### 安全特性
- ✅ JWT令牌认证
- ✅ 密码bcrypt加密
- ✅ SQL注入防护（GORM参数化查询）
- ✅ XSS防护
- ✅ CSRF防护
- ✅ 敏感词过滤
- ✅ 请求频率限制（Redis实现）
- ✅ IP归属地查询

## 项目结构

```
/workspace
├── backend/                    # 后端代码
│   ├── cmd/
│   │   └── main.go            # 程序入口
│   ├── internal/
│   │   ├── config/            # 配置管理
│   │   ├── database/          # 数据库连接
│   │   ├── handler/           # HTTP处理器
│   │   ├── middleware/        # 中间件
│   │   ├── model/             # 数据模型
│   │   ├── service/           # 业务逻辑
│   │   └── utils/             # 工具函数
│   ├── static/uploads/        # 上传文件存储
│   └── go.mod                 # Go模块定义
├── frontend/                   # 前端代码
│   ├── src/
│   │   ├── api/               # API接口
│   │   ├── assets/            # 静态资源
│   │   ├── components/        # 组件
│   │   ├── router/            # 路由
│   │   ├── store/             # 状态管理
│   │   ├── utils/             # 工具函数
│   │   └── views/             # 页面视图
│   └── public/                # 公共资源
├── deploy/                     # 部署配置
│   ├── init.sql               # 数据库初始化脚本
│   ├── docker-compose.yml     # Docker编排
│   └── Caddyfile              # Caddy配置
└── docs/                       # 文档
    └── README.md              # 项目说明
```

## 快速开始

### 环境要求
- Docker 20.10+
- Docker Compose 2.0+
- Debian 11/12 云服务器

### 部署步骤

1. **克隆项目到服务器**
```bash
cd /var/www
git clone <your-repo> blog-cms
cd blog-cms
```

2. **配置环境变量**
```bash
cp .env.example .env
# 编辑.env文件，修改数据库密码、JWT密钥等
```

3. **启动服务**
```bash
docker-compose up -d
```

4. **查看日志**
```bash
docker-compose logs -f
```

5. **访问网站**
- 前台：https://your-domain.com
- 后台：https://your-domain.com/admin

### 默认账号
- 用户名：admin
- 密码：admin123 (首次登录后请立即修改)

## API文档

### 文章相关
- `GET /api/articles` - 获取文章列表
- `GET /api/articles/:slug` - 获取文章详情
- `POST /api/articles` - 创建文章 (需认证)
- `PUT /api/articles/:id` - 更新文章 (需认证)
- `DELETE /api/articles/:id` - 删除文章 (需认证)
- `POST /api/articles/:id/like` - 点赞文章
- `POST /api/articles/:id/share` - 分享文章

### 评论相关
- `GET /api/comments/article/:article_id` - 获取文章评论
- `POST /api/comments/article/:article_id` - 发表评论
- `DELETE /api/comments/:id` - 删除评论 (管理员)

### 用户相关
- `POST /api/users/register` - 注册
- `POST /api/users/login` - 登录
- `POST /api/users/sms/send` - 发送短信验证码
- `GET /api/users/profile` - 获取个人信息 (需认证)

## 数据库设计

详见 `deploy/init.sql`，包含以下核心表：
- users - 用户表
- categories - 分类表
- articles - 文章表
- comments - 评论表
- sensitive_words - 敏感词库
- sms_codes - 短信验证码
- visit_logs - 访问日志
- likes - 点赞记录

## 敏感词库

项目使用 [Sensitive-lexicon](https://github.com/konsheng/Sensitive-lexicon) 中文敏感词库，已导入数据库。管理员可在后台动态添加/删除敏感词。

## 安全建议

1. **生产环境配置**
   - 修改默认管理员密码
   - 设置强JWT密钥
   - 启用防火墙只开放必要端口
   - 定期备份数据库

2. **短信服务**
   - 接入阿里云/腾讯云短信服务
   - 设置发送频率限制防刷

3. **文件上传**
   - 限制文件大小和类型
   - 使用独立OSS存储

4. **监控告警**
   - 配置日志收集
   - 设置异常告警

## 许可证

MIT License

## 联系方式

如有问题请提交Issue或联系开发者。
