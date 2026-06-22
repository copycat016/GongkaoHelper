# Backend Development

当前已实现后端基础骨架，以及 LLM Provider / Model、Prompt 模板的配置类接口。
已实现错题库、番茄钟记录、学习日志、学习计划、日历聚合、音乐歌单与上传、OCR（百度 OCR）、PDF 文本解析测试、申论 PDF 结构化与批改等业务接口。

## 目录结构

```txt
backend/
├─ cmd/server/main.go
├─ internal/config
├─ internal/database
├─ internal/middleware
├─ internal/response
├─ internal/routes
├─ internal/handlers
├─ internal/services
├─ internal/models
├─ uploads
├─ docs
├─ .env.example
├─ docker-compose.yml
├─ go.mod
└─ go.sum
```

## 1. 选择数据库

默认使用 SQLite，本地无需启动数据库服务：

```bash
cd backend
cp .env.example .env
```

`.env` 中保持：

```env
DB_DRIVER=sqlite
SQLITE_PATH=./data/gkweb.db
```

如需切换 PostgreSQL，将 `.env` 改为：

```env
DB_DRIVER=postgres
```

然后启动 PostgreSQL：

```bash
cd backend
docker compose --env-file .env up -d
```

检查容器：

```bash
docker compose --env-file .env ps
```

## 2. 启动后端

后端配置全部来自环境变量。可以手动导出，也可以让 shell 读取 `.env`：

```bash
cd backend
set -a
source .env
set +a
go run ./cmd/server
```

默认服务地址：

```txt
http://localhost:21080
```

## 2.1 启动前端联调

前端 Vite 已配置 `/api` 代理到 `http://localhost:21080`。先启动后端，再在项目根目录启动前端：

```bash
cd ..
npm run dev -- --host 0.0.0.0 --port 21073
```

打开：

```txt
http://localhost:21073/
```

当前已接入真实后端的页面：

- LLM 配置
- Prompt 配置
- 错题库（后端接口已完成，前端仍使用 mock 数据展示）
- 番茄钟
- 学习日志 / 学习计划 / 日历
- 音乐播放器（歌单、上传、播放）
- OCR 配置与识别（百度 OCR）
- PDF 解析测试
- 申论 PDF 上传、解析、chunk/section 分类、题目组装与批改（LLM 批改已接入真实模型，无 LLM 时回退启发式评分）

## 3. 测试接口

除 `/api/health`、`/api/db/ping`、`/api/auth/login` 外，业务接口需要先登录并携带 Bearer Token：

```bash
TOKEN=$(curl -s -X POST http://localhost:21080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"123456"}' \
  | sed -n 's/.*"access_token":"\([^"]*\)".*/\1/p')

curl -H "Authorization: Bearer $TOKEN" http://localhost:21080/api/llm/providers
```

健康检查：

```bash
curl http://localhost:21080/api/health
```

数据库连接检查：

```bash
curl http://localhost:21080/api/db/ping
```

LLM Provider：

```bash
curl http://localhost:21080/api/llm/providers

curl -X POST http://localhost:21080/api/llm/providers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "OpenAI",
    "provider_type": "openai-compatible",
    "base_url": "https://api.openai.com/v1",
    "api_key": "replace-with-env-or-secret-later",
    "enabled": true,
    "note": "高质量任务"
  }'
```

LLM Model：

```bash
curl -X POST http://localhost:21080/api/llm/models \
  -H "Content-Type: application/json" \
  -d '{
    "provider_id": 1,
    "provider": "OpenAI",
    "name": "gpt-example",
    "alias": "高质量解析",
    "cost_level": "高",
    "speed_level": "中",
    "quality_level": "高",
    "enabled": true,
    "use_quality": true,
    "use_essay": true
  }'
```

Prompt：

