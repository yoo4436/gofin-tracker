import { createRouter, createWebHistory } from 'vue-router'
import ReportsView from '../views/ReportsView.vue'
import ReportDetailView from '../views/ReportDetailView.vue'

const routes = [
  { path: '/', component: () => import('../views/DashboardView.vue') },
  { path: '/reports', component: ReportsView },
  { path: '/reports/:id', component: ReportDetailView },
  { path: '/news', component: () => import('../views/NewsView.vue') },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
})
