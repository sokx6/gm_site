<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { getImages, updateImage, deleteImage, type ImageData } from '@/api/images'
import { getAlbums, type AlbumData } from '@/api/albums'

// ── State ──
const images = ref<ImageData[]>([])
const albums = ref<AlbumData[]>([])
const loading = ref(true)
const error = ref('')
const currentPage = ref(1)
const totalPages = ref(0)
const total = ref(0)
const pageSize = 20

// Editing state
const editingId = ref<number | null>(null)
const editTitle = ref('')
const editTags = ref('')
const editAlbumId = ref<number | null>(null)
const savingEdit = ref(false)
const editError = ref('')

// Delete confirm
const deletingId = ref<number | null>(null)
const deleting = ref(false)

// ── Fetch ──
async function fetchImages() {
  loading.value = true
  error.value = ''
  try {
    const res = await getImages({ page: currentPage.value, page_size: pageSize })
    images.value = res.data.data
    total.value = res.data.total
    totalPages.value = res.data.total_pages
  } catch (e: any) {
    error.value = e?.response?.data?.message || '加载图片列表失败'
  } finally {
    loading.value = false
  }
}

async function fetchAlbums() {
  try {
    const res = await getAlbums()
    albums.value = res.data
  } catch {
    // albums not critical
  }
}

onMounted(() => {
  fetchImages()
  fetchAlbums()
})

// ── Pagination ──
function goPage(page: number) {
  if (page < 1 || page > totalPages.value) return
  currentPage.value = page
  fetchImages()
}

const pages = computed(() => {
  const p: number[] = []
  const start = Math.max(1, currentPage.value - 2)
  const end = Math.min(totalPages.value, currentPage.value + 2)
  for (let i = start; i <= end; i++) p.push(i)
  return p
})

// ── Album name lookup ──
function albumName(id: number | null): string {
  if (!id) return '—'
  const a = albums.value.find(x => x.id === id)
  return a ? a.name : `#${id}`
}

// ── Edit ──
function startEdit(img: ImageData) {
  editingId.value = img.id
  editTitle.value = img.title
  editTags.value = (img.tags ?? []).join(', ')
  editAlbumId.value = img.album_id
  editError.value = ''
}

function cancelEdit() {
  editingId.value = null
  editError.value = ''
}

async function saveEdit(id: number) {
  savingEdit.value = true
  editError.value = ''
  try {
    const tags = editTags.value
      .split(',')
      .map(t => t.trim())
      .filter(Boolean)
    await updateImage(id, {
      title: editTitle.value,
      tags,
      album_id: editAlbumId.value ?? undefined,
    })
    // Update local state
    const idx = images.value.findIndex(i => i.id === id)
    if (idx >= 0) {
      images.value[idx] = {
        ...images.value[idx],
        title: editTitle.value,
        tags,
        album_id: editAlbumId.value ?? images.value[idx].album_id,
      }
    }
    editingId.value = null
  } catch (e: any) {
    editError.value = e?.response?.data?.message || '保存失败'
  } finally {
    savingEdit.value = false
  }
}

// ── Delete ──
function confirmDelete(id: number) {
  deletingId.value = id
}

function cancelDelete() {
  deletingId.value = null
}

async function doDelete(id: number) {
  deleting.value = true
  try {
    await deleteImage(id)
    images.value = images.value.filter(i => i.id !== id)
    deletingId.value = null
  } catch (e: any) {
    error.value = e?.response?.data?.message || '删除失败'
    deletingId.value = null
  } finally {
    deleting.value = false
  }
}

// ── Format user ID ──
function userLabel(userId: number): string {
  return `UID:${userId}`
}
</script>

