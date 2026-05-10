# 代码审查与修复计划

## TL;DR

> **Quick Summary**: 全栈代码审查发现 3 个 Critical、5 个 High、4 个 Medium 问题。优先修复安全漏洞（密钥泄露、JWT 跨站），其次是 bug 修复和性能优化，最后 git commit 规范化。

> **Estimated Effort**: Medium
> **Parallel Execution**: YES — 3 waves

---

## 审查结果

### 🔴 Critical (立即修复)

| # | 文件 | 问题 | 修复 |
|---|------|------|------|
| C1 | `*.yaml` | config.local.yaml/config.syk91.yaml 含真实密码被 git 追踪 | 加入 .gitignore，`git rm --cached` |
| C2 | `main.go` | 三站点共享 JWT 密钥，token 跨站互认 | 每站点生成独立密钥，或改为依赖数据库用户验证 |
| C3 | `auth.ts:40` | `atob()` 解码 JWT 无签名验证，前端判断管理员角色不可信 | 改为调用 `/api/auth/me` 接口验证 |

### 🟠 High (优先修复)

| # | 文件 | 问题 | 修复 |
|---|------|------|------|
| H1 | `comment.go:105` | goroutine 无 panic 恢复，email panics 静默吞掉 | 加 `defer recover()` |
| H2 | `comment.go` | 调试日志未清理（10+ 处 `logger.L.Info`） | 改为 Debug 级别或移除 |
| H3 | `UploadModal.vue` | `URL.createObjectURL` 在组件卸载时可能泄漏 | 加 `onUnmounted` 清理 |
| H4 | `ImagePopup.vue:184` | `scheduleReappear()` 关闭的弹窗 5-13s 后自动重开 | 移除或改为用户可配置 |
| H5 | `client.ts:47` | refresh 失败后 `window.location.hash` 强制跳转丢失状态 | 先清 state 再跳转，加用户提示 |

### 🟡 Medium（建议修复）

| # | 文件 | 问题 | 修复 |
|---|------|------|------|
| M1 | `config.go` | CORS 多域名逗号字符串可能解析失败 | 改为 `[]string` 或空格分隔 |
| M2 | `image.go:172` | `ListImages` 缺少 OptionalAuth 导致 viewerID 始终为 0 | 确认 middleware 链正确 |
| M3 | `ws client` | WebSocket 连接无心跳机制，代理可能断开 | 加 ping/pong |
| M4 | `stats.go` | `GetNewMembersCount` 每次广播都查 DB | 缓存 1 分钟 |

---

## 修复任务

### Wave 1: Security (3 tasks, parallel)
- [ ] C1: .gitignore + git rm secrets
- [ ] C2: 独立 JWT 密钥 (或改为 32 位随机 key)
- [ ] C3: 后端加 `GET /api/auth/me`，前端改为 API 验证角色

### Wave 2: Bug fixes (3 tasks, parallel)
- [ ] H1: 评论 goroutine panic recovery
- [ ] H2: 清理调试日志 (comment.go + 其他)
- [ ] H3: UploadModal URL cleanup + H4 popup auto-reopen

### Wave 3: Git commit + deploy
- [ ] 清理 session 文件 + 被追踪的临时文件
- [ ] 规范 commit message
- [ ] 部署到三站点

---

## Commit Strategy
1. `fix(security): remove tracked secrets, generate unique JWT keys`
2. `fix(backend): add panic recovery to comment goroutine, cleanup debug logs`
3. `fix(frontend): fix URL.createObjectURL leak, remove popup auto-reopen`
4. `chore: gitignore session files, clean up tracked artifacts`

---

## Success Criteria
- [ ] `config.local.yaml` 不再被 git 追踪
- [ ] JWT 密钥三个站点各不相同
- [ ] 前端角色验证改用 API 而非本地解码
- [ ] 评论 goroutine 不会因 panic 静默死掉
- [ ] `git status` 干净无敏感文件
