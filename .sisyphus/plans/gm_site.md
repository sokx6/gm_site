# GM Site — 群友图片展示网站

## TL;DR

> **Quick Summary**: 前后端分离的赛博朋克风格图片展示网站，Vue 3 + Go/Echo + SQLite，支持用户注册登录（双Token）、图片上传至兰空图床、可拖拽弹窗大图浏览、WebSocket实时数据、管理员后台管理。

> **Deliverables**:
> - Go 后端 API 服务器（Echo + SQLite + JWT双Token + WebSocket）
> - Vue 3 前端 SPA（赛博朋克风格，粒子背景，浮动弹窗）
> - 管理员后台（用户审核 + 图片/相册/评论管理）
> - 数据库迁移脚本（SQLite）
> - 配置文件（SMTP + 兰空图床 + 管理员邮箱）
> - 完整 TDD 测试套件（Go testify + Vue vitest）

> **Estimated Effort**: Extra Large (45+ tasks)
> **Parallel Execution**: YES — 8 waves, max 9 concurrent
> **Critical Path**: T2→T4→T5→T6→T8→T9→T12→T18→T30→T31→T42→F1-F4

---

## Context

### Original Request
构建群友图片展示网站，样式参考 `gm_site_demo.html`。前端 Vue，后端 Go+Echo+SQLite。核心功能：图片展示弹窗、实时用户数据（非假数据）、注册登录（邮箱+双Token）、管理员图片CRUD、普通用户管理自己的图片。图片上传后端转兰空图床链接。SMTP和兰空图床配置化。

### Interview Summary
**Key Discussions**:
- Vue 3 + Composition API + Vite，纯手写 CSS 还原赛博朋克风格
- 保留 Demo 核心风格（彩虹边框、粒子背景、跑马灯），移除假广告/安全认证等装饰
- 实时数据：在线人数 + 访客计数 + 新注册数，WebSocket 推送
- TDD 测试驱动开发（Go: testing+testify, Vue: vitest）
- 独立管理后台页面，配置文件中指定管理员邮箱
- 图片需要相册/标签分类 + 评论 + 搜索
- 部署：本地开发环境；163邮箱SMTP；兰空图床已有实例；Git版本控制
- 图片上传限制 10MB

**Research Findings**:
- **Lsky Pro API**: `POST /api/v1/upload` multipart/form-data，响应 `data.links.url` 即图片链接；Bearer Token 认证
- **Echo v4**: 稳定版，推荐 golang-jwt/jwt/v5、modernc.org/sqlite（纯Go）、golang-migrate/migrate、gomail SMTP
- **标准 Go 布局**: `cmd/server`, `internal/{config,handler,middleware,model,repository,service,websocket}`, `migrations/`

### Metis Review
**Identified Gaps** (addressed):
- 部署环境 → 本地开发机，无 Docker
- 兰空图床状态 → 已有实例，配置化
- SMTP → 163邮箱，配置化
- 管理员账号 → 配置文件指定管理员邮箱
- 图片大小限制 → 10MB
- WebSocket 心跳 → 30s 默认
- 评论删除 → 图片作者和管理员可删
- 安全运行天数起始日期 → 配置项

**Auto-Resolved from Metis**:
- 用户注册验证码 → 不需要验证码（邮箱验证 + 管理员审核）
- 用户注销 → Access Token 短期自动过期，Refresh Token 可主动撤销
- 并发上传 → 串行处理（SQLite 写锁限制）
- 前端路由 → Vue Router hash 模式（简单部署）

---

## Work Objectives

### Core Objective
构建一个功能完整的赛博朋克风格群友图片展示网站，实现用户体系、图片管理、实时数据和后台管理。

### Concrete Deliverables
- `backend/` — Go Echo 后端项目（API、WebSocket、数据库）
- `frontend/` — Vue 3 + Vite 前端项目
- `backend/migrations/` — SQLite 迁移脚本
- `backend/config.yaml` — 配置文件模板
- `.git/` — Git 仓库（初始提交）

### Definition of Done
- [ ] `go run ./cmd/server` 启动后端，所有 API 端点可访问
- [ ] `cd frontend && npm run dev` 启动前端，页面正常渲染
- [ ] 访客可浏览图片，无需登录
- [ ] 用户可注册 → 管理员收到邮件 → 管理员后台审核 → 用户登录
- [ ] 登录用户可上传图片、编辑/删除自己的图片
- [ ] 管理员可增删改查所有图片
- [ ] 图片弹窗可拖拽，关闭后自动重新出现
- [ ] WebSocket 实时推送在线人数等
- [ ] `go test ./...` 全部通过
- [ ] `cd frontend && npm test` 全部通过

### Must Have
- 双Token认证（Access 15min + Refresh 7d）
- 访客可浏览，登录后可上传
- 兰空图床图片转换
- 管理员审核注册 + 后台管理图片
- Demo 风格弹窗（可拖拽、自动复活）
- 配置文件驱动（SMTP、兰空图床、管理员邮箱、群友名称）

### Must NOT Have (Guardrails)
- 禁止在代码中硬编码密钥/密码 → 全部走配置文件 + 环境变量
- 禁止跨用户越权操作 → 中间件强制校验 owner/admin
- 禁止前端直接调用兰空图床 → 必须走后端代理
- 禁止 Demo 中的假广告/360认证等虚假内容
- 禁止 Refresh Token 不过期
- 禁止 SQL 注入 → 参数化查询
- 禁止评论无审核地展示 → 前端 XSS 防护

---

## Verification Strategy

> **ZERO HUMAN INTERVENTION** — ALL verification is agent-executed. No exceptions.

### Test Decision
- **Infrastructure exists**: NO（从零搭建）
- **Automated tests**: TDD（测试驱动开发）
- **Framework**: Go: `testing` + `testify` | Vue: `vitest` + `@vue/test-utils`
- **TDD Flow**: 每个任务先写 RED（失败测试）→ GREEN（最小实现）→ REFACTOR

### QA Policy
每个任务包含 Agent-Executed QA Scenarios，证据保存至 `.sisyphus/evidence/task-{N}-{scenario-slug}.{ext}`。

- **API/Backend**: Bash (curl) — 发请求、断言状态码和响应字段
- **Frontend/UI**: Playwright — 打开浏览器、导航、填写表单、断言DOM、截图
- **CLI**: Bash — 运行命令、断言输出

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Start Immediately — 项目初始化 + 基础设施):
├── T1: Git 仓库初始化 [quick]
├── T2: Go 项目脚手架 [quick]
├── T3: Vue 项目脚手架 [quick]
├── T4: 配置文件模块 [quick]
├── T5: 数据库连接 + 迁移运行器 [quick]
├── T6: SQLite 迁移脚本 [quick]
├── T7: Go 数据模型定义 [quick]

Wave 2 (After Wave 1 — 后端认证体系):
├── T8: User Repository [quick]
├── T9: JWT 服务（生成/验证/刷新）[quick]
├── T10: Auth 中间件 [quick]
├── T11: 邮箱服务（SMTP TLS）[quick]
├── T12: 注册接口 [quick]
├── T13: 登录 + Token 刷新接口 [quick]
├── T14: 管理员用户审核接口 [quick]

Wave 3 (After Wave 2 — 后端图片核心):
├── T15: Album Repository + CRUD 接口 [quick]
├── T16: 兰空图床客户端 [deep]
├── T17: 图片上传接口 [deep]
├── T18: 图片 CRUD 接口（权限控制）[deep]
├── T19: Comment Repository + CRUD 接口 [quick]
├── T20: 图片搜索接口 [quick]

Wave 4 (After Wave 1 — 后端 WebSocket + 前端基础):
├── T21: WebSocket Hub + 客户端管理 [deep]
├── T22: 在线统计追踪 + WebSocket 广播 [quick]
├── T23: 全局 CSS（Demo 风格迁移）[visual-engineering]
├── T24: 粒子背景组件 [visual-engineering]
├── T25: 彩虹横幅 + 跑马灯组件 [visual-engineering]
├── T26: 页面布局框架（三栏）[visual-engineering]

Wave 5 (After Wave 4 — 前端核心组件):
├── T27: 图片卡片 + 画廊网格 [visual-engineering]
├── T28: 图片浮动弹窗（可拖拽+自动复活）[visual-engineering]
├── T29: 相册/标签筛选组件 [visual-engineering]
├── T30: 搜索栏组件 [visual-engineering]
├── T31: 评论区组件 [visual-engineering]
├── T32: 统计栏组件（WebSocket连接）[visual-engineering]
├── T33: 访客计数器 + 悬浮按钮 [visual-engineering]

Wave 6 (After Wave 5 — 前端视图 + API层):
├── T34: API 客户端 + Token 拦截器 [quick]
├── T35: Auth Store (Pinia) + 路由守卫 [quick]
├── T36: 登录/注册/待审核页面 [visual-engineering]
├── T37: 首页视图（组装所有组件）[visual-engineering]

Wave 7 (After Wave 6 — 管理后台):
├── T38: 管理员图片管理视图 [visual-engineering]
├── T39: 管理员用户管理视图 [visual-engineering]
├── T40: 管理员相册管理视图 [quick]

Wave 8 (After Wave 7 — 集成 + 测试):
├── T41: 前端-后端联调 + CORS 配置 [deep]
├── T42: 端到端流程测试 [quick]
├── T43: Git 初始提交 [git]

