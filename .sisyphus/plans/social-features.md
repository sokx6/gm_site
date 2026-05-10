# 好友系统 + 隐私 + 评论回复 + 邮件通知

## TL;DR

> **Quick Summary**: 完整社交功能——好友请求/接受、图片隐私（公开/好友/私密）、评论回复、全链路邮件通知、游客只看公开图。涉及后端 8 个新文件 + 改造 10 个文件，前端 5 个新组件 + 改造 8 个文件。

> **Estimated Effort**: Extra Large
> **Parallel Execution**: YES — 5 waves
> **Critical Path**: Wave 1 (DB + models) → Wave 2 (services) → Wave 3 (handlers + email) → Wave 4 (frontend) → Wave 5 (QA)

---

## Context

### 需求清单

| # | 需求 | 类型 |
|---|------|------|
| 1 | 用户间发送好友请求 | 新建 |
| 2 | 好友请求邮件通知被请求者 | 新建 |
| 3 | 接受/拒绝后通知请求者 | 新建 |
| 4 | 注册审核通过/拒绝邮件通知 | 改造 |
| 5 | 图片设为私密/好友可见/公开 | 新建 |
| 6 | 图片上传后可改隐私 | 改造 |
| 7 | 新建好友相册/私密相册 | 新建 |
| 8 | 评论回复 + 被回复人邮件通知 | 新建 |
| 9 | 修复评论过多弹窗溢出 | Bugfix |
| 10 | 游客只能看公开图片 | 改造 |
| 11 | 管理员便捷管理所有用户图片/评论/相册 | 改造 |

---

## Execution Strategy

```
Wave 1 (DB Foundation — 2 tasks):
├── Task 1: Database migration — 4 new tables + alter existing
└── Task 2: Models — 4 new models + update Image/Album/Comment

Wave 2 (Business Logic — 3 tasks, parallel):
├── Task 3: Friend service — request/accept/reject/list
├── Task 4: Privacy filter — image/album list filtered by visibility
└── Task 5: Comment reply — nested reply model + service

Wave 3 (API + Email — 4 tasks, parallel after Wave 2):
├── Task 6: Friend API handlers — routes + registration in main.go
├── Task 7: Email templates — 6 notification types
├── Task 8: Update image/album handlers — privacy field
└── Task 9: Approval email — send on approve/reject

Wave 4 (Frontend — 6 tasks, parallel):
├── Task 10: Friend management UI — requests list, accept/reject, friend list
├── Task 11: Privacy selector component — upload + edit forms
├── Task 12: Comment reply UI — nested replies, reply button
├── Task 13: Popup overflow fix — scrollable comment area
├── Task 14: Guest filter — hide non-public images for guests
└── Task 15: Admin dashboard — 统一管理面板（图片/评论/相册列表 + 搜索 + 批量操作）

Wave 5 (Integration — 2 tasks):
├── Task 16: go build + npm build + deploy
└── Task 17: End-to-end QA — Playwright test all flows
```

---

## Database Changes

### New Tables
```sql
-- 好友关系
CREATE TABLE friends (
  id INTEGER PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id),
  friend_id INTEGER NOT NULL REFERENCES users(id),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(user_id, friend_id)
);

-- 好友请求
CREATE TABLE friend_requests (
  id INTEGER PRIMARY KEY,
  from_user_id INTEGER NOT NULL REFERENCES users(id),
  to_user_id INTEGER NOT NULL REFERENCES users(id),
  status TEXT NOT NULL DEFAULT 'pending', -- pending/accepted/rejected
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 通知
CREATE TABLE notifications (
  id INTEGER PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id),
  type TEXT NOT NULL, -- friend_request/friend_accepted/friend_rejected/register_approved/register_rejected/comment_reply
  title TEXT NOT NULL,
  content TEXT NOT NULL,
  related_id INTEGER, -- friend_request_id or comment_id or null
  is_read INTEGER DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Alter Existing
```sql
ALTER TABLE images ADD COLUMN privacy TEXT NOT NULL DEFAULT 'public'; -- public/friends/private
ALTER TABLE albums ADD COLUMN privacy TEXT NOT NULL DEFAULT 'public';
ALTER TABLE albums ADD COLUMN is_friend_album INTEGER DEFAULT 0;
ALTER TABLE comments ADD COLUMN parent_id INTEGER REFERENCES comments(id);
```

---

## New API Endpoints

| Method | Path | Auth | 说明 |
|--------|------|------|------|
| POST | `/api/friends/request` | Yes | 发送好友请求 |
| PUT | `/api/friends/request/:id/accept` | Yes | 接受 |
| PUT | `/api/friends/request/:id/reject` | Yes | 拒绝 |
| GET | `/api/friends/requests` | Yes | 好友请求列表 |
| GET | `/api/friends` | Yes | 好友列表 |
| DELETE | `/api/friends/:id` | Yes | 删除好友 |
| GET | `/api/notifications` | Yes | 通知列表 |
| PUT | `/api/notifications/:id/read` | Yes | 标记已读 |
| PUT | `/api/images/:id/privacy` | Yes | 修改图片隐私 |
| PUT | `/api/albums/:id/privacy` | Yes | 修改相册隐私 |
| POST | `/api/comments/:id/reply` | Yes | 回复评论 |

---

## Frontend New Components

| 组件 | 说明 |
|------|------|
| `FriendPanel.vue` | 好友管理面板（请求列表 + 好友列表） |
| `PrivacyBadge.vue` | 隐私标签（公开/好友/私密） |
| `PrivacySelector.vue` | 上传/编辑时的隐私下拉选择 |
| `NotificationBell.vue` | 通知铃铛 + 红点 + 下拉列表 |
| `ReplyForm.vue` | 评论内嵌回复表单 |

---

## Must Have
- 好友请求双向邮件通知
- 注册审批邮件通知
- 图片三级隐私（公开/好友/私密），上传后可改
- 评论回复功能 + 被回复人邮件通知
- 弹窗评论区域可滚动（修复溢出）
- 游客只能看到 privacy=public 的图片

## Must NOT Have
- 不修改前端路由结构
- 不引入第三方推送服务
- 不做实时通知（WebSocket 推送日后补）
- 不做私信/聊天功能

---

## TODOs

---

## Final Verification Wave

---

## Commit Strategy
- **1-2**: `feat(backend): add friend system database migration and models`
- **3-5**: `feat(backend): add friend service, privacy filter, comment reply`
- **6-9**: `feat(backend): add friend/notification API handlers and email templates`
- **10-14**: `feat(frontend): add friend panel, privacy selector, comment reply, notifications`
- **15-16**: `verify: go build + npm build + Playwright QA`

---

## Success Criteria
- [ ] 用户 A 向用户 B 发好友请求 → B 收到邮件
- [ ] B 接受 → A 收到通知邮件
- [ ] 管理员审批注册 → 用户收到通知邮件
- [ ] 图片设为"好友可见" → 游客看不到，好友能看到
- [ ] 图片设为"私密" → 只有上传者能看到
- [ ] 评论回复 → 被回复人收到邮件
- [ ] 多条评论时弹窗可滚动不出屏
- [ ] `go build` + `npm run build` 通过
