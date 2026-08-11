<template>
  <div class="report-detail-container">
    <router-link class="back-btn" to="/reports">
      ← 返回日報列表
    </router-link>

    <div v-if="loading" class="report-article detail-skeleton" aria-label="正在載入日報">
      <span class="skeleton skeleton-meta" aria-hidden="true"></span>
      <span class="skeleton skeleton-title" aria-hidden="true"></span>
      <span class="skeleton skeleton-divider" aria-hidden="true"></span>
      <span v-for="index in 5" :key="index" class="skeleton skeleton-line" aria-hidden="true"></span>
    </div>

    <div v-else-if="error" class="state-panel" role="alert">
      <span class="state-icon" aria-hidden="true">!</span>
      <h2>無法顯示這篇日報</h2>
      <p>{{ error }}</p>
      <button type="button" class="retry-btn" @click="loadReport(route.params.id)">重新載入</button>
    </div>

    <article v-else-if="report" class="report-article">
      <header class="report-header">
        <span v-if="report.is_premium" class="badge-premium">會員獨享</span>
        <h1 class="report-title">{{ report.title }}</h1>
        <p class="report-meta">📅 發布時間：{{ formattedPublishedAt }}</p>
      </header>

      <hr class="divider" />

      <div class="markdown-body" v-html="parsedContent"></div>
    </article>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onBeforeUnmount } from 'vue'
import { useRoute } from 'vue-router'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { fetchReport, PublicApiError, type DailyReportDetail } from '../api/reports'

const route = useRoute()
const report = ref<DailyReportDetail | null>(null)
const loading = ref(true)
const error = ref('')
let requestController: AbortController | null = null

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

const formattedPublishedAt = computed(() => {
  if (!report.value) return ''
  const date = new Date(report.value.published_at)
  if (Number.isNaN(date.getTime())) return '日期未提供'
  return new Intl.DateTimeFormat('zh-TW', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  }).format(date)
})

const loadReport = async (routeId: string | string[] | undefined) => {
  requestController?.abort()
  report.value = null
  error.value = ''
  loading.value = true

  const rawId = Array.isArray(routeId) ? routeId[0] : routeId
  const reportId = rawId && /^\d+$/.test(rawId) ? Number(rawId) : Number.NaN
  if (!Number.isSafeInteger(reportId) || reportId < 1) {
    error.value = '日報網址無效'
    loading.value = false
    return
  }

  const controller = new AbortController()
  requestController = controller

  try {
    report.value = await fetchReport(reportId, controller.signal)
  } catch (cause) {
    if (cause instanceof DOMException && cause.name === 'AbortError') return
    error.value = cause instanceof PublicApiError && cause.status === 404
      ? '找不到該篇日報，可能尚未發布或已下架。'
      : cause instanceof Error ? cause.message : '讀取日報失敗'
  } finally {
    if (requestController === controller) loading.value = false
  }
}

watch(() => route.params.id, loadReport, { immediate: true })
onBeforeUnmount(() => requestController?.abort())
</script>

<style scoped>
.report-detail-container {
  max-width: 1100px;
  margin: 0 auto;
  padding: 24px clamp(12px, 3vw, 36px) 48px;
  color: #e0e3eb;
}

.back-btn {
  display: inline-flex;
  align-items: center;
  background-color: #2a2e39;
  color: #d1d4dc;
  border: 1px solid #363a45;
  padding: 8px 16px;
  border-radius: 6px;
  cursor: pointer;
  margin-bottom: 20px;
  font-size: 0.9rem;
  text-decoration: none;
  transition: all 0.2s ease;
}

.back-btn:hover {
  background-color: #363a45;
  color: #ffffff;
}

.report-article {
  background: linear-gradient(145deg, #202532 0%, #1b202b 100%);
  border-radius: 14px;
  padding: clamp(24px, 4vw, 48px);
  border: 1px solid #303746;
  box-shadow: 0 18px 45px rgba(0, 0, 0, 0.22);
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

.markdown-body :deep(ul),
.markdown-body :deep(ol) {
  margin: 12px 0 18px;
  padding-left: 1.65rem;
}

.markdown-body :deep(li) {
  padding-left: 0.25rem;
}

.markdown-body :deep(li + li) {
  margin-top: 8px;
}

.markdown-body :deep(blockquote) {
  width: 100%;
  border-left: 4px solid #4f7cff;
  background: linear-gradient(135deg, rgba(41, 98, 255, 0.16), rgba(41, 98, 255, 0.07));
  padding: 18px clamp(18px, 3vw, 28px);
  margin: 20px 0 24px;
  border-radius: 0 10px 10px 0;
  color: #e6e9f0;
}

.markdown-body :deep(blockquote > :last-child) {
  margin-bottom: 0;
}

.markdown-body :deep(blockquote ul),
.markdown-body :deep(blockquote ol) {
  padding-left: 1.4rem;
}

.markdown-body :deep(a) {
  color: #2962ff;
  text-decoration: none;
}

.markdown-body :deep(a:hover) {
  text-decoration: underline;
}

.state-panel {
  min-height: 320px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 40px;
  border: 1px dashed #363b49;
  border-radius: 14px;
  background-color: rgba(30, 34, 45, 0.55);
}

.state-panel h2 {
  margin: 14px 0 6px;
  color: #ffffff;
  font-size: 1.25rem;
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

.retry-btn {
  margin-top: 18px;
  padding: 9px 16px;
  border: 1px solid #2962ff;
  border-radius: 8px;
  background-color: #2962ff;
  color: #ffffff;
  cursor: pointer;
}

.detail-skeleton {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.skeleton {
  display: block;
  height: 15px;
  border-radius: 999px;
  background: linear-gradient(90deg, #282d39 25%, #343a48 50%, #282d39 75%);
  background-size: 200% 100%;
  animation: shimmer 1.4s infinite;
}

.skeleton-meta { width: 22%; }
.skeleton-title { width: min(680px, 82%); height: 32px; }
.skeleton-divider { width: 100%; height: 1px; margin: 8px 0; }
.skeleton-line { width: 100%; }
.skeleton-line:last-child { width: 64%; }

@keyframes shimmer {
  to { background-position: -200% 0; }
}

@media (prefers-reduced-motion: reduce) {
  .skeleton { animation: none; }
}

@media (max-width: 640px) {
  .report-detail-container {
    padding-top: 16px;
  }

  .report-article {
    border-radius: 10px;
  }

  .report-title {
    font-size: 1.5rem;
  }

  .markdown-body {
    font-size: 1rem;
  }
}
</style>