<template>
  <div class="admin-sub-page">
    <!-- Header -->
    <div class="page-header">
      <router-link to="/admin" class="neon-link back-link">← 返回管理</router-link>
      <h1 class="page-title glow-text">🖼️ 图片管理</h1>
      <p class="page-subtitle">共 {{ total }} 张图片</p>
    </div>

    <!-- Error -->
    <div v-if="error" class="page-error">
      <span class="error-icon">⚠</span>
      {{ error }}
      <button class="dismiss-btn" @click="error = ''">×</button>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="loading-box neon-box">
      <span class="loading-text">加载中...</span>
    </div>

    <!-- Table -->
    <div v-else class="table-wrap">
      <table class="neon-table">
        <thead>
          <tr>
            <th class="col-thumb">缩略图</th>
            <th class="col-title">标题</th>
            <th class="col-user">上传者</th>
            <th class="col-album">相册</th>
            <th class="col-tags">标签</th>
            <th class="col-date">创建时间</th>
            <th class="col-actions">操作</th>
          </tr>
        </thead>
        <tbody>
          <template v-for="img in images" :key="img.id">
            <!-- Normal row -->
            <tr v-if="editingId !== img.id">
              <td class="col-thumb">
                <img
                  v-if="img.thumbnail_url"
                  :src="img.thumbnail_url"
                  :alt="img.title"
                  class="thumb-img"
                />
                <span v-else class="thumb-placeholder">—</span>
              </td>
              <td class="col-title">{{ img.title || '无标题' }}</td>
              <td class="col-user">{{ userLabel(img.user_id) }}</td>
              <td class="col-album">{{ albumName(img.album_id) }}</td>
              <td class="col-tags">
                <span v-if="img.tags?.length" class="tag-list">
                  <span v-for="t in img.tags" :key="t" class="tag-chip">{{ t }}</span>
                </span>
                <span v-else class="text-muted">—</span>
              </td>
              <td class="col-date">{{ new Date(img.created_at).toLocaleDateString('zh-CN') }}</td>
              <td class="col-actions">
                <button class="action-btn action-btn--edit" @click="startEdit(img)">编辑</button>
                <button class="action-btn action-btn--delete" @click="confirmDelete(img.id)">删除</button>
              </td>
            </tr>

            <!-- Edit row -->
            <tr v-else class="edit-row">
              <td class="col-thumb">
                <img
                  v-if="img.thumbnail_url"
                  :src="img.thumbnail_url"
                  :alt="img.title"
                  class="thumb-img thumb-img--dim"
                />
              </td>
              <td class="col-title">
                <input v-model="editTitle" class="edit-input" placeholder="标题" />
              </td>
              <td class="col-user">{{ userLabel(img.user_id) }}</td>
              <td class="col-album">
                <select v-model.number="editAlbumId" class="edit-select">
                  <option :value="null">无相册</option>
                  <option v-for="a in albums" :key="a.id" :value="a.id">{{ a.name }}</option>
                </select>
              </td>
              <td class="col-tags">
                <input v-model="editTags" class="edit-input" placeholder="标签,逗号分隔" />
              </td>
              <td class="col-date">{{ new Date(img.created_at).toLocaleDateString('zh-CN') }}</td>
              <td class="col-actions">
                <button class="action-btn action-btn--save" :disabled="savingEdit" @click="saveEdit(img.id)">
                  {{ savingEdit ? '保存中...' : '保存' }}
                </button>
                <button class="action-btn action-btn--cancel" @click="cancelEdit">取消</button>
                <span v-if="editError" class="edit-error-inline">{{ editError }}</span>
              </td>
            </tr>
          </template>

          <!-- Empty -->
          <tr v-if="images.length === 0">
            <td colspan="7" class="empty-cell">暂无图片数据</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination -->
    <div v-if="!loading && totalPages > 1" class="pagination">
      <button :disabled="currentPage === 1" @click="goPage(1)">|«</button>
      <button :disabled="currentPage === 1" @click="goPage(currentPage - 1)">«</button>
      <button
        v-for="p in pages"
        :key="p"
        :class="{ active: p === currentPage }"
        @click="goPage(p)"
      >
        {{ p }}
      </button>
      <button :disabled="currentPage === totalPages" @click="goPage(currentPage + 1)">»</button>
      <button :disabled="currentPage === totalPages" @click="goPage(totalPages)">»|</button>
    </div>

    <!-- Delete confirm modal -->
    <Teleport to="body">
      <div v-if="deletingId !== null" class="modal-overlay" @click.self="cancelDelete">
        <div class="modal-card neon-box--red">
          <div class="scanlines" aria-hidden="true"></div>
          <h3 class="modal-title">⚠ 确认删除</h3>
          <p class="modal-body">确定要删除这张图片吗？此操作不可撤销。</p>
          <div class="modal-actions">
            <button class="action-btn action-btn--delete" :disabled="deleting" @click="doDelete(deletingId!)">
              {{ deleting ? '删除中...' : '确认删除' }}
            </button>
            <button class="action-btn action-btn--cancel" @click="cancelDelete">取消</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
/* ── Page layout ───────────────────────────────────── */
.admin-sub-page {
  min-height: 100vh;
  background: var(--bg-primary);
  padding: 24px 20px 60px;
  max-width: 1400px;
  margin: 0 auto;
}

/* ── Header ───────────────────────────────────────── */
.page-header {
  margin-bottom: 24px;
}

.back-link {
  display: inline-block;
  margin-bottom: 12px;
}

.page-title {
  font-family: var(--font-display);
  font-size: 28px;
  margin: 0 0 4px;
  color: #fff;
  text-shadow:
    0 0 10px var(--neon-pink),
    0 0 20px var(--neon-cyan),
    3px 3px 0 var(--neon-red),
    -3px -3px 0 var(--neon-blue);
  letter-spacing: 4px;
}

