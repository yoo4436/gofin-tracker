<template>
  <div class="report-detail-container">
    <button class="back-btn" @click="$router.push('/reports')">
      ← 返回日報列表
    </button>

    <div v-if="loading" class="loading-state">
      <p>⏳ 載入日報內容中...</p>
    </div>

    <div v-else-if="error" class="error-state">
      <p>❌ {{ error }}</p>
    </div>

    <article v-else-if="report" class="report-article">
      <header class="report-header">
        <span v-if="report.is_premium" class="badge-premium">會員獨享</span>
        <h1 class="report-title">{{ report.title }}</h1>
        <p class="report-meta">
          📅 發布時間：{{ new Date(report.published_at).toLocaleDateString() }}
        </p>
      </header>

      <hr class="divider" />

      <div class="markdown-body" v-html="parsedContent"></div>
    </article>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { marked } from 'marked'
import DOMPurify from 'dompurify'

interface DailyReport {
  id: number
  title: string
  summary: string
  content: string
  cover_image_url: string
  is_premium: boolean
  published_at: string
}

const route = useRoute()
const report = ref<DailyReport | null>(null)
const loading = ref(true)
const error = ref('')

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

// 設定 marked 讓 Markdown 裡的超連結自動在「新頁籤」開啟
const renderer = new marked.Renderer()
renderer.link = ({ href, title, text }) => {
  return `<a href="${href}" title="${title || ''}" target="_blank" rel="noopener noreferrer">${text}</a>`
}
marked.setOptions({ renderer })

// 安全轉換與防 XSS 過濾
const parsedContent = computed(() => {
  if (!report.value?.content) return ''
  const rawHtml = marked(report.value.content) as string
  return DOMPurify.sanitize(rawHtml)
})

onMounted(async () => {
  try {
    const reportId = route.params.id
    const res = await fetch(`${API_BASE_URL}/api/v1/reports/${reportId}`)
    
    if (!res.ok) {
      throw new Error('找不到該篇日報內容')
    }
    
    report.value = await res.json()
  } catch (err: any) {
    error.value = err.message || '讀取日報失敗'
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.report-detail-container {
  max-width: 900px;
  margin: 0 auto;
  padding: 24px;
  color: #e0e3eb;
}

.back-btn {
  background-color: #2a2e39;
  color: #d1d4dc;
  border: 1px solid #363a45;
  padding: 8px 16px;
  border-radius: 6px;
  cursor: pointer;
  margin-bottom: 20px;
  font-size: 0.9rem;
  transition: all 0.2s ease;
}

.back-btn:hover {
  background-color: #363a45;
  color: #ffffff;
}

.report-article {
  background-color: #1e222d;
  border-radius: 10px;
  padding: 32px;
  border: 1px solid #2a2e39;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.2);
}

.report-title {
  font-size: 1.8rem;
  color: #ffffff;
  margin-bottom: 12px;
  line-height: 1.4;
}

.report-meta {
  color: #787b86;
  font-size: 0.9rem;
}

.badge-premium {
  display: inline-block;
  background-color: #f6358a;
  color: #ffffff;
  font-size: 0.75rem;
  padding: 2px 8px;
  border-radius: 4px;
  margin-bottom: 8px;
  font-weight: bold;
}

.divider {
  border: 0;
  height: 1px;
  background-color: #2a2e39;
  margin: 24px 0;
}

/* Markdown 樣式微調 */
.markdown-body {
  line-height: 1.8;
  font-size: 1.05rem;
  text-align: left;
}

.markdown-body :deep(h1),
.markdown-body :deep(h2),
.markdown-body :deep(h3) {
  color: #ffffff;
  margin-top: 24px;
  margin-bottom: 12px;
}

.markdown-body :deep(p) {
  margin-bottom: 16px;
}

.markdown-body :deep(blockquote) {
  border-left: 4px solid #2962ff;
  background-color: rgba(41, 98, 255, 0.1);
  padding: 12px 16px;
  margin: 16px 0;
  border-radius: 0 6px 6px 0;
}

.markdown-body :deep(a) {
  color: #2962ff;
  text-decoration: none;
}

.markdown-body :deep(a:hover) {
  text-decoration: underline;
}

.loading-state, .error-state {
  text-align: center;
  padding: 40px;
  font-size: 1.1rem;
}
</style>