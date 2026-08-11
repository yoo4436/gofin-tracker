<template>
  <div class="reports-container">
    <header class="page-header">
      <div>
        <span class="eyebrow">DAILY INSIGHTS</span>
        <h2>每日市場投資分析</h2>
      </div>
      <p>掌握市場脈動、產業趨勢與重要事件。</p>
    </header>

    <div v-if="loading" class="report-grid" aria-label="正在載入日報">
      <div v-for="index in 6" :key="index" class="report-card skeleton-card" aria-hidden="true">
        <span class="skeleton skeleton-short"></span>
        <span class="skeleton skeleton-title"></span>
        <span class="skeleton skeleton-line"></span>
        <span class="skeleton skeleton-line"></span>
      </div>
    </div>

    <div v-else-if="error" class="state-panel" role="alert">
      <span class="state-icon">!</span>
      <h3>日報暫時無法載入</h3>
      <p>{{ error }}</p>
      <button type="button" class="retry-btn" @click="loadReports">重新載入</button>
    </div>

    <div v-else-if="reports.length === 0" class="state-panel">
      <span class="state-icon muted">—</span>
      <h3>目前還沒有已發布的日報</h3>
      <p>新內容發布後會顯示在這裡。</p>
    </div>

    <div v-else class="report-grid">
      <router-link
        v-for="report in reports" 
        :key="report.id" 
        :to="`/reports/${report.id}`"
        class="report-card"
      >
        <span v-if="report.is_premium" class="badge-paid">會員獨享</span>
        <h3>{{ report.title }}</h3>
        <p class="date">{{ formatPublishedAt(report.published_at) }}</p>
        <p class="summary">{{ report.summary }}</p>
        <span class="read-more">閱讀日報 <span aria-hidden="true">→</span></span>
      </router-link>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { fetchReports, type DailyReportSummary } from '../api/reports'

const reports = ref<DailyReportSummary[]>([])
const loading = ref(true)
const error = ref('')
let requestController: AbortController | null = null

const formatPublishedAt = (value: string) => {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '發布日期未提供'
  return new Intl.DateTimeFormat('zh-TW', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  }).format(date)
}

const loadReports = async () => {
  requestController?.abort()
  const controller = new AbortController()
  requestController = controller
  loading.value = true
  error.value = ''

  try {
    reports.value = await fetchReports(controller.signal)
  } catch (cause) {
    if (cause instanceof DOMException && cause.name === 'AbortError') return
    error.value = cause instanceof Error ? cause.message : '請稍後再試。'
  } finally {
    if (requestController === controller) loading.value = false
  }
}

onMounted(loadReports)
onBeforeUnmount(() => requestController?.abort())
</script>

<style scoped>
.reports-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 28px 10px 56px;
  color: #e0e3eb;
}

.page-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 24px;
  margin-bottom: 28px;
}

.eyebrow {
  display: block;
  margin-bottom: 5px;
  color: #6f92ff;
  font-size: 0.7rem;
  font-weight: 800;
  letter-spacing: 0.13em;
}

.page-header h2 {
  color: #ffffff;
  font-size: clamp(1.55rem, 2.4vw, 2rem);
  line-height: 1.25;
}

.page-header p {
  max-width: 390px;
  color: #969aa7;
  font-size: 0.95rem;
  text-align: right;
}

.report-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
}

.report-card {
  display: flex;
  flex-direction: column;
  min-height: 230px;
  background-color: #1e222d;
  border: 1px solid #2a2e39;
  border-radius: 12px;
  padding: 24px;
  cursor: pointer;
  color: inherit;
  text-decoration: none;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.1);
  transition: transform 0.2s, border-color 0.2s, box-shadow 0.2s;
  position: relative;
}

.report-card:hover {
  transform: translateY(-4px);
  border-color: #2962ff;
  box-shadow: 0 16px 36px rgba(0, 0, 0, 0.2);
}

.report-card h3 {
  color: #ffffff;
  font-size: 1.15rem;
  line-height: 1.4;
  padding-right: 68px;
  margin-bottom: 8px;
}

.badge-paid {
  position: absolute;
  top: 15px;
  right: 15px;
  background-color: #f6358a;
  color: #ffffff;
  font-size: 0.75rem;
  padding: 2px 6px;
  border-radius: 4px;
}

.date {
  color: #787b86;
  font-size: 0.85rem;
  margin: 8px 0;
}

.summary {
  color: #b2b5be;
  font-size: 0.95rem;
  line-height: 1.5;
  margin-bottom: 20px;
  display: -webkit-box;
  overflow: hidden;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 3;
}

.read-more {
  margin-top: auto;
  color: #6f92ff;
  font-size: 0.88rem;
  font-weight: 700;
}

.state-panel {
  min-height: 300px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 32px;
  border: 1px dashed #363b49;
  border-radius: 14px;
  background-color: rgba(30, 34, 45, 0.55);
  text-align: center;
}

.state-panel h3 {
  margin: 14px 0 6px;
  color: #ffffff;
}

.state-panel p {
  color: #969aa7;
}

.state-icon {
  width: 40px;
  height: 40px;
  display: grid;
  place-items: center;
  border-radius: 50%;
  background-color: rgba(246, 53, 138, 0.14);
  color: #ff6cae;
  font-size: 1.15rem;
  font-weight: 800;
}

.state-icon.muted {
  background-color: #2a2e39;
  color: #969aa7;
}

.retry-btn {
  margin-top: 18px;
  padding: 9px 16px;
  border: 1px solid #2962ff;
  border-radius: 8px;
  background-color: #2962ff;
  color: #ffffff;
  cursor: pointer;
}

.skeleton-card {
  cursor: default;
  gap: 14px;
}

.skeleton {
  display: block;
  height: 14px;
  border-radius: 999px;
  background: linear-gradient(90deg, #282d39 25%, #343a48 50%, #282d39 75%);
  background-size: 200% 100%;
  animation: shimmer 1.4s infinite;
}

.skeleton-short { width: 28%; }
.skeleton-title { width: 78%; height: 22px; }
.skeleton-line { width: 100%; }

@keyframes shimmer {
  to { background-position: -200% 0; }
}

@media (prefers-reduced-motion: reduce) {
  .skeleton { animation: none; }
}

@media (max-width: 640px) {
  .reports-container {
    padding-top: 18px;
  }

  .page-header {
    align-items: flex-start;
    flex-direction: column;
    gap: 8px;
  }

  .page-header p {
    text-align: left;
  }

  .report-grid {
    grid-template-columns: 1fr;
  }
}
</style>
