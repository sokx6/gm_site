# 多弹窗 + 评论内嵌

## TL;DR

> **Quick Summary**: 将单图片弹窗改为多窗口模式（可同时打开多个），将 CommentSection 从页面底部移入每个弹窗内部。

> **Deliverables**:
> - 可同时打开多个图片弹窗，各自独立拖拽和关闭
> - 每个弹窗底部显示该图片的评论区

> **Estimated Effort**: Quick
> **Parallel Execution**: YES — 2 tasks in 1 wave
> **Critical Path**: Task 1 → Task 2 (sequential)

---

## 改动范围

| 文件 | 改动 | 量 |
|------|------|-----|
| `HomeView.vue` | `selectedImage` → 数组，`v-for` 多弹窗，移除 CommentSection | ~20 行 |
| `ImagePopup.vue` | 导入 CommentSection，嵌到 `.popup-body` 底部，接收 auth props | ~10 行 |

---

## TODOs

- [ ] 1. HomeView — 多弹窗支持 + 移除底部 CommentSection

  **What to do**:
  1. 将 `selectedImage` 从 `ref<GalleryImage | null>(null)` 改为 `ref<GalleryImage[]>([])`
  2. 移除 `showPopup`、`showComments` computed、`popupImage` computed
  3. `onCardClick(id)`: 查找图片是否已打开 → 已打开则聚焦该弹窗，否则 push 到数组
  4. 新增 `onPopupClose(imageId)`: 从数组中移除对应图片
  5. 模板：`<ImagePopup>` 从单 `v-if` 改为 `v-for="img in selectedImages"`，加 `:key="img.id"`
  6. 移除 `CommentSection` 的 import 和模板调用（不再在 HomeView 中渲染）
  7. ImagePopup 需要传入 auth props：`:is-logged-in="auth.isLoggedIn"` `:current-user-id="auth.user?.id ?? 0"` `:is-admin="auth.isAdmin"`

  **Must NOT do**:
  - 不修改 GalleryCard / GalleryGrid / 画廊逻辑
  - 不碰 WebSocket / 统计代码

- [ ] 2. ImagePopup — 内嵌 CommentSection

  **What to do**:
  1. Import `CommentSection` from `@/components/CommentSection.vue`
  2. 新增 props: `isLoggedIn?: boolean`, `currentUserId?: number`, `isAdmin?: boolean`
  3. 在 `.popup-body` 末尾（`.popup-meta` 之后）添加 `<CommentSection>`，传入 `image-id`、`is-logged-in`、`current-user-id`、`is-admin`
  4. CommentSection 需要包一层容器防止键盘事件穿透，加 `@click.stop` 防止拖拽误触

  **Must NOT do**:
  - 不修改 CommentSection.vue 内部代码
  - 不修改弹窗的拖拽/关闭逻辑

---

## Success Criteria
- [ ] 点击两张不同图片 → 屏幕出现两个可拖拽弹窗
- [ ] 每个弹窗底部有独立的评论输入区
- [ ] 关闭一个弹窗不影响另一个
- [ ] `npm run build` 通过