.page-subtitle {
  font-family: var(--font-mono);
  font-size: 13px;
  color: var(--neon-cyan);
  margin: 0;
}

.neon-link {
  font-family: var(--font-mono);
  font-size: 14px;
  color: var(--neon-cyan);
  text-decoration: none;
  text-shadow: 0 0 8px var(--neon-cyan);
  transition: color 0.3s, text-shadow 0.3s;
}

.neon-link:hover {
  color: var(--neon-yellow);
  text-shadow: 0 0 12px var(--neon-yellow);
}

/* ── Error / Loading ──────────────────────────────── */
.page-error {
  padding: 10px 14px;
  background: rgba(255, 0, 0, 0.12);
  border: 1px solid var(--neon-red);
  color: var(--neon-red);
  font-family: var(--font-mono);
  font-size: 13px;
  box-shadow: 0 0 12px rgba(255, 0, 0, 0.25);
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  animation: error-shake 0.4s ease;
}

.error-icon { font-size: 16px; flex-shrink: 0; }

.dismiss-btn {
  margin-left: auto;
  background: none;
  border: none;
  color: var(--neon-red);
  font-size: 18px;
  cursor: pointer;
  font-family: var(--font-mono);
}

@keyframes error-shake {
  0%, 100% { transform: translateX(0); }
  20%      { transform: translateX(-6px); }
  40%      { transform: translateX(6px); }
  60%      { transform: translateX(-4px); }
  80%      { transform: translateX(4px); }
}

.loading-box {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 40px 20px;
}

.loading-text {
  font-family: var(--font-mono);
  font-size: 14px;
  color: var(--neon-green);
  animation: blink 1.5s ease-in-out infinite;
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50%      { opacity: 0; }
}

/* ── Table ────────────────────────────────────────── */
.table-wrap {
  overflow-x: auto;
}

.neon-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  font-family: var(--font-mono);
  font-size: 13px;
  background: #000;
  border: 2px solid var(--neon-green);
  box-shadow: var(--glow-green), inset 0 0 20px rgba(0, 255, 0, 0.03);
}

.neon-table thead th {
  background: rgba(0, 255, 0, 0.06);
  color: var(--neon-green);
  font-weight: 400;
  text-transform: uppercase;
  letter-spacing: 2px;
  padding: 12px 10px;
  border-bottom: 2px solid var(--neon-green);
  text-align: left;
  white-space: nowrap;
  text-shadow: 0 0 6px var(--neon-green);
}

.neon-table tbody td {
  padding: 10px;
  border-bottom: 1px solid rgba(0, 255, 0, 0.12);
  color: #ccc;
  vertical-align: middle;
}

.neon-table tbody tr:hover {
  background: rgba(0, 255, 0, 0.03);
}

.neon-table tbody tr:last-child td {
  border-bottom: none;
}

/* Columns */
.col-thumb { width: 80px; text-align: center; }
.col-title { min-width: 140px; }
.col-user { width: 80px; }
.col-album { width: 100px; }
.col-tags { min-width: 120px; }
.col-date { width: 110px; white-space: nowrap; }
.col-actions { width: 150px; white-space: nowrap; }

/* Thumbnail */
.thumb-img {
  width: 64px;
  height: 64px;
  object-fit: cover;
  border: 1px solid rgba(0, 255, 0, 0.2);
  border-radius: 2px;
}

.thumb-img--dim {
  opacity: 0.5;
}

.thumb-placeholder {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
  border: 1px dashed rgba(0, 255, 0, 0.15);
  color: #555;
  font-size: 11px;
}

