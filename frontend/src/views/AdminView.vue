<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
</script>

<template>
  <div class="admin-page">
    <!-- Ambient orbs -->
    <div class="ambient-orb ambient-orb--pink" aria-hidden="true"></div>
    <div class="ambient-orb ambient-orb--cyan" aria-hidden="true"></div>

    <!-- Header -->
    <div class="admin-header">
      <h1 class="admin-title glow-text">⚡ 管理后台 ⚡</h1>
      <p class="admin-subtitle">
        SYSTEM ADMIN PANEL // 管理员 {{ auth.user?.nickname || auth.user?.email }}
      </p>
    </div>

    <!-- Navigation cards -->
    <div class="admin-nav">
      <router-link to="/admin/images" class="admin-nav-card neon-box--pink">
        <div class="scanlines" aria-hidden="true"></div>
        <span class="nav-icon">🖼️</span>
        <h2 class="nav-title">图片管理</h2>
        <p class="nav-desc">查看、编辑、删除所有图片</p>
        <span class="nav-arrow">→</span>
      </router-link>

      <router-link to="/admin/users" class="admin-nav-card neon-box--red">
        <div class="scanlines" aria-hidden="true"></div>
        <span class="nav-icon">👥</span>
        <h2 class="nav-title">用户管理</h2>
        <p class="nav-desc">审核待批准的用户申请</p>
        <span class="nav-arrow">→</span>
      </router-link>

      <router-link to="/admin/albums" class="admin-nav-card neon-box--green">
        <div class="scanlines" aria-hidden="true"></div>
        <span class="nav-icon">📁</span>
        <h2 class="nav-title">相册管理</h2>
        <p class="nav-desc">创建、编辑、删除相册</p>
        <span class="nav-arrow">→</span>
      </router-link>
    </div>

    <!-- Back to home -->
    <div class="admin-footer">
      <router-link to="/" class="neon-link">← 返回首页</router-link>
    </div>
  </div>
</template>

<style scoped>
/* ── Page layout ───────────────────────────────────── */
.admin-page {
  min-height: 100vh;
  background: var(--bg-primary);
  position: relative;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 40px 20px;
}

/* ── Ambient orbs ─────────────────────────────────── */
.ambient-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(120px);
  opacity: 0.12;
  pointer-events: none;
  z-index: 0;
}

.ambient-orb--pink {
  width: 500px;
  height: 500px;
  background: var(--neon-pink);
  top: -150px;
  right: -150px;
  animation: orb-drift 8s ease-in-out infinite alternate;
}

.ambient-orb--cyan {
  width: 400px;
  height: 400px;
  background: var(--neon-cyan);
  bottom: -100px;
  left: -100px;
  animation: orb-drift 10s ease-in-out infinite alternate-reverse;
}

@keyframes orb-drift {
  0%   { transform: translate(0, 0) scale(1); }
  100% { transform: translate(40px, -30px) scale(1.15); }
}

/* ── Header ───────────────────────────────────────── */
.admin-header {
  text-align: center;
  margin-bottom: 48px;
  position: relative;
  z-index: 1;
}

.admin-title {
  font-family: var(--font-display);
  font-size: 42px;
  margin: 0 0 8px;
  color: #fff;
  text-shadow:
    0 0 10px var(--neon-pink),
    0 0 20px var(--neon-cyan),
    0 0 40px var(--neon-yellow),
    3px 3px 0 var(--neon-red),
    -3px -3px 0 var(--neon-blue);
  letter-spacing: 8px;
}

.admin-subtitle {
  font-family: var(--font-mono);
  font-size: 13px;
  color: var(--neon-cyan);
  text-transform: uppercase;
  letter-spacing: 4px;
  margin: 0;
}

/* ── Navigation cards ─────────────────────────────── */
.admin-nav {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 24px;
  width: 100%;
  max-width: 960px;
  position: relative;
  z-index: 1;
}

.admin-nav-card {
  position: relative;
  background: #000;
  border: 3px double var(--neon-pink);
  box-shadow:
    var(--glow-pink),
    0 0 60px rgba(255, 0, 255, 0.1);
  padding: 32px 28px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  text-decoration: none;
  transition: transform 0.3s ease, box-shadow 0.3s ease;
  overflow: hidden;
}

.admin-nav-card:hover {
  transform: translateY(-4px);
  box-shadow:
    0 0 30px var(--neon-pink),
    0 0 60px rgba(255, 0, 255, 0.25);
}

.admin-nav-card.neon-box--pink {
  border-color: var(--neon-pink);
  box-shadow: var(--glow-pink), 0 0 60px rgba(255, 0, 255, 0.1);
}

.admin-nav-card.neon-box--pink:hover {
  box-shadow: 0 0 30px var(--neon-pink), 0 0 60px rgba(255, 0, 255, 0.25);
}

.admin-nav-card.neon-box--red {
  border-color: var(--neon-red);
  box-shadow: var(--glow-red), 0 0 60px rgba(255, 0, 0, 0.1);
}

.admin-nav-card.neon-box--red:hover {
  box-shadow: 0 0 30px var(--neon-red), 0 0 60px rgba(255, 0, 0, 0.25);
}

.admin-nav-card.neon-box--green {
  border-color: var(--neon-green);
  box-shadow: var(--glow-green), 0 0 60px rgba(0, 255, 0, 0.08);
}

.admin-nav-card.neon-box--green:hover {
  box-shadow: 0 0 30px var(--neon-green), 0 0 60px rgba(0, 255, 0, 0.2);
}

/* Scanlines on cards */
.scanlines {
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

.nav-icon {
  font-size: 48px;
  position: relative;
  z-index: 3;
  filter: drop-shadow(0 0 8px var(--neon-pink));
}

.neon-box--red .nav-icon {
  filter: drop-shadow(0 0 8px var(--neon-red));
}

.neon-box--green .nav-icon {
  filter: drop-shadow(0 0 8px var(--neon-green));
}

.nav-title {
  font-family: var(--font-display);
  font-size: 22px;
  color: #fff;
  margin: 4px 0 0;
  text-shadow: 0 0 8px rgba(255, 255, 255, 0.3);
  position: relative;
  z-index: 3;
}

.nav-desc {
  font-family: var(--font-mono);
  font-size: 13px;
  color: #888;
  text-align: center;
  position: relative;
  z-index: 3;
  margin: 0;
}

.nav-arrow {
  font-family: var(--font-mono);
  font-size: 18px;
  color: var(--neon-cyan);
  margin-top: 4px;
  position: relative;
  z-index: 3;
  transition: transform 0.3s ease;
}

.admin-nav-card:hover .nav-arrow {
  transform: translateX(6px);
}

/* ── Footer ───────────────────────────────────────── */
.admin-footer {
  margin-top: 48px;
  position: relative;
  z-index: 1;
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

/* ── Responsive ───────────────────────────────────── */
@media (max-width: 768px) {
  .admin-title {
    font-size: 28px;
    letter-spacing: 4px;
  }

  .admin-nav {
    grid-template-columns: 1fr;
    max-width: 400px;
  }

  .admin-nav-card {
    padding: 24px 20px;
  }

  .nav-icon {
    font-size: 36px;
  }
}
</style>
