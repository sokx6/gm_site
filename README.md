# GM Site —— 群友图片展示网站

赛博朋克风格的群友图片展示网站，支持用户注册登录、图片上传至兰空图床、可拖拽弹窗浏览、实时数据展示。

> ⚠️ **声明**：本项目由 AI（OpenCode + Claude/DeepSeek）全自动生成，包括前后端代码、数据库设计、测试用例及文档。人工仅提供需求描述和配置调整。
> 🤖 AI-Generated | 🛠 Go + Vue 3 + SQLite | 📡 Lsky Pro | 🔐 JWT 双Token

## 技术栈

- **后端**: Go + Echo v4 + SQLite（modernc.org/sqlite，纯 Go 无 CGO 依赖）
- **前端**: Vue 3 + Vite + TypeScript（静态 SPA）
- **图床**: Lsky Pro（兰空图床）API 代理
- **认证**: JWT 双 Token 机制（Access Token 15min + Refresh Token 7d）
- **邮件**: SMTP TLS（支持 163、QQ 等邮箱）
- **实时通信**: WebSocket 在线人数/统计

---

## 目录结构

```
gm_site/
├── backend/                    # Go 后端
│   ├── cmd/server/main.go      # 程序入口
│   ├── internal/               # 内部包
│   │   ├── config/             # 配置加载（viper）
│   │   ├── database/           # 数据库初始化 + 迁移
│   │   ├── deps/               # 依赖注入容器
│   │   ├── handler/            # HTTP 处理器（路由控制器）
│   │   ├── middleware/         # JWT、CORS 等中间件
│   │   ├── model/              # 数据模型
│   │   ├── repository/         # 数据访问层
│   │   ├── service/            # 业务逻辑层
│   │   └── websocket/          # WebSocket 管理
│   ├── migrations/             # SQLite 数据库迁移脚本
│   ├── config.yaml             # 默认配置文件（模板）
│   ├── config.local.yaml       # 本地配置（不纳入版本控制）
│   ├── .env.example            # 环境变量参考
│   ├── go.mod / go.sum         # Go 依赖
│   └── server.exe              # 编译产物（不纳入版本控制）
├── frontend/                   # Vue 3 前端
│   ├── src/
│   │   ├── api/                # API 请求封装（axios）
│   │   ├── assets/             # 静态资源（CSS、图片等）
│   │   ├── components/         # 可复用组件
│   │   ├── composables/        # Vue 组合式函数
│   │   ├── router/             # 路由配置
│   │   ├── stores/             # Pinia 状态管理
│   │   ├── views/              # 页面组件
│   │   ├── App.vue             # 根组件
│   │   ├── main.ts             # 入口文件
│   │   └── style.css           # 全局样式
│   ├── public/                 # 静态资源（不经过编译）
│   ├── dist/                   # 构建产物（不纳入版本控制）
│   ├── vite.config.ts          # Vite 配置（含开发代理）
│   ├── package.json            # Node 依赖
│   └── index.html              # HTML 入口
├── .gitignore
└── README.md
```

---

## 环境要求

| 依赖 | 版本要求 | 说明 |
|------|----------|------|
| Go | >= 1.21 | 编译后端 |
| Node.js | >= 18 | 编译前端 |
| 兰空图床实例 | 已部署 | 图片存储后端 |
| SMTP 邮箱账号 | 163/QQ 等 | 发送注册验证/通知邮件 |

---

## 快速开始（开发环境）

### 1. 克隆项目

```bash
git clone <your-repo-url> gm_site
cd gm_site
```

### 2. 配置后端

```bash
cd backend

# 复制默认配置为本地配置（config.local.yaml 不纳入版本控制）
cp config.yaml config.local.yaml
```

编辑 `config.local.yaml`，将占位符替换为真实配置。核心字段如下：

```yaml
server:
  port: 1323                    # 后端监听端口
  host: "0.0.0.0"

database:
  path: "gm_site.db"            # SQLite 数据库文件路径

jwt:
  access_secret: "至少32位随机字符串，生产环境必须更换"
  refresh_secret: "至少32位随机字符串，与 access_secret 不同"
  access_expire: "15m"          # Access Token 有效期
  refresh_expire: "168h"        # Refresh Token 有效期（7天）

smtp:
  host: "smtp.163.com"          # SMTP 服务器
  port: 587                     # TLS 端口
  username: "your-email@163.com"
  password: "授权码（非邮箱登录密码）"
  from: "your-email@163.com"
  admin_email: "admin@example.com"
  use_tls: true

lsky:
  base_url: "https://your-lsky-instance.com"  # 兰空图床地址，末尾不带斜杠
  email: "lsky@example.com"
  password: "兰空图床登录密码"
  token: ""                     # 留空，程序启动时自动获取

admin:
  email: "admin@example.com"    # 使用此邮箱注册的用户自动成为管理员

site:
  name: "群友名称"               # 站点展示名称
  start_date: "2024-01-01"      # 安全运行起始日期

upload:
  max_size_mb: 10               # 单张图片上传大小限制（MB）
```

