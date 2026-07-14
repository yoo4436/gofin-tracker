<template>
  <div class="search-container" ref="searchContainerRef">
    <div class="search-box" @click="openDropdown" :class="{ 'is-open': isOpen }">
      <span class="icon">🔍</span>
      <input 
        type="text" 
        v-model="searchQuery" 
        @input="handleSearch"
        placeholder="搜尋代號或名稱 (例如: BTC)"
        :readonly="!isOpen"
      />
    </div>

    <div class="dropdown-menu" v-if="isOpen">
      <div class="dropdown-header">
        <span class="col-code">代號</span>
        <span class="col-name">名稱</span>
        <span class="col-exchange">交易所</span>
      </div>
      
      <ul class="symbol-list">
        <li v-if="isLoading" class="state-msg">載入中...</li>
        <li v-else-if="symbols.length === 0" class="state-msg">找不到商品</li>
        
        <li 
          v-else 
          v-for="item in symbols" 
          :key="item.id" 
          @click="selectSymbol(item)"
          class="symbol-item"
        >
          <span class="col-code">{{ item.symbol_code }}</span>
          <span class="col-name">{{ item.name }}</span>
          <span class="col-exchange"><span class="badge">{{ item.exchange_name }}</span></span>
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';

// 對應 Go 後端的資料結構
interface SymbolItem {
  id: number;
  symbol_code: string;
  name: string;
  market_type: string;
  exchange_name: string;
}

// 宣告往外傳遞的事件 (當選中商品時，通知父元件 App.vue)
const emit = defineEmits(['select']);

const isOpen = ref(false);
const searchQuery = ref('');
const symbols = ref<SymbolItem[]>([]);
const isLoading = ref(false);
const searchContainerRef = ref<HTMLElement | null>(null);

let debounceTimer: ReturnType<typeof setTimeout> | null = null;
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

// 📡 呼叫後端 API
const fetchSymbols = async (query = '') => {
  isLoading.value = true;
  try {
    const url = query 
      ? `${API_BASE_URL}/api/v1/symbols?q=${encodeURIComponent(query)}`
      : `${API_BASE_URL}/api/v1/symbols`;
      
    const res = await fetch(url);
    if (!res.ok) throw new Error('API 請求失敗');
    symbols.value = await res.json();
  } catch (error) {
    console.error('獲取商品清單失敗:', error);
  } finally {
    isLoading.value = false;
  }
};

// ⌨️ 處理使用者輸入 (防抖機制 Debounce)
const handleSearch = () => {
  if (debounceTimer) clearTimeout(debounceTimer);
  debounceTimer = setTimeout(() => {
    fetchSymbols(searchQuery.value);
  }, 300); // 停頓 300 毫秒才發送請求給後端
};

// 🖱️ 打開選單
const openDropdown = () => {
  if (!isOpen.value) {
    isOpen.value = true;
    searchQuery.value = ''; // 打開時清空輸入框，準備讓使用者搜尋
    fetchSymbols();         // 預設抓取全部列表
  }
};

// ✅ 選擇商品
const selectSymbol = (item: SymbolItem) => {
  searchQuery.value = item.symbol_code; // 把選中的代號填入輸入框
  isOpen.value = false;
  emit('select', item); // 通知父元件切換圖表資料
};

// 🛡️ 點擊元件外部自動關閉
const handleClickOutside = (event: MouseEvent) => {
  if (searchContainerRef.value && !searchContainerRef.value.contains(event.target as Node)) {
    isOpen.value = false;
  }
};

onMounted(() => {
  document.addEventListener('click', handleClickOutside);
});

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside);
});
</script>

<style scoped>
.search-container {
  position: relative;
  width: 320px;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Arial, sans-serif;
}

/* 搜尋框設計 */
.search-box {
  display: flex;
  align-items: center;
  background-color: #ffffff;
  border: 1px solid #d1d4dc;
  border-radius: 8px;
  padding: 8px 12px;
  cursor: text;
  transition: border-color 0.2s;
}
.search-box.is-open { border-color: #2962ff; box-shadow: 0 0 0 2px rgba(41, 98, 255, 0.2); }
.icon { font-size: 16px; margin-right: 8px; color: #787b86; }
.search-box input { border: none; outline: none; width: 100%; font-size: 16px; color: #131722; cursor: pointer; }
.search-box.is-open input { cursor: text; }

/* 下拉選單設計 */
.dropdown-menu {
  position: absolute;
  top: 110%;
  left: 0;
  width: 100%;
  background-color: #ffffff;
  border: 1px solid #e0e3eb;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  z-index: 100;
  overflow: hidden;
}

.dropdown-header {
  display: flex;
  padding: 10px 16px;
  background-color: #f8f9fd;
  border-bottom: 1px solid #e0e3eb;
  font-size: 12px;
  color: #787b86;
  font-weight: 600;
}

.symbol-list { list-style: none; margin: 0; padding: 0; max-height: 300px; overflow-y: auto; }
.state-msg { padding: 16px; text-align: center; color: #787b86; font-size: 14px; }

.symbol-item {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  cursor: pointer;
  border-bottom: 1px solid #f0f3fa;
  transition: background-color 0.2s;
}
.symbol-item:hover { background-color: #f0f3fa; }

/* 欄位排版 */
.col-code { width: 35%; font-weight: 700; color: #131722; }
.col-name { width: 45%; color: #787b86; font-size: 13px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.col-exchange { width: 20%; text-align: right; }

.badge {
  background-color: #e3f2fd;
  color: #2962ff;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: bold;
}
</style>