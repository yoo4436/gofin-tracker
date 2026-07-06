<template>
  <div class="container">
    <div class="header-zone">
      <h2>BTC/USDT 歷史走勢與技術指標 (Vue 3 + TS)</h2>
      <button class="toggle-btn" :class="{ 'btn-active': showMacd }" @click="handleToggleMacd">
        {{ showMacd ? '📊 隱藏 MACD 指標' : '📊 顯示 MACD 指標' }}
      </button>
    </div>
    
    <div ref="chartContainer" class="chart-box"></div>
    
    <div ref="macdContainer" class="macd-box" :style="{ display: showMacd ? 'block' : 'none' }"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue';
// 💡 引入 5.x 所需的各類圖表組件型態
import { createChart, CandlestickSeries, LineSeries, HistogramSeries } from 'lightweight-charts';
import type { CandlestickData, LineData, HistogramData, Time, IChartApi } from 'lightweight-charts';

// 對應 Go 後端全新的資料結構
interface ApiResponse {
  time: number;
  open: number;
  high: number;
  low: number;
  close: number;
  dif: number;
  dea: number;
  hist: number;
}

const chartContainer = ref<HTMLDivElement | null>(null);
const macdContainer = ref<HTMLDivElement | null>(null);
const showMacd = ref(true); // 預設開啟指標

// 將圖表實體提升至全域，以便 resize 函式存取
let mainChart: IChartApi | null = null;
let macdChart: IChartApi | null = null;

// 💡 處理按鈕切換開關的邏輯
const handleToggleMacd = async () => {
  showMacd.value = !showMacd.value;
  
  // 關鍵點：必須等 Vue 的 DOM 渲染更新完（nextTick），再命令畫布重新計算寬度
  await nextTick();
  if (mainChart && chartContainer.value) {
    mainChart.resize(chartContainer.value.clientWidth, 400);
  }
  if (macdChart && macdContainer.value && showMacd.value) {
    macdChart.resize(macdContainer.value.clientWidth, 150);
  }
};

