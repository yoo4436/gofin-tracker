<template>
  <div class="container">
    <div class="header-zone">
      <h2>BTC/USDT 歷史走勢與全指標儀表板</h2>
      <div class="btn-group">
        <button class="toggle-btn" :class="{ 'btn-active': showMa }" @click="handleToggleMa">〰️ MA 均線</button>
        <button class="toggle-btn" :class="{ 'btn-active': showBb }" @click="handleToggleBb">🌀 布林通道</button>
        <button class="toggle-btn" :class="{ 'btn-active': showMacd }" @click="handleToggleMacd">📊 MACD 指標</button>
        <button class="toggle-btn" :class="{ 'btn-active': showRsi }" @click="handleToggleRsi">📈 RSI 指標</button>
      </div>
    </div>

    <div ref="chartContainer" class="chart-box">
      <div class="chart-legend">
        <span class="legend-title">BTC/USDT 1D</span>
        <span v-if="showMa" class="legend-item ma7">MA(7)</span>
        <span v-if="showMa" class="legend-item ma25">MA(25)</span>
        <span v-if="showBb" class="legend-item bb">BB(20, 2)</span>
      </div>
    </div>

    <div ref="matchContainer" class="macd-box" :style="{ display: showMacd ? 'block' : 'none' }"></div>
    <div ref="rsiContainer" class="rsi-box" :style="{ display: showRsi ? 'block' : 'none' }"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue';
import { createChart, CandlestickSeries, LineSeries, HistogramSeries } from 'lightweight-charts';
import type { Time, IChartApi, ISeriesApi } from 'lightweight-charts';

interface ApiResponse {
  time: number;
  open: number;
  high: number;
  low: number;
  close: number;
  dif: number;
  dea: number;
  hist: number;
  ma7: number;
  ma25: number;
  bbiUpper: number;
  bbiMiddle: number;
  bbiLower: number;
  rsi: number;
}

const chartContainer = ref<HTMLDivElement | null>(null);
const matchContainer = ref<HTMLDivElement | null>(null);
const rsiContainer = ref<HTMLDivElement | null>(null);

// 四大開關狀態
const showMa = ref(true);
const showBb = ref(true);
const showMacd = ref(true);
const showRsi = ref(true);

let mainChart: IChartApi | null = null;
let macdChart: IChartApi | null = null;
let rsiChart: IChartApi | null = null;

// 💡 將主圖的線條變數提升到外層，這樣按鈕函數才抓得到它們來設定隱藏
let ma7Series: ISeriesApi<"Line"> | null = null;
let ma25Series: ISeriesApi<"Line"> | null = null;
let bbUpperSeries: ISeriesApi<"Line"> | null = null;
let bbMiddleSeries: ISeriesApi<"Line"> | null = null;
let bbLowerSeries: ISeriesApi<"Line"> | null = null;

const resizeDashboard = () => {
  if (mainChart && chartContainer.value) mainChart.resize(chartContainer.value.clientWidth, 400);
  if (macdChart && matchContainer.value && showMacd.value) macdChart.resize(matchContainer.value.clientWidth, 120);
  if (rsiChart && rsiContainer.value && showRsi.value) rsiChart.resize(rsiContainer.value.clientWidth, 120);
};

// 💡 透過 applyOptions({ visible: false }) 實現主圖線條的動態開關
const handleToggleMa = () => {
  showMa.value = !showMa.value;
  if (ma7Series) ma7Series.applyOptions({ visible: showMa.value });
  if (ma25Series) ma25Series.applyOptions({ visible: showMa.value });
};

const handleToggleBb = () => {
  showBb.value = !showBb.value;
  if (bbUpperSeries) bbUpperSeries.applyOptions({ visible: showBb.value });
  if (bbMiddleSeries) bbMiddleSeries.applyOptions({ visible: showBb.value });
  if (bbLowerSeries) bbLowerSeries.applyOptions({ visible: showBb.value });
};

const handleToggleMacd = async () => {
  showMacd.value = !showMacd.value;
  await nextTick();
  resizeDashboard();
};

const handleToggleRsi = async () => {
  showRsi.value = !showRsi.value;
  await nextTick();
  resizeDashboard();
};

