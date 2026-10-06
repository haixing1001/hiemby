# emby-server

Go 语言编写的 Emby 兼容媒体服务器。播放使用直连/重定向，**不做视频转码**。

## 功能

- SQLite 数据库（pure Go，无 CGO，多架构友好）
- Emby 兼容 API，可连接常用 Emby 客户端
- Vue3 + Vite Web UI：海报墙、影视详情、播放器、后台管理
- 媒体库全量扫描 / 增量刷新 / 目录管理
- NFO 解析、本地海报/背景图读取
- ffprobe 媒体信息提取（分辨率、编码、时长）
- TMDB 元数据刮削
- 外挂字幕自动识别，服务端转 WebVTT
- 用户管理、播放权限、设备数量限制
- 文件管理（浏览/删除）
- 实时日志（SSE）
- 多架构 Docker 镜像：amd64 / arm64

## 快速开始（Docker Compose）

```bash
ADMIN_PASSWORD=yourpass MEDIA_PATH=/path/to/media \
  docker compose up -d --build
```

打开 http://localhost:8096 ，用 `admin` / 密码登录。

多架构构建：

```bash
docker buildx build --platform linux/amd64,linux/arm64 -t emby-server:latest .
```

## 本地开发

```bash
# 后端
go build -o emby-server .
./emby-server -addr :8096 -data ./data -admin-password admin123

# 前端
cd web && npm install && npm run dev   # 开发
npm run build                          # 产物到 web/dist，后端自动托管
```

## 目录结构

```
internal/
  config/   配置（flags + 环境变量）
  db/       SQLite（用户、token、媒体库、媒体项、字幕、播放记录）
  api/      Emby 兼容 API + 管理 API
  scanner/  媒体库扫描（全量/增量）、ffprobe 提取
  nfo/      Kodi NFO 解析
  tmdb/     TMDB 刮削客户端
  subs/     外挂字幕发现
  log/      日志 + 实时广播
web/        Vue3 + Vite 前端
```

## 环境变量

| 变量 | 说明 |
|---|---|
| ADMIN_PASSWORD | 初始 admin 密码 |
| TMDB_API_KEY | TMDB 刮削密钥 |
| MEDIA_DIRS | 逗号分隔的媒体目录（首次启动自动建库） |
