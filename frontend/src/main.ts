import { createApp } from 'vue'
import App from './App.vue'
import { router } from './router' // 1. 引入路由對照設定
import './style.css'

const app = createApp(App)
app.use(router) // 2. 註冊 Vue Router
app.mount('#app')