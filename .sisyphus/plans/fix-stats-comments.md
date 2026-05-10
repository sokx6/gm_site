# 修复统计数据显示和评论功能

## TL;DR

> **Quick Summary**: 修复 5 个前端统计与评论组件的 6 个缺陷 — 全部是数据通道断连问题（硬编码、字段名不匹配、路由路径错、缺少认证头），后端数据正确，不动后端。

> **Deliverables**:
> - TopNotifyBar 显示实时数据（在线人数/新增成员/运行天数）
> - VisitorCounter 显示累计访客数
> - StatsBar WebSocket 字段正确映射
> - 评论提交/删除/加载正常

> **Estimated Effort**: Quick
> **Parallel Execution**: YES — 2 waves
> **Critical Path**: Task 1 → Task 5 (Wave 1) → 验证

---

## Context

### Original Request
前端统计组件（在线人数、新增成员、安全运行天数、访客计数）全部显示 `--` 或 `0`，评论功能无法使用。

### Interview Summary
**Key Discussions**:
- TopNotifyBar props 硬编码 `"--"`，无数据绑定
- VisitorCounter 的 `visitorCount` ref 从未更新
- StatsBar WebSocket 消息字段名称与后端不匹配
- CommentSection: DELETE 路由路径错、提交不用 API client 缺 JWT、响应解析层级错

**Research Findings**:
- 后端 `StatsService.BroadcastStats()` 每 10s 广播 `{"type":"stats","data":{"online":N,"visitors":N,"newMembers":N,"safeDays":N}}`
- StatsBar 前端读 `msg.online` 而非 `msg.data.online`，字段 `totalVisitors`/`uptimeDays` 与后端 `visitors`/`safeDays` 不对应
- CommentSection 用裸 `fetch()` 而非经 JWT 拦截器的 `apiClient`
- 评论 DELETE: 前端路径 `/api/images/:id/comments/:cid`，后端 `/api/comments/:cid`

### Metis Review
**Identified Gaps** (addressed):
- WebSocket 连接被破坏风险: StatsBar 和 HomeView 各自独立建 WS 连接 — 修复方案将共享单个 WebSocket 连接
- 评论 POST 的 API 响应解析: `data.data.comments` 而非 `data.comments`
- TopNotifyBar 和 VisitorCounter 无数据源 — 从 WebSocket stats 消息统一获取

---

## Work Objectives

### Core Objective
打通前端所有统计与评论组件的数据通道，使实时统计和评论功能正常工作。

### Concrete Deliverables
- `HomeView.vue` — 添加 WebSocket 连接、更新 TopNotifyBar 绑定、VisitorCounter 绑定
- `StatsBar.vue` — 修复 WebSocket 消息字段映射
- `CommentSection.vue` — 修复路由、认证、响应解析
- `TopNotifyBar.vue` — 无需改动（纯展示组件，props 即可）
- `VisitorCounter.vue` — 无需改动（纯展示组件，props 即可）

### Definition of Done
- [ ] 页面刷新后 TopNotifyBar 数字不是 `--`（由 WebSocket 消息更新）
- [ ] StatsBar 侧栏四个统计值在数秒内更新为真实数字
- [ ] VisitorCounter 访客数 > 0
- [ ] 登录用户可提交评论，评论列表正确显示
- [ ] 管理员/作者可删除评论

### Must Have
- WebSocket 连接复用：不创建第二个 WebSocket 连接
- 评论采用 JWT 认证（使用 apiClient 或自行附加 Bearer token）
- 所有数据路径仅前端改动，零后端变更

### Must NOT Have (Guardrails)
- 不修改后端任何代码
- 不新增后端 API 端点
- 不创建新的 WebSocket 端点
- 不破坏已有的图片上传/展示功能
- 不修改 GalleryCard / GalleryGrid / ImagePopup（只读组件，不涉及）

---

## Verification Strategy

> **ZERO HUMAN INTERVENTION** — ALL verification is agent-executed.

### Test Decision
- **Infrastructure exists**: YES (vitest, vue-test-utils)
- **Automated tests**: TESTS-AFTER — 先修复，后补测试
- **Framework**: vitest

### QA Policy
每任务含 Agent-Executed QA Scenarios。前端用 Playwright 验证；API 用 curl 验证。

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Start Immediately — 3 independent tasks):
├── Task 1: HomeView — Add shared WebSocket + bind TopNotifyBar/VisitorCounter [quick]
├── Task 2: StatsBar — Fix WebSocket field mapping [quick]
└── Task 3: CommentSection — Fix route + auth + response parsing [quick]

