<template>
  <div class="news-page">
    <header class="news-hero">
      <span class="eyebrow">MARKET NEWSROOM</span>
      <h1>市場新聞</h1>
      <p>一次掌握台股、國際情勢與美股的重要動態。</p>

      <form class="search-form" role="search" @submit.prevent="submitSearch">
        <label class="search-field">
          <span class="search-icon" aria-hidden="true"></span>
          <span class="sr-only">搜尋新聞</span>
          <input
            v-model="searchInput"
            type="search"
            maxlength="80"
            placeholder="搜尋公司、產業或事件，例如：台積電、AI、聯準會"
            autocomplete="off"
          />
        </label>
        <button type="submit" :disabled="loading">搜尋</button>
        <button v-if="activeQuery" class="clear-button" type="button" :disabled="loading" @click="clearSearch">清除</button>
      </form>

      <p v-if="activeQuery && !loading" class="search-summary">
        「{{ activeQuery }}」共找到 {{ feed?.articles.length ?? 0 }} 則分類結果
      </p>
    </header>

    <div v-if="error" class="state-panel" role="alert">
      <span class="state-mark">!</span>
      <div>
        <h2>新聞暫時載入失敗</h2>
        <p>{{ error }}</p>
      </div>
      <button type="button" @click="loadNews(activeQuery)">重新載入</button>
    </div>

    <div v-else class="news-columns" :aria-busy="loading">
      <section v-for="section in sections" :key="section.key" class="news-section">
        <header class="section-header">
          <div>
            <span class="section-dot" :class="`dot-${section.key}`"></span>
            <h2>{{ section.label }}</h2>
          </div>
          <span class="article-count">{{ section.items.length }} 則</span>
        </header>

        <div v-if="loading" class="article-list" aria-label="新聞載入中">
          <div v-for="index in 4" :key="index" class="article-card skeleton-card" aria-hidden="true">
            <span class="skeleton skeleton-image"></span>
            <span class="skeleton skeleton-meta"></span>
            <span class="skeleton skeleton-title"></span>
            <span class="skeleton skeleton-title short"></span>
          </div>
        </div>

        <div v-else-if="section.items.length === 0" class="empty-state">
          <span>暫無符合新聞</span>
          <small v-if="activeQuery">試著使用較短或不同的關鍵字</small>
        </div>

        <div v-else class="article-list">
          <a
            v-for="article in section.items"
            :key="article.id"
            class="article-card"
            :href="article.url"
            target="_blank"
            rel="noopener noreferrer"
          >
            <div v-if="article.image_url" class="article-image-wrap">
              <img :src="article.image_url" alt="" class="article-image" loading="lazy" />
            </div>
            <div v-else class="article-image-wrap image-placeholder" aria-hidden="true"><span>GoFin</span></div>
            <div class="article-body">
              <p class="article-meta">
                <span>{{ article.source }}</span>
                <time :datetime="article.published_at">{{ formatTime(article.published_at) }}</time>
              </p>
              <h3>{{ article.title }}</h3>
              <span class="read-link">閱讀原文 <span aria-hidden="true">↗</span></span>
            </div>
          </a>
        </div>
      </section>
    </div>

    <footer v-if="feed && !loading && !error" class="source-note">
      <span>新聞由鉅亨網公開端點彙整，內容與圖片版權歸原發布者所有；點擊後前往原始新聞頁面。</span>
      <span>更新時間：{{ formatFetchedAt(feed.fetched_at) }}</span>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { fetchNews, type NewsCategory, type NewsFeed } from '../api/news'

const searchInput = ref('')
const activeQuery = ref('')
const feed = ref<NewsFeed | null>(null)
const loading = ref(true)
const error = ref('')
let requestController: AbortController | null = null

const sectionDefinitions: Array<{ key: NewsCategory; label: string }> = [
  { key: 'taiwan', label: '台股' },
  { key: 'international', label: '國際情勢' },
  { key: 'us', label: '美股' },
]

const sections = computed(() => sectionDefinitions.map(section => ({
  ...section,
  items: (feed.value?.articles ?? []).filter(article => article.category === section.key),
})))

const loadNews = async (query: string) => {
  requestController?.abort()
  const controller = new AbortController()
  requestController = controller
  loading.value = true
  error.value = ''

  try {
    feed.value = await fetchNews(query, controller.signal)
  } catch (cause) {
    if (cause instanceof DOMException && cause.name === 'AbortError') return
    error.value = cause instanceof Error ? cause.message : '新聞載入失敗，請稍後再試。'
  } finally {
    if (requestController === controller) loading.value = false
  }
}

