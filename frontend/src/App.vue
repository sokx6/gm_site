<script setup lang="ts">
import ParticleBg from './components/ParticleBg.vue'
import LayoutShell from './components/LayoutShell.vue'
import RainbowBanner from './components/RainbowBanner.vue'
import Marquee from './components/Marquee.vue'
import TopNotifyBar from './components/TopNotifyBar.vue'
</script>

<template>
  <ParticleBg />
  <LayoutShell>
    <template #top>
      <TopNotifyBar
        online-count="42"
        new-members="7"
        uptime="365"
      />
    </template>

    <template #left-sidebar>
      <Marquee text="🔥 群友风采 · 奥力给 · 冲冲冲 🔥" color="pink" />
      <div class="sidebar-placeholder">
        <p>📌 导航菜单</p>
        <p>📌 快捷入口</p>
      </div>
    </template>

    <template #main>
      <RainbowBanner title="群友风采" subtitle="Welcome to the Club" />
      <div class="main-placeholder neon-box" style="margin-top: 16px; padding: 24px;">
        <p>🎯 主要内容区域 — 群友图片展示</p>
      </div>
    </template>

    <template #right-sidebar>
      <Marquee text="⚡ SYSTEM ONLINE · 系统运行中 ⚡" color="green" />
      <div class="sidebar-placeholder">
        <p>📊 实时数据</p>
        <p>📊 公告栏</p>
      </div>
    </template>
  </LayoutShell>
</template>

<style scoped>
.sidebar-placeholder {
  margin-top: 12px;
  padding: 8px;
  font-family: var(--font-mono);
  font-size: 13px;
  color: var(--neon-green);
  line-height: 2;
}

.main-placeholder p {
  font-family: var(--font-display);
  color: var(--neon-yellow);
  text-align: center;
}
</style>
