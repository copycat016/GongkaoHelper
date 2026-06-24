#!/usr/bin/env bash
# 构建单文件可执行程序：前端打包后嵌入 Go 后端，产出 backend/gkweb
# 用法：bash deploy/build.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "[1/3] 构建前端 (npm run build)..."
npm install
npm run build

echo "[2/3] 把前端 dist/ 嵌入后端 web/ 目录..."
WEB_DIR="backend/cmd/server/web"
find "$WEB_DIR" -mindepth 1 ! -name .gitkeep -delete
cp -r dist/* "$WEB_DIR"/

echo "[3/3] 编译后端单文件..."
cd backend
go build -o gkweb ./cmd/server

echo ""
echo "✅ 完成 -> backend/gkweb"
echo "   运行：cd backend && set -a && source .env && set +a && ./gkweb"