onMounted(async () => {
  if (!chartContainer.value || !macdContainer.value) return;

  // ---------------------------------------------------
  // 1. 初始化【主 K 線圖】
  // ---------------------------------------------------
  mainChart = createChart(chartContainer.value, {
    width: chartContainer.value.clientWidth,
    height: 400,
    layout: { background: { color: '#ffffff' }, textColor: '#333333' },
    grid: { vertLines: { color: '#f0f0f0' }, horzLines: { color: '#f0f0f0' } },
    // 💡 密技：隱藏主圖的時間軸，交給最下方的 MACD 顯示，畫面才不會有重複的日期，變得很精簡乾淨
    timeScale: { borderColor: '#cccccc', visible: false }, 
  });

  const candlestickSeries = mainChart.addSeries(CandlestickSeries, {
    upColor: '#26a69a', downColor: '#ef5350',
    borderUpColor: '#26a69a', borderDownColor: '#ef5350',
    wickUpColor: '#26a69a', wickDownColor: '#ef5350',
  });

  // ---------------------------------------------------
  // 2. 初始化【MACD 子指標圖】
  // ---------------------------------------------------
  macdChart = createChart(macdContainer.value, {
    width: macdContainer.value.clientWidth,
    height: 150, // 指標圖稍微矮一點
    layout: { background: { color: '#ffffff' }, textColor: '#333333' },
    grid: { vertLines: { color: '#f0f0f0' }, horzLines: { color: '#f0f0f0' } },
    timeScale: { borderColor: '#cccccc', timeVisible: true }, // 由最底部的圖表負責秀時間軸
  });

  // 使用 5.x 全新語法建立快線(DIF)、慢線(DEA)與柱狀圖(HIST)
  const difSeries = macdChart.addSeries(LineSeries, { color: '#2196F3', lineWidth: 2 }); // 經典藍線
  const deaSeries = macdChart.addSeries(LineSeries, { color: '#FF9800', lineWidth: 2 }); // 經典橘線
  const histSeries = macdChart.addSeries(HistogramSeries, { base: 0 });                   // 能量柱狀圖

  // ---------------------------------------------------
  // 3. 核心大絕招：互相同步兩張畫布的左右拖拽與縮放
  // ---------------------------------------------------
  let isSyncing = false;
  
  mainChart.timeScale().subscribeVisibleLogicalRangeChange((range) => {
    if (isSyncing || !macdChart || !range) return;
    isSyncing = true;
    macdChart.timeScale().setVisibleLogicalRange(range);
    isSyncing = false;
  });

  macdChart.timeScale().subscribeVisibleLogicalRangeChange((range) => {
    if (isSyncing || !mainChart || !range) return;
    isSyncing = true;
    mainChart.timeScale().setVisibleLogicalRange(range);
    isSyncing = false;
  });

  // 監聽 RWD 瀏覽器視窗縮放
  window.addEventListener('resize', () => {
    if (chartContainer.value && mainChart) mainChart.resize(chartContainer.value.clientWidth, 400);
    if (macdContainer.value && macdChart) macdChart.resize(macdContainer.value.clientWidth, 150);
  });

  // ---------------------------------------------------
  // 4. 連線 Go 後端 API 並進行資料清洗與填入
  // ---------------------------------------------------
  try {
    const response = await fetch('http://localhost:8080/api/v1/klines');
    if (!response.ok) throw new Error('無法取得後端 API 資料');
    const rawData: ApiResponse[] = await response.json();

    // 格式化主 K 線
    const klineData: CandlestickData[] = rawData.map(item => ({
      time: item.time as Time, open: item.open, high: item.high, low: item.low, close: item.close
    }));

    // 格式化 MACD 快慢線
    const difData: LineData[] = rawData.map(item => ({ time: item.time as Time, value: item.dif }));
    const deaData: LineData[] = rawData.map(item => ({ time: item.time as Time, value: item.dea }));
    
    // 💡 格式化 MACD 柱狀圖：根據正負值動態給予紅綠顏色
    const histData: HistogramData[] = rawData.map(item => ({
      time: item.time as Time,
      value: item.hist,
      color: item.hist >= 0 ? '#26a69a' : '#ef5350' // 零軸以上綠色，零軸以下紅色
    }));

    // 將資料分別餵入各自的 Series
    candlestickSeries.setData(klineData);
    difSeries.setData(difData);
    deaSeries.setData(deaData);
    histSeries.setData(histData);

    // 撐滿圖表內容，並做第一次強制同步
    mainChart.timeScale().fitContent();
    const logicalRange = mainChart.timeScale().getVisibleLogicalRange();
    if (logicalRange) macdChart.timeScale().setVisibleLogicalRange(logicalRange);

  } catch (error) {
    console.error('資料流對接失敗:', error);
  }
});
</script>

<style scoped>
.container {
  padding: 20px;
  background-color: #f8f9fa;
  font-family: Arial, sans-serif;
}
.header-zone {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}
h2 {
  color: #333;
  margin: 0;
}
/* 💡 TradingView 風格的高質感按鈕 */
.toggle-btn {
  padding: 8px 16px;
  font-size: 14px;
  font-weight: bold;
  background-color: #ffffff;
  color: #333333;
  border: 1px solid #cccccc;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
}
.toggle-btn:hover {
  background-color: #f5f5f5;
  border-color: #a0a0a0;
}
.btn-active {
  background-color: #e3f2fd;
  color: #2196f3;
  border-color: #2196f3;
}
.chart-box {
  width: 100%;
  height: 400px;
  background-color: white;
  border-top-left-radius: 8px;
  border-top-right-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  overflow: hidden;
  border-bottom: 1px solid #f0f0f0;
}
.macd-box {
  width: 100%;
  height: 150px;
  background-color: white;
  border-bottom-left-radius: 8px;
  border-bottom-right-radius: 8px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}
</style>