> **提示**: 敏感字段（密码、密钥）可通过环境变量覆盖，避免写入配置文件。详见[配置说明](#配置说明)。

### 3. 启动后端

```bash
# 确保当前目录为 backend/
go run ./cmd/server
```

首次启动会自动：
- 创建 SQLite 数据库文件（`gm_site.db`）
- 执行数据库迁移（创建 users、albums、images 等表）
- 连接兰空图床并获取 API Token

后端启动后访问 `http://localhost:1323`。

### 4. 启动前端

打开新终端：

```bash
cd frontend
npm install
npm run dev
```

前端开发服务器启动后访问 `http://localhost:5173`，API 请求会自动代理到后端 `http://localhost:1323`（参见 `vite.config.ts` 中的 proxy 配置）。

### 5. 首次使用

1. 使用 `admin.email` 配置的邮箱在前端注册账号，自动获得管理员权限
2. 管理员登录后，可在后台审核其他注册用户
3. 审核通过的用户即可上传图片，后端代理上传至兰空图床并返回链接

---

## 配置说明

配置文件位于 `backend/config.yaml`（模板）或 `backend/config.local.yaml`（本地覆盖）。使用 [viper](https://github.com/spf13/viper) 加载，支持 YAML 格式。

### 字段详解

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `server.port` | int | 是 | 后端 HTTP 服务端口，默认 1323 |
| `server.host` | string | 是 | 监听地址，`0.0.0.0` 监听所有网卡 |
| `database.path` | string | 是 | SQLite 数据库文件路径，相对于工作目录 |
| `jwt.access_secret` | string | 是 | Access Token 签名密钥，至少 32 位随机字符 |
| `jwt.refresh_secret` | string | 是 | Refresh Token 签名密钥，与 access_secret 不同 |
| `jwt.access_expire` | string | 是 | Access Token 过期时间，格式 `数字+单位`（s/m/h），如 `15m` |
| `jwt.refresh_expire` | string | 是 | Refresh Token 过期时间，如 `168h`（7 天） |
| `smtp.host` | string | 是 | SMTP 服务器地址，如 `smtp.163.com` |
| `smtp.port` | int | 是 | SMTP 端口，TLS 通常为 587 |
| `smtp.username` | string | 是 | SMTP 登录用户名（完整邮箱地址） |
| `smtp.password` | string | 是 | SMTP 密码（163/QQ 邮箱为授权码，非登录密码） |
| `smtp.from` | string | 是 | 发件人地址，通常与 username 相同 |
| `smtp.admin_email` | string | 是 | 管理员邮箱，接收新用户注册通知 |
| `smtp.use_tls` | bool | 是 | 是否启用 TLS 加密 |
| `lsky.base_url` | string | 是 | 兰空图床 API 地址，末尾不带斜杠 |
| `lsky.email` | string | 是 | 兰空图床登录邮箱 |
| `lsky.password` | string | 是 | 兰空图床登录密码 |
| `lsky.token` | string | 否 | 兰空图床 API Token，留空则启动时自动获取 |
| `admin.email` | string | 是 | 管理员邮箱，该邮箱注册的用户自动获得管理员权限 |
| `site.name` | string | 是 | 站点名称（群友昵称），展示在页面标题等位置 |
| `site.start_date` | string | 是 | 安全运行起始日期，格式 `YYYY-MM-DD`，用于计算运行天数 |
| `upload.max_size_mb` | int | 是 | 单张图片上传大小限制（MB） |

### 环境变量覆盖

以下敏感字段支持通过环境变量覆盖（环境变量优先级高于配置文件）：

| 环境变量 | 对应配置字段 |
|----------|-------------|
| `GM_SMTP_PASSWORD` | `smtp.password` |
| `GM_JWT_ACCESS_SECRET` | `jwt.access_secret` |
| `GM_JWT_REFRESH_SECRET` | `jwt.refresh_secret` |
| `GM_LSKY_PASSWORD` | `lsky.password` |

使用示例（Linux/macOS）：

```bash
export GM_SMTP_PASSWORD="your-auth-code"
export GM_JWT_ACCESS_SECRET="a-very-long-random-string-at-least-32-chars"
export GM_JWT_REFRESH_SECRET="another-long-random-string-32-chars"
export GM_LSKY_PASSWORD="your-lsky-password"

go run ./cmd/server
```

Windows（PowerShell）：

```powershell
$env:GM_SMTP_PASSWORD="your-auth-code"
$env:GM_JWT_ACCESS_SECRET="a-very-long-random-string-at-least-32-chars"
$env:GM_JWT_REFRESH_SECRET="another-long-random-string-32-chars"
$env:GM_LSKY_PASSWORD="your-lsky-password"

go run ./cmd/server
```

也可创建 `backend/.env` 文件（参见 `backend/.env.example`），程序启动时自动加载。

---

## 生产部署

### 方式一：直接部署（Nginx 反向代理）

#### 构建后端

```bash
cd backend
go build -o gm_site_server ./cmd/server
```

#### 构建前端

```bash
cd frontend
npm install
npm run build        # 产出 dist/ 目录
```

#### Nginx 配置

将以下配置保存为 `/etc/nginx/sites-available/gm_site`：

```nginx
server {
    listen 80;
    server_name your-domain.com;

    # 前端静态文件
    root /path/to/gm_site/frontend/dist;
    index index.html;

    # API 代理
    location /api/ {
        proxy_pass http://127.0.0.1:1323;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # WebSocket 代理
    location /api/ws {
        proxy_pass http://127.0.0.1:1323;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # SPA 前端路由：所有非 API 请求返回 index.html
    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

启用配置：

```bash
sudo ln -s /etc/nginx/sites-available/gm_site /etc/nginx/sites-enabled/
sudo nginx -t          # 测试配置
sudo nginx -s reload   # 重载 Nginx
```

#### 启动后端

```bash
cd backend
./gm_site_server
```

推荐配合 `screen` 或 `tmux` 保持后台运行，或使用 systemd（见下节）。

---

### 方式二：systemd 服务（推荐）

创建服务文件 `/etc/systemd/system/gm_site.service`：

```ini
[Unit]
Description=GM Site Server
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/path/to/gm_site/backend
ExecStart=/path/to/gm_site/backend/gm_site_server
Restart=on-failure
RestartSec=5

# 环境变量（敏感信息通过环境变量注入）
Environment="GM_SMTP_PASSWORD=your-smtp-auth-code"
Environment="GM_JWT_ACCESS_SECRET=your-access-secret"
Environment="GM_JWT_REFRESH_SECRET=your-refresh-secret"
Environment="GM_LSKY_PASSWORD=your-lsky-password"

[Install]
WantedBy=multi-user.target
```

启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable gm_site
sudo systemctl start gm_site
sudo systemctl status gm_site   # 查看运行状态
```

常用管理命令：

```bash
sudo systemctl stop gm_site     # 停止服务
sudo systemctl restart gm_site  # 重启服务
sudo journalctl -u gm_site -f   # 查看实时日志
```

---

## 首次运行流程

1. **启动后端**：程序自动创建 SQLite 数据库并执行迁移脚本
2. **管理员注册**：使用配置的 `admin.email` 邮箱在前端注册账号，该账号自动获得管理员权限
3. **用户审核**：管理员登录后进入后台，审核其他注册用户
4. **上传图片**：审核通过的用户可在前端上传图片，后端将图片代理上传至兰空图床，返回可访问的图片链接

---

## 环境变量完整参考

| 变量名 | 说明 | 必填 |
|--------|------|------|
| `GM_SMTP_PASSWORD` | SMTP 邮箱密码/授权码 | 推荐 |
| `GM_JWT_ACCESS_SECRET` | JWT Access Token 签名密钥（至少 32 字符） | 推荐 |
| `GM_JWT_REFRESH_SECRET` | JWT Refresh Token 签名密钥（至少 32 字符） | 推荐 |
| `GM_LSKY_PASSWORD` | 兰空图床登录密码 | 推荐 |

---

## 常见问题

### 端口被占用

修改 `config.local.yaml` 中 `server.port` 为其他端口（如 `1324`），同时更新前端 `vite.config.ts` 中的代理目标地址。

### CORS 错误

开发环境下 Vite 已配置代理（`vite.config.ts`），前端请求 `/api` 会被代理到后端。如果仍然出现 CORS 错误，检查：

1. 后端是否在 `http://localhost:1323` 启动
2. `vite.config.ts` 中 proxy target 是否正确

### 兰空图床上传失败

1. 确认 `lsky.base_url` 可访问，末尾不要带斜杠
2. 确认兰空图床账号密码正确
3. 检查后端启动日志中的 Token 获取状态
4. 验证图片大小是否超过 `upload.max_size_mb` 限制

### 邮件发送失败

1. 确认 SMTP 服务器地址和端口正确：
   - 163 邮箱：`smtp.163.com:587`
   - QQ 邮箱：`smtp.qq.com:587`
2. 确认使用的是**授权码**而非邮箱登录密码（163/QQ 邮箱需在设置中生成）
3. 确认 `smtp.use_tls` 设置为 `true`
4. 检查服务器防火墙是否放行 SMTP 端口（587）

### 数据库文件权限

确保后端运行用户对 `database.path` 指定的路径有读写权限。如果使用 SQLite，数据库文件所在的**目录**也需要写权限（用于创建 WAL 日志文件）。

### 前端构建报错

```bash
cd frontend
rm -rf node_modules package-lock.json
npm install
npm run build
```

确保 Node.js 版本 >= 18。

---

## 关于本项目

本项目为 **AI 全自动生成** 的全栈项目，从需求分析 → 架构设计 → 代码实现 → 测试编写 → 部署文档，全部由 AI 完成。

- **AI 模型**: Claude (Anthropic) + DeepSeek
- **AI 工作流**: OpenCode + Sisyphus 多智能体协作系统
- **人工参与**: 需求描述、配置填写、代码审查

如有问题或建议，欢迎提 Issue。