Wave FINAL (After ALL tasks):
├── Task F1: Playwright smoke test — verify stats display [visual-engineering]
├── Task F2: Playwright smoke test — verify comment CRUD [visual-engineering]
└── Task F3: Verify no regressions (image upload, gallery browse) [visual-engineering]
```

Critical Path: All Wave 1 tasks → Wave FINAL
Max Concurrent: 3 (Wave 1)

---

## TODOs

- [x] 1. HomeView — Add shared WebSocket connection and bind TopNotifyBar + VisitorCounter

  **What to do**:
  - 在 `<script setup>` 中添加 WebSocket 连接逻辑（参考 StatsBar.vue 的 connect 模式，复制核心代码）
  - 创建响应式 ref：`topOnline`、`topNewMembers`、`topUptime`、`visitorCount`，初始值均为 `'--'` 或 `0`
  - WebSocket `onmessage` 中解析 `msg.data` 更新这些 ref：
    - `topOnline` ← `msg.data.online`
    - `topNewMembers` ← `msg.data.newMembers`
    - `topUptime` ← `msg.data.safeDays`
    - `visitorCount` ← `msg.data.visitors`
  - 模板中 TopNotifyBar props 改为绑定 ref：
    - `:online-count="topOnline"`
    - `:new-members="topNewMembers"`
    - `:uptime="topUptime"`
  - `onMounted` 中调 connect，`onUnmounted` 中清理 WS 连接
  - **Key constraint**: 不要导入 StatsBar，自己维护一个独立的 WebSocket 实例（StatsBar 已有自己的一份，不动它避免耦合）

  **Must NOT do**:
  - 不要修改 TopNotifyBar.vue 或 VisitorCounter.vue 内部代码
  - 不要创建共享的 WebSocket composable（保持简单，每个组件自己管理）

  **Recommended Agent Profile**:
  > `quick` — 单文件改动，复制已有模式，纯连线工作
  - **Category**: `quick`
  - **Skills**: `[]`
  - **Skills Evaluated but Omitted**: `frontend-ui-ux` (无需视觉设计)

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 2, 3)
  - **Blocks**: None
  - **Blocked By**: None (can start immediately)

  **References**:
  - `frontend/src/components/StatsBar.vue:9-45` — WebSocket connect/reconnect 模式直接复用
  - `frontend/src/components/TopNotifyBar.vue:1-14` — props 接口定义
  - `frontend/src/components/VisitorCounter.vue:4-13` — props 接口定义
  - `frontend/src/views/HomeView.vue:221-225` — TopNotifyBar 当前硬编码调用
  - `frontend/src/views/HomeView.vue:287` — VisitorCounter 当前调用
  - `backend/internal/service/stats.go:21-28` — WebSocket 消息的 StatsData 结构体（字段名对照）

  **Acceptance Criteria**:
  - [ ] 页面加载后 TopNotifyBar 显示的数字在 WebSocket 连接后从 `--` 变为实际数字
  - [ ] VisitorCounter 从 `0` 变为 `visitors` 实际数字
  - [ ] 组件销毁时 WebSocket 正确清理（无内存泄漏）

  **QA Scenarios**:

  ```
  Scenario: 页面刷新后统计数字由 WebSocket 更新
    Tool: Playwright
    Preconditions: 后端运行，数据库有 visitor 记录
    Steps:
      1. 导航到 https://qgx91.xyz
      2. 等待 5 秒（WebSocket 连接 + 首次广播）
      3. 找到 .top-notify-bar strong 元素 — 断言 textContent 不包含 "--"
      4. 找到 .visitor-counter .digit-group — 断言 .digit 子元素存在且非全 "0"
    Expected Result: TopNotifyBar 和 VisitorCounter 显示非占位符数字
    Failure Indicators: 仍显示 "--" 或 "0"（WebSocket 未连接或消息解析失败）
    Evidence: .sisyphus/evidence/task-1-stats-update.png

  Scenario: WebSocket 断线后自动重连
    Tool: Playwright + 服务器端 kill WebSocket
    Preconditions: 页面已加载且 WebSocket 连接建立
    Steps:
      1. 导航到 https://qgx91.xyz，等待 3 秒建立连接
      2. 在服务器端重启后端进程（断掉 WS）
      3. 等待 10 秒观察页面
    Expected Result: 3 秒内自动重连，统计数据恢复
    Failure Indicators: 数字永久停留在旧值或变为 "--"
    Evidence: .sisyphus/evidence/task-1-reconnect.png
  ```

  **Commit**: YES
  - Message: `fix(frontend): connect HomeView to WebSocket for live stats`
  - Files: `frontend/src/views/HomeView.vue`

- [x] 2. StatsBar — Fix WebSocket message field name mapping

  **What to do**:
  - 修改 `StatsBar.vue` 中 `ws.onmessage` 回调（第 22-34 行）
  - 将 `msg.online` 改为 `msg.data?.online ?? msg.online`
  - 将 `msg.totalVisitors` 改为 `msg.data?.visitors ?? msg.totalVisitors`
  - 将 `msg.uptimeDays` 改为 `msg.data?.safeDays ?? msg.uptimeDays`
  - `msg.newMembers` 保持不变（路径正确 `msg.data?.newMembers ?? msg.newMembers`）
  - 使用可选链 `msg.data?.` 保持向后兼容（如果将来消息格式变化不会崩溃）

  **Must NOT do**:
  - 不要改动 WebSocket 连接逻辑
  - 不新增 ref 变量

  **Recommended Agent Profile**:
  > `quick` — 4 行字段名修正，影响范围极小
  - **Category**: `quick`
  - **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 3)
  - **Blocks**: None
  - **Blocked By**: None

  **References**:
  - `frontend/src/components/StatsBar.vue:22-34` — 当前 onmessage handler
  - `backend/internal/service/stats.go:18-29` — StatsMessage + StatsData 结构（JSON tag 对照）

  **Acceptance Criteria**:
  - [ ] 页面加载后 StatsBar 的四个统计值 3 秒内由 `--` 变为真实数字
  - [ ] 字段值正确：在线人数、累计访客、新增成员、运行天数均非 `--`

  **QA Scenarios**:

  ```
  Scenario: StatsBar 在 WebSocket 消息到达后显示真实数据
    Tool: Playwright
    Preconditions: 后端运行
    Steps:
      1. 导航到 https://qgx91.xyz
      2. 等待 5 秒
      3. 在 .stats-bar 中找到所有 strong 元素
      4. 断言所有 strong 的 textContent 不包含 "--"
    Expected Result: 四个统计值均为数字
    Failure Indicators: 仍有 "--" 显示
    Evidence: .sisyphus/evidence/task-2-statsbar.png
  ```

  **Commit**: YES
  - Message: `fix(frontend): fix StatsBar WebSocket field name mapping`
  - Files: `frontend/src/components/StatsBar.vue`

- [x] 3. CommentSection — Fix route, auth, and response parsing

  **What to do**:
  1. **DELETE 路由修复** (line 88): 将 `/api/images/${props.imageId}/comments/${commentId}` 改为 `/api/comments/${commentId}`
  2. **POST 请求加 Authorization** (line 70-74): 
     - 从 localStorage 读取 `accessToken`
     - 在 headers 中添加 `Authorization: Bearer ${token}`
  3. **响应解析修复** (line 49-50):
     - 将 `data.comments` 改为 `data.data?.comments`（处理 APIResponse 包装层）
     - 同时保持 `Array.isArray(data)` 兜底

  **Must NOT do**:
  - 不改为使用 apiClient（避免与现有 fetch 模式不一致）
  - 不修改 CommentSection 的模板部分

  **Recommended Agent Profile**:
  > `quick` — 3 处小改，纯补丁
  - **Category**: `quick`
  - **Skills**: `[]`

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2)
  - **Blocks**: None
  - **Blocked By**: None

  **References**:
  - `frontend/src/components/CommentSection.vue:40-62` — fetchComments（响应解析）
  - `frontend/src/components/CommentSection.vue:65-84` — submitComment（POST 请求）
  - `frontend/src/components/CommentSection.vue:86-95` — deleteComment（DELETE 请求）
  - `frontend/src/api/client.ts:8-12` — 请求拦截器模式（Bearer token 附加方式）
  - `backend/cmd/server/main.go:136-137` — 评论路由定义
  - `backend/internal/handler/comment.go:106-112` — ListByImage 响应格式

  **Acceptance Criteria**:
  - [ ] 登录用户可成功提交评论（POST 200）
  - [ ] 评论列表正确显示已提交的评论
  - [ ] 管理员/作者可删除评论（DELETE 200）
  - [ ] 未登录用户提交评论返回 401（graceful handling）

  **QA Scenarios**:

  ```
  Scenario: 登录用户提交评论后列表显示
    Tool: Playwright
    Preconditions: 已登录，弹出图片详情弹窗（CommentSection 可见）
    Steps:
      1. 登录后导航到首页，点击一张图片打开弹窗
      2. 在 .comment-textarea 中输入 "测试评论内容"
      3. 点击提交按钮（查找包含 "发表" 文字的 button）
      4. 等待 2 秒
      5. 查找 .comment-nickname — 断言存在且内容非空
      6. 查找评论内容 — 断言包含 "测试评论内容"
    Expected Result: 新评论出现在评论列表中
    Failure Indicators: 评论不出现或 alert 弹出错误
    Evidence: .sisyphus/evidence/task-3-comment-crud.png

  Scenario: 删除自己的评论
    Tool: Playwright
    Preconditions: 已登录（管理员或评论作者），评论列表有评论
    Steps:
      1. 在上述场景的评论列表中，找到删除按钮
      2. 点击删除按钮
      3. 等待 1 秒
      4. 断言该评论已从 DOM 中移除（.comment-nickname 计数减少）
    Expected Result: 评论被删除
    Failure Indicators: 评论仍在列表中
    Evidence: .sisyphus/evidence/task-3-comment-delete.png
  ```

  **Commit**: YES
  - Message: `fix(frontend): fix CommentSection route, auth, and response parsing`
  - Files: `frontend/src/components/CommentSection.vue`

---

## Final Verification Wave

> 3 review activities run sequentially (not parallel — each is a Playwright script).

- [x] F1. **Stats Verification** ✅ — PASS (6/167/1/0 all real numbers)
  Navigate to https://qgx91.xyz, wait 5s for WebSocket data. Verify:
  - TopNotifyBar: all 3 values not `--`
  - StatsBar: all 4 values not `--`
  - VisitorCounter: shows > 0
  Capture full-page screenshot.
  Output: `Stats [PASS/FAIL] | Evidence: .sisyphus/evidence/final-stats.png`

- [x] F2. **Comment CRUD Verification** ✅ — PASS (submit/display/delete all working)
  Login → click image → submit comment → verify visible → delete comment → verify removed.
  Output: `Comments [PASS/FAIL] | Evidence: .sisyphus/evidence/final-comments.png`

- [x] F3. **Regression Check** ✅ — PASS (gallery, popup "locxl", TOP button all OK)
  Verify image gallery loads, click image opens popup (with correct uploader name), TOP button works.
  Output: `Regression [PASS/FAIL] | Evidence: .sisyphus/evidence/final-regression.png`

---

## Commit Strategy

- **1**: `fix(frontend): connect HomeView to WebSocket for live stats` — `frontend/src/views/HomeView.vue`
- **2**: `fix(frontend): fix StatsBar WebSocket field name mapping` — `frontend/src/components/StatsBar.vue`
- **3**: `fix(frontend): fix CommentSection route, auth, and response parsing` — `frontend/src/components/CommentSection.vue`

---

## Success Criteria

### Verification Commands
```bash
# Build: must pass
cd frontend && npm run build