onMounted(async () => {
  if (!chartContainer.value || !matchContainer.value || !rsiContainer.value) return;

  const commonGrid = { vertLines: { color: '#f2f2f2' }, horzLines: { color: '#f2f2f2' } };

  // 1. 建立主圖
  mainChart = createChart(chartContainer.value, {
    width: chartContainer.value.clientWidth,
    height: 400,
    layout: { background: { color: '#ffffff' }, textColor: '#333333' },
    grid: commonGrid,
    timeScale: { borderColor: '#cccccc', visible: false },
  });

  const candlestickSeries = mainChart.addSeries(CandlestickSeries, {
    upColor: '#26a69a', downColor: '#ef5350', borderUpColor: '#26a69a', borderDownColor: '#ef5350', wickUpColor: '#26a69a', wickDownColor: '#ef5350',
  });

  // 綁定到外層變數
  ma7Series = mainChart.addSeries(LineSeries, { color: '#ba55d3', lineWidth: 1 });
  ma25Series = mainChart.addSeries(LineSeries, { color: '#4169e1', lineWidth: 1 });
  bbUpperSeries = mainChart.addSeries(LineSeries, { color: '#9e9e9e', lineWidth: 1, lineStyle: 2 });
  bbMiddleSeries = mainChart.addSeries(LineSeries, { color: '#ffb300', lineWidth: 1, lineStyle: 2 });
  bbLowerSeries = mainChart.addSeries(LineSeries, { color: '#9e9e9e', lineWidth: 1, lineStyle: 2 });

  // 2. 建立 MACD
  macdChart = createChart(matchContainer.value, { width: matchContainer.value.clientWidth, height: 120, layout: { background: { color: '#ffffff' }, textColor: '#333333' }, grid: commonGrid, timeScale: { borderColor: '#cccccc', visible: false } });
  const difSeries = macdChart.addSeries(LineSeries, { color: '#2196F3', lineWidth: 2 });
  const deaSeries = macdChart.addSeries(LineSeries, { color: '#FF9800', lineWidth: 2 });
  const histSeries = macdChart.addSeries(HistogramSeries, { base: 0 });

  // 3. 建立 RSI
  rsiChart = createChart(rsiContainer.value, { width: rsiContainer.value.clientWidth, height: 120, layout: { background: { color: '#ffffff' }, textColor: '#333333' }, grid: commonGrid, timeScale: { borderColor: '#cccccc', timeVisible: true } });
  const rsiSeries = rsiChart.addSeries(LineSeries, { color: '#e91e63', lineWidth: 2 });
  const rsi30Series = rsiChart.addSeries(LineSeries, { color: '#b0bec5', lineWidth: 1, lineStyle: 3 });
  const rsi70Series = rsiChart.addSeries(LineSeries, { color: '#b0bec5', lineWidth: 1, lineStyle: 3 });

  // 4. 同步縮放
  let isSyncing = false;
  const syncRange = (range: any, targetCharts: (IChartApi | null)[]) => {
    if (isSyncing || !range) return;
    isSyncing = true;
    targetCharts.forEach(chart => { if (chart) chart.timeScale().setVisibleLogicalRange(range); });
    isSyncing = false;
  };
  mainChart.timeScale().subscribeVisibleLogicalRangeChange((range) => syncRange(range, [macdChart, rsiChart]));
  macdChart.timeScale().subscribeVisibleLogicalRangeChange((range) => syncRange(range, [mainChart, rsiChart]));
  rsiChart.timeScale().subscribeVisibleLogicalRangeChange((range) => syncRange(range, [mainChart, macdChart]));
  window.addEventListener('resize', resizeDashboard);

  // 5. 抓取資料並繪圖
  try {
    const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
    const response = await fetch(`${API_BASE_URL}/api/v1/klines`);
    if (!response.ok) throw new Error('API 失敗');
    const rawData: ApiResponse[] = await response.json();
    const times = rawData.map(item => item.time as Time);

    candlestickSeries.setData(rawData.map(item => ({ time: item.time as Time, open: item.open, high: item.high, low: item.low, close: item.close })));

    ma7Series.setData(rawData.map((item, i) => ({ time: times[i], value: item.ma7 })));
    ma25Series.setData(rawData.map((item, i) => ({ time: times[i], value: item.ma25 })));
    bbUpperSeries.setData(rawData.map((item, i) => ({ time: times[i], value: item.bbiUpper })));
    bbMiddleSeries.setData(rawData.map((item, i) => ({ time: times[i], value: item.bbiMiddle })));
    bbLowerSeries.setData(rawData.map((item, i) => ({ time: times[i], value: item.bbiLower })));

    difSeries.setData(rawData.map((item, i) => ({ time: times[i], value: item.dif })));
    deaSeries.setData(rawData.map((item, i) => ({ time: times[i], value: item.dea })));
    histSeries.setData(rawData.map((item, i) => ({ time: times[i], value: item.hist, color: item.hist >= 0 ? '#26a69a' : '#ef5350' })));

    rsiSeries.setData(rawData.map((item, i) => ({ time: times[i], value: item.rsi })));
    rsi30Series.setData(rawData.map((_, i) => ({ time: times[i], value: 30 })));
    rsi70Series.setData(rawData.map((_, i) => ({ time: times[i], value: 70 })));

    mainChart.timeScale().fitContent();
    const logicalRange = mainChart.timeScale().getVisibleLogicalRange();
    if (logicalRange) {
      if (macdChart) macdChart.timeScale().setVisibleLogicalRange(logicalRange);
      if (rsiChart) rsiChart.timeScale().setVisibleLogicalRange(logicalRange);
    }
  } catch (error) {
    console.error('抓取失敗:', error);
  }
});
</script>

<style scoped>
/* 樣式保持原樣即可 */
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

.btn-group {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

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
  position: relative;
  /* 👈 這行是關鍵，讓內部圖例有定位基準 */
  width: 100%;
  height: 400px;
  background-color: white;
  border-top-left-radius: 8px;
  border-top-right-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  overflow: hidden;
  border-bottom: 1px solid #f0f0f0;
}

/* 新增圖例的樣式 */
.chart-legend {
  position: absolute;
  top: 12px;
  left: 15px;
  z-index: 10;
  /* 確保它浮在畫布的最上層 */
  display: flex;
  gap: 12px;
  font-size: 13px;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Arial, sans-serif;

  /* 這個屬性超重要：讓滑鼠游標「穿透」文字，才不會擋住圖表的拖曳跟十字線！ */
  pointer-events: none;
}

.legend-title {
  font-weight: 700;
  color: #131722;
}

.legend-item {
  font-weight: 600;
}

/* 顏色與你 Go 後端 / TS 設定的顏色完全對齊 */
.legend-item.ma7 {
  color: #ba55d3;
}

/* 紫色 */
.legend-item.ma25 {
  color: #4169e1;
}

/* 藍色 */
.legend-item.bb {
  color: #ffb300;
}

/* 以布林中軌的黃色做代表 */

.macd-box,
.rsi-box {
  width: 100%;
  height: 120px;
  background-color: white;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.05);
  overflow: hidden;
  border-bottom: 1px solid #f0f0f0;
}

.rsi-box {
  border-bottom-left-radius: 8px;
  border-bottom-right-radius: 8px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}
</style>