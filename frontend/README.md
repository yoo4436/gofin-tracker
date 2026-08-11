# GoFin-Tracker Frontend

Vue 3、TypeScript 與 Vite 組成的公開使用者介面。

## Local development

複製 `.env.example` 為 `.env.local`，依環境設定公開 API 位址：

```env
VITE_API_BASE_URL="http://localhost:8080"
```

```bash
npm install
npm run dev
```

## Daily report API boundary

前端只使用下列公開、唯讀 API：

- `GET /api/v1/reports`
- `GET /api/v1/reports/:id`

管理端的產生與發布 API 不屬於前端功能。不得在 `frontend/` 的程式碼、環境變數或部署設定中加入任何管理憑證；管理 Token 只能由後端或安全的管理工具持有。
