# AdminImagesView 图片列表不显示

## 根因
后端 API `/api/images` 返回格式：
```json
{"code":200,"data":{"list":[...images],"total":N,"page":N,"limit":N}}
```
前端第 34 行取 `res.data.data`（undefined），应该取 `res.data.list`。同时 `total_pages` 后端没返回，需前端计算。

## 改动 (2 行)
| 行 | 当前 | 改为 |
|----|------|------|
| 34 | `images.value = res.data.data` | `images.value = res.data.list` |
| 36 | `totalPages.value = res.data.total_pages` | `totalPages.value = Math.ceil(res.data.total / pageSize)` |
