<script setup lang="ts">
withDefaults(
  defineProps<{
    onlineCount?: string
    newMembers?: string
    uptime?: string
  }>(),
  {
    onlineCount: '--',
    newMembers: '--',
    uptime: '--',
  }
)
</script>

<template>
  <div class="top-notify-bar neon-box">
    <span class="stat-item">
      🟢 在线人数: <strong>{{ onlineCount }}</strong>
    </span>
    <span class="stat-sep">|</span>
    <span class="stat-item">
      👤 新增成员: <strong>{{ newMembers }}</strong>
    </span>
    <span class="stat-sep">|</span>
    <span class="stat-item">
      ⏱ 安全运行天数: <strong>{{ uptime }}</strong>
    </span>
  </div>
</template>

<style scoped>
.top-notify-bar {
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
  .top-notify-bar {
    font-size: 11px;
    gap: 6px;
    padding: 6px 8px;
  }
}
</style>
