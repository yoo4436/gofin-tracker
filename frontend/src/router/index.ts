import { createRouter, createWebHistory } from 'vue-router'
import ReportsView from '../views/ReportsView.vue'
import ReportDetailView from '../views/ReportDetailView.vue'

const routes = [
  { path: '/', component: () => import('../views/DashboardView.vue') }, // 原本的 K 線圖
  { path: '/reports', component: ReportsView },                        // 日報列表頁
  { path: '/reports/:id', component: ReportDetailView }               // 單篇日報閱讀頁
]

export const router = createRouter({
  history: createWebHistory(),
  routes
})