Wave FINAL (After ALL tasks — 4 parallel reviews):
├── F1: Plan Compliance Audit (oracle)
├── F2: Code Quality Review (unspecified-high)
├── F3: Real Manual QA (unspecified-high + playwright)
├── F4: Scope Fidelity Check (deep)
```

### Dependency Matrix

| Task | Depends On | Blocks | Wave |
|------|-----------|--------|------|
| T1 | - | T43 | 1 |
| T2 | - | T8-T22 | 1 |
| T3 | - | T23-T40 | 1 |
| T4 | - | T11, T16, T17, T34 | 1 |
| T5 | T2 | T6, T8, T15, T19 | 1 |
| T6 | T5 | T8, T15, T19 | 1 |
| T7 | - | T8-T22, T34 | 1 |
| T8 | T5, T6, T7 | T12, T13, T14, T22 | 2 |
| T9 | T7 | T10, T12, T13 | 2 |
| T10 | T9 | T12-T14, T18, T34 | 2 |
| T11 | T4, T7 | T12 | 2 |
| T12 | T8, T9, T10, T11 | T35, T42 | 2 |
| T13 | T8, T9 | T34, T35, T42 | 2 |
| T14 | T8, T10 | T39, T42 | 2 |
| T15 | T5, T6, T7 | T17, T18, T29, T40 | 3 |
| T16 | T4, T7 | T17 | 3 |
| T17 | T4, T15, T16 | T18, T27, T38 | 3 |
| T18 | T10, T15, T17 | T27, T38, T42 | 3 |
| T19 | T5, T6, T7 | T31, T38 | 3 |
| T20 | T7, T15, T18 | T30, T37 | 3 |
| T21 | T2, T7 | T22, T32 | 4 |
| T22 | T8, T21 | T32 | 4 |
| T23 | T3 | T24-T33 | 4 |
| T24 | T23 | T26, T37 | 4 |
| T25 | T23 | T26, T37 | 4 |
| T26 | T24, T25 | T27-T33, T37 | 4 |
| T27 | T17, T18, T26 | T28, T37 | 5 |
| T28 | T26, T27 | T37 | 5 |
| T29 | T15, T26 | T37 | 5 |
| T30 | T20, T26 | T37 | 5 |
| T31 | T19, T26 | T37 | 5 |
| T32 | T21, T22, T26 | T37 | 5 |
| T33 | T26 | T37 | 5 |
| T34 | T4, T10, T13 | T35-T40 | 6 |
| T35 | T12, T13, T34 | T36, T37 | 6 |
| T36 | T34, T35 | T37 | 6 |
| T37 | T24-T36 | T41, T42 | 6 |
| T38 | T18, T19, T34, T35 | T41 | 7 |
| T39 | T14, T34, T35 | T41 | 7 |
| T40 | T15, T34, T35 | T41 | 7 |
| T41 | T37-T40 | T42 | 8 |
| T42 | T12, T13, T18, T41 | F1-F4 | 8 |
| T43 | T1, T42 | - | 8 |
| F1 | T43 | - | FINAL |
| F2 | T43 | - | FINAL |
| F3 | T43 | - | FINAL |
| F4 | T43 | - | FINAL |

### Agent Dispatch Summary

| Wave | Tasks | Profiles |
|------|-------|----------|
| 1 | 7 (T1-T7) | quick×7 |
| 2 | 7 (T8-T14) | quick×7 |
| 3 | 6 (T15-T20) | deep×3, quick×3 |
| 4 | 6 (T21-T26) | deep×1, quick×1, visual-engineering×4 |
| 5 | 7 (T27-T33) | visual-engineering×7 |
| 6 | 4 (T34-T37) | quick×2, visual-engineering×2 |
| 7 | 3 (T38-T40) | visual-engineering×2, quick×1 |
| 8 | 3 (T41-T43) | deep×1, quick×1, git×1 |
| FINAL | 4 (F1-F4) | oracle×1, unspecified-high×2, deep×1 |

---

## TODOs

- [x] 1. **Git 仓库初始化**

  **What to do**:
  - `git init` 在项目根目录
  - 创建 `.gitignore`（Go + Vue + Node 标准忽略规则，含 `.env`, `*.db`, `node_modules/`, `dist/`）
  - 创建 `README.md`（项目简介）

  **Must NOT do**:
  - 不要提交 `.env` 或任何含密钥的文件

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 纯初始化操作，无复杂逻辑

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with T2-T7)

  **Acceptance Criteria**:
  - [ ] `.git/` 目录存在
  - [ ] `.gitignore` 包含 `*.db`, `.env`, `node_modules/`, `dist/`, `*.exe`

  **QA Scenarios**:
  ```
  Scenario: Git repo initialized correctly
    Tool: Bash
    Steps:
      1. git status → 显示 "No commits yet" 且有未跟踪文件
      2. Test-Path .gitignore → true
    Expected Result: Git仓库就绪，.gitignore 存在
    Evidence: .sisyphus/evidence/task-1-git-init.txt
  ```

  **Commit**: NO（由 T43 统一初始提交）

- [x] 2. **Go 项目脚手架**

  **What to do**:
  - 创建 `backend/` 目录结构：`cmd/server/`, `internal/{config,handler,middleware,model,repository,service,websocket}`, `migrations/`
  - `go mod init gm_site` 初始化模块
  - 安装依赖：`echo` (v4), `golang-jwt/jwt/v5`, `modernc.org/sqlite`, `golang-migrate/migrate/v4`, `gomail`, `gorilla/websocket`, `testify`, `viper`(配置), `go-playground/validator`
  - 创建 `cmd/server/main.go` 骨架（Echo 实例创建、基础路由、优雅关闭）
  - 创建 `internal/config/config.go` stub

  **Must NOT do**:
  - 不要包含业务逻辑
  - 不要写 handler 实现

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 标准Go项目初始化，按约定创建目录和依赖

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with T1, T3-T7)
  - **Blocks**: T8-T22

  **Acceptance Criteria**:
  - [ ] `go build ./cmd/server` 编译成功
  - [ ] 目录结构符合标准 Go 布局

  **QA Scenarios**:
  ```
  Scenario: Go project compiles
    Tool: Bash
    Steps:
      1. cd backend && go build ./cmd/server
      2. 检查编译产物存在
    Expected Result: 编译成功，无错误
    Evidence: .sisyphus/evidence/task-2-go-build.txt
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 3. **Vue 3 + Vite 项目脚手架**

  **What to do**:
  - `npm create vite@latest frontend -- --template vue-ts` 创建项目
  - 安装依赖：`vue-router`, `pinia`, `axios`, `vitest`, `@vue/test-utils`, `jsdom`
  - 创建目录：`src/{components,views,stores,router,api,composables,assets/styles}`
  - 配置 `vite.config.ts`（dev 代理到后端 `localhost:1323`）
  - 创建 `tsconfig.json`（strict 模式）
  - 配置 `vitest` (vitest.config.ts)

  **Must NOT do**:
  - 不要安装 UI 组件库
  - 不要写组件实现

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 标准 Vue 项目初始化

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with T1-T2, T4-T7)
  - **Blocks**: T23-T40

  **Acceptance Criteria**:
  - [ ] `cd frontend && npm install` 成功
  - [ ] `cd frontend && npm run dev` 启动无错误
  - [ ] `cd frontend && npm run build` 构建成功

  **QA Scenarios**:
  ```
  Scenario: Vue project dev server starts
    Tool: Bash
    Steps:
      1. cd frontend && timeout 10 npm run dev (检查启动成功输出)
      2. curl http://localhost:5173 → 返回 HTML
    Expected Result: Dev server 启动，返回 Vue app HTML
    Evidence: .sisyphus/evidence/task-3-vue-init.txt
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 4. **配置管理模块**

  **What to do**:
  - 实现 `internal/config/config.go`，使用 `viper` 读取 `config.yaml`
  - 配置结构体包含：
    ```go
    Server:  { Port, Host }
    Database: { Path (sqlite文件路径) }
    JWT:     { AccessSecret, RefreshSecret, AccessExpire(15m), RefreshExpire(168h) }
    SMTP:    { Host, Port(587), Username, Password, From, AdminEmail, UseTLS }
    Lsky:    { BaseURL, Email, Password, Token(auto-fetched) }
    Admin:   { Email (管理员邮箱) }
  Site:    { Name (群友名称/站点名，默认"顾夏"), StartDate (安全运行起始日) }
  Upload:  { MaxSizeMB(10) }
  ```
  - 注意：`Site.Name` 为可配置的群友名称，"顾夏"仅为示例默认值。前端所有展示名称位置（横幅、跑马灯、计数器、页面标题）均应从后端 API 获取或读取此配置。
  - 创建 `config.yaml` 模板文件（含注释说明各字段）
  - 创建 `.env.example`（敏感配置建议用环境变量覆盖）

  **Must NOT do**:
  - 不要在代码中硬编码任何密钥/密码
  - 不要提交 `config.yaml` 中的真实密码

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 配置结构体定义 + viper 集成，标准操作

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with T1-T3, T5-T7)
  - **Blocks**: T11, T16, T17, T34

  **Acceptance Criteria**:
  - [ ] `go test ./internal/config -run TestLoadConfig` PASS（TDD：先写测试）
  - [ ] 配置文件模板可被解析

  **QA Scenarios**:
  ```
  Scenario: Config loads from yaml file
    Tool: Bash
    Steps:
      1. cd backend && go test ./internal/config -v -run TestLoadConfig
    Expected Result: 所有配置字段正确解析
    Evidence: .sisyphus/evidence/task-4-config-test.txt
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 5. **数据库连接 + 迁移运行器**

  **What to do**:
  - 实现 `internal/database/database.go`：初始化 `modernc.org/sqlite` 连接、连接池配置
  - 实现迁移运行器：集成 `golang-migrate/migrate/v4` + sqlite3 driver，`database.RunMigrations()` 函数
  - 在 `main.go` 中调用：加载配置 → 连接数据库 → 运行迁移
  - TDD：先写 `TestDatabaseConnection` 和 `TestRunMigrations`

  **Must NOT do**:
  - 不要写具体的迁移 SQL（留给 T6）
  - 不要使用 CGO 版本的 sqlite driver

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 数据库连接初始化 + migrate 集成，标准操作

  **Parallelization**:
  - **Can Run In Parallel**: YES（可并行于 T6）
  - **Parallel Group**: Wave 1 (with T1-T4, T6-T7)
  - **Depends On**: T2
  - **Blocks**: T6, T8, T15, T19

  **Acceptance Criteria**:
  - [ ] `go test ./internal/database -run TestDatabaseConnection` PASS
  - [ ] `go test ./internal/database -run TestRunMigrations` PASS（空迁移目录时）

  **QA Scenarios**:
  ```
  Scenario: Database file created on startup
    Tool: Bash
    Steps:
      1. cd backend && go run ./cmd/server (后台启动)
      2. Test-Path gm_site.db → true
    Expected Result: SQLite 数据库文件自动创建
    Evidence: .sisyphus/evidence/task-5-db-connect.txt
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 6. **SQLite 迁移脚本**

  **What to do**:
  - 在 `backend/migrations/` 创建迁移文件（up + down）：
    - `001_create_users.up.sql` / `.down.sql`：users 表（id, email, password_hash, nickname, role[user/admin], status[pending/approved/rejected], created_at, updated_at）
    - `002_create_albums.up.sql` / `.down.sql`：albums 表（id, name, description, created_by, created_at）
    - `003_create_images.up.sql` / `.down.sql`：images 表（id, album_id FK, title, description, tags(JSON array), lsky_url, thumbnail_url, uploaded_by FK, created_at, updated_at）
    - `004_create_comments.up.sql` / `.down.sql`：comments 表（id, image_id FK, user_id FK, content, created_at）
  - 创建 `internal/database/seed.go`：种子数据函数（可选初始管理员，由配置指定）

  **Must NOT do**:
  - 不要在迁移中硬编码具体用户数据
  - 不要忘记外键约束和索引

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: SQL 脚本编写，确定性工作

  **Parallelization**:
  - **Can Run In Parallel**: YES（可与 T5 并行）
  - **Parallel Group**: Wave 1 (with T1-T5, T7)
  - **Depends On**: T5
  - **Blocks**: T8, T15, T19

  **Acceptance Criteria**:
  - [ ] 迁移脚本 up+down 各 4 对文件
  - [ ] `go test ./internal/database -run TestRunMigrations` PASS（含 migration 文件后）

  **QA Scenarios**:
  ```
  Scenario: All migrations run successfully
    Tool: Bash
    Steps:
      1. rm -f gm_site.db
      2. cd backend && go test ./internal/database -v -run TestRunMigrations
      3. 查询 sqlite3 gm_site.db ".tables" → 包含 users, albums, images, comments
    Expected Result: 4张表创建成功，含外键和索引
    Evidence: .sisyphus/evidence/task-6-migrations.txt
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 7. **Go 数据模型定义**

  **What to do**:
  - 在 `internal/model/` 创建：
    - `user.go`：User struct + UserRole(admin/user) + UserStatus(pending/approved/rejected) 常量
    - `image.go`：Image struct（含 Tags []string JSON field）
    - `album.go`：Album struct
    - `comment.go`：Comment struct
  - 创建 `internal/model/request.go`：注册/登录/上传/CreateImage/UpdateImage 等请求 struct（含 validate tags）
  - 创建 `internal/model/response.go`：统一响应格式 `{code, message, data}` + 各接口响应 struct
  - 创建 `internal/model/token.go`：TokenPair(AccessToken, RefreshToken, ExpiresIn)

  **Must NOT do**:
  - 不要引入 ORM（GORM等）— 使用原生 struct tag `db:"column_name"`
  - 不要在 model 层写业务逻辑

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 纯结构体定义，无复杂逻辑

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with T1-T6)
  - **Blocks**: T8-T22, T34

  **Acceptance Criteria**:
  - [ ] `go build ./...` 编译通过
  - [ ] 所有 struct 含 validate tags（email, required, min/max 等）

  **QA Scenarios**:
  ```
  Scenario: Models compile and validate
    Tool: Bash
    Steps:
      1. cd backend && go build ./internal/model
    Expected Result: 编译无错误
    Evidence: .sisyphus/evidence/task-7-models.txt
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 8. **User Repository**

  **What to do**:
  - 实现 `internal/repository/user.go`：
    - `Create(user *model.User) error` — INSERT
    - `FindByEmail(email string) (*model.User, error)` — SELECT by email
    - `FindByID(id int64) (*model.User, error)` — SELECT by id
    - `UpdateStatus(id int64, status model.UserStatus) error` — 审核用
    - `ListPending() ([]model.User, error)` — 管理员查看待审核列表
    - `CountByStatus(status model.UserStatus) (int, error)` — 统计
  - TDD：先写 `user_repository_test.go`（使用内存 SQLite）

  **Must NOT do**:
  - 不要在此层做密码哈希 — 那是 service 层
  - 不要返回纯文本密码

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 标准 CRUD repository，SQL 参数化查询

  **Parallelization**:
  - **Can Run In Parallel**: YES（T8-T11 可并行）
  - **Parallel Group**: Wave 2 (with T9-T11)
  - **Depends On**: T5, T6, T7
  - **Blocks**: T12, T13, T14, T22

  **Acceptance Criteria**:
  - [ ] `go test ./internal/repository -run TestUserRepo` PASS（至少6个测试）
  - [ ] 参数化查询无 SQL 注入风险

  **QA Scenarios**:
  ```
  Scenario: Create and find user
    Tool: Bash
    Steps:
      1. cd backend && go test ./internal/repository -v -run TestUserRepo
    Expected Result: 所有 user repo 测试通过
    Evidence: .sisyphus/evidence/task-8-user-repo.txt
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 9. **JWT 服务（生成/验证/刷新）**

  **What to do**:
  - 实现 `internal/service/jwt.go`：
    - `GenerateAccessToken(userID int64, role string) (string, time.Time, error)` — 15min 有效期
    - `GenerateRefreshToken(userID int64) (string, error)` — 7d 有效期
    - `ValidateAccessToken(tokenStr string) (*Claims, error)` — 解析+验证
    - `ValidateRefreshToken(tokenStr string) (int64, error)` — 返回 userID
    - `GenerateTokenPair(userID, role) (*model.TokenPair, error)` — 一次生成 pair
  - Claims 结构体含：UserID, Role, exp, iat
  - 使用 `golang-jwt/jwt/v5`，密钥从 config 读取
  - TDD：测试 token 生成/验证/过期

  **Must NOT do**:
  - 不要硬编码密钥
  - 不要 Access Token 超过 30min
  - 不要在 Claims 中存敏感信息（密码等）

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: JWT 标准模式，清晰输入输出

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with T8, T10-T11)
  - **Depends On**: T7
  - **Blocks**: T10, T12, T13

  **Acceptance Criteria**:
  - [ ] `go test ./internal/service -run TestJWT` PASS
  - [ ] 过期 token 验证返回错误

  **QA Scenarios**:
  ```
  Scenario: Token generation and validation
    Tool: Bash
    Steps:
      1. cd backend && go test ./internal/service -v -run TestJWT
    Expected Result: 生成→验证→过期 全流程通过
    Evidence: .sisyphus/evidence/task-9-jwt.txt
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 10. **Auth 中间件**

  **What to do**:
  - 实现 `internal/middleware/auth.go`：
    - `AuthRequired()` — 验证 Access Token，注入 user info 到 context
    - `AdminRequired()` — 验证 role==admin
    - `OptionalAuth()` — 有 token 则注入，无 token 则继续（访客模式）
    - Context key 常量：`UserIDKey`, `UserRoleKey`
  - TDD：测试带/不带 token、过期 token、错误 token

  **Must NOT do**:
  - 不要在中间件中调用数据库 — 只验证 JWT
  - 不要对 OPTIONS 请求做 auth 检查

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Echo 中间件标准模式

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with T8-T9, T11)
  - **Depends On**: T9
  - **Blocks**: T12-T14, T18, T34

  **Acceptance Criteria**:
  - [ ] `go test ./internal/middleware -run TestAuthMiddleware` PASS
  - [ ] 无 token 返回 401
  - [ ] 错误 token 返回 401
  - [ ] 正确 token 放行并在 context 注入 UserID+Role

  **QA Scenarios**:
  ```
  Scenario: Auth middleware blocks unauthenticated
    Tool: Bash (curl)
    Steps:
      1. curl -X GET http://localhost:1323/api/protected → 401
      2. curl -H "Authorization: Bearer invalid" → 401
      3. curl -H "Authorization: Bearer <valid_token>" → 200
    Expected Result: 401/200 符合预期
    Evidence: .sisyphus/evidence/task-10-auth-middleware.txt
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 11. **邮箱服务（SMTP TLS）**

  **What to do**:
  - 实现 `internal/service/email.go`：
    - `SendAdminNotification(newUserEmail, nickname string) error` — 新用户注册通知管理员
    - `SendApprovalNotification(userEmail string, approved bool) error` — 审核结果通知用户
    - 使用 `gomail` 库，SMTP TLS 587 端口
    - 配置从 viper 读取：Host, Port, Username, Password, From, AdminEmail
  - 实现 `internal/service/email_mock.go` — 开发/测试用 mock（记录到 log）
  - TDD：测试邮件构造逻辑（用 mock）

  **Must NOT do**:
  - 不要在代码中硬编码 SMTP 密码
  - 邮箱模板内容不要过度设计 — 简单文本即可

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: gomail 标准用法

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with T8-T10)
  - **Depends On**: T4, T7
  - **Blocks**: T12

  **Acceptance Criteria**:
  - [ ] `go test ./internal/service -run TestEmail` PASS（mock mode）
  - [ ] 真实 SMTP 连接测试（config 中 UseTLS=true）

  **QA Scenarios**:
  ```
  Scenario: Email service with mock
    Tool: Bash
    Steps:
      1. cd backend && go test ./internal/service -v -run TestEmailMock
    Expected Result: Mock 邮件"发送"成功，无 panic
    Evidence: .sisyphus/evidence/task-11-email.txt
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 12. **注册接口**

  **What to do**:
  - 实现 `internal/handler/auth.go` — `Register` 函数：
    - POST `/api/auth/register` — 接收 email, password, nickname
    - 密码 bcrypt 哈希（`golang.org/x/crypto/bcrypt`）
    - 验证邮箱唯一性
    - 检查是否为管理员邮箱（config.Admin.Email）→ role=admin
    - 创建用户 status=pending
    - 异步发送邮件通知管理员（goroutine）
    - 返回成功消息（不返回 token，需审核通过才可登录）
  - TDD：测试正常注册、重复邮箱、无效邮箱格式

  **Must NOT do**:
  - 不要在注册时直接激活用户
  - 不要注册后直接返回 token

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 标准注册 handler，输入验证+数据库写入+邮件通知

  **Parallelization**:
  - **Can Run In Parallel**: NO（依赖 T8-T11）
  - **Parallel Group**: Wave 2, Sequential after T8-T11
  - **Depends On**: T8, T9, T10, T11
  - **Blocks**: T35, T42

  **Acceptance Criteria**:
  - [ ] `go test ./internal/handler -run TestRegister` PASS
  - [ ] 重复邮箱返回 409

  **QA Scenarios**:
  ```
  Scenario: User registers successfully
    Tool: Bash (curl)
    Steps:
      1. curl -X POST http://localhost:1323/api/auth/register \
         -H "Content-Type: application/json" \
         -d '{"email":"test@example.com","password":"Test123!","nickname":"测试用户"}'
      2. 断言: HTTP 201, body.message 含 "审核"
    Expected Result: 返回 201，提示等待审核
    Failure: 200/500/409
    Evidence: .sisyphus/evidence/task-12-register.txt

  Scenario: Duplicate email rejected
    Tool: Bash (curl)
    Steps:
      1. 再次注册相同邮箱
      2. 断言: HTTP 409
    Expected Result: 409 冲突
    Evidence: .sisyphus/evidence/task-12-register-dup.txt
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 13. **登录 + Token 刷新接口**

  **What to do**:
  - `POST /api/auth/login` — 验证 email+password → 检查 status==approved → 返回 TokenPair
  - `POST /api/auth/refresh` — 验证 RefreshToken → 生成新 TokenPair（Refresh Token Rotation）
  - `POST /api/auth/logout` — 可选：记录已撤销的 RefreshToken（简易方案：直接让前端删除 token）
  - TDD：正常登录、错误密码、未审核用户登录拒绝、token 刷新、过期 refresh token

  **Must NOT do**:
  - 不要允许未审核(pending)用户登录
  - 不要允许被拒绝(rejected)用户登录

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 标准登录/刷新 handler

  **Parallelization**:
  - **Can Run In Parallel**: NO（依赖 T8-T9）
  - **Parallel Group**: Wave 2, Sequential after T8-T10
  - **Depends On**: T8, T9, T10
  - **Blocks**: T34, T35, T42

  **Acceptance Criteria**:
  - [ ] `go test ./internal/handler -run TestLogin` PASS
  - [ ] `go test ./internal/handler -run TestRefreshToken` PASS
  - [ ] 未审核用户返回 403

  **QA Scenarios**:
  ```
  Scenario: Approved user logs in
    Tool: Bash (curl)
    Steps:
      1. 先注册并手动将用户 status 改为 approved (sqlite3)
      2. curl -X POST /api/auth/login -d '{"email":"ok@test.com","password":"Test123!"}'
      3. 断言: HTTP 200, body.data.access_token 存在, body.data.refresh_token 存在
    Expected Result: 返回双 token
    Evidence: .sisyphus/evidence/task-13-login.txt

  Scenario: Pending user login rejected
    Tool: Bash (curl)
    Steps:
      1. curl -X POST /api/auth/login -d '{"email":"pending@test.com","password":"Test123!"}'
      2. 断言: HTTP 403
    Expected Result: 403 "账号待审核"
    Evidence: .sisyphus/evidence/task-13-login-pending.txt

  Scenario: Token refresh
    Tool: Bash (curl)
    Steps:
      1. 用登录返回的 refresh_token 调用 POST /api/auth/refresh
      2. 断言: HTTP 200, 返回新的 access_token + refresh_token
    Expected Result: Token 刷新成功
    Evidence: .sisyphus/evidence/task-13-refresh.txt
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 14. **管理员用户审核接口**

  **What to do**:
  - `GET /api/admin/users/pending` — AdminRequired 中间件，返回待审核用户列表
  - `PUT /api/admin/users/:id/approve` — 批准用户（status→approved）
  - `PUT /api/admin/users/:id/reject` — 拒绝用户（status→rejected）
  - 审核后发送邮件通知用户（调用 T11 邮件服务）
  - TDD：管理员可审核、非管理员被拒

  **Must NOT do**:
  - 不要允许普通用户调用这些接口
  - 不要重复审核已处理用户

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 标准 CRUD handler + 权限检查

  **Parallelization**:
  - **Can Run In Parallel**: NO（依赖 T8, T10, T11）
  - **Parallel Group**: Wave 2, Sequential
  - **Depends On**: T8, T10, T11
  - **Blocks**: T39, T42

  **Acceptance Criteria**:
  - [ ] `go test ./internal/handler -run TestAdminReview` PASS
  - [ ] 普通用户访问返回 403

  **QA Scenarios**:
  ```
  Scenario: Admin approves user
    Tool: Bash (curl)
    Steps:
      1. curl -H "Authorization: Bearer <admin_token>" \
         -X PUT /api/admin/users/3/approve
      2. 断言: HTTP 200, user.status = "approved"
    Expected Result: 用户状态更新为 approved
    Evidence: .sisyphus/evidence/task-14-admin-approve.txt

  Scenario: Non-admin rejected
    Tool: Bash (curl)
    Steps:
      1. curl -H "Authorization: Bearer <user_token>" \
         -X PUT /api/admin/users/3/approve
      2. 断言: HTTP 403
    Expected Result: 403 Forbidden
    Evidence: .sisyphus/evidence/task-14-admin-forbidden.txt
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 15. **Album Repository + CRUD 接口**

  **What to do**:
  - 实现 `internal/repository/album.go`：Create, FindAll, FindByID, Update, Delete（CASCADE删除关联图片？先做软保护：有图片的相册不可删除）
  - 实现 `internal/handler/album.go`：
    - `GET /api/albums` — 公开，所有相册列表
    - `POST /api/albums` — AuthRequired，创建相册
    - `PUT /api/albums/:id` — 相册创建者或管理员
    - `DELETE /api/albums/:id` — 相册创建者或管理员
  - TDD

  **Must NOT do**:
  - 删除相册时不要级联删除图片（先设保护）
  - 不要允许非创建者编辑相册

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 标准 CRUD + 权限检查

  **Parallelization**:
  - **Can Run In Parallel**: YES（T15-T16 可并行）
  - **Parallel Group**: Wave 3 (with T16)
  - **Depends On**: T5, T6, T7
  - **Blocks**: T17, T18, T29, T40

  **Acceptance Criteria**:
  - [ ] `go test ./internal/repository -run TestAlbumRepo` PASS
  - [ ] `go test ./internal/handler -run TestAlbumHandler` PASS

  **QA Scenarios**:
  ```
  Scenario: Create and list albums
    Tool: Bash (curl)
    Steps:
      1. POST /api/albums -d '{"name":"顾夏帅照"}' → 201
      2. GET /api/albums → 200, 列表包含刚创建的相册
    Expected Result: 相册 CRUD 正常
    Evidence: .sisyphus/evidence/task-15-album.txt
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 16. **兰空图床客户端**

  **What to do**:
  - 实现 `internal/service/lsky.go`：
    - `NewLskyClient(baseURL, email, password string) (*LskyClient, error)` — 自动获取 token
    - `UploadImage(file multipart.File, filename string) (url string, error)` — 构造 multipart 请求上传到兰空
    - `DeleteImage(url string) error` — 根据 URL 删除兰空图片（可选）
    - Token 过期自动重新获取
    - HTTP client 超时设置（上传 30s）
  - 参考兰空 API：POST `/api/v1/upload` multipart，Authorization: Bearer token
  - TDD：用 httptest mock 兰空 API

  **Must NOT do**:
  - 不要在前端暴露兰空 API 地址/token
  - 不要硬编码兰空凭证

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: 需要理解兰空 API 详细规范、multipart 代理上传、token 管理

  **Skills**: []
  - 纯 Go 实现，无需额外 skill

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with T15)
  - **Depends On**: T4, T7
  - **Blocks**: T17

  **Acceptance Criteria**:
  - [ ] `go test ./internal/service -run TestLskyClient` PASS（mock server）
  - [ ] 上传返回的 URL 格式正确（以 http 开头）

  **QA Scenarios**:
  ```
  Scenario: Lsky client uploads via mock
    Tool: Bash
    Steps:
      1. cd backend && go test ./internal/service -v -run TestLskyUpload
      2. 检查返回 URL 非空且以 http 开头
    Expected Result: Mock 上传成功，返回 URL
    Evidence: .sisyphus/evidence/task-16-lsky-mock.txt
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 17. **图片上传接口**

  **What to do**:
  - 实现 `internal/handler/image.go` — `Upload` 函数：
    - `POST /api/images/upload` — AuthRequired，multipart/form-data
    - 接收：file（图片文件）、title、album_id(可选)、tags(逗号分隔)
    - 验证文件大小 ≤ config.Upload.MaxSizeMB
    - 验证 MIME 类型（image/jpeg, image/png, image/gif, image/webp）
    - 调用 T16 兰空客户端上传 → 获取 URL
    - 存入 images 表（lsky_url, uploaded_by=当前用户ID）
    - 返回创建的 image 记录
  - TDD：测试上传成功、文件过大、非图片类型、未登录

  **Must NOT do**:
  - 不要在后端本地保存图片文件（直接 pipe 到兰空）
  - 不要限制只允许特定图片格式（开放常见格式即可）

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: multipart 文件处理 + 兰空 pipe + 事务写入 DB

  **Parallelization**:
  - **Can Run In Parallel**: NO（依赖 T15, T16）
  - **Parallel Group**: Wave 3, Sequential after T15-T16
  - **Depends On**: T4, T15, T16
  - **Blocks**: T18, T27, T38

  **Acceptance Criteria**:
  - [ ] `go test ./internal/handler -run TestImageUpload` PASS
  - [ ] 10MB 以内图片上传成功
  - [ ] 超大文件返回 413

  **QA Scenarios**:
  ```
  Scenario: Upload image successfully
    Tool: Bash (curl)
    Preconditions: 用户已登录，有 token
    Steps:
      1. 创建测试图片: 1KB JPEG
      2. curl -X POST /api/images/upload \
         -H "Authorization: Bearer <token>" \
         -F "file=@test.jpg" -F "title=测试图片" -F "album_id=1" -F "tags=顾夏,帅"
      3. 断言: HTTP 201, body.data.lsky_url 非空
    Expected Result: 图片上传成功，返回兰空URL
    Evidence: .sisyphus/evidence/task-17-upload.txt

  Scenario: Oversized file rejected
    Tool: Bash (curl)
    Steps:
      1. curl -F "file=@large_15mb.jpg" → 断言 HTTP 413
    Expected Result: 413 Payload Too Large
    Evidence: .sisyphus/evidence/task-17-upload-oversize.txt
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 18. **图片 CRUD 接口（权限控制）**

  **What to do**:
  - `GET /api/images` — 公开，分页（page, limit），支持 ?album_id= & ?tags= & ?search= 过滤
  - `GET /api/images/:id` — 公开，单张图片详情（含评论数）
  - `PUT /api/images/:id` — AuthRequired，仅上传者或管理员
  - `DELETE /api/images/:id` — AuthRequired，仅上传者或管理员
  - 实现 `internal/repository/image.go`：带过滤的分页查询
  - TDD：CRUD 各场景 + 权限边界

  **Must NOT do**:
  - 不要让普通用户编辑/删除他人图片
  - 删除图片时考虑是否也删兰空图床上的原图（先不删兰空）

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: 复杂查询 + 权限矩阵 + 分页

  **Parallelization**:
  - **Can Run In Parallel**: NO（依赖 T10, T15, T17）
  - **Parallel Group**: Wave 3, Sequential
  - **Depends On**: T10, T15, T17
  - **Blocks**: T27, T38, T42

  **Acceptance Criteria**:
  - [ ] `go test ./internal/handler -run TestImageCRUD` PASS
  - [ ] 普通用户编辑他人图片返回 403
  - [ ] 管理员可编辑任意图片
  - [ ] 分页参数 page=2&limit=10 正确

  **QA Scenarios**:
  ```
  Scenario: User edits own image
    Tool: Bash (curl)
    Steps:
      1. PUT /api/images/5 -H "Authorization: Bearer <owner_token>" \
         -d '{"title":"新标题","tags":["新标签"]}' → 200
    Expected Result: 图片信息更新成功
    Evidence: .sisyphus/evidence/task-18-edit-own.txt

  Scenario: User cannot edit others' image
    Tool: Bash (curl)
    Steps:
      1. PUT /api/images/5 -H "Authorization: Bearer <other_token>" → 403
    Expected Result: 403 Forbidden
    Evidence: .sisyphus/evidence/task-18-edit-other.txt

  Scenario: Public image listing
    Tool: Bash (curl)
    Steps:
      1. GET /api/images?page=1&limit=8 (无需 token) → 200, 返回图片数组
    Expected Result: 访客可浏览图片列表
    Evidence: .sisyphus/evidence/task-18-public-list.txt
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 19. **Comment Repository + CRUD 接口**

  **What to do**:
  - 实现 `internal/repository/comment.go`：Create, FindByImageID, Delete
  - 实现 `internal/handler/comment.go`：
    - `GET /api/images/:id/comments` — 公开，分页
    - `POST /api/images/:id/comments` — AuthRequired
    - `DELETE /api/comments/:id` — 评论作者或图片作者或管理员
  - TDD

  **Must NOT do**:
  - 不要在返回评论时暴露用户密码
  - 评论内容做基本 XSS 防护（后端 strip HTML tags）

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 简单 CRUD + 权限检查

  **Parallelization**:
  - **Can Run In Parallel**: YES（T19-T20 可并行）
  - **Parallel Group**: Wave 3 (with T20)
  - **Depends On**: T5, T6, T7
  - **Blocks**: T31, T38

  **Acceptance Criteria**:
  - [ ] `go test ./internal/handler -run TestComment` PASS
  - [ ] HTML 标签被 strip

  **QA Scenarios**:
  ```
  Scenario: Post and read comments
    Tool: Bash (curl)
    Steps:
      1. POST /api/images/1/comments -H "Auth: <token>" \
         -d '{"content":"顾夏最帅！"}' → 201
      2. GET /api/images/1/comments → 200, 包含刚发的评论
    Expected Result: 评论发布和读取正常
    Evidence: .sisyphus/evidence/task-19-comment.txt

  Scenario: XSS filtered
    Tool: Bash (curl)
    Steps:
      1. POST comment with '<script>alert(1)</script>' → 201
      2. GET 评论 → content 不含 <script>
    Expected Result: HTML 被过滤
    Evidence: .sisyphus/evidence/task-19-xss.txt
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 20. **图片搜索接口**

  **What to do**:
  - `GET /api/images/search?q=keyword` — 搜索 title, description, tags（LIKE 查询）
  - 扩展 T18 的 repository 添加 `SearchImages` 方法
  - 返回分页结果
  - TDD：中文搜索、空结果、特殊字符

  **Must NOT do**:
  - 不要实现全文搜索引擎（SQLite LIKE 即可）
  - 搜索不要区分大小写

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: SQL LIKE 查询 + handler

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with T19)
  - **Depends On**: T7, T15, T18
  - **Blocks**: T30, T37

  **Acceptance Criteria**:
  - [ ] `go test ./internal/handler -run TestSearch` PASS
  - [ ] 搜索 "顾夏" 返回匹配结果
  - [ ] 搜索 "不存在的" 返回空数组

  **QA Scenarios**:
  ```
  Scenario: Search images by keyword
    Tool: Bash (curl)
    Steps:
      1. GET /api/images/search?q=顾夏 → 200, 返回匹配图片
      2. GET /api/images/search?q=xyznotfound → 200, 空数组
    Expected Result: 搜索功能正常
    Evidence: .sisyphus/evidence/task-20-search.txt
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 21. **WebSocket Hub + 客户端管理**

  **What to do**:
  - 实现 `internal/websocket/hub.go`：
    - Hub struct：clients map, register/unregister/broadcast channels
    - `Run()` goroutine 主循环
  - 实现 `internal/websocket/client.go`：
    - Client struct：conn, hub, send channel
    - `ReadPump()` / `WritePump()` goroutines
    - 心跳检测（30s ping/pong）
  - `GET /ws` — WebSocket 升级端点，OptionalAuth（有token则关联用户）
  - 使用 `gorilla/websocket`
  - TDD：测试连接/断开/broadcast

  **Must NOT do**:
  - 不要在 WebSocket 连接中传输敏感信息
  - 不要忘记设置读写超时

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: gorilla/websocket 完整实现，含并发安全和 goroutine 管理

  **Parallelization**:
  - **Can Run In Parallel**: NO（依赖 T2, T7）
  - **Parallel Group**: Wave 4
  - **Depends On**: T2, T7
  - **Blocks**: T22, T32

  **Acceptance Criteria**:
  - [ ] `go test ./internal/websocket -run TestHub` PASS
  - [ ] 多客户端连接/断开不 panic
  - [ ] Broadcast 所有客户端收到消息

  **QA Scenarios**:
  ```
  Scenario: WebSocket connect and receive broadcast
    Tool: Bash (websocat 或 wscat)
    Steps:
      1. 启动后端
      2. websocat ws://localhost:1323/ws → 连接成功
      3. 通过 HTTP API 触发 broadcast → WebSocket 客户端收到消息
    Expected Result: 客户端收到 JSON 消息
    Evidence: .sisyphus/evidence/task-21-websocket.txt
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 22. **在线统计追踪 + WebSocket 广播**

  **What to do**:
  - 实现 `internal/service/stats.go`：
    - `IncrementVisitor()` — 每次页面访问调用，写入 visitors 表或内存计数
    - `GetOnlineCount()` — 从 Hub 获取连接数
    - `GetVisitorCount()` — 从 DB 获取累计访问
    - `GetNewMembersCount()` — 本周新增 approved 用户数
    - `GetSafeDays()` — 从 config.Site.StartDate 计算到今天的天数
    - `BroadcastStats()` — 定时(10s)向所有 WebSocket 客户端推送统计 JSON
  - 在 main.go 中启动定时广播 goroutine
  - WebSocket 消息格式：`{"type":"stats","data":{"online":N,"visitors":N,"newMembers":N,"safeDays":N}}`
  - TDD

  **Must NOT do**:
  - 不要在每次 HTTP 请求中广播（用定时器）
  - 不要把 visitor count 存内存（存 DB，SQLite 持久化）

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 简单的计数逻辑 + 定时广播

  **Parallelization**:
  - **Can Run In Parallel**: NO（依赖 T8, T21）
  - **Parallel Group**: Wave 4
  - **Depends On**: T8, T21
  - **Blocks**: T32

  **Acceptance Criteria**:
  - [ ] `go test ./internal/service -run TestStats` PASS
  - [ ] WebSocket 客户端每 10s 收到 stats 消息

  **QA Scenarios**:
  ```
  Scenario: Stats broadcast received
    Tool: Bash (websocat)
    Steps:
      1. 连接 WebSocket → 等待 10s
      2. 检查收到的 JSON: type=="stats", data.online 为数字
    Expected Result: 收到实时统计数据
    Evidence: .sisyphus/evidence/task-22-stats.txt
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 23. **全局 CSS（Demo 风格迁移）**

  **What to do**:
  - 创建 `frontend/src/assets/styles/main.css`：
    - CSS Reset
    - 赛博朋克色彩变量（`:root` — 霓虹色系：neon-red, neon-green, neon-yellow, neon-pink, neon-cyan）
    - 暗色背景 `#0a0a0a`
    - 字体：`"Microsoft YaHei", "SimHei", "PingFang SC", sans-serif`
    - 动画 keyframes：`rainbow-bg`, `glitch`（降低频率，不用 `infinite` 改 `5s` 间隔）, `marquee`, `blink`（降低频率）, `pulse-border`
    - 通用样式类：`.rainbow-border`, `.glow-text`, `.neon-box`
    - 响应式断点：768px
  - 从 `gm_site_demo.html` 提取并优化 CSS

  **Must NOT do**:
  - 不要保留 `infinite` 闪烁动画在关键交互元素上（影响可访问性）
  - 不要直接复制所有 CSS — 只保留核心风格，简化过度动画

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: CSS 设计系统构建，需要视觉审美和 Demo 风格还原

  **Skills**: [`frontend-ui-ux`]
    - `frontend-ui-ux`: 赛博朋克风格视觉设计

  **Parallelization**:
  - **Can Run In Parallel**: YES（T23-T26 可并行于 T21-T22）
  - **Parallel Group**: Wave 4 (with T24-T26)
  - **Depends On**: T3
  - **Blocks**: T24-T33

  **Acceptance Criteria**:
  - [ ] CSS 变量定义完整
  - [ ] 页面背景 `#0a0a0a` 渲染正确

  **QA Scenarios**:
  ```
  Scenario: Global styles apply correctly
    Tool: Playwright
    Steps:
      1. 打开 http://localhost:5173
      2. 检查 body background-color: rgb(10, 10, 10)
      3. 检查 CSS 变量 --neon-green 存在
    Expected Result: 暗色背景 + CSS 变量可用
    Evidence: .sisyphus/evidence/task-23-global-css.png
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 24. **粒子背景组件**

  **What to do**:
  - 创建 `frontend/src/components/ParticleBg.vue`：
    - 40个随机 emoji（'🔥','💪','👑','💩','🐂','🍺','⚡','💀','🤡','🎯'）
    - 随机 left 位置（0-100%）
    - 随机动画时长（4-14s）和延迟（0-8s）
    - CSS 动画：从 -100px 降落到 105vh，旋转 720deg
    - `position: fixed` 覆盖全屏，`pointer-events: none`
    - 使用 Vue 的 `v-for` + CSS custom properties 动态设置

  **Must NOT do**:
  - 不要在 JS 中操作大量 DOM — 用 CSS animation
  - 不要阻塞页面渲染

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: CSS 动画 + Vue 动态渲染

  **Skills**: [`frontend-ui-ux`]
    - `frontend-ui-ux`: 粒子动画视觉效果

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with T23, T25-T26)
  - **Depends On**: T23
  - **Blocks**: T26, T37

  **Acceptance Criteria**:
  - [ ] 组件渲染 40 个 emoji span
  - [ ] 粒子动画流畅无卡顿
  - [ ] 不响应鼠标事件（pointer-events: none）

  **QA Scenarios**:
  ```
  Scenario: Particles render and animate
    Tool: Playwright
    Steps:
      1. 打开首页
      2. 检查 #particles 容器存在
      3. 检查至少有 30 个 .star 元素
      4. 等待 2s → 粒子位置改变（动画生效）
    Expected Result: 40个emoji粒子漂浮动画
    Evidence: .sisyphus/evidence/task-24-particles.png
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 25. **彩虹横幅 + 跑马灯组件**

  **What to do**:
  - 创建 `frontend/src/components/RainbowBanner.vue`：
    - 彩虹渐变动画背景（300%宽度，2s 循环）
    - glitch 文字效果（减少频率：5s 间隔，非 infinite）
    - 标题 + 副标题（props 可配置）
    - 4px ridge 黄边框 + 霓虹阴影
  - 创建 `frontend/src/components/Marquee.vue`：
    - props: `text` (string), `color` ('green'|'pink'|'yellow')
    - CSS 动画：12s linear infinite 从右到左滚动
    - `white-space: nowrap`, `overflow: hidden`

  **Must NOT do**:
  - 不要 infinite glitch（改为 5s 间隔循环）
  - 跑马灯不要使用 `<marquee>` 标签（已废弃）

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: Demo 风格动画组件

  **Skills**: [`frontend-ui-ux`]
    - `frontend-ui-ux`: 霓虹彩虹视觉效果

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with T23-T24, T26)
  - **Depends On**: T23
  - **Blocks**: T26, T37

  **Acceptance Criteria**:
  - [ ] RainbowBanner 渲染彩虹背景 + glitch 标题
  - [ ] Marquee 文字持续滚动
  - [ ] 颜色变体正确（绿/粉/黄）

  **QA Scenarios**:
  ```
  Scenario: Banner renders with rainbow border
    Tool: Playwright
    Steps:
      1. 打开首页
      2. 检查 .top-banner 元素可见
      3. 检查 border 颜色包含 #ff0 (黄色) ridge 样式
    Expected Result: 彩虹横幅正常
    Evidence: .sisyphus/evidence/task-25-banner.png

  Scenario: Marquee scrolls continuously
    Tool: Playwright
    Steps:
      1. 找到 .marquee-content 元素
      2. 记录初始 transform/position
      3. 等待 2s → transform 已改变
    Expected Result: 跑马灯持续滚动
    Evidence: .sisyphus/evidence/task-25-marquee.png
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 26. **页面布局框架（三栏）**

  **What to do**:
  - 创建 `frontend/src/components/LayoutShell.vue`：
    - 三栏布局：左侧 sidebar（导航+统计）、中间 main-area（图片区）、右侧 sidebar（公告+实时计数）
    - CSS Grid 或 Flexbox：`grid-template-columns: 220px 1fr 260px`
    - 响应式：≤768px 变为单栏堆叠
  - 创建 `frontend/src/App.vue`：使用 LayoutShell + router-view
  - 顶部通知栏：统计条（在线人数、新增成员、安全运行天数）
    - 创建 `frontend/src/components/TopNotifyBar.vue`
  - 页脚：访客计数器

  **Must NOT do**:
  - 不要在布局中硬编码内容
  - 不要忘记 responsive 断点

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: 页面布局架构设计

  **Skills**: [`frontend-ui-ux`]
    - `frontend-ui-ux`: 三栏布局 + 响应式设计

  **Parallelization**:
  - **Can Run In Parallel**: YES（依赖 T24-T25 完成）
  - **Parallel Group**: Wave 4
  - **Depends On**: T24, T25
  - **Blocks**: T27-T33, T37

  **Acceptance Criteria**:
  - [ ] 页面呈三栏布局
  - [ ] ≤768px 时变为单栏
  - [ ] TopNotifyBar 显示统计占位

  **QA Scenarios**:
  ```
  Scenario: Three-column layout renders
    Tool: Playwright
    Steps:
      1. 打开首页
      2. 检查页面宽度 > 1100px → 三栏并排
      3. 设置 viewport 375px → 单栏堆叠
    Expected Result: 响应式三栏布局
    Evidence: .sisyphus/evidence/task-26-layout.png
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 27. **图片卡片 + 画廊网格**

  **What to do**:
  - 创建 `frontend/src/components/GalleryCard.vue`：
    - props: `image: Image`（id, title, lsky_url, tags, uploaded_by, created_at）
    - 卡片样式：渐变背景边框、hover 缩放+发光效果
    - 显示：缩略图（lsky_url）、标题、标签 badge、上传时间
    - emit: `@click` → 触发弹窗
  - 创建 `frontend/src/components/GalleryGrid.vue`：
    - props: `images: Image[]`
    - CSS Grid：`grid-template-columns: repeat(auto-fill, minmax(220px, 1fr))`
    - gap, padding
    - 空状态："暂无图片，快来上传吧！"
  - TDD：vitest 测试组件渲染

  **Must NOT do**:
  - 不要在卡片组件内处理路由跳转
  - 图片加载失败时要有 fallback 占位

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: 图片卡片 UI + Grid 布局

  **Skills**: [`frontend-ui-ux`, `playwright`]
    - `frontend-ui-ux`: 卡片视觉设计
    - `playwright`: 浏览器端QA验证

  **Parallelization**:
  - **Can Run In Parallel**: YES（T27-T33 全可并行）
  - **Parallel Group**: Wave 5 (with T28-T33)
  - **Depends On**: T17, T18, T26
  - **Blocks**: T28, T37

  **Acceptance Criteria**:
  - [ ] 组件渲染图片卡片网格
  - [ ] hover 效果可见
  - [ ] 空数组时显示空状态提示

  **QA Scenarios**:
  ```
  Scenario: Gallery grid displays images
    Tool: Playwright
    Steps:
      1. 确保后端有至少 2 张图片
      2. 打开首页
      3. 检查 .gallery-card 元素数量 ≥ 2
      4. 检查每个卡片有 img[src] 和标题文字
    Expected Result: 图片卡片正确渲染
    Evidence: .sisyphus/evidence/task-27-gallery.png

  Scenario: Empty state
    Tool: Playwright
    Steps:
      1. 清空数据库图片
      2. 打开首页 → 显示 "暂无图片"
    Expected Result: 空状态提示
    Evidence: .sisyphus/evidence/task-27-empty.png
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 28. **图片浮动弹窗（可拖拽+自动复活）**

  **What to do**:
  - 创建 `frontend/src/components/ImagePopup.vue`：
    - props: `image: Image | null`, `visible: boolean`
    - 样式：Demo 风格浮动弹窗（`position: fixed`, `border: 3px ridge #ff0`, 霓虹阴影）
    - 大图显示：`<img :src="image.lsky_url">` 最大宽度 80vw
    - 可拖拽：mousedown on header → mousemove 更新 left/top
    - 关闭按钮：scale(0) 动画 → setTimeout → scale(1) 自动复活（5-13s 随机延迟）
    - 图片信息：标题、上传者、标签、时间
    - emit: `@close`
  - 使用 `useDraggable` composable 封装拖拽逻辑

  **Must NOT do**:
  - 不要用 v-if 销毁组件（关闭只是隐藏，用 v-show）
  - 拖拽不要超出视口边界

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: 复杂交互组件：拖拽 + 动画 + 自动复活

  **Skills**: [`frontend-ui-ux`, `playwright`]
    - `frontend-ui-ux`: 浮动弹窗视觉效果
    - `playwright`: 拖拽交互QA

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 5 (with T27, T29-T33)
  - **Depends On**: T26, T27
  - **Blocks**: T37

  **Acceptance Criteria**:
  - [ ] 弹窗打开显示大图
  - [ ] 可拖拽移动
  - [ ] 关闭后 5-13s 自动重新出现
  - [ ] 不超出视口边界

  **QA Scenarios**:
  ```
  Scenario: Popup opens and displays image
    Tool: Playwright
    Steps:
      1. 点击任意 GalleryCard
      2. 检查 .float-popup 可见 → 含 <img> 标签
      3. 检查弹窗有 rainbow 边框样式
    Expected Result: 浮动弹窗显示大图
    Evidence: .sisyphus/evidence/task-28-popup-open.png

  Scenario: Popup is draggable
    Tool: Playwright
    Steps:
      1. 弹窗打开后，mousedown on .popup-header
      2. mousemove 200px right, 100px down
      3. mouseup → 检查弹窗 left/top 已改变
    Expected Result: 弹窗位置改变
    Evidence: .sisyphus/evidence/task-28-drag.png

  Scenario: Popup auto-reappears after close
    Tool: Playwright
    Steps:
      1. 点击 .popup-close → 弹窗消失
      2. 等待 6-14s → 弹窗重新出现
    Expected Result: 弹窗自动复活
    Evidence: .sisyphus/evidence/task-28-reappear.png
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 29. **相册/标签筛选组件**

  **What to do**:
  - 创建 `frontend/src/components/AlbumFilter.vue`：
    - 水平标签列表：显示所有相册（从 API `GET /api/albums` 获取）
    - 选中态高亮（霓虹绿边框）
    - emit: `@select(albumId | null)` — null 表示全部
    - "全部" 默认选项
  - 父组件监听 `@select` → 重新请求 `/api/images?album_id=N`
  - vitest 测试

  **Must NOT do**:
  - 不要在组件内 hardcode 相册列表

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: 标签选择 UI 组件

  **Skills**: [`frontend-ui-ux`]
    - `frontend-ui-ux`: 霓虹风格标签设计

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 5 (with T27-T28, T30-T33)
  - **Depends On**: T15, T26
  - **Blocks**: T37

  **Acceptance Criteria**:
  - [ ] 相册列表正确渲染
  - [ ] 点击切换相册 → emit 事件
  - [ ] 选中态样式不同

  **QA Scenarios**:
  ```
  Scenario: Filter by album
    Tool: Playwright
    Steps:
      1. 打开首页 → 相册标签可见（至少"全部"）
      2. 点击某个相册标签
      3. 图片列表更新为对应相册的图片
    Expected Result: 相册筛选功能正常
    Evidence: .sisyphus/evidence/task-29-filter.png
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 30. **搜索栏组件**

  **What to do**:
  - 创建 `frontend/src/components/SearchBar.vue`：
    - 输入框 + 搜索按钮（霓虹绿边框样式）
    - 防抖 300ms → emit `@search(keyword)`
    - 支持回车搜索
  - 父组件监听 → 调用 `/api/images/search?q=keyword`
  - vitest 测试 debounce

  **Must NOT do**:
  - 不要在每次输入字符时发请求（debounce）

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: 搜索输入 UI + debounce

  **Skills**: [`frontend-ui-ux`]
    - `frontend-ui-ux`: 搜索栏视觉设计

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 5 (with T27-T29, T31-T33)
  - **Depends On**: T20, T26
  - **Blocks**: T37

  **Acceptance Criteria**:
  - [ ] 输入文字 300ms 后 emit search 事件
  - [ ] 回车键也触发搜索

  **QA Scenarios**:
  ```
  Scenario: Search triggers after debounce
    Tool: Playwright
    Steps:
      1. 在搜索框输入 "顾夏"
      2. 等待 400ms → 图片列表更新为搜索结果
    Expected Result: 搜索功能正常
    Evidence: .sisyphus/evidence/task-30-search.png
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 31. **评论区组件**

  **What to do**:
  - 创建 `frontend/src/components/CommentSection.vue`：
    - props: `imageId: number`
    - 评论列表（GET `/api/images/:id/comments`）
    - 评论输入框（登录用户可见）+ 提交按钮
    - 未登录显示 "登录后评论"
    - 删除按钮（评论作者/管理员可见）
    - 分页"加载更多"
    - vitest 测试

  **Must NOT do**:
  - 不要在未登录状态显示输入框
  - 评论内容做 XSS 过滤（v-text 而非 v-html）

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: 评论 UI + 权限控制显示

  **Skills**: [`frontend-ui-ux`]
    - `frontend-ui-ux`: 评论区视觉设计

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 5 (with T27-T30, T32-T33)
  - **Depends On**: T19, T26
  - **Blocks**: T37

  **Acceptance Criteria**:
  - [ ] 评论列表正确渲染
  - [ ] 登录用户显示输入框
  - [ ] 未登录显示提示文字
  - [ ] 评论作者可删除

  **QA Scenarios**:
  ```
  Scenario: Guest sees login prompt
    Tool: Playwright
    Steps:
      1. 未登录状态打开图片评论区
      2. 检查文本 "登录后评论" 可见
      3. 输入框不存在
    Expected Result: 访客无法评论
    Evidence: .sisyphus/evidence/task-31-guest.png

  Scenario: Logged-in user posts comment
    Tool: Playwright
    Steps:
      1. 登录后 → 打开图片弹窗
      2. 输入评论 "好图！" → 点击提交
      3. 评论出现在列表中
    Expected Result: 评论发布成功
    Evidence: .sisyphus/evidence/task-31-comment.png
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 32. **统计栏组件（WebSocket 连接）**

  **What to do**:
  - 创建 `frontend/src/components/StatsBar.vue`：
    - 显示在线人数、累计访客、新注册数、安全运行天数
    - 使用 `useWebSocket` composable 连接 WebSocket
    - 接收 stats 消息更新显示
    - 数字跳动动画（可选）
  - 创建 `frontend/src/composables/useWebSocket.ts`：
    - 连接管理（自动重连）
    - 消息解析
    - Vue ref 绑定

  **Must NOT do**:
  - 不要在连接断开时页面崩溃
  - 不要忘记 cleanup（onUnmounted 断开连接）

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: WebSocket 前端集成 + 统计显示

  **Skills**: [`frontend-ui-ux`]
    - `frontend-ui-ux`: 统计数字视觉展示

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 5 (with T27-T31, T33)
  - **Depends On**: T21, T22, T26
  - **Blocks**: T37

  **Acceptance Criteria**:
  - [ ] 数字随 WebSocket 消息更新
  - [ ] 断开连接后自动重连
  - [ ] 初始显示 "--" 占位

  **QA Scenarios**:
  ```
  Scenario: Stats update via WebSocket
    Tool: Playwright
    Steps:
      1. 打开首页 → 统计栏显示初始值或 "--"
      2. 等待 10s → 数字更新为非占位值
      3. 打开第二个标签页 → 在线人数+1
    Expected Result: 实时数据更新
    Evidence: .sisyphus/evidence/task-32-stats.png
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 33. **访客计数器 + 悬浮按钮**

  **What to do**:
  - 创建 `frontend/src/components/VisitorCounter.vue`：
    - 底部显示 "你是顾夏第 N 个舔狗"（N 从 WebSocket stats 获取）
    - 数字分位显示，monospace 字体，霓虹绿样式
  - 创建 `frontend/src/components/FloatingButtons.vue`：
    - 右下角固定悬浮按钮组
    - "TOP" — scrollToTop
    - 登录/注册按钮（未登录时）
    - 上传按钮（登录后）

  **Must NOT do**:
  - 不要保留 Demo 的 "VIP" "客服" 等假按钮
  - 不要使用 `alert()` 作为交互

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: 底部计数器 + 悬浮按钮 UI

  **Skills**: [`frontend-ui-ux`]
    - `frontend-ui-ux`: 浮动按钮和计数器设计

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 5 (with T27-T32)
  - **Depends On**: T26
  - **Blocks**: T37

  **Acceptance Criteria**:
  - [ ] 计数器数字分位显示
  - [ ] "TOP" 按钮点击回到顶部
  - [ ] 登录/未登录状态下按钮不同

  **QA Scenarios**:
  ```
  Scenario: Visitor counter displays
    Tool: Playwright
    Steps:
      1. 打开首页 → 底部计数器可见
      2. 检查数字用 <span> 独立显示（分位）
    Expected Result: 计数器正常渲染
    Evidence: .sisyphus/evidence/task-33-counter.png

  Scenario: TOP button scrolls to top
    Tool: Playwright
    Steps:
      1. 滚动到底部
      2. 点击 "TOP" 按钮
      3. 检查 window.scrollY === 0
    Expected Result: 回到顶部
    Evidence: .sisyphus/evidence/task-33-top.png
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 34. **API 客户端 + Token 拦截器**

  **What to do**:
  - 创建 `frontend/src/api/client.ts`：
    - axios 实例，baseURL 从 vite 代理或环境变量
    - 请求拦截器：自动附加 `Authorization: Bearer <access_token>`
    - 响应拦截器：401 → 自动尝试 refresh token → 重试原请求 → refresh 也失败则跳转登录
    - token 存储：localStorage（accessToken, refreshToken）
  - 创建 `frontend/src/api/auth.ts`：register, login, refresh, logout API 函数
  - 创建 `frontend/src/api/images.ts`：upload, list, get, update, delete, search API 函数
  - 创建 `frontend/src/api/stats.ts`：getStats API 函数
  - vitest 测试拦截器逻辑

  **Must NOT do**:
  - 不要在拦截器中无限循环重试
  - 不要将 token 存 sessionStorage（关闭标签页即丢失）

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: axios 封装 + 拦截器，标准模式

  **Parallelization**:
  - **Can Run In Parallel**: NO（依赖 T4, T10, T13）
  - **Parallel Group**: Wave 6
  - **Depends On**: T4, T10, T13
  - **Blocks**: T35-T40

  **Acceptance Criteria**:
  - [ ] 401 自动 refresh token
  - [ ] refresh 失败跳转 /login
  - [ ] 请求自动携带 token

  **QA Scenarios**:
  ```
  Scenario: Token auto-refresh on 401
    Tool: Playwright
    Steps:
      1. 登录后 → 手动修改 localStorage accessToken 为过期值
      2. 发起 API 请求 → 拦截器自动 refresh
      3. 请求成功返回数据
    Expected Result: Token 自动刷新
    Evidence: .sisyphus/evidence/task-34-refresh.txt
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 35. **Auth Store (Pinia) + 路由守卫**

  **What to do**:
  - 创建 `frontend/src/stores/auth.ts`：
    - state: user, isLoggedIn, isAdmin, isPending
    - actions: login, register, logout, fetchUser, refreshToken
    - getters: isAuthenticated, userRole
    - 持久化：token 存 localStorage，页面刷新恢复登录态
  - 创建 `frontend/src/router/index.ts`：
    - 路由：`/` (Home), `/login`, `/register`, `/pending` (待审核), `/admin` (管理后台), `/admin/images`, `/admin/users`
    - 路由守卫 `beforeEach`：需要登录的路由重定向到 /login
    - Admin 路由检查 isAdmin
    - hash 模式（`createWebHashHistory`）

  **Must NOT do**:
  - 不要在路由守卫中做 API 请求（用 store 状态判断）

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Pinia store + Vue Router 配置

  **Parallelization**:
  - **Can Run In Parallel**: NO（依赖 T12, T13, T34）
  - **Parallel Group**: Wave 6
  - **Depends On**: T12, T13, T34
  - **Blocks**: T36, T37

  **Acceptance Criteria**:
  - [ ] 登录后 store 状态更新
  - [ ] 未登录访问 /admin → 重定向 /login
  - [ ] 页面刷新后保持登录态

  **QA Scenarios**:
  ```
  Scenario: Auth guard redirects
    Tool: Playwright
    Steps:
      1. 未登录时导航到 /#/admin
      2. 检查 URL 变为 /#/login
    Expected Result: 重定向到登录页
    Evidence: .sisyphus/evidence/task-35-guard.png

  Scenario: Login persists across refresh
    Tool: Playwright
    Steps:
      1. 登录成功后刷新页面
      2. 检查导航栏显示已登录状态（非登录按钮）
    Expected Result: 登录态保持
    Evidence: .sisyphus/evidence/task-35-persist.png
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 36. **登录/注册/待审核页面**

  **What to do**:
  - 创建 `frontend/src/views/LoginView.vue`：
    - 邮箱 + 密码表单，赛博朋克风格输入框（霓虹绿边框+发光）
    - 提交 → auth store login → 成功后跳转首页
    - 错误提示（密码错误、账号未审核等）
    - 链接到注册页
  - 创建 `frontend/src/views/RegisterView.vue`：
    - 邮箱 + 密码 + 昵称表单
    - 提交 → auth store register → 成功后提示 "请等待管理员审核"
    - 链接到登录页
  - 创建 `frontend/src/views/PendingView.vue`：
    - 审核中状态页面："您的账号正在审核中，请耐心等待管理员审核"
    - 赛博朋克风格等待动画
    - 重新检查状态按钮（轮询）
  - vitest 测试

  **Must NOT do**:
  - 登录失败不要暴露具体是"用户不存在"还是"密码错误"（统一提示）

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: 三个认证相关页面的视觉设计

  **Skills**: [`frontend-ui-ux`, `playwright`]
    - `frontend-ui-ux`: 表单和等待页设计
    - `playwright`: 表单交互QA

  **Parallelization**:
  - **Can Run In Parallel**: YES（三个页面可并行开发）
  - **Parallel Group**: Wave 6
  - **Depends On**: T34, T35
  - **Blocks**: T37

  **Acceptance Criteria**:
  - [ ] 登录成功跳转首页
  - [ ] 注册成功提示等待审核
  - [ ] 表单验证（空字段、无效邮箱）

  **QA Scenarios**:
  ```
  Scenario: Login flow
    Tool: Playwright
    Steps:
      1. 导航到 /#/login
      2. 填写 email: "approved@test.com", password: "Test123!"
      3. 点击 "登录" 按钮
      4. 断言: URL 变为 /#/
      5. 导航栏显示已登录状态
    Expected Result: 登录成功进入首页
    Evidence: .sisyphus/evidence/task-36-login.png

  Scenario: Register flow
    Tool: Playwright
    Steps:
      1. 导航到 /#/register
      2. 填写: "new@test.com", "Test123!", "新用户"
      3. 点击注册 → 显示 "请等待管理员审核"
    Expected Result: 注册成功提示
    Evidence: .sisyphus/evidence/task-36-register.png
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 37. **首页视图（组装所有组件）**

  **What to do**:
  - 创建 `frontend/src/views/HomeView.vue`：
    - 组合所有组件：ParticleBg, RainbowBanner, Marquee(×3), TopNotifyBar(StatsBar), LayoutShell
    - 中间区域：SearchBar + AlbumFilter + GalleryGrid
    - ImagePopup（全局单例，通过 provide/inject 或 store 管理）
    - 根据 GalleryCard click → 打开对应图片弹窗
    - API 调用：`getImages({ page, album_id, search })` 加载数据
    - 搜索/筛选联动刷新图片列表
    - 分页"加载更多"按钮
  - 创建 `frontend/src/stores/images.ts`：图片列表状态管理

  **Must NOT do**:
  - 不要在 HomeView 中实现图片上传（留给 FloatingButtons 触发，后续可加模态框）

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: 首页整体组装，多个组件集成

  **Skills**: [`frontend-ui-ux`, `playwright`]
    - `frontend-ui-ux`: 首页整体视觉
    - `playwright`: 集成QA

  **Parallelization**:
  - **Can Run In Parallel**: NO（依赖所有前端组件）
  - **Parallel Group**: Wave 6, LAST in wave
  - **Depends On**: T24-T36
  - **Blocks**: T41, T42

  **Acceptance Criteria**:
  - [ ] 首页完整渲染：粒子 + 横幅 + 跑马灯 + 统计 + 图片网格 + 弹窗
  - [ ] 搜索/筛选联动
  - [ ] 点击图片打开弹窗
  - [ ] 加载更多分页

  **QA Scenarios**:
  ```
  Scenario: Full home page renders
    Tool: Playwright
    Steps:
      1. 打开 http://localhost:5173
      2. 检查: ParticleBg 可见, RainbowBanner 可见, Marquee 可见 ×3
      3. 检查: 图片画廊渲染 > 0 张卡片
      4. 检查: StatsBar 有数字
    Expected Result: 首页完整渲染
    Evidence: .sisyphus/evidence/task-37-home.png

  Scenario: Click image opens popup
    Tool: Playwright
    Steps:
      1. 点击第一个 GalleryCard
      2. 检查 ImagePopup 可见，含大图
    Expected Result: 弹窗打开显示大图
    Evidence: .sisyphus/evidence/task-37-popup.png
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 38. **管理员图片管理视图**

  **What to do**:
  - 创建 `frontend/src/views/AdminImagesView.vue`：
    - 表格展示所有图片：缩略图、标题、上传者、相册、标签、时间
    - 操作：编辑（弹窗修改标题/标签/相册）、删除（确认对话框）
    - 分页
    - 批量删除（可选）
    - 赛博朋克风格表格（暗色背景、霓虹边框）
  - 路由：`/admin/images`
  - vitest 测试

  **Must NOT do**:
  - 不要展示非管理员可访问的内容
  - 删除操作必须有确认步骤

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: 管理后台图片表格 UI

  **Skills**: [`frontend-ui-ux`, `playwright`]
    - `frontend-ui-ux`: 管理员表格设计
    - `playwright`: 管理操作QA

  **Parallelization**:
  - **Can Run In Parallel**: YES（T38-T40 可并行）
  - **Parallel Group**: Wave 7 (with T39-T40)
  - **Depends On**: T18, T19, T34, T35
  - **Blocks**: T41

  **Acceptance Criteria**:
  - [ ] 管理员可查看所有图片
  - [ ] 可编辑/删除任意图片
  - [ ] 删除有确认对话框

  **QA Scenarios**:
  ```
  Scenario: Admin views all images
    Tool: Playwright
    Steps:
      1. 管理员登录 → 导航到 /#/admin/images
      2. 检查表格包含所有用户的图片
      3. 检查编辑/删除按钮可见
    Expected Result: 管理员图片管理正常
    Evidence: .sisyphus/evidence/task-38-admin-images.png

  Scenario: Admin deletes image
    Tool: Playwright
    Steps:
      1. 点击某图片的删除按钮
      2. 确认对话框出现 → 点击确认
      3. 图片从列表消失
    Expected Result: 删除成功
    Evidence: .sisyphus/evidence/task-38-delete.png
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 39. **管理员用户管理视图**

  **What to do**:
  - 创建 `frontend/src/views/AdminUsersView.vue`：
    - 待审核用户列表（GET `/api/admin/users/pending`）
    - 每个用户卡片：邮箱、昵称、注册时间
    - 操作：批准按钮（绿色）、拒绝按钮（红色）
    - 已审核用户列表（可选）
  - 路由：`/admin/users`
  - vitest 测试

  **Must NOT do**:
  - 不要显示用户密码
  - 批准/拒绝操作不可撤销（或加二次确认）

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: 用户审核管理 UI

  **Skills**: [`frontend-ui-ux`, `playwright`]
    - `frontend-ui-ux`: 审核卡片设计
    - `playwright`: 审核操作QA

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 7 (with T38, T40)
  - **Depends On**: T14, T34, T35
  - **Blocks**: T41

  **Acceptance Criteria**:
  - [ ] 管理员可查看待审核用户
  - [ ] 批准/拒绝操作正常
  - [ ] 操作后列表更新

  **QA Scenarios**:
  ```
  Scenario: Admin approves user
    Tool: Playwright
    Steps:
      1. 管理员登录 → /#/admin/users
      2. 待审核列表有用户
      3. 点击 "批准" → 用户从待审核列表消失
      4. 该用户现在可以登录
    Expected Result: 用户审核流程正常
    Evidence: .sisyphus/evidence/task-39-approve.png
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 40. **管理员相册管理视图**

  **What to do**:
  - 创建 `frontend/src/views/AdminAlbumsView.vue`：
    - 相册列表 + 创建新相册
    - 编辑相册名称/描述
    - 删除相册（检查无图片关联）
  - 路由：`/admin/albums`
  - vitest 测试

  **Must NOT do**:
  - 不要在有图片的相册上直接删除（提示错误）

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 简单 CRUD 管理页面

  **Skills**: [`frontend-ui-ux`]
    - `frontend-ui-ux`: 管理表单设计

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 7 (with T38-T39)
  - **Depends On**: T15, T34, T35
  - **Blocks**: T41

  **Acceptance Criteria**:
  - [ ] 相册 CRUD 正常
  - [ ] 删除有图片的相册提示错误

  **QA Scenarios**:
  ```
  Scenario: Admin manages albums
    Tool: Playwright
    Steps:
      1. 管理员 → /#/admin/albums
      2. 创建新相册 "测试相册"
      3. 编辑相册名称
      4. 删除空相册成功
    Expected Result: 相册管理正常
    Evidence: .sisyphus/evidence/task-40-albums.png
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 41. **前端-后端联调 + CORS 配置**

  **What to do**:
  - 实现 `internal/middleware/cors.go`：CORS 中间件
    - 允许 origin：`http://localhost:5173`
    - 允许 methods：GET, POST, PUT, DELETE, OPTIONS
    - 允许 headers：Content-Type, Authorization
    - 允许 credentials
  - 验证前端所有 API 调用正常工作
  - 验证 WebSocket 连接（CORS 对 WS 的影响）
  - 修复联调中发现的问题

  **Must NOT do**:
  - 不要在生产环境允许 `*` origin

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: 全栈联调，可能涉及多处修复

  **Skills**: [`playwright`]
    - `playwright`: 端到端测试验证

  **Parallelization**:
  - **Can Run In Parallel**: NO（依赖所有前后端）
  - **Parallel Group**: Wave 8
  - **Depends On**: T37-T40
  - **Blocks**: T42

  **Acceptance Criteria**:
  - [ ] 前端可正常调用所有后端 API
  - [ ] WebSocket 连接成功
  - [ ] 无 CORS 错误

  **QA Scenarios**:
  ```
  Scenario: Full API integration
    Tool: Playwright
    Steps:
      1. 前端首页加载 → 图片列表正常
      2. 注册 → 成功
      3. 管理员审核 → 用户可登录
      4. 登录 → 上传图片 → 弹窗显示
      5. WebSocket 统计更新
    Expected Result: 全流程无 CORS/网络错误
    Evidence: .sisyphus/evidence/task-41-integration.png
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 42. **端到端流程测试**

  **What to do**:
  - TDD：编写完整的 E2E 测试场景：
    - 访客浏览：打开首页 → 查看图片 → 点击弹窗 → 关闭弹窗 → 搜索图片
    - 注册流程：注册 → 提示等待审核
    - 审核流程：管理员登录 → 审核用户 → 用户登录
    - 上传流程：登录 → 上传图片 → 图片出现在列表中
    - 权限测试：用户A无法编辑用户B的图片
    - Token 刷新：Access Token 过期自动刷新
  - 后端：Go 集成测试（`internal/handler` 的 HTTP test）
  - 前端：Playwright E2E 测试

  **Must NOT do**:
  - 不要跳过边界条件测试

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 测试编写，以现有功能为基础

  **Skills**: [`playwright`]
    - `playwright`: 浏览器E2E测试

  **Parallelization**:
  - **Can Run In Parallel**: NO（依赖所有前后端）
  - **Parallel Group**: Wave 8
  - **Depends On**: T12, T13, T18, T41
  - **Blocks**: T43, F1-F4

  **Acceptance Criteria**:
  - [ ] `go test ./...` 全部通过
  - [ ] `cd frontend && npx playwright test` 全部通过
  - [ ] 覆盖所有关键用户流程

  **QA Scenarios**:
  ```
  Scenario: Full guest browsing flow
    Tool: Playwright
    Steps:
      1. 打开首页
      2. 检查图片列表非空
      3. 点击图片 → 弹窗打开
      4. 关闭弹窗
      5. 搜索 "顾夏" → 结果更新
    Expected Result: 访客浏览全流程正常
    Evidence: .sisyphus/evidence/task-42-guest-flow.png

  Scenario: Full auth and upload flow
    Tool: Playwright
    Steps:
      1. 注册新用户 → 成功提示
      2. 管理员登录 → 审核该用户
      3. 用户登录 → 上传图片
      4. 图片出现在首页
      5. 用户编辑自己的图片 → 成功
      6. 用户尝试删除他人图片 → 失败
    Expected Result: 完整认证+上传流程正常
    Evidence: .sisyphus/evidence/task-42-auth-flow.png
  ```

  **Commit**: NO（由 T43 统一提交）

- [x] 43. **Git 初始提交**

  **What to do**:
  - `git add .` 添加所有项目文件
  - 确保 `.gitignore` 正确过滤（`.env`, `*.db`, `node_modules/`, `dist/`）
  - `git commit -m "feat: initial GM Site project — Vue 3 + Go/Echo + SQLite"` 
  - 提交信息参考 Commit Strategy 章节

  **Must NOT do**:
  - 不要提交任何含密钥/密码的文件
  - 不要提交 `gm_site.db`
  - 不要提交 `node_modules/`

  **Recommended Agent Profile**:
  - **Category**: `git`
    - Reason: Git 操作

  **Skills**: [`git-master`]
    - `git-master`: Git 提交操作

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 8
  - **Depends On**: T1, T42

  **Acceptance Criteria**:
  - [ ] `git log` 显示初始提交
  - [ ] `.gitignore` 生效（无敏感文件被提交）
  - [ ] 提交包含所有源文件

  **QA Scenarios**:
  ```
  Scenario: Git commit is clean
    Tool: Bash
    Steps:
      1. git status → "nothing to commit, working tree clean"
      2. git log -1 → 显示 feat commit
      3. 检查提交中不含 .env 或 *.db
    Expected Result: 干净初始提交
    Evidence: .sisyphus/evidence/task-43-git.txt
  ```

  **Commit**: YES（自身即为提交）
  - Message: `feat: initial GM Site project — Vue 3 + Go/Echo + SQLite`
  - Files: 所有项目源文件
  - Pre-commit: `go test ./... && cd frontend && npm test`

---

## Final Verification Wave

> 4 review agents run in PARALLEL. ALL must APPROVE. Present consolidated results to user and get explicit "okay" before completing.
>
> **Do NOT auto-proceed after verification. Wait for user's explicit approval before marking work complete.**

- [x] F1. **Plan Compliance Audit** — `oracle`
  Read the plan end-to-end. For each "Must Have": verify implementation exists (read file, curl endpoint, run command). For each "Must NOT Have": search codebase for forbidden patterns — reject with file:line if found. Check evidence files exist in .sisyphus/evidence/. Compare deliverables against plan.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

- [x] F2. **Code Quality Review** — `unspecified-high`
  Run `go vet ./...` + `golangci-lint` + `go test ./...` for backend. Run `npx tsc --noEmit` + `npx eslint` + `npm test` for frontend. Review all changed files for: `as any`/`@ts-ignore`, empty catches, console.log in prod, commented-out code, unused imports. Check AI slop: excessive comments, over-abstraction, generic names (data/result/item/temp).
  Output: `Build [PASS/FAIL] | Lint [PASS/FAIL] | Tests [N pass/N fail] | VERDICT`

- [x] F3. **Real Manual QA** — `unspecified-high` (+ `playwright` skill)
  Start from clean state. Execute EVERY QA scenario from EVERY task — follow exact steps, capture evidence. Test cross-task integration (register→approve→login→upload→view popup→comment→search). Test edge cases: empty state, invalid input, rapid actions, token expiry. Save to `.sisyphus/evidence/final-qa/`.
  Output: `Scenarios [N/N pass] | Integration [N/N] | VERDICT`

- [x] F4. **Scope Fidelity Check** — `deep`
  For each task: read "What to do", read actual diff (git log/diff). Verify 1:1 — everything in spec was built (no missing), nothing beyond spec was built (no creep). Check "Must NOT do" compliance. Detect cross-task contamination: Task N touching Task M's files. Flag unaccounted changes.
  Output: `Tasks [N/N compliant] | Contamination [CLEAN/N issues] | VERDICT`

---

## Commit Strategy

- **T1-T7**: `feat(init): project scaffolding` — backend/, frontend/, go.mod, package.json
- **T8-T14**: `feat(auth): user registration and JWT authentication` — internal/auth/...
- **T15-T20**: `feat(images): upload, CRUD, comments, search` — internal/image/...
- **T21-T22**: `feat(ws): WebSocket real-time stats` — internal/websocket/...
- **T23-T26**: `feat(ui): layout and decorative components` — frontend/src/components/...
- **T27-T33**: `feat(ui): gallery, popup, search, comments` — frontend/src/components/...
- **T34-T37**: `feat(ui): auth pages and home view` — frontend/src/views/...
- **T38-T40**: `feat(admin): admin management views` — frontend/src/views/admin/...
- **T41-T42**: `feat(integration): e2e tests and cors` — tests/, config
- **T43**: `chore: initial git commit`

---

## Success Criteria

### Verification Commands
```bash
# Backend
cd backend && go test ./...          # Expected: all pass
cd backend && go run ./cmd/server     # Expected: server starts on :1323

# Frontend
cd frontend && npm test              # Expected: all pass
cd frontend && npm run build         # Expected: build succeeds
```

### Final Checklist
- [ ] All "Must Have" present
- [ ] All "Must NOT Have" absent
- [ ] All go tests pass
- [ ] All vitest tests pass
- [ ] Frontend builds without errors
- [ ] Guest can browse images
- [ ] User can register, get approved, login, upload
- [ ] Admin can CRUD all images
- [ ] Popup is draggable and auto-reappears
- [ ] WebSocket stats are live
