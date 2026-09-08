export type NewsCategory = 'taiwan' | 'international' | 'us'

export interface NewsArticle {
  id: number
  title: string
  url: string
  image_url: string
  source: string
  category: NewsCategory
  published_at: string
}

export interface NewsFeed {
  query: string
  fetched_at: string
  articles: NewsArticle[]
}

const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080').replace(/\/+$/, '')

const isArticle = (value: unknown): value is NewsArticle => {
  if (!value || typeof value !== 'object') return false
  const article = value as Record<string, unknown>
  return Number.isSafeInteger(article.id)
    && typeof article.title === 'string'
    && typeof article.url === 'string'
    && typeof article.image_url === 'string'
    && typeof article.source === 'string'
    && ['taiwan', 'international', 'us'].includes(String(article.category))
    && typeof article.published_at === 'string'
}

const isFeed = (value: unknown): value is NewsFeed => {
  if (!value || typeof value !== 'object') return false
  const feed = value as Record<string, unknown>
  return typeof feed.query === 'string'
    && typeof feed.fetched_at === 'string'
    && Array.isArray(feed.articles)
    && feed.articles.every(isArticle)
}

export const fetchNews = async (query: string, signal?: AbortSignal): Promise<NewsFeed> => {
  const params = new URLSearchParams()
  if (query.trim()) params.set('q', query.trim())
  const suffix = params.size > 0 ? `?${params.toString()}` : ''

  let response: Response
  try {
    response = await fetch(`${API_BASE_URL}/api/v1/news${suffix}`, {
      method: 'GET',
      headers: { Accept: 'application/json' },
      signal,
    })
  } catch (cause) {
    if (cause instanceof DOMException && cause.name === 'AbortError') throw cause
    throw new Error('目前無法連上新聞服務，請稍後再試。')
  }

  if (!response.ok) {
    throw new Error(response.status === 400
      ? '搜尋關鍵字過長，請縮短後再試。'
      : '新聞來源暫時無法使用，請稍後再試。')
  }

  const payload: unknown = await response.json()
  if (!isFeed(payload)) throw new Error('新聞資料格式異常，請稍後再試。')
  return payload
}