# API: comment endpoint must be reachable
curl -s http://127.0.0.1:1323/api/images/1/comments | python3 -c "import json,sys; print(json.load(sys.stdin)['code'])"
# Expected: 200
```

### Final Checklist
- [ ] `npm run build` 通过（0 TS error）
- [ ] TopNotifyBar 显示真实数字（非 `--`）
- [ ] StatsBar 显示真实数字（非 `--`）
- [ ] VisitorCounter 显示 > 0
- [ ] 评论提交/列表/删除均正常
- [ ] 图片上传仍可用
- [ ] 图片浏览仍可用

---

## Commit Strategy

- **1**: `fix(frontend): connect HomeView to WebSocket for live stats` — HomeView.vue
- **2**: `fix(frontend): fix StatsBar WebSocket field name mapping` — StatsBar.vue
- **3**: `fix(frontend): fix CommentSection route, auth, and response parsing` — CommentSection.vue

---

## Success Criteria

### Verification Commands
```bash
# Stats: check WebSocket message handling in browser console
# Comments: curl POST /api/images/1/comments with valid JWT → 200
```

### Final Checklist
- [ ] TopNotifyBar shows real numbers (not `--`)
- [ ] StatsBar shows real numbers (not `--`)
- [ ] VisitorCounter shows > 0
- [ ] Comments: submit, list, delete all work
- [ ] Image upload still works
- [ ] Image gallery still displays
