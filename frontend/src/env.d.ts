/// <reference types="vite/client" />

// 宣告所有 .vue 檔案的型別，讓 TypeScript 可以正常 import Vue 元件
declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}