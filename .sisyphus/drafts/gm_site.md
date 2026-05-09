# Draft: GM Site - 群友图片展示网站

## 需求总结 (Requirements)
- 前后端分离架构：Vue前端 + Go/Echo后端 + SQLite数据库
- 样式参考：`gm_site_demo.html`（赛博朋克/土味酷炫风格）
- 核心功能：图片展示弹窗、实时数据、注册登录

## Demo 结构分析 (from gm_site_demo.html)

### 页面布局
1. **粒子背景**：随机emoji漂浮动画
2. **顶部横幅**：彩虹渐变背景、glitch文字效果
3. **跑马灯**：3条不同颜色的滚动文字
4. **数据统计栏**：在线人数、新增成员、安全运行天数（假数据）
5. **中间图片区**：8张SVG占位图卡片
6. **侧边栏**：公告、实时计数、安全检测
7. **浮动弹窗**：3个可拖拽弹窗，关闭后5-13秒自动重新出现
8. **悬浮按钮**：客服、TOP、VIP

### 需要从假数据改为真实数据的功能
- 在线人数统计 → 后端实时统计
- 图片卡片 → 真实用户上传图片（通过兰空图床转链接）
- 弹窗点击 → 实际登录/注册/图片查看功能
- 访客计数器 → 真实访问统计

## 技术决策 (已确认)
- [x] Vue版本：**Vue 3 + Composition API**
- [x] 前端构建工具：**Vite**
- [x] UI风格：**保留Demo核心风格，适当简化**（保留彩虹边框、色彩体系，减少闪烁频率）
- [x] 实时数据：**在线人数 + 访客计数 + 新注册数**

## 技术决策 (已确认)
- [x] Vue版本：**Vue 3 + Composition API**
- [x] 前端构建工具：**Vite**
- [x] UI风格：**保留Demo核心风格，适当简化**（保留彩虹边框、色彩体系，减少闪烁频率）
- [x] 实时数据：**在线人数 + 访客计数 + 新注册数**
- [x] 测试策略：**TDD（测试驱动开发）**
- [x] 管理员面板：**独立管理后台页面**
- [x] 图片弹窗：**类似Demo可拖拽浮动弹窗 + 大图展示**
- [x] UI组件库：**纯手写CSS**（还原赛博朋克风格）
- [x] 测试框架：**Go: testing+testify | Vue: vitest**
- [x] 注册审核：**邮箱通知管理员 + 后台审核 + 用户待审核状态页**
- [x] Go项目结构：**标准Go布局** (cmd/, internal/, migrations/)
- [x] 数据库设计：**我来设计，计划中包含完整表结构**
- [x] Demo装饰：**移除假内容**（假广告、360认证等）
- [x] 图片分类：**需要相册/标签**，前端可筛选
- [x] 排除范围：**只排除用户个人主页**，评论和搜索保留
- [x] 实时推送：**WebSocket实时推送**在线人数等

## 功能边界
### IN（范围内）
- 注册/登录（邮箱 + 双Token认证）
- 访客浏览（无需登录）
- 管理员审核注册 + 后台管理图片CRUD
- 普通用户管理自己上传的图片
- 图片上传 → 兰空图床转换链接
- 可拖拽浮动弹窗查看大图
- 图片相册/标签分类筛选
- 图片评论
- 图片搜索
- WebSocket实时推送：在线人数、访客计数、新注册数
- SMTP邮件配置 + 兰空图床配置

### OUT（明确排除）
- 用户个人主页/头像
- 点赞/收藏功能
- 支付/会员系统
- 假广告/假认证等Demo装饰元素

## 功能边界
### IN（范围内）
- 用户注册登录（邮箱 + 管理员审核）
- 双Token认证（Refresh + Access）
- 访客可浏览
- 图片上传 → 兰空图床转换
- 图片弹窗展示
- 实时数据（在线人数等）
- 管理员CRUD所有图片
- 普通用户CRUD自己上传的图片
- 配置项管理（SMTP、兰空图床）

### OUT（明确排除）
- 待确认

## 配置项
1. SMTP邮箱配置（TLS安全连接）
2. 兰空图床配置（API地址、Token等）

## Research 结论

### Lsky Pro API (兰空图床)
- **认证**: `POST /api/v1/tokens` (email+password) → bearer token
- **上传**: `POST /api/v1/upload` multipart/form-data → 响应 `data.links.url` 即为图片链接
- **管理**: 支持 list/delete/update 图片、相册(album)管理
- **策略**: 支持多种存储策略(strategy)，如本地/OSS/S3等
- **推荐**: 使用 V1 API（开源免费版），V2 是付费版

### Echo + JWT + SQLite 最佳实践
- **Echo**: v4 稳定版；v5 2026年发布但仍推荐 v4
- **JWT**: `golang-jwt/jwt/v5` — Access Token (短期15min) + Refresh Token (长期7d) 双Token
- **SQLite**: `modernc.org/sqlite` (纯Go无CGO) 推荐；`golang-migrate/migrate` 做迁移
- **SMTP**: `gomail` 或 `net/smtp` + TLS
- **文件上传**: Echo `FormFile()` → 构造multipart请求pipe到兰空图床
- **项目结构**: 推荐标准Go布局 `cmd/`, `internal/`, `migrations/`
