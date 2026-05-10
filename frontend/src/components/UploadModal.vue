<script setup lang="ts">
import { ref, watch, onUnmounted } from 'vue'
import type { AlbumData } from '@/api/albums'
import { uploadImage } from '@/api/images'

// ── Props ──────────────────────────────────────────────────
const props = defineProps<{
  visible: boolean
  albums: AlbumData[]
}>()

// ── Emits ──────────────────────────────────────────────────
const emit = defineEmits<{
  close: []
  uploaded: []
}>()

// ── State ──────────────────────────────────────────────────
const file = ref<File | null>(null)
const previewUrl = ref<string | null>(null)
const title = ref('')
const tags = ref('')
const albumId = ref('')
const privacy = ref('public')
const uploading = ref(false)
const dragOver = ref(false)

// ── File handling ──────────────────────────────────────────
function revokePreview() {
  if (previewUrl.value) {
    URL.revokeObjectURL(previewUrl.value)
    previewUrl.value = null
  }
}

function setFile(f: File) {
  revokePreview()
  file.value = f
  previewUrl.value = URL.createObjectURL(f)
  // Pre-fill title from filename (strip extension)
  title.value = f.name.replace(/\.[^.]+$/, '')
}

function onFileChange(e: Event) {
  const input = e.target as HTMLInputElement
  const f = input.files?.[0]
  if (f) setFile(f)
}

function onDrop(e: DragEvent) {
  dragOver.value = false
  const f = e.dataTransfer?.files?.[0]
  if (f) setFile(f)
}

function onDragOver(e: DragEvent) {
  e.preventDefault()
  dragOver.value = true
}

function onDragLeave() {
  dragOver.value = false
}

function removeFile() {
  revokePreview()
  file.value = null
  title.value = ''
}

// ── Upload ─────────────────────────────────────────────────
async function doUpload() {
  if (!file.value || !title.value.trim()) return

  uploading.value = true
  try {
    const formData = new FormData()
    formData.append('file', file.value)
    formData.append('title', title.value.trim())
    if (tags.value.trim()) {
      formData.append('tags', tags.value.trim())
    }
    if (albumId.value) {
      formData.append('album_id', albumId.value)
    }
    if (privacy.value !== 'public') {
      formData.append('privacy', privacy.value)
    }

    await uploadImage(formData)
    emit('uploaded')
    resetAndClose()
  } catch (error: any) {
    const message = error?.response?.data?.message || '上传失败'
    window.alert(message)
  } finally {
    uploading.value = false
  }
}

// ── Close / Reset ──────────────────────────────────────────
function resetAndClose() {
  revokePreview()
  file.value = null
  title.value = ''
  tags.value = ''
  albumId.value = ''
  privacy.value = 'public'
  dragOver.value = false
  emit('close')
}

function onOverlayClick() {
  resetAndClose()
}

// Cleanup preview URL when modal closes
watch(() => props.visible, (vis) => {
  if (!vis) revokePreview()
})

// Revoke object URL on unmount to prevent memory leak
onUnmounted(() => {
  if (previewUrl.value) {
    URL.revokeObjectURL(previewUrl.value)
    previewUrl.value = null
  }
})
</script>

