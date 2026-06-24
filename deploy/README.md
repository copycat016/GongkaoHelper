# 部署指南（VPS + Tailscale）

本项目最终产物是**一个可执行文件**：前端被嵌进 Go 后端，启动后在 `:21080` 上同时提供页面和 API。
不需要 nginx，不需要 Node 跑在服务器上。访问走 Tailscale 私有网络，**无需域名、无需公网 HTTPS、无需开放公网端口**。

```
你的设备 ──(Tailscale)──> VPS 上的 gkweb(:21080) ──> SQLite 文件
```

---

## 一次性准备（VPS 上）

```bash
# 1. 装运行/构建依赖
sudo apt update
sudo apt install -y golang poppler-utils git     # poppler-utils 用于 PDF 解析
curl -fsSL https://nodejs.org/... 或用 nvm 装 Node 20+   # 仅构建时需要

# 2. 装 Tailscale 并加入你的网络
curl -fsSL https://tailscale.com/install.sh | sh
sudo tailscale up        # 按提示在浏览器登录，授权这台机器

# 3. 拿到这台机器的 Tailscale 地址
tailscale ip -4          # 形如 100.x.y.z
```

## 部署步骤

```bash
# 1. 拉代码
sudo git clone https://github.com/copycat016/GongkaoHelper.git /opt/gkweb
cd /opt/gkweb

# 2. 配置后端环境变量（公网/共享环境必须设强密码）
cp backend/.env.example backend/.env
nano backend/.env
#   GIN_MODE=release
#   JWT_SECRET=<粘贴一长串随机字符，可用 `openssl rand -hex 32` 生成>
#   AUTH_BOOTSTRAP_PASSWORD=<你的管理员初始密码>
#   其余默认（SQLite）即可

# 3. 一键构建出单文件
bash deploy/build.sh        # 产出 /opt/gkweb/backend/gkweb

# 4. 装成开机自启服务
sudo cp deploy/gkweb.service /etc/systemd/system/gkweb.service
sudo systemctl daemon-reload
sudo systemctl enable --now gkweb
journalctl -u gkweb -f      # 看日志，确认启动成功（Ctrl+C 退出查看）
```

## 访问

- **HTTP（最简单）**：在你任意一台装了 Tailscale 的设备上浏览器打开
  `http://<上面的 100.x.y.z>:21080`
- **HTTPS（推荐，浏览器显示安全锁）**：在 VPS 上执行一次
  ```bash
  sudo tailscale serve --bg 21080
  ```
  之后用 `https://<主机名>.<你的 tailnet>.ts.net` 访问，证书由 Tailscale 自动签发。

首次用 `.env` 里设置的 `admin` / `AUTH_BOOTSTRAP_PASSWORD` 登录。

---

## 以后更新到新版本

```bash
cd /opt/gkweb
git pull
bash deploy/build.sh
sudo systemctl restart gkweb
```

## 数据与备份

- 数据库：`backend/data/gkweb.db`（SQLite 单文件）
- 上传文件：`backend/uploads/`
- 备份就是把这两样复制走即可。
