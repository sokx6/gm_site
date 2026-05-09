<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  getAlbums,
  createAlbum,
  updateAlbum,
  deleteAlbum,
  type AlbumData,
} from '@/api/albums'

// ── State ──
const albums = ref<AlbumData[]>([])
const loading = ref(true)
const error = ref('')
const actionMsg = ref('')

// Create form
const showCreate = ref(false)
const newName = ref('')
const newDesc = ref('')
const creating = ref(false)

// Edit state
const editingId = ref<number | null>(null)
const editName = ref('')
const editDesc = ref('')
const savingEdit = ref(false)

// Delete confirm
const deletingId = ref<number | null>(null)
const deleting = ref(false)

// ── Fetch ──
async function fetchAlbums() {
  loading.value = true
  error.value = ''
  try {
    const res = await getAlbums()
    albums.value = res.data
  } catch (e: any) {
    error.value = e?.response?.data?.message || '加载相册列表失败'
  } finally {
    loading.value = false
  }
}

onMounted(fetchAlbums)

// ── Create ──
function toggleCreate() {
  showCreate.value = !showCreate.value
  newName.value = ''
  newDesc.value = ''
  error.value = ''
}

async function handleCreate() {
  if (!newName.value.trim()) {
    error.value = '相册名称不能为空'
    return
  }
  creating.value = true
  error.value = ''
  try {
    const res = await createAlbum({
      name: newName.value.trim(),
      description: newDesc.value.trim() || undefined,
    })
    albums.value.unshift(res.data)
    actionMsg.value = `已创建相册「${res.data.name}」`
    toggleCreate()
  } catch (e: any) {
    error.value = e?.response?.data?.message || '创建失败'
  } finally {
    creating.value = false
  }
}

// ── Edit ──
function startEdit(album: AlbumData) {
  editingId.value = album.id
  editName.value = album.name
  editDesc.value = album.description || ''
  error.value = ''
}

function cancelEdit() {
  editingId.value = null
  error.value = ''
}