const submitSearch = () => {
  const query = searchInput.value.trim()
  activeQuery.value = query
  void loadNews(query)
}

const clearSearch = () => {
  searchInput.value = ''
  activeQuery.value = ''
  void loadNews('')
}

const formatTime = (value: string) => {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '時間未提供'
  return new Intl.DateTimeFormat('zh-TW', {
    month: 'numeric', day: 'numeric', hour: '2-digit', minute: '2-digit',
  }).format(date)
}

const formatFetchedAt = (value: string) => {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '時間未提供'
  return new Intl.DateTimeFormat('zh-TW', {
    year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit',
  }).format(date)
}

onMounted(() => loadNews(''))
onBeforeUnmount(() => requestController?.abort())
</script>

<style scoped>
.news-page {
  max-width: 1480px;
  margin: 0 auto;
  padding: clamp(18px, 2.4vw, 34px) clamp(4px, 1.5vw, 18px) 48px;
}

.news-hero {
  margin-bottom: 28px;
  padding: clamp(22px, 3vw, 36px);
  border: 1px solid #303646;
  border-radius: 18px;
  background: radial-gradient(circle at 88% 20%, rgba(80, 107, 255, 0.22), transparent 32%), linear-gradient(135deg, #202532 0%, #181d28 100%);
  box-shadow: 0 22px 50px rgba(0, 0, 0, 0.18);
}

.eyebrow {
  display: block;
  margin-bottom: 6px;
  color: #7898ff;
  font-size: 0.7rem;
  font-weight: 850;
  letter-spacing: 0.14em;
}

.news-hero h1 {
  color: #ffffff;
  font-size: clamp(1.8rem, 3.2vw, 2.65rem);
  line-height: 1.15;
  letter-spacing: -0.035em;
}

.news-hero > p:not(.search-summary) {
  margin-top: 9px;
  color: #a4a9b6;
  font-size: 0.98rem;
}

.search-form {
  max-width: 850px;
  display: flex;
  gap: 10px;
  margin-top: 22px;
}

.search-field {
  min-width: 0;
  flex: 1;
  position: relative;
}

.search-icon {
  position: absolute;
  top: 50%;
  left: 17px;
  width: 15px;
  height: 15px;
  border: 2px solid #818898;
  border-radius: 50%;
  transform: translateY(-58%);
  pointer-events: none;
}

.search-icon::after {
  content: '';
  position: absolute;
  right: -5px;
  bottom: -4px;
  width: 6px;
  height: 2px;
  border-radius: 2px;
  background: #818898;
  transform: rotate(45deg);
}

.search-field input {
  width: 100%;
  min-height: 48px;
  padding: 0 18px 0 47px;
  border: 1px solid #3c4353;
  border-radius: 11px;
  background: #111620;
  color: #ffffff;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.search-field input::placeholder { color: #747b8b; }

.search-field input:focus {
  border-color: #5c82ff;
  box-shadow: 0 0 0 4px rgba(41, 98, 255, 0.14);
  outline: none;
}

.search-form button,
.state-panel button {
  min-height: 48px;
  padding: 0 21px;
  border: 1px solid #4771ff;
  border-radius: 11px;
  background: #2962ff;
  color: #ffffff;
  font-weight: 750;
  cursor: pointer;
}

.search-form button:disabled { cursor: wait; opacity: 0.65; }
.search-form .clear-button { border-color: #424958; background: transparent; color: #b9bec9; }
.search-summary { margin-top: 13px; color: #9ea5b3; font-size: 0.88rem; }

.news-columns {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 18px;
}

.news-section {
  min-width: 0;
  border: 1px solid #2c3240;
  border-radius: 15px;
  background: #191e29;
  overflow: hidden;
}

.section-header {
  min-height: 62px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 0 18px;
  border-bottom: 1px solid #2c3240;
}

.section-header > div { display: flex; align-items: center; gap: 9px; }
.section-header h2 { color: #ffffff; font-size: 1.05rem; }

.section-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  box-shadow: 0 0 0 5px rgba(255, 255, 255, 0.04);
}

.dot-taiwan { background: #36c5a4; }
.dot-international { background: #f5b84b; }
.dot-us { background: #6e8fff; }
.article-count { color: #7f8695; font-size: 0.78rem; }
.article-list { display: flex; flex-direction: column; }

.article-card {
  display: grid;
  grid-template-columns: 112px minmax(0, 1fr);
  gap: 14px;
  padding: 16px;
  border-bottom: 1px solid #292f3c;
  color: inherit;
  text-decoration: none;
  transition: background-color 0.2s ease;
}

.article-card:last-child { border-bottom: 0; }
a.article-card:hover { background: #202633; }

.article-image-wrap {
  width: 112px;
  aspect-ratio: 16 / 10;
  align-self: start;
  overflow: hidden;
  border-radius: 8px;
  background: #111620;
}

.article-image {
  width: 100%;
  height: 100%;
  display: block;
  object-fit: cover;
  transition: transform 0.25s ease;
}

a.article-card:hover .article-image { transform: scale(1.035); }

.image-placeholder {
  display: grid;
  place-items: center;
  background: linear-gradient(145deg, #242c3e, #171c28);
  color: #66718a;
  font-size: 0.7rem;
  font-weight: 800;
  letter-spacing: 0.08em;
}

.article-body { min-width: 0; }

.article-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 6px;
  color: #7f8797;
  font-size: 0.73rem;
  white-space: nowrap;
  overflow: hidden;
}

.article-meta span { overflow: hidden; text-overflow: ellipsis; }
.article-meta time::before { content: '·'; margin-right: 6px; }

.article-card h3 {
  display: -webkit-box;
  overflow: hidden;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 3;
  color: #eef0f5;
  font-size: 0.94rem;
  line-height: 1.45;
}

.read-link { display: inline-block; margin-top: 9px; color: #7695ff; font-size: 0.75rem; font-weight: 700; }

.empty-state {
  min-height: 240px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 7px;
  padding: 30px;
  color: #9299a8;
  text-align: center;
}

.empty-state small { color: #686f7e; }

.state-panel {
  min-height: 110px;
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 20px;
  padding: 20px;
  border: 1px solid rgba(255, 108, 142, 0.28);
  border-radius: 13px;
  background: rgba(104, 29, 50, 0.18);
}

.state-panel > div { flex: 1; }
.state-panel h2 { margin-bottom: 4px; color: #ffffff; font-size: 1rem; }
.state-panel p { color: #b5bbc6; font-size: 0.88rem; }

.state-mark {
  width: 34px;
  height: 34px;
  display: grid;
  place-items: center;
  flex: 0 0 auto;
  border-radius: 50%;
  background: rgba(246, 53, 138, 0.16);
  color: #ff7cab;
  font-weight: 850;
}

.skeleton-card { min-height: 115px; grid-template-rows: repeat(3, auto); }

.skeleton {
  display: block;
  border-radius: 999px;
  background: linear-gradient(90deg, #252b37 25%, #313846 50%, #252b37 75%);
  background-size: 200% 100%;
  animation: shimmer 1.35s infinite;
}

.skeleton-image { grid-row: 1 / 4; width: 112px; height: 70px; border-radius: 8px; }
.skeleton-meta { width: 45%; height: 9px; }
.skeleton-title { width: 100%; height: 13px; }
.skeleton-title.short { width: 76%; }

.source-note {
  display: flex;
  justify-content: space-between;
  gap: 18px;
  margin-top: 20px;
  padding: 0 4px;
  color: #727987;
  font-size: 0.75rem;
  line-height: 1.55;
}

.source-note span:last-child { flex: 0 0 auto; }

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

@keyframes shimmer { to { background-position: -200% 0; } }

@media (prefers-reduced-motion: reduce) {
  .skeleton { animation: none; }
  .article-image { transition: none; }
}

@media (max-width: 1120px) {
  .news-columns { grid-template-columns: 1fr; }
}

@media (max-width: 640px) {
  .news-page { padding-top: 10px; }
  .news-hero { border-radius: 13px; }
  .search-form { flex-wrap: wrap; }
  .search-field { flex-basis: 100%; }
  .search-form button { flex: 1; }

  .article-card,
  .skeleton-card {
    grid-template-columns: 94px minmax(0, 1fr);
    gap: 12px;
    padding: 13px;
  }

  .article-image-wrap,
  .skeleton-image { width: 94px; }
  .article-card h3 { font-size: 0.9rem; }

  .source-note,
  .state-panel { align-items: flex-start; flex-direction: column; }
}
</style>
