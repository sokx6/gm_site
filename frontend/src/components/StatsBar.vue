<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const onlineCount = ref('--')
const totalVisitors = ref('--')
const newMembers = ref('--')
const uptimeDays = ref('--')

let ws: WebSocket | null = null
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
let destroyed = false

function connect() {
  if (destroyed) return
  const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
  const url = `${protocol}//${location.host}/ws`

  ws = new WebSocket(url)

  ws.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data)
      if (msg.type === 'stats') {
        if (msg.online !== undefined) onlineCount.value = String(msg.online)
        if (msg.totalVisitors !== undefined) totalVisitors.value = String(msg.totalVisitors)
        if (msg.newMembers !== undefined) newMembers.value = String(msg.newMembers)
        if (msg.uptimeDays !== undefined) uptimeDays.value = String(msg.uptimeDays)
      }
    } catch {
      // Ignore invalid JSON
    }
  }

  ws.onclose = () => {
    if (!destroyed) {
      reconnectTimer = setTimeout(connect, 3000)
    }
  }

  ws.onerror = () => {
    ws?.close()
  }
}

onMounted(() => {
  connect()
})

onUnmounted(() => {
  destroyed = true
  if (reconnectTimer) clearTimeout(reconnectTimer)
  if (ws) {
    ws.onclose = null
    ws.onerror = null
    ws.close()
    ws = null
  }
})
</script>

<template>
  <div class="stats-bar neon-box">
    <span class="stat-item">
      🟢 在线人数: <strong>{{ onlineCount }}</strong>
    </span>
    <span class="stat-sep">|</span>
    <span class="stat-item">
      📊 累计访客: <strong>{{ totalVisitors }}</strong>
    </span>
    <span class="stat-sep">|</span>
    <span class="stat-item">
      👤 新增成员: <strong>{{ newMembers }}</strong>
    </span>
    <span class="stat-sep">|</span>
    <span class="stat-item">
      ⏱ 安全运行天数: <strong>{{ uptimeDays }}</strong>
    </span>
  </div>
</template>

<style scoped>
.stats-bar {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 12px;
  padding: 8px 16px;
  font-family: var(--font-mono);
  font-size: 13px;
  color: var(--neon-green);
  flex-wrap: wrap;
}

.stat-item strong {
  color: var(--neon-yellow);
  text-shadow: 0 0 6px var(--neon-yellow);
}

.stat-sep {
  color: var(--neon-green);
  opacity: 0.4;
}

@media (max-width: 768px) {
  .stats-bar {
    font-size: 11px;
    gap: 6px;
    padding: 6px 8px;
  }
}
</style>
