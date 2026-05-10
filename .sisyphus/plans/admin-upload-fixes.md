# Admin 图片显示修复 + 上传选项增强

## 3 个问题

| # | 问题 | 文件 | 根因 | 修复 |
|---|------|------|------|------|
| 1 | 缩略图不显示 | AdminImagesView.vue:195 | `img.thumbnail_url` 为空字符串 | 改用 `img.lsky_url` |
| 2 | UID:undefined | AdminImagesView.vue:203,231 | 读 `img.user_id`，后端返回 `uploaded_by` | 改用 `img.uploaded_by` |
| 3 | 上传无相册/标签 | HomeView.vue:218-242 | `onUploadFileChange` 只取 title | 加 prompt 取 tags，加 album 选择 |

## 改动

### AdminImagesView.vue
- 行 195-196: `v-if="img.thumbnail_url" :src="img.thumbnail_url"` → `v-if="img.lsky_url" :src="img.lsky_url"`
- 行 203: `userLabel(img.user_id)` → `userLabel(img.uploaded_by)`
- 行 231: `userLabel(img.user_id)` → `userLabel(img.uploaded_by)`
- 行 149: `userLabel(userId: number | undefined)` → `userLabel(userId: number)`

### HomeView.vue (onUploadFileChange)
- 在 title prompt 后加 tags prompt: `const tagsStr = window.prompt('请输入标签（逗号分隔，可选）')`
- 如果 tagsStr 非空，`formData.append('tags', tagsStr)`
- album: 当前已有 AlbumFilter 组件的数据源，可以在上传时加 album 选择。或者用 `window.prompt('请输入相册ID（可选，留空跳过）')` 简化
