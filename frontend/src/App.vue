<template>
  <div class="app-layout">
    <header class="navbar">
      <div class="brand" @click="$router.push('/')">
        <span class="logo-icon">📈</span>
        <span class="brand-title">GoFin-Tracker 指數追蹤系統</span>
      </div>

      <nav class="nav-links">
        <router-link to="/" class="nav-item">
          <span class="icon">📊</span> 即時看盤
        </router-link>
        <router-link to="/reports" class="nav-item">
          <span class="icon">📰</span> 投資日報
        </router-link>
      </nav>
    </header>

    <main class="main-content">
      <router-view />
    </main>
  </div>
</template>

<script setup ts>
// 通用外殼，保持簡潔
</script>

<style scoped>
.app-layout {
  width: 100%;
  min-height: 100vh;
  background-color: #131722;
  color: #d1d4dc;
  display: flex;
  flex-direction: column;
}

/* 上方導覽列滿版 */
.navbar {
  height: 60px;
  background-color: #1e222d;
  border-bottom: 1px solid #2a2e39;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  position: sticky;
  top: 0;
  z-index: 1000;
}

.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
}

.logo-icon {
  font-size: 1.4rem;
}

.brand-title {
  font-size: 1.1rem;
  font-weight: 700;
  color: #ffffff;
}

.nav-links {
  display: flex;
  gap: 8px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: 6px;
  color: #787b86;
  text-decoration: none;
  font-size: 0.95rem;
  transition: all 0.2s ease;
}

.nav-item:hover {
  color: #e0e3eb;
  background-color: #2a2e39;
}

.nav-item.router-link-exact-active {
  color: #ffffff;
  background-color: #2962ff;
  font-weight: 600;
}

/* 主內容區域：移除過度邊距，提供寬敞閱讀與看盤視角 */
.main-content {
  flex: 1;
  width: 100%;
  padding: 16px;
  box-sizing: border-box;
}
</style>