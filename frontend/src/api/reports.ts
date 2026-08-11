export interface DailyReportSummary {
  id: number
  title: string
  summary: string
  cover_image_url: string
  is_premium: boolean
  published_at: string
}

export interface DailyReportDetail extends DailyReportSummary {
  content: string
}

const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080').replace(/\/+$/, '')

export class PublicApiError extends Error {
  readonly status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'PublicApiError'
    this.status = status
  }
}

const getErrorMessage = async (response: Response) => {
  if (response.status >= 500) {
    return '日報服務暫時無法使用，請稍後再試。'
  }

  try {
    const body = await response.json() as { error?: unknown }
    return typeof body.error === 'string' && body.error.trim()
      ? body.error
      : `日報服務回應錯誤 (${response.status})`
  } catch {
    return `日報服務回應錯誤 (${response.status})`
  }
}

const isReportSummary = (value: unknown): value is DailyReportSummary => {
  if (!value || typeof value !== 'object') return false
  const report = value as Record<string, unknown>

  return Number.isSafeInteger(report.id)
    && typeof report.title === 'string'
    && typeof report.summary === 'string'
    && typeof report.cover_image_url === 'string'
    && typeof report.is_premium === 'boolean'
    && typeof report.published_at === 'string'
}

const isReportDetail = (value: unknown): value is DailyReportDetail =>
  isReportSummary(value) && typeof Reflect.get(value, 'content') === 'string'

const fetchPublicJson = async <T>(path: string, signal?: AbortSignal): Promise<T> => {
  let response: Response
  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      method: 'GET',
      headers: { Accept: 'application/json' },
      signal,
    })
  } catch (cause) {
    if (cause instanceof DOMException && cause.name === 'AbortError') throw cause
    throw new PublicApiError('無法連線到日報服務，請檢查網路後再試。', 0)
  }

  if (!response.ok) {
    throw new PublicApiError(await getErrorMessage(response), response.status)
  }

  return response.json() as Promise<T>
}

export const fetchReports = async (signal?: AbortSignal) => {
  const reports = await fetchPublicJson<unknown>('/api/v1/reports', signal)

  // Go 的 nil slice 會序列化成 null，公開列表在沒有資料時視為空陣列。
  if (reports === null) return []
  if (!Array.isArray(reports) || !reports.every(isReportSummary)) {
    throw new PublicApiError('日報列表格式不正確', 502)
  }
  return reports
}

export const fetchReport = async (id: number, signal?: AbortSignal) => {
  const report = await fetchPublicJson<unknown>(`/api/v1/reports/${id}`, signal)
  if (!isReportDetail(report)) {
    throw new PublicApiError('日報內容格式不正確', 502)
  }
  return report
}
