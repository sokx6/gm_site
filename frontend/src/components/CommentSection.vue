<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import apiClient from '@/api/client'

interface Comment {
  id: number
  nickname: string
  content: string
  user_id: number
  created_at: string
  parent_id?: number
  reply_count?: number
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
const replyTo = ref<{ id: number; nickname: string } | null>(null)
const replyMap = ref<Record<number, Comment[]>>({})
const replyContent = ref('')
const replySubmitting = ref(false)
const expandedReplies = ref<Set<number>>(new Set())

function formatTime(iso: string): string {
  const d = new Date(iso)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

async function fetchComments(append = false) {
  try {
    const res = await apiClient.get(`/api/images/${props.imageId}/comments`, {
      params: { page: page.value, per_page: 10 }
    })
    const data = res.data
    const list: Comment[] = Array.isArray(data) ? data : (data.data?.comments ?? [])
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
    await apiClient.post(`/api/images/${props.imageId}/comments`, { content })
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
    await apiClient.delete(`/api/comments/${commentId}`)
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

async function fetchReplies(parentId: number) {
  try {
    const res = await apiClient.get(`/api/images/${props.imageId}/comments`, {
      params: { parent_id: parentId }
    })
    const data = res.data
    const list: Comment[] = Array.isArray(data) ? data : (data.data?.comments ?? [])
    replyMap.value[parentId] = list
  } catch {
    replyMap.value[parentId] = []
  }
}

async function submitReply(parentId: number) {
  const content = replyContent.value.trim()
  if (!content) return
  replySubmitting.value = true
  try {
    await apiClient.post(`/api/comments/${parentId}/reply`, { content, parent_id: parentId })
    replyContent.value = ''
    replyTo.value = null
    await fetchReplies(parentId)
    expandedReplies.value = new Set([...expandedReplies.value, parentId])
    // Refresh the parent comment to get updated reply_count
    await fetchComments()
  } catch {
    // Graceful degradation
  } finally {
    replySubmitting.value = false
  }
}

async function toggleReplies(commentId: number) {
  if (expandedReplies.value.has(commentId)) {
    const next = new Set(expandedReplies.value)
    next.delete(commentId)
    expandedReplies.value = next
  } else {
    if (!replyMap.value[commentId]) {
      await fetchReplies(commentId)
    }
    expandedReplies.value = new Set([...expandedReplies.value, commentId])
  }
}

function startReply(comment: Comment) {
  replyTo.value = { id: comment.id, nickname: comment.nickname }
  replyContent.value = ''
  nextTick(() => {
    const el = document.getElementById(`reply-form-${comment.id}`)
    if (el) el.scrollIntoView({ behavior: 'smooth', block: 'center' })
  })
}

function cancelReply() {
  replyTo.value = null
  replyContent.value = ''
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
        <div class="comment-actions">
          <button
            v-if="canDelete(comment)"
            class="comment-delete"
            @click="deleteComment(comment.id)"
          >
            删除
          </button>
          <button
            v-if="props.isLoggedIn"
            class="comment-reply-btn"
            @click="startReply(comment)"
          >
            回复
          </button>
        </div>

        <!-- Reply count / toggle -->
        <div
          v-if="(comment.reply_count ?? 0) > 0"
          class="reply-count"
        >
          <button
            class="reply-toggle"
            @click="toggleReplies(comment.id)"
          >
            {{ expandedReplies.has(comment.id) ? '收起回复' : `${comment.reply_count} 条回复` }}
          </button>
        </div>

        <!-- Nested replies -->
        <div
          v-if="expandedReplies.has(comment.id)"
          class="replies-container"
        >
          <div
            v-for="reply in replyMap[comment.id]"
            :key="reply.id"
            class="reply-item"
          >
            <div class="comment-header">
              <span class="comment-nickname">{{ reply.nickname }}</span>
              <span class="comment-time">{{ formatTime(reply.created_at) }}</span>
            </div>
            <p class="comment-content" v-text="reply.content" />
            <button
              v-if="canDelete(reply)"
              class="comment-delete"
              @click="deleteComment(reply.id)"
            >
              删除
            </button>
          </div>
        </div>

        <!-- Inline reply form -->
        <div
          v-if="replyTo?.id === comment.id"
          :id="`reply-form-${comment.id}`"
          class="reply-form"
        >
          <div class="reply-to-hint">回复 @{{ replyTo.nickname }}:</div>
          <textarea
            v-model="replyContent"
            class="comment-textarea neon-box"
            placeholder="写下你的回复…"
            rows="2"
            maxlength="500"
          />
          <div class="reply-form-actions">
            <button class="reply-cancel" @click="cancelReply">取消</button>
            <button
              class="comment-submit"
              :disabled="replySubmitting || !replyContent.trim()"
              @click="submitReply(comment.id)"
            >
              {{ replySubmitting ? '提交中…' : '回复' }}
            </button>
          </div>
        </div>
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

.comment-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 6px;
}

.comment-reply-btn {
  background: none;
  border: none;
  color: #555;
  font-family: var(--font-mono);
  font-size: 12px;
  cursor: pointer;
  padding: 2px 6px;
  transition: color 0.2s;
}

.comment-reply-btn:hover {
  color: var(--neon-pink);
}

.reply-count {
  margin-top: 4px;
}

.reply-toggle {
  background: none;
  border: none;
  color: #888;
  font-family: var(--font-mono);
  font-size: 12px;
  cursor: pointer;
  padding: 2px 0;
  transition: color 0.2s;
}

.reply-toggle:hover {
  color: var(--neon-cyan);
}

.replies-container {
  margin-left: 12px;
  border-left: 1px solid #1a1a1a;
  padding-left: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.reply-item {
  padding: 8px 0;
  border-bottom: 1px solid #1a1a1a;
}

.reply-item:last-child {
  border-bottom: none;
}

.reply-form {
  margin-top: 8px;
  margin-left: 12px;
  padding: 10px;
  border: 1px solid #222;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.reply-to-hint {
  font-family: var(--font-mono);
  font-size: 12px;
  color: var(--neon-pink);
}

.reply-form-actions {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 8px;
}

.reply-cancel {
  background: none;
  border: 1px solid #333;
  color: #888;
  font-family: var(--font-mono);
  font-size: 12px;
  padding: 4px 12px;
  cursor: pointer;
  transition: color 0.2s, border-color 0.2s;
}

.reply-cancel:hover {
  color: #fff;
  border-color: #555;
}

.comment-delete {
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
