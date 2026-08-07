<template>
  <div class="reports-container">
    <h2>📈 每日市場投資分析</h2>
    <div class="report-grid">
      <div 
        v-for="report in reports" 
        :key="report.id" 
        class="report-card"
        @click="$router.push(`/reports/${report.id}`)"
      >
        <span v-if="report.is_premium" class="badge-paid">會員獨享</span>
        <h3>{{ report.title }}</h3>
        <p class="date">{{ new Date(report.published_at).toLocaleDateString() }}</p>
        <p class="summary">{{ report.summary }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

interface DailyReport {
  id: number
  title: string
  summary: string
  cover_image_url: string
  is_premium: boolean
  published_at: string
}

const reports = ref<DailyReport[]>([])
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

onMounted(async () => {
  try {
    const res = await fetch(`${API_BASE_URL}/api/v1/reports`)
    if (res.ok) {
      reports.value = await res.json()
    }
  } catch (err) {
    console.error('無法載入日報列表:', err)
  }
})
</script>

<style scoped>
.reports-container {
  max-width: 1200px; /* 控制文章列表閱讀適中寬度 */
  margin: 0 auto;
  padding: 20px 10px;
  color: #e0e3eb;
}

.reports-container h2 {
  color: #ffffff !important; /* 確保主標題為亮白色 */
  font-size: 1.6rem;
  margin-bottom: 24px;
}

.report-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
}

.report-card {
  background-color: #1e222d;
  border: 1px solid #2a2e39;
  border-radius: 8px;
  padding: 20px;
  cursor: pointer;
  transition: transform 0.2s, border-color 0.2s;
  position: relative;
}

.report-card:hover {
  transform: translateY(-4px);
  border-color: #2962ff;
}

.report-card h3 {
  color: #ffffff !important; /* 確保卡片標題為亮白色 */
  font-size: 1.15rem;
  line-height: 1.4;
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
}
</style>