<template>
  <Teleport to="body">
    <div
      v-if="visible"
      class="upload-overlay"
      @click.self="onOverlayClick"
      @keydown.escape="resetAndClose"
    >
      <div class="upload-modal" :class="{ 'upload-modal--dragover': dragOver }">
        <!-- Header -->
        <h2 class="modal-title">上传图片</h2>

        <!-- Drop zone -->
        <div
          class="drop-zone"
          :class="{
            'drop-zone--active': dragOver,
            'drop-zone--has-file': previewUrl
          }"
          @drop.prevent="onDrop"
          @dragover="onDragOver"
          @dragleave="onDragLeave"
        >
          <template v-if="previewUrl">
            <img :src="previewUrl" class="preview-thumb" alt="Preview" />
            <button type="button" class="remove-preview" @click="removeFile">
              ✕
            </button>
          </template>
          <template v-else>
            <div class="drop-placeholder">
              <span class="drop-icon">{{ dragOver ? '⬇' : '📁' }}</span>
              <span class="drop-text">拖拽文件到这里 或点击选择</span>
              <span class="drop-hint">支持 JPG / PNG / GIF / WebP</span>
            </div>
            <input
              type="file"
              accept="image/*"
              class="file-input"
              @change="onFileChange"
            />
          </template>
        </div>

        <!-- Title -->
        <label class="field-label">标题</label>
        <input
          v-model="title"
          type="text"
          class="cyber-input"
          placeholder="输入图片标题"
        />

        <!-- Tags -->
        <label class="field-label">标签</label>
        <input
          v-model="tags"
          type="text"
          class="cyber-input"
          placeholder="标签（逗号分隔，可选）"
        />

        <!-- Album dropdown -->
        <label class="field-label">相册</label>
        <select v-model="albumId" class="cyber-select">
          <option value="">无相册</option>
          <option
            v-for="a in props.albums"
            :key="a.id"
            :value="a.id"
          >
            {{ a.name }}
          </option>
        </select>

        <!-- Privacy -->
        <label class="field-label">权限</label>
        <select v-model="privacy" class="cyber-select">
          <option value="public">公开</option>
          <option value="friends">好友可见</option>
          <option value="private">私密</option>
        </select>

        <!-- Actions -->
        <div class="modal-actions">
          <button
            type="button"
            class="btn-cancel"
            :disabled="uploading"
            @click="resetAndClose"
          >
            取消
          </button>
          <button
            type="button"
            class="btn-upload"
            :disabled="!file || !title.trim() || uploading"
            @click="doUpload"
          >
            <span v-if="uploading" class="spinner"></span>
            {{ uploading ? '上传中...' : '上传' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
/* ── Overlay ─────────────────────────────────────────────── */
.upload-overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(4px);
}

/* ── Modal box ───────────────────────────────────────────── */
.upload-modal {
  width: 460px;
  max-width: 95vw;
  max-height: 90vh;
  overflow-y: auto;
  background: var(--bg-primary);
  border: 2px solid var(--neon-green);
  box-shadow:
    var(--glow-green),
    0 0 60px rgba(0, 255, 0, 0.15),
    inset 0 0 30px rgba(0, 255, 0, 0.05);
  padding: 28px 24px 24px;
  animation: modal-pop 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.upload-modal--dragover {
  border-color: var(--neon-cyan);
  box-shadow:
    0 0 15px var(--neon-cyan),
    0 0 60px rgba(0, 255, 255, 0.2),
    inset 0 0 30px rgba(0, 255, 255, 0.08);
}

@keyframes modal-pop {
  0% {
    opacity: 0;
    transform: scale(0.9) translateY(20px);
  }
  100% {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

/* ── Title ───────────────────────────────────────────────── */
.modal-title {
  font-family: var(--font-mono);
  font-size: 22px;
  color: var(--neon-green);
  text-align: center;
  text-shadow: 0 0 10px var(--neon-green), 0 0 30px rgba(0, 255, 0, 0.5);
  margin: 0 0 20px;
  letter-spacing: 6px;
  text-transform: uppercase;
  animation: text-flicker 3s infinite;
}

@keyframes text-flicker {
  0%, 19%, 21%, 23%, 25%, 54%, 56%, 100% {
    opacity: 1;
  }
  20%, 24%, 55% {
    opacity: 0.6;
  }
}

/* ── Drop zone ───────────────────────────────────────────── */
.drop-zone {
  position: relative;
  border: 2px dashed #333;
  border-radius: 8px;
  min-height: 140px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition:
    border-color 0.2s,
    box-shadow 0.2s;
  overflow: hidden;
  margin-bottom: 16px;
}

.drop-zone:hover,
.drop-zone--active {
  border-color: var(--neon-cyan);
  box-shadow: 0 0 12px rgba(0, 255, 255, 0.2);
}

.drop-zone--has-file {
  border-color: var(--neon-green);
  box-shadow: var(--glow-green);
  cursor: default;
}

.file-input {
  position: absolute;
  inset: 0;
  opacity: 0;
  cursor: pointer;
}

/* ── Drop placeholder ────────────────────────────────────── */
.drop-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  pointer-events: none;
  padding: 20px;
}

.drop-icon {
  font-size: 36px;
  opacity: 0.7;
}

.drop-text {
  font-family: var(--font-mono);
  font-size: 14px;
  color: #888;
}

.drop-hint {
  font-family: var(--font-mono);
  font-size: 11px;
  color: #555;
}

/* ── Preview thumbnail ───────────────────────────────────── */
.preview-thumb {
  max-width: 100%;
  max-height: 260px;
  object-fit: contain;
  pointer-events: none;
}

.remove-preview {
  position: absolute;
  top: 6px;
  right: 6px;
  width: 28px;
  height: 28px;
  border: 1px solid var(--neon-pink);
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.85);
  color: var(--neon-pink);
  font-size: 14px;
  font-family: var(--font-mono);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition:
    box-shadow 0.2s,
    transform 0.2s;
}

.remove-preview:hover {
  box-shadow: 0 0 8px var(--neon-pink);
  transform: scale(1.1);
}

/* ── Labels ──────────────────────────────────────────────── */
.field-label {
  display: block;
  font-family: var(--font-mono);
  font-size: 12px;
  color: var(--neon-cyan);
  text-transform: uppercase;
  letter-spacing: 2px;
  margin-bottom: 4px;
}

/* ── Inputs ──────────────────────────────────────────────── */
.cyber-input,
.cyber-select {
  width: 100%;
  padding: 8px 12px;
  font-family: var(--font-mono);
  font-size: 14px;
  color: #fff;
  background: #111;
  border: 1px solid #333;
  border-radius: 4px;
  outline: none;
  transition:
    border-color 0.2s,
    box-shadow 0.2s;
  margin-bottom: 12px;
}

.cyber-input:focus,
.cyber-select:focus {
  border-color: var(--neon-green);
  box-shadow: var(--glow-green);
}

.cyber-input::placeholder {
  color: #555;
  font-family: var(--font-mono);
}

.cyber-select {
  appearance: none;
  cursor: pointer;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='8'%3E%3Cpath d='M1 1l5 5 5-5' stroke='%230f0' fill='none' stroke-width='1.5'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
  padding-right: 32px;
}

.cyber-select option {
  background: #0a0a0a;
  color: #fff;
}

/* ── Actions ─────────────────────────────────────────────── */
.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid #222;
}

/* ── Buttons ─────────────────────────────────────────────── */
.btn-cancel,
.btn-upload {
  font-family: var(--font-mono);
  font-size: 13px;
  padding: 8px 24px;
  border-radius: 4px;
  cursor: pointer;
  text-transform: uppercase;
  letter-spacing: 2px;
  transition:
    background 0.2s,
    box-shadow 0.2s,
    opacity 0.2s;
  outline: none;
}

.btn-cancel {
  background: transparent;
  color: #555;
  border: 1px solid #333;
}

.btn-cancel:hover:not(:disabled) {
  color: #aaa;
  border-color: #555;
}

.btn-upload {
  background: rgba(0, 255, 0, 0.1);
  color: var(--neon-green);
  border: 1px solid var(--neon-green);
  box-shadow: 0 0 8px rgba(0, 255, 0, 0.2);
}

.btn-upload:hover:not(:disabled) {
  background: rgba(0, 255, 0, 0.2);
  box-shadow: var(--glow-green);
}

.btn-upload:active:not(:disabled) {
  background: rgba(0, 255, 0, 0.3);
}

.btn-cancel:disabled,
.btn-upload:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

/* ── Spinner ─────────────────────────────────────────────── */
.spinner {
  display: inline-block;
  width: 12px;
  height: 12px;
  border: 2px solid rgba(0, 255, 0, 0.3);
  border-top-color: var(--neon-green);
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
  vertical-align: middle;
  margin-right: 6px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* ── Scrollbar (cyberpunk) ───────────────────────────────── */
.upload-modal::-webkit-scrollbar {
  width: 6px;
}

.upload-modal::-webkit-scrollbar-track {
  background: var(--bg-primary);
}

.upload-modal::-webkit-scrollbar-thumb {
  background: #333;
  border-radius: 3px;
}

.upload-modal::-webkit-scrollbar-thumb:hover {
  background: var(--neon-green);
}
</style>
