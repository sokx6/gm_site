# 浮动按钮增加管理入口

## TL;DR
Admin 用户登入后，浮动按钮中"上传"下方多一个"管理"按钮，点击跳转 `/#/admin`。

## 改动

| 文件 | 改动 |
|------|------|
| `FloatingButtons.vue` | 加 `isAdmin` prop + `admin` emit + 按钮 + 样式 |
| `HomeView.vue` | 传 `:is-admin` + 加 `@admin` 路由跳转 |
