<script setup lang="ts">
import { ref, watch, computed, onUnmounted } from 'vue'

// ── Image data model (mirrors GalleryCard.vue) ────────
interface Image {
  id: number
  title: string
  lsky_url: string
  tags: string[]
  uploaded_by: number
  created_at: string
}

// ── Props & Emits ─────────────────────────────────────
const props = defineProps<{
  image: Image
  visible: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

// ── Internal state ────────────────────────────────────
const showPopup = ref(props.visible)
const closing = ref(false)
const opening = ref(false)
const firstOpen = ref(true)

let reappearTimer: ReturnType<typeof setTimeout> | null = null
let imgLoadError = ref(false)

// ── Position / drag state ─────────────────────────────
const popupLeft = ref(0)
const popupTop = ref(0)

const isDragging = ref(false)
const dragStartX = ref(0)
const dragStartY = ref(0)
const popupStartLeft = ref(0)
const popupStartTop = ref(0)

// ── Watch visible prop ────────────────────────────────
watch(
  () => props.visible,
  (val) => {
    if (val) {
      if (firstOpen.value) {
        randomPosition()
        firstOpen.value = false
      }
      showPopup.value = true
      closing.value = false
      // Trigger open animation on next tick so v-show has taken effect
      requestAnimationFrame(() => {
        opening.value = true
        setTimeout(() => {
          opening.value = false
        }, 500)
      })
    } else {
      closing.value = true
      setTimeout(() => {
        showPopup.value = false
        closing.value = false
        scheduleReappear()
      }, 300)
    }
  },
  { immediate: true }
)

// Cleanup timer on unmount
onUnmounted(() => {
  if (reappearTimer) clearTimeout(reappearTimer)
  document.removeEventListener('mousemove', onMouseMove)
  document.removeEventListener('mouseup', onMouseUp)
})

// ── Format helpers ────────────────────────────────────
function formatDate(iso: string): string {
  try {
    const d = new Date(iso)
    return d.toLocaleDateString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    })
  } catch {
    return iso
  }
}

const tagColors = [
  'var(--neon-pink)',
  'var(--neon-cyan)',
  'var(--neon-green)',
  'var(--neon-yellow)',
  'var(--neon-orange)',
  'var(--neon-purple)',
]

function tagColor(idx: number): string {
  return tagColors[idx % tagColors.length]
}

// ── Random position ───────────────────────────────────
function randomPosition(): void {
  // Estimate popup size (will be clamped after render too)
  const estimatedW = Math.min(window.innerWidth * 0.8, 600)
  const estimatedH = Math.min(window.innerHeight * 0.7, 500)
  const padding = 40

  popupLeft.value =
    padding + Math.random() * (window.innerWidth - estimatedW - padding * 2)
  popupTop.value =
    padding + Math.random() * (window.innerHeight - estimatedH - padding * 2)
}

// ── Clamp position to viewport ────────────────────────
function clampPosition(): void {
  // Read actual popup dimensions from DOM
  const el = document.querySelector('.image-popup') as HTMLElement | null
  if (!el) return

  const pw = el.offsetWidth
  const ph = el.offsetHeight
  const maxX = window.innerWidth - pw - 10
  const maxY = window.innerHeight - ph - 10

  popupLeft.value = Math.max(10, Math.min(popupLeft.value, maxX))
  popupTop.value = Math.max(10, Math.min(popupTop.value, maxY))
}

// ── Drag handlers ─────────────────────────────────────
function onMouseDown(e: MouseEvent): void {
  // Only drag from header
  const target = e.target as HTMLElement
  if (!target.closest('.popup-header')) return

  isDragging.value = true
  dragStartX.value = e.clientX
  dragStartY.value = e.clientY
  popupStartLeft.value = popupLeft.value
  popupStartTop.value = popupTop.value

  document.addEventListener('mousemove', onMouseMove)
  document.addEventListener('mouseup', onMouseUp)
}

function onMouseMove(e: MouseEvent): void {
  if (!isDragging.value) return

  const dx = e.clientX - dragStartX.value
  const dy = e.clientY - dragStartY.value

  popupLeft.value = popupStartLeft.value + dx
  popupTop.value = popupStartTop.value + dy

  // Clamp after moving
  clampPosition()
}

function onMouseUp(): void {
  isDragging.value = false
  document.removeEventListener('mousemove', onMouseMove)
  document.removeEventListener('mouseup', onMouseUp)
}

// ── Close & auto-reappear ─────────────────────────────
function handleClose(): void {
  closing.value = true
  setTimeout(() => {
    showPopup.value = false
    closing.value = false
    emit('close')
    scheduleReappear()
  }, 300)
}

function scheduleReappear(): void {
  if (reappearTimer) clearTimeout(reappearTimer)
  const delay = 5000 + Math.random() * 8000 // 5-13s
  reappearTimer = setTimeout(() => {
    showPopup.value = true
    requestAnimationFrame(() => {
      opening.value = true
      setTimeout(() => {
        opening.value = false
      }, 500)
    })
  }, delay)
}

// ── Computed popup style ──────────────────────────────
const popupStyle = computed(() => ({
  left: `${popupLeft.value}px`,
  top: `${popupTop.value}px`,
}))
</script>

