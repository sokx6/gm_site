# 上传界面优化

## TL;DR
替换 `window.prompt()` 弹窗 → 上传弹窗（UploadModal），集成标题、标签、相册选择（名称非ID）。

## 改为

| 新建/改造 | 文件 |
|-----------|------|
| 新建 | `components/UploadModal.vue` — 上传弹窗组件 |
| 改造 | `views/HomeView.vue` — 移除 prompt，集成 UploadModal |

## 流程
```
点击"上传"按钮 → 打开 UploadModal
├── 拖拽/点击选文件
├── 标题输入框（预填文件名）
├── 标签输入框（逗号分隔，可选）
├── 相册下拉选择（显示名称，传 album_id）
└── [确认上传] / [取消]
```

## UploadModal 设计
- Props: `visible`, `albums: AlbumData[]`
- Emits: `close`, `uploaded`
- 文件选择后预览缩略图
- 相册下拉从 API 获取列表（onMounted fetch）
- 上传时调用 `uploadImage(formData)`
- 上传成功后 emit `uploaded` + 关闭