async function saveEdit(id: number) {
  if (!editName.value.trim()) {
    error.value = '相册名称不能为空'
    return
  }
  savingEdit.value = true
  error.value = ''
  try {
    const res = await updateAlbum(id, {
      name: editName.value.trim(),
      description: editDesc.value.trim() || undefined,
    })
    const idx = albums.value.findIndex(a => a.id === id)
    if (idx >= 0) {
      albums.value[idx] = {
        ...albums.value[idx],
        name: res.data.name,
        description: res.data.description,
        updated_at: res.data.updated_at,
      }
    }
    actionMsg.value = `已更新相册「${res.data.name}」`
    editingId.value = null
  } catch (e: any) {
    error.value = e?.response?.data?.message || '保存失败'
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

async function handleDelete(id: number) {
  deleting.value = true
  try {
    const album = albums.value.find(a => a.id === id)
    await deleteAlbum(id)
    albums.value = albums.value.filter(a => a.id !== id)
    actionMsg.value = `已删除相册「${album?.name || id}」`
    deletingId.value = null
  } catch (e: any) {
    error.value = e?.response?.data?.message || '删除失败'
    deletingId.value = null
  } finally {
    deleting.value = false
  }
}

// ── Format ──
function formatDate(d: string): string {
  return new Date(d).toLocaleDateString('zh-CN')
}
</script>

<template>
  <div class="admin-sub-page">
    <!-- Header -->
    <div class="page-header">
      <router-link to="/admin" class="neon-link back-link">← 返回管理</router-link>
      <h1 class="page-title glow-text">📁 相册管理</h1>
      <p class="page-subtitle">共 {{ albums.length }} 个相册</p>
    </div>

    <!-- Error -->
    <div v-if="error" class="page-error">
      <span class="error-icon">⚠</span>
      {{ error }}
      <button class="dismiss-btn" @click="error = ''">×</button>
    </div>

    <!-- Success -->
    <div v-if="actionMsg" class="page-success">
      <span class="success-icon">✓</span>
      {{ actionMsg }}
      <button class="dismiss-btn dismiss-btn--green" @click="actionMsg = ''">×</button>
    </div>

    <!-- Create button -->
    <div class="toolbar">
      <button v-if="!showCreate" class="action-btn action-btn--create" @click="toggleCreate">
        + 新建相册
      </button>
    </div>

    <!-- Create form -->
    <div v-if="showCreate" class="create-card neon-box">
      <div class="scanlines" aria-hidden="true"></div>
      <h3 class="card-title">新建相册</h3>
      <div class="form-row">
        <div class="form-group">
          <label class="form-label">名称</label>
          <input
            v-model="newName"
            class="neon-input"
            placeholder="相册名称"
            :disabled="creating"
            @keyup.enter="handleCreate"
          />
        </div>
        <div class="form-group">
          <label class="form-label">描述（可选）</label>
          <input
            v-model="newDesc"
            class="neon-input"
            placeholder="相册描述"
            :disabled="creating"
            @keyup.enter="handleCreate"
          />
        </div>
      </div>
      <div class="form-actions">
        <button class="action-btn action-btn--save" :disabled="creating" @click="handleCreate">
          {{ creating ? '创建中...' : '创建' }}
        </button>
        <button class="action-btn action-btn--cancel" @click="toggleCreate">取消</button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="loading-box neon-box">
      <span class="loading-text">加载中...</span>
    </div>

    <!-- Album list -->
    <template v-else>
      <div v-if="albums.length === 0" class="empty-state">
        <span class="empty-icon">📂</span>
        <p>暂无相册，点击上方按钮创建第一个相册</p>
      </div>

      <div class="album-cards">
        <div
          v-for="album in albums"
          :key="album.id"
          class="album-card neon-box"
        >
          <div class="scanlines" aria-hidden="true"></div>

          <!-- Display mode -->
          <template v-if="editingId !== album.id">
            <div class="album-info">
              <div class="album-name">
                {{ album.name }}
                <span v-if="album.image_count > 0" class="album-count">
                  {{ album.image_count }} 张图片
                </span>
              </div>
              <p v-if="album.description" class="album-desc">{{ album.description }}</p>
              <p class="album-meta">创建于 {{ formatDate(album.created_at) }}</p>
            </div>
            <div class="album-actions">
              <button class="action-btn action-btn--edit" @click="startEdit(album)">编辑</button>
              <button class="action-btn action-btn--delete" @click="confirmDelete(album.id)">删除</button>
            </div>
          </template>

          <!-- Edit mode -->
          <template v-else>
            <div class="album-info">
              <div class="form-group edit-group">
                <label class="form-label">名称</label>
                <input v-model="editName" class="edit-input" placeholder="相册名称" />
              </div>
              <div class="form-group edit-group">
                <label class="form-label">描述</label>
                <input v-model="editDesc" class="edit-input" placeholder="相册描述（可选）" />
              </div>
            </div>
            <div class="album-actions">
              <button class="action-btn action-btn--save" :disabled="savingEdit" @click="saveEdit(album.id)">
                {{ savingEdit ? '保存中...' : '保存' }}
              </button>
              <button class="action-btn action-btn--cancel" @click="cancelEdit">取消</button>
            </div>
          </template>
        </div>
      </div>
    </template>

    <!-- Delete confirm modal -->
    <Teleport to="body">
      <div v-if="deletingId !== null" class="modal-overlay" @click.self="cancelDelete">
        <div class="modal-card neon-box--red">
          <div class="scanlines" aria-hidden="true"></div>
          <h3 class="modal-title">⚠ 确认删除</h3>
          <p class="modal-body">确定要删除这个相册吗？此操作不可撤销。</p>
          <div class="modal-actions">
            <button
              class="action-btn action-btn--delete"
              :disabled="deleting"
              @click="handleDelete(deletingId!)"
            >
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
  max-width: 800px;
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

/* ── Messages ─────────────────────────────────────── */
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

.page-success {
  padding: 10px 14px;
  background: rgba(0, 255, 0, 0.08);
  border: 1px solid var(--neon-green);
  color: var(--neon-green);
  font-family: var(--font-mono);
  font-size: 13px;
  box-shadow: 0 0 12px rgba(0, 255, 0, 0.2);
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
}

.error-icon, .success-icon { font-size: 16px; flex-shrink: 0; }

.dismiss-btn {
  margin-left: auto;
  background: none;
  border: none;
  color: var(--neon-red);
  font-size: 18px;
  cursor: pointer;
  font-family: var(--font-mono);
}

.dismiss-btn--green { color: var(--neon-green); }

@keyframes error-shake {
  0%, 100% { transform: translateX(0); }
  20%      { transform: translateX(-6px); }
  40%      { transform: translateX(6px); }
  60%      { transform: translateX(-4px); }
  80%      { transform: translateX(4px); }
}

/* ── Loading ──────────────────────────────────────── */
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

/* ── Toolbar ──────────────────────────────────────── */
.toolbar {
  margin-bottom: 20px;
}

/* ── Create card ──────────────────────────────────── */
.create-card {
  position: relative;
  background: #000;
  border: 2px double var(--neon-green);
  box-shadow: var(--glow-green), 0 0 30px rgba(0, 255, 0, 0.06);
  padding: 24px;
  margin-bottom: 24px;
}

.create-card .scanlines {
  position: absolute;
  inset: 0;
  pointer-events: none;
  background: repeating-linear-gradient(
    0deg,
    transparent,
    transparent 2px,
    rgba(0, 0, 0, 0.1) 2px,
    rgba(0, 0, 0, 0.1) 4px
  );
  z-index: 2;
}

.card-title {
  font-family: var(--font-display);
  font-size: 16px;
  color: var(--neon-green);
  margin: 0 0 16px;
  text-shadow: 0 0 8px var(--neon-green);
  position: relative;
  z-index: 3;
}

.form-row {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
  position: relative;
  z-index: 3;
}

.form-group {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-label {
  font-family: var(--font-mono);
  font-size: 12px;
  color: var(--neon-cyan);
  text-transform: uppercase;
  letter-spacing: 1px;
}

.neon-input {
  font-family: var(--font-mono);
  font-size: 14px;
  padding: 10px 12px;
  background: #0a0a0a;
  color: var(--neon-green);
  border: 1px solid var(--neon-green);
  box-shadow:
    inset 0 0 6px rgba(0, 255, 0, 0.1),
    0 0 8px rgba(0, 255, 0, 0.15);
  outline: none;
  transition: border-color 0.3s, box-shadow 0.3s;
}

.neon-input::placeholder {
  color: rgba(0, 255, 0, 0.2);
}

.neon-input:focus {
  border-color: var(--neon-cyan);
  box-shadow:
    inset 0 0 8px rgba(0, 255, 255, 0.15),
    0 0 12px var(--neon-cyan);
}

.neon-input:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.form-actions {
  display: flex;
  gap: 8px;
  position: relative;
  z-index: 3;
}

/* ── Empty state ──────────────────────────────────── */
.empty-state {
  padding: 60px 20px;
  text-align: center;
  color: #555;
  font-family: var(--font-mono);
  font-size: 14px;
}

.empty-icon {
  font-size: 48px;
  display: block;
  margin-bottom: 12px;
  opacity: 0.4;
}

/* ── Album cards ──────────────────────────────────── */
.album-cards {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.album-card {
  position: relative;
  background: #000;
  border: 2px double var(--neon-green);
  box-shadow: var(--glow-green), 0 0 30px rgba(0, 255, 0, 0.04);
  padding: 20px 24px;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
  transition: box-shadow 0.3s ease;
}

.album-card:hover {
  box-shadow:
    0 0 20px var(--neon-green),
    0 0 40px rgba(0, 255, 0, 0.1);
}

.scanlines {
  position: absolute;
  inset: 0;
  pointer-events: none;
  background: repeating-linear-gradient(
    0deg,
    transparent,
    transparent 2px,
    rgba(0, 0, 0, 0.1) 2px,
    rgba(0, 0, 0, 0.1) 4px
  );
  z-index: 2;
}

.album-info {
  position: relative;
  z-index: 3;
  flex: 1;
  min-width: 0;
}

.album-name {
  font-family: var(--font-display);
  font-size: 18px;
  color: #fff;
  margin-bottom: 6px;
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.album-count {
  font-family: var(--font-mono);
  font-size: 12px;
  color: var(--neon-cyan);
  border: 1px solid var(--neon-cyan);
  padding: 2px 8px;
  box-shadow: 0 0 6px rgba(0, 255, 255, 0.15);
}

.album-desc {
  font-family: var(--font-mono);
  font-size: 13px;
  color: #888;
  margin: 0 0 4px;
  line-height: 1.4;
}

.album-meta {
  font-family: var(--font-mono);
  font-size: 11px;
  color: #555;
  margin: 0;
}

.album-actions {
  position: relative;
  z-index: 3;
  display: flex;
  gap: 8px;
  flex-shrink: 0;
  align-items: flex-start;
}

/* ── Edit inline ──────────────────────────────────── */
.edit-group {
  margin-bottom: 8px;
}

.edit-input {
  font-family: var(--font-mono);
  font-size: 13px;
  padding: 6px 10px;
  background: #0a0a0a;
  color: var(--neon-green);
  border: 1px solid var(--neon-green);
  box-shadow:
    inset 0 0 6px rgba(0, 255, 0, 0.1),
    0 0 6px rgba(0, 255, 0, 0.15);
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

/* ── Action buttons ───────────────────────────────── */
.action-btn {
  font-family: var(--font-mono);
  font-size: 12px;
  padding: 6px 14px;
  background: #000;
  cursor: pointer;
  border: 1px solid;
  text-transform: uppercase;
  letter-spacing: 1px;
  transition: all 0.2s ease;
}

.action-btn--create {
  color: var(--neon-green);
  border-color: var(--neon-green);
  box-shadow: var(--glow-green);
  font-size: 13px;
  padding: 8px 20px;
  letter-spacing: 2px;
}

.action-btn--create:hover {
  background: rgba(0, 255, 0, 0.1);
  box-shadow:
    0 0 16px var(--neon-green),
    0 0 32px rgba(0, 255, 0, 0.2);
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

/* ── Modal ────────────────────────────────────────── */
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
  font-size: 13px;
  padding: 8px 16px;
}

/* ── Responsive ───────────────────────────────────── */
@media (max-width: 600px) {
  .form-row {
    flex-direction: column;
  }

  .album-card {
    flex-direction: column;
    padding: 16px;
  }

  .album-actions {
    width: 100%;
    justify-content: flex-end;
  }
}
</style>
