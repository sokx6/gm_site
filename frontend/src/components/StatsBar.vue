<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { onlineCount, totalVisitors, newMembers, uptimeDays, connect, disconnect } from '@/composables/useStatsWebSocket'

onMounted(() => {
  connect()
})

onUnmounted(() => {
  disconnect()
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
