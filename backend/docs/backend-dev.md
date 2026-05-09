# Backend Development

当前已实现后端基础骨架，以及 LLM Provider / Model、Prompt 模板的配置类接口。
已实现错题库、番茄钟记录、学习日志、学习计划和日历聚合接口。尚未实现 OCR、题库、PDF、申论批改等业务接口。

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

## 1. 启动 PostgreSQL

```bash
cd backend
cp .env.example .env
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

## 3. 测试接口

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
| `DB_HOST` | PostgreSQL 主机 | `localhost` |
| `DB_PORT` | PostgreSQL 宿主机端口 | `21432` |
| `DB_USER` | 数据库用户 | `gkweb` |
| `DB_PASSWORD` | 数据库密码 | `gkweb_password` |
| `DB_NAME` | 数据库名 | `gkweb` |
| `DB_SSLMODE` | SSL 模式 | `disable` |
| `DB_TIMEZONE` | 数据库时区 | `Asia/Shanghai` |

## 5. 下一阶段建议

下一阶段建议继续做细节完善：

1. 把前端错题库页接到真实 `/api/mistakes`。
2. 给学习计划增加编辑弹窗和阶段筛选。
3. 给学习日志增加周/月统计。
4. 开始接 OCR 识题流程。