```bash
curl -X POST http://localhost:21080/api/prompts \
  -H "Content-Type: application/json" \
  -d '{
    "task_type": "申论批改",
    "name": "申论分项评分",
    "system_prompt": "你是严谨的申论批改老师。",
    "user_prompt": "请批改以下答案：{{answer}}",
    "variables": "answer: 用户答案",
    "default_model": "高质量解析",
    "version": "v1.0",
    "enabled": true
  }'
```

错题库：

```bash
curl http://localhost:21080/api/mistakes

curl -X POST http://localhost:21080/api/mistakes \
  -H "Content-Type: application/json" \
  -d '{
    "subject": "行测",
    "question_type": "资料分析",
    "sub_type": "增长率",
    "title": "资料分析增长率误选",
    "stem": "根据材料计算同比增长率。",
    "options": "{\"A\":\"10%\",\"B\":\"12.5%\",\"C\":\"15%\",\"D\":\"18%\"}",
    "correct_answer": "B",
    "user_answer": "C",
    "analysis": "增长率公式使用错误。",
    "error_reason": "计算错误",
    "mastery": "模糊",
    "tags": "[\"计算\",\"资料分析\"]",
    "source": "手动录入",
    "note": "复盘增长率公式"
  }'

curl -X POST http://localhost:21080/api/mistakes/1/review \
  -H "Content-Type: application/json" \
  -d '{
    "mastery": "基本掌握",
    "next_review_at": "2026-05-12T09:00:00+08:00",
    "note": "已复习一次"
  }'
```

番茄钟：

```bash
curl -X POST http://localhost:21080/api/pomodoro/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "task_type": "行测刷题",
    "task_name": "资料分析专项",
    "mode": "focus",
    "planned_minutes": 25,
    "actual_minutes": 25,
    "completed_at": "2026-05-08T16:30:00+08:00"
  }'

curl http://localhost:21080/api/pomodoro/stats/today
```

学习日志 / 计划 / 日历：

```bash
curl http://localhost:21080/api/logs
curl http://localhost:21080/api/logs/stats

curl -X POST http://localhost:21080/api/plans \
  -H "Content-Type: application/json" \
  -d '{
    "title": "资料分析专项",
    "plan_type": "每日计划",
    "subject": "行测",
    "target_min": 50,
    "due_date": "2026-05-08T23:59:00+08:00",
    "priority": "高",
    "status": "进行中"
  }'

curl http://localhost:21080/api/calendar/events?month=2026-05
```

OCR 识别：

```bash
curl -X POST http://localhost:21080/api/ocr/recognize \
  -F "file=@/path/to/image.png" \
  -F "scene=printed" \
  -F "engine=general_basic"
```

PDF 解析测试：

```bash
curl -X POST http://localhost:21080/api/pdf/parse-test \
  -F "file=@/path/to/test.pdf"
```

申论文档：

```bash
curl http://localhost:21080/api/essay/documents

curl -X POST http://localhost:21080/api/essay/documents \
  -F "file=@/path/to/essay.pdf" \
  -F "title=2026年申论真题" \
  -F "document_role=combined"

# 解析（可通过 boundary_model_id 指定 LLM 边界识别模型）
curl -X POST http://localhost:21080/api/essay/documents/1/parse

# 调试边界识别
curl -X POST http://localhost:21080/api/essay/documents/1/debug-boundary \
  -H "Content-Type: application/json" \
  -d '{"model_id": 1}'

# 查看 sections（语义区域，主流程使用）
curl http://localhost:21080/api/essay/documents/1/sections

# 查看 chunks（固定长度片段，兼容旧流程）
curl http://localhost:21080/api/essay/documents/1/chunks

curl -X POST http://localhost:21080/api/essay/documents/1/classify \
  -H "Content-Type: application/json" \
  -d '{"model_id": 1}'

curl -X POST http://localhost:21080/api/essay/documents/1/assemble

curl http://localhost:21080/api/essay/documents/1/questions

# 手动新增 / 编辑 / 删除题目
curl -X POST http://localhost:21080/api/essay/questions \
  -H "Content-Type: application/json" \
  -d '{"document_id": 1, "question_no": "1", "title": "第一题", "question_type": "归纳概括题", "question_text": "请概括材料中的问题。", "max_score": 20, "word_limit": 300}'

curl -X PUT http://localhost:21080/api/essay/questions/1 \
  -H "Content-Type: application/json" \
  -d '{"question_no": "1", "title": "第一题", "question_type": "提出对策题", "question_text": "请提出对策。", "max_score": 25, "word_limit": 400, "scoring_rubric": "按要点给分"}'

curl -X DELETE http://localhost:21080/api/essay/questions/1

# 编辑分段与重置题目关联
curl -X PUT http://localhost:21080/api/essay/sections/1 \
  -H "Content-Type: application/json" \
  -d '{"section_type": "material", "title": "材料一", "content": "材料内容", "related_question_nos": "1"}'

curl -X POST http://localhost:21080/api/essay/questions/1/relations \
  -H "Content-Type: application/json" \
  -d '{"material_ids": [1, 2], "answer_ids": [5]}'

# 批改（LLM 批改已接入，无 LLM 时回退启发式评分）
curl -X POST http://localhost:21080/api/essay/questions/1/review \
  -H "Content-Type: application/json" \
  -d '{"model_id": 1, "user_answer": "我的作答内容"}'
```