/* Tags */
.tag-list {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.tag-chip {
  display: inline-block;
  padding: 2px 6px;
  background: rgba(0, 255, 0, 0.08);
  border: 1px solid rgba(0, 255, 0, 0.2);
  font-size: 11px;
  color: var(--neon-green);
  border-radius: 2px;
}

.text-muted {
  color: #555;
}

/* ── Edit row ─────────────────────────────────────── */
.edit-row {
  background: rgba(0, 255, 0, 0.04) !important;
}

.edit-row td {
  border-bottom-color: var(--neon-cyan) !important;
}

.edit-input {
  font-family: var(--font-mono);
  font-size: 13px;
  padding: 6px 8px;
  background: #0a0a0a;
  color: var(--neon-green);
  border: 1px solid var(--neon-green);
  box-shadow: inset 0 0 6px rgba(0, 255, 0, 0.1), 0 0 6px rgba(0, 255, 0, 0.15);
  outline: none;
  width: 100%;
  transition: border-color 0.3s, box-shadow 0.3s;
}

.edit-input:focus {
  border-color: var(--neon-cyan);
  box-shadow:
    inset 0 0 8px rgba(0, 255, 255, 0.15),
    0 0 12px var(--neon-cyan);
}

.edit-select {
  font-family: var(--font-mono);
  font-size: 13px;
  padding: 6px 8px;
  background: #0a0a0a;
  color: var(--neon-green);
  border: 1px solid var(--neon-green);
  box-shadow: 0 0 6px rgba(0, 255, 0, 0.15);
  outline: none;
  width: 100%;
  cursor: pointer;
}

.edit-select option {
  background: #0a0a0a;
  color: var(--neon-green);
}

.edit-error-inline {
  display: block;
  font-size: 11px;
  color: var(--neon-red);
  margin-top: 4px;
}

/* ── Action buttons ───────────────────────────────── */
.action-btn {
  font-family: var(--font-mono);
  font-size: 12px;
  padding: 5px 12px;
  background: #000;
  cursor: pointer;
  border: 1px solid;
  text-transform: uppercase;
  letter-spacing: 1px;
  transition: all 0.2s ease;
  margin-right: 4px;
}

.action-btn--edit {
  color: var(--neon-cyan);
  border-color: var(--neon-cyan);
  box-shadow: 0 0 6px rgba(0, 255, 255, 0.2);
}

.action-btn--edit:hover {
  background: rgba(0, 255, 255, 0.1);
  box-shadow: 0 0 12px var(--neon-cyan);
}

.action-btn--delete {
  color: var(--neon-red);
  border-color: var(--neon-red);
  box-shadow: 0 0 6px rgba(255, 0, 0, 0.2);
}

.action-btn--delete:hover {
  background: rgba(255, 0, 0, 0.1);
  box-shadow: 0 0 12px var(--neon-red);
}

.action-btn--save {
  color: var(--neon-green);
  border-color: var(--neon-green);
  box-shadow: 0 0 6px rgba(0, 255, 0, 0.2);
}

.action-btn--save:hover:not(:disabled) {
  background: rgba(0, 255, 0, 0.1);
  box-shadow: 0 0 12px var(--neon-green);
}

.action-btn--cancel {
  color: #888;
  border-color: #444;
}

.action-btn--cancel:hover {
  color: #fff;
  border-color: #888;
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* ── Pagination ───────────────────────────────────── */
.pagination {
  display: flex;
  justify-content: center;
  gap: 4px;
  margin-top: 24px;
}

.pagination button {
  font-family: var(--font-mono);
  font-size: 13px;
  padding: 6px 12px;
  background: #000;
  color: var(--neon-cyan);
  border: 1px solid var(--neon-cyan);
  box-shadow: 0 0 6px rgba(0, 255, 255, 0.15);
  cursor: pointer;
  transition: all 0.2s ease;
}

.pagination button:hover:not(:disabled) {
  background: rgba(0, 255, 255, 0.1);
  box-shadow: 0 0 12px var(--neon-cyan);
}

.pagination button.active {
  background: var(--neon-cyan);
  color: #000;
  font-weight: bold;
  box-shadow: 0 0 16px var(--neon-cyan);
}

.pagination button:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

/* ── Empty ────────────────────────────────────────── */
.empty-cell {
  text-align: center;
  padding: 40px 20px !important;
  color: #555;
  font-size: 14px;
}

/* ── Modal (confirm delete) ───────────────────────── */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.75);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-card {
  position: relative;
  background: #000;
  border: 3px double var(--neon-red);
  box-shadow:
    var(--glow-red),
    0 0 60px rgba(255, 0, 0, 0.15);
  padding: 32px 28px;
  max-width: 420px;
  width: 90%;
}

.modal-card .scanlines {
  position: absolute;
  inset: 0;
  pointer-events: none;
  background: repeating-linear-gradient(
    0deg,
    transparent,
    transparent 2px,
    rgba(0, 0, 0, 0.12) 2px,
    rgba(0, 0, 0, 0.12) 4px
  );
  z-index: 2;
}

.modal-title {
  font-family: var(--font-display);
  font-size: 20px;
  color: var(--neon-red);
  margin: 0 0 12px;
  text-shadow: 0 0 8px var(--neon-red);
  position: relative;
  z-index: 3;
}

.modal-body {
  font-family: var(--font-mono);
  font-size: 14px;
  color: #aaa;
  margin: 0 0 20px;
  position: relative;
  z-index: 3;
}

.modal-actions {
  display: flex;
  gap: 8px;
  position: relative;
  z-index: 3;
}

.modal-actions .action-btn {
  margin-right: 0;
  font-size: 13px;
  padding: 8px 16px;
}
</style>
