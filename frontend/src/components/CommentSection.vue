<script setup lang="ts">
import { ref, onMounted } from 'vue'

interface Comment {
  id: number
  nickname: string
  content: string
  user_id: number
  created_at: string
}

const props = withDefaults(
  defineProps<{
    imageId: number
    isLoggedIn?: boolean
    currentUserId?: number
    isAdmin?: boolean
  }>(),
  {
    isLoggedIn: false,
    currentUserId: 0,
    isAdmin: false,
  }
)

const comments = ref<Comment[]>([])
const loading = ref(true)
const newComment = ref('')
const submitting = ref(false)
const page = ref(1)
const hasMore = ref(true)
const loadMoreLoading = ref(false)

function formatTime(iso: string): string {
  const d = new Date(iso)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

async function fetchComments(append = false) {
  try {
    const res = await fetch(
      `/api/images/${props.imageId}/comments?page=${page.value}&per_page=10`
    )
    if (!res.ok) {
      if (!append) comments.value = []
      return
    }
    const data = await res.json()
    const list: Comment[] = Array.isArray(data) ? data : (data.comments ?? [])
    if (append) {
      comments.value.push(...list)
    } else {
      comments.value = list
    }
    hasMore.value = list.length >= 10
  } catch {
    if (!append) comments.value = []
  } finally {
    loading.value = false
    loadMoreLoading.value = false
  }
}

async function submitComment() {
  const content = newComment.value.trim()
  if (!content) return
  submitting.value = true
  try {
    const res = await fetch(`/api/images/${props.imageId}/comments`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ content }),
    })
    if (!res.ok) return
    newComment.value = ''
    page.value = 1
    await fetchComments()
  } catch {
    // Graceful degradation
  } finally {
    submitting.value = false
  }
}

async function deleteComment(commentId: number) {
  try {
    await fetch(`/api/images/${props.imageId}/comments/${commentId}`, {
      method: 'DELETE',
    })
    comments.value = comments.value.filter((c) => c.id !== commentId)
  } catch {
    // Graceful degradation
  }
}

function canDelete(comment: Comment): boolean {
  return props.isAdmin || comment.user_id === props.currentUserId
}

async function loadMore() {
  loadMoreLoading.value = true
  page.value++
  await fetchComments(true)
}

onMounted(() => {
  fetchComments()
})
</script>

<template>
  <div class="comment-section">
    <!-- Comment input for logged-in users -->
    <div v-if="props.isLoggedIn" class="comment-form">
      <textarea
        v-model="newComment"
        class="comment-textarea neon-box"
        placeholder="写下你的评论…"
        rows="3"
        maxlength="500"
      />
      <button
        class="comment-submit"
        :disabled="submitting || !newComment.trim()"
        @click="submitComment"
      >
        {{ submitting ? '提交中…' : '提交评论' }}
      </button>
    </div>
    <div v-else class="comment-login-hint neon-box">
      登录后评论
    </div>

    <!-- Loading state -->
    <div v-if="loading" class="comment-loading">加载中…</div>

    <!-- Comment list -->
    <template v-else>
      <div v-if="comments.length === 0" class="comment-empty">
        暂无评论
      </div>
      <div
        v-for="comment in comments"
        :key="comment.id"
        class="comment-item"
      >
        <div class="comment-header">
          <span class="comment-nickname">{{ comment.nickname }}</span>
          <span class="comment-time">{{ formatTime(comment.created_at) }}</span>
        </div>
        <p class="comment-content" v-text="comment.content" />
        <button
          v-if="canDelete(comment)"
          class="comment-delete"
          @click="deleteComment(comment.id)"
        >
          删除
        </button>
      </div>

      <!-- Load more -->
      <div v-if="hasMore" class="comment-more">
        <button
          class="load-more-btn"
          :disabled="loadMoreLoading"
          @click="loadMore"
        >
          {{ loadMoreLoading ? '加载中…' : '加载更多' }}
        </button>
      </div>
    </template>
  </div>
</template>

<style scoped>
.comment-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.comment-form {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.comment-textarea {
  width: 100%;
  background: #000;
  color: #fff;
  font-family: var(--font-mono);
  font-size: 13px;
  padding: 8px;
  resize: vertical;
  outline: none;
}

.comment-textarea::placeholder {
  color: #444;
}

.comment-submit {
  align-self: flex-end;
  padding: 6px 16px;
  font-family: var(--font-mono);
  font-size: 13px;
  color: var(--neon-green);
  background: transparent;
  border: 2px solid var(--neon-green);
  cursor: pointer;
  transition: background 0.2s, box-shadow 0.2s;
}

.comment-submit:hover:not(:disabled) {
  background: rgba(0, 255, 0, 0.1);
  box-shadow: var(--glow-green);
}

.comment-submit:disabled {
  color: #555;
  border-color: #333;
  cursor: not-allowed;
}

.comment-login-hint {
  text-align: center;
  color: #666;
  font-family: var(--font-mono);
  font-size: 14px;
}

.comment-loading,
.comment-empty {
  text-align: center;
  color: #555;
  font-family: var(--font-mono);
  font-size: 13px;
  padding: 12px 0;
}

.comment-item {
  padding: 10px 0;
  border-bottom: 1px solid #1a1a1a;
  position: relative;
}

.comment-header {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: 6px;
}

.comment-nickname {
  font-family: var(--font-mono);
  font-size: 13px;
  color: var(--neon-cyan);
  text-shadow: 0 0 4px var(--neon-cyan);
}

.comment-time {
  font-family: var(--font-mono);
  font-size: 11px;
  color: #555;
}

.comment-content {
  font-family: var(--font-display);
  font-size: 14px;
  color: #ccc;
  line-height: 1.5;
  word-break: break-word;
}

.comment-delete {
  position: absolute;
  top: 10px;
  right: 0;
  background: none;
  border: none;
  color: #555;
  font-family: var(--font-mono);
  font-size: 12px;
  cursor: pointer;
  padding: 2px 6px;
  transition: color 0.2s;
}

.comment-delete:hover {
  color: var(--neon-red);
  text-shadow: 0 0 4px var(--neon-red);
}

.comment-more {
  text-align: center;
  padding: 8px 0;
}

.load-more-btn {
  background: none;
  border: 1px solid #333;
  color: #888;
  font-family: var(--font-mono);
  font-size: 13px;
  padding: 6px 20px;
  cursor: pointer;
  transition: color 0.2s, border-color 0.2s, box-shadow 0.2s;
}

.load-more-btn:hover:not(:disabled) {
  color: var(--neon-green);
  border-color: var(--neon-green);
  box-shadow: var(--glow-green);
}

.load-more-btn:disabled {
  color: #444;
  cursor: wait;
}
</style>
