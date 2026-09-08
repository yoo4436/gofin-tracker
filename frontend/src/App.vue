<template>
  <div class="app-layout">
    <header class="navbar">
      <router-link class="brand" to="/" aria-label="回到市場總覽">
        <span class="logo-icon" aria-hidden="true">GF</span>
        <span class="brand-title">GoFin Tracker</span>
      </router-link>

      <nav class="nav-links" aria-label="主要功能">
        <router-link to="/" class="nav-item">市場總覽</router-link>
        <router-link to="/reports" class="nav-item">每日報告</router-link>
        <router-link to="/news" class="nav-item">新聞</router-link>
      </nav>
    </header>

    <main class="main-content">
      <router-view />
    </main>
  </div>
</template>

<style scoped>
.app-layout {
  width: 100%;
  min-height: 100vh;
  background-color: #131722;
  color: #d1d4dc;
  display: flex;
  flex-direction: column;
}

.navbar {
  min-height: 64px;
  background-color: rgba(24, 28, 39, 0.96);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid #2a2e39;
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: clamp(18px, 3vw, 42px);
  padding: 0 clamp(14px, 3vw, 32px);
  position: sticky;
  top: 0;
  z-index: 1000;
}

.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 0 0 auto;
  color: #ffffff;
  text-decoration: none;
}

.logo-icon {
  width: 34px;
  height: 34px;
  display: grid;
  place-items: center;
  border-radius: 9px;
  background: linear-gradient(135deg, #2962ff, #7547ff);
  color: #ffffff;
  font-size: 0.72rem;
  font-weight: 900;
  letter-spacing: -0.03em;
  box-shadow: 0 7px 18px rgba(41, 98, 255, 0.28);
}

.brand-title {
  font-size: 1.05rem;
  font-weight: 750;
  letter-spacing: -0.01em;
}

.nav-links {
  min-width: 0;
  display: flex;
  align-items: center;
  align-self: stretch;
  gap: 4px;
  overflow-x: auto;
  scrollbar-width: none;
}

.nav-links::-webkit-scrollbar { display: none; }

.nav-item {
  height: 100%;
  display: flex;
  align-items: center;
  flex: 0 0 auto;
  padding: 0 14px;
  border-bottom: 2px solid transparent;
  color: #9297a6;
  text-decoration: none;
  font-size: 0.92rem;
  font-weight: 650;
  transition: color 0.2s ease, border-color 0.2s ease, background-color 0.2s ease;
}

.nav-item:hover {
  color: #e7e9ef;
  background-color: rgba(255, 255, 255, 0.035);
}

.nav-item.router-link-exact-active {
  color: #ffffff;
  border-bottom-color: #5f85ff;
  background-color: rgba(41, 98, 255, 0.09);
}

.main-content {
  flex: 1;
  width: 100%;
  padding: 16px;
}

@media (max-width: 640px) {
  .navbar {
    min-height: 58px;
    gap: 10px;
    padding: 0 10px;
  }

  .brand-title { display: none; }

  .nav-item {
    padding: 0 11px;
    font-size: 0.86rem;
  }

  .main-content { padding: 10px; }
}
</style>