<template>
  <div
    v-show="showPopup"
    class="image-popup"
    :class="{
      'image-popup--closing': closing,
      'image-popup--opening': opening,
    }"
    :style="popupStyle"
    @mousedown="onMouseDown"
  >
    <!-- Header (drag handle) -->
    <div class="popup-header">
      {{ props.image.title }}
    </div>

    <!-- Close button -->
    <button class="popup-close" @click="handleClose" title="关闭">&times;</button>

    <!-- Body -->
    <div class="popup-body">
      <div class="popup-img-wrap">
        <img
          v-if="!imgLoadError"
          :src="props.image.lsky_url"
          :alt="props.image.title"
          class="popup-img"
          @error="imgLoadError = true"
        />
        <div v-else class="popup-img-fallback">
          ⚠ 图片加载失败
        </div>
      </div>

      <div class="popup-meta">
        <h2 class="popup-title">{{ props.image.title }}</h2>

        <!-- Tags -->
        <div v-if="props.image.tags.length" class="popup-tags">
          <span
            v-for="(tag, i) in props.image.tags"
            :key="tag"
            class="popup-tag"
            :style="{ color: tagColor(i), borderColor: tagColor(i) }"
          >
            {{ tag }}
          </span>
        </div>

        <div class="popup-info">
          <span class="popup-uploader">
            👤 上传者: User#{{ props.image.uploaded_by }}
          </span>
          <span class="popup-date">
            📅 {{ formatDate(props.image.created_at) }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* ── Root popup ─────────────────────────────────────── */
.image-popup {
  position: fixed;
  z-index: 999;
  width: min(560px, 90vw);
  background: linear-gradient(180deg, #1a0000, #0a0a0a);
  border: 3px ridge var(--neon-yellow);
  border-radius: 6px;
  box-shadow:
    0 0 30px var(--neon-red),
    0 0 60px var(--neon-yellow);
  overflow: hidden;
  font-family: var(--font-display);
  user-select: none;
}

/* ── Open animation ─────────────────────────────────── */
.image-popup--opening {
  animation: popup-scale-in 0.5s ease-out;
}

/* ── Close animation ────────────────────────────────── */
.image-popup--closing {
  transform: scale(0);
  opacity: 0;
  transition:
    transform 0.3s ease-in,
    opacity 0.3s ease-in;
}

/* ── Header (drag handle) ───────────────────────────── */
.popup-header {
  background: linear-gradient(90deg, var(--neon-red), var(--neon-yellow), var(--neon-red));
  background-size: 200% 100%;
  color: #fff;
  padding: 10px 14px;
  font-weight: bold;
  font-size: 15px;
  text-align: center;
  cursor: move;
  text-shadow: 1px 1px 2px #000;
  letter-spacing: 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  padding-right: 40px; /* room for close button */
}

/* ── Close button ───────────────────────────────────── */
.popup-close {
  position: absolute;
  top: 6px;
  right: 8px;
  width: 24px;
  height: 24px;
  background: var(--neon-red);
  color: #fff;
  font-weight: bold;
  font-size: 18px;
  cursor: pointer;
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 3px;
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
  padding: 0;
  z-index: 2;
  transition:
    background 0.2s,
    color 0.2s;
}

.popup-close:hover {
  background: var(--neon-yellow);
  color: #000;
}

/* ── Body ───────────────────────────────────────────── */
.popup-body {
  padding: 16px;
}

/* ── Image wrapper ──────────────────────────────────── */
.popup-img-wrap {
  width: 100%;
  max-height: 60vh;
  overflow: hidden;
  background: #000;
  border-radius: 4px;
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.popup-img {
  max-width: 100%;
  max-height: 60vh;
  object-fit: contain;
  display: block;
}

.popup-img-fallback {
  padding: 32px 16px;
  color: var(--neon-red);
  font-family: var(--font-mono);
  font-size: 14px;
  text-align: center;
}

/* ── Meta section ───────────────────────────────────── */
.popup-meta {
  color: #ccc;
}

.popup-title {
  font-family: var(--font-display);
  font-size: 16px;
  color: var(--neon-yellow);
  margin: 0 0 10px;
  text-shadow: 0 0 6px var(--neon-yellow);
  font-weight: bold;
  letter-spacing: 1px;
}

/* ── Tags ───────────────────────────────────────────── */
.popup-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 10px;
}

.popup-tag {
  font-family: var(--font-mono);
  font-size: 11px;
  padding: 2px 8px;
  border: 1px solid;
  border-radius: 3px;
  background: rgba(0, 0, 0, 0.4);
}

/* ── Info row ───────────────────────────────────────── */
.popup-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-family: var(--font-mono);
  font-size: 12px;
  color: #888;
  padding-top: 8px;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}

.popup-uploader {
  color: var(--neon-cyan);
}

.popup-date {
  color: #888;
}

/* ── Open keyframe ──────────────────────────────────── */
@keyframes popup-scale-in {
  0% {
    transform: scale(0);
    opacity: 0;
  }
  60% {
    transform: scale(1.05);
  }
  100% {
    transform: scale(1);
    opacity: 1;
  }
}

/* ── Responsive ─────────────────────────────────────── */
@media (max-width: 640px) {
  .image-popup {
    width: min(95vw, 400px);
  }

  .popup-header {
    font-size: 13px;
    padding: 8px 10px 8px 10px;
  }

  .popup-body {
    padding: 10px;
  }

  .popup-img-wrap,
  .popup-img {
    max-height: 45vh;
  }

  .popup-title {
    font-size: 14px;
  }
}
</style>
