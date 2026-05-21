# 快速启动

## 技术栈

- **前端**：Vite + React 19 + Ant Design 6 + React Router 7
- **后端**：Go + Gin + GORM
- **数据库**：SQLite（默认）/ PostgreSQL
- **PDF 解析**：Poppler `pdftotext`（优先）+ Go `pdf` 库（回退）
- **OCR**：百度 OCR

## 1. 启动数据库（可选）

默认使用 SQLite，无需额外配置：

```bash
cd backend
cp .env.example .env
# 保持 DB_DRIVER=sqlite 即可
```

如需 PostgreSQL：

```bash
cd backend
cp .env.example .env
# 修改 .env：DB_DRIVER=postgres
docker compose --env-file .env up -d
```

## 2. 启动后端

```bash
cd backend
set -a && source .env && set +a
go run ./cmd/server
```

默认监听：`http://localhost:21080`

## 3. 启动前端

```bash
npm install
npm run dev -- --host 0.0.0.0 --port 21073
```

访问：`http://localhost:21073`

## 端口约定

| 服务 | 端口 |
|------|------|
| 前端 Vite | `21073` |
| 后端 Gin | `21080` |
| PostgreSQL | `21432` |

## 验证命令

```bash
# 后端测试
cd backend && go test ./...

# 前端构建
npm run build

# 接口冒烟
curl http://localhost:21080/api/health
curl http://localhost:21080/api/db/ping
```

## PDF 解析依赖（可选但推荐）

```bash
# Debian/Ubuntu
sudo apt install poppler-utils

# Fedora
sudo dnf install poppler-utils

# Arch
sudo pacman -S poppler

# 验证
pdftotext -v
```

未安装时会自动回退到 Go 库，但中文 PDF 稳定性较差；扫描件 PDF 必须走 OCR。

## OCR 配置

百度 OCR AK/SK 在 `/ai?tab=ocr` 配置页填写，后端保存到本地配置文件，不会提交到 Git。

## 主要页面

| 路由 | 页面 |
|------|------|
| `/` | 首页：今日 / 本周 / 阶段规划（三层任务规划 + DDL + 今日专注概况） |
| `/intake` | 统一录入器（图片 OCR / PDF / 粘贴文本） |
| `/essay` | 申论 PDF 结构化与批改 |
| `/pomodoro` | 番茄钟 |
| `/music` | 音乐播放器 |
| `/study` | 学习日志（含手动补登） |
| `/ai` | 配置（LLM / Prompt / OCR / PDF / 备份） |

> `/ocr`、`/questions`、`/mistakes` 旧入口会自动重定向到 `/intake`，`/plans`、`/calendar`、`/logs` 重定向到 `/study`。

## 下一步

- 维护开发参考：[`MAINTENANCE.md`](MAINTENANCE.md)
- 待办清单：[`TODO.md`](TODO.md)
- 后端接口示例：[`../backend/docs/backend-dev.md`](../backend/docs/backend-dev.md)