成功响应格式示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "status": "ok"
  },
  "request_id": "..."
}
```

## 4. 环境变量

| 变量 | 说明 | 默认示例 |
| --- | --- | --- |
| `SERVER_PORT` | 后端监听端口 | `21080` |
| `GIN_MODE` | Gin 运行模式 | `debug` |
| `JWT_SECRET` | JWT 签名密钥；`GIN_MODE=release` 时必须设置 | 空 |
| `JWT_EXPIRE_HOURS` | 登录 token 有效小时数 | `168` |
| `AUTH_MODE` | 鉴权模式，当前默认单人模式 | `single` |
| `AUTH_BOOTSTRAP_USERNAME` | 空库首次管理员账号 | `admin` |
| `AUTH_BOOTSTRAP_PASSWORD` | 空库首次管理员密码；release 模式必须设置强密码 | 空 |
| `AUTH_BOOTSTRAP_DISPLAY_NAME` | 空库首次管理员显示名 | `Owner` |
| `DB_DRIVER` | 数据库类型，支持 `sqlite` / `postgres` / `postgresql` | `sqlite` |
| `SQLITE_PATH` | SQLite 数据库文件路径 | `./data/gkweb.db` |
| `DB_HOST` | PostgreSQL 主机 | `localhost` |
| `DB_PORT` | PostgreSQL 宿主机端口 | `21432` |
| `DB_USER` | 数据库用户 | `gkweb` |
| `DB_PASSWORD` | 数据库密码 | `gkweb_password` |
| `DB_NAME` | 数据库名 | `gkweb` |
| `DB_SSLMODE` | SSL 模式 | `disable` |
| `DB_TIMEZONE` | 数据库时区 | `Asia/Shanghai` |

## 5. PDF 解析依赖说明

申论和 PDF 测试功能依赖文本解析，当前策略如下：

- 优先使用系统命令 `pdftotext`（来自 Poppler）。
- 找不到时回退到 Go 库 `github.com/ledongthuc/pdf`。
- 扫描件 PDF 文本层不可用，必须走 OCR 识别。

建议安装 Poppler：

```bash
# Debian / Ubuntu
sudo apt install poppler-utils

# Fedora
sudo dnf install poppler-utils

# Arch
sudo pacman -S poppler
```

验证：

```bash
pdftotext -v
```

## 6. 下一阶段建议

下一阶段建议继续做细节完善：

1. 申论 Prompt 按题型分发，并允许单题自定义模板。
2. 批改记录保存 prompt、raw request 和 raw response，方便回放排错。
3. 给录入器补扫描件 PDF 转图片 + OCR 流程。
4. 实现 OCR AI 修正接口和 Prompt 测试接口。
5. 引入用户认证和 root/admin 权限模型，替换当前 `X-User-ID` 临时方案。
