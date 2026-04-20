# Jievow Calendar API

一个公开的日期/阴历/节气查询 API 服务。

## 快速开始

```bash
# 构建
go build -o bin/server ./cmd/server/

# 启动（默认 8080 端口）
./bin/server

# 指定端口
PORT=8900 ./bin/server
./bin/server -port 8900
```

## API

### 查询日期信息

```
GET /api/v1/date/{date}?fields=basic,lunar,solar_term
```

**路径参数：** `date` — 公历日期，格式 `YYYY-MM-DD`

**查询参数：** `fields` — 逗号分隔的字段组，默认 `basic,lunar,solar_term`

| 字段组 | 返回内容 |
|--------|---------|
| `basic` | 公历日期、星期几 |
| `lunar` | 阴历年月日、闰月标记、年月日干支 |
| `solar_term` | 节气名称（非节气日为 null） |
| `supplement` | 节气补充信息（预留，当前为 null） |

### 示例

```bash
# 节气日
curl http://localhost:8900/api/v1/date/2026-04-20
```

```json
{
  "data_version": "2025a",
  "date": "2026-04-20",
  "weekday": "星期一",
  "lunar": {
    "year": 2026,
    "month": 3,
    "day": 4,
    "is_leap_month": false,
    "year_ganzhi": "丙午",
    "month_ganzhi": "壬辰",
    "day_ganzhi": "甲子"
  },
  "solar_term": {
    "name": "谷雨"
  }
}
```

```bash
# 只要基础信息
curl http://localhost:8900/api/v1/date/2025-01-29?fields=basic
```

```json
{
  "data_version": "2025a",
  "date": "2025-01-29",
  "weekday": "星期三"
}
```

### 错误响应

| 状态码 | 场景 |
|--------|------|
| 400 | 日期格式无效、未知字段名 |
| 404 | 日期超出支持范围 |

```json
{"error": "invalid_date", "message": "日期格式应为 YYYY-MM-DD"}
```

## 支持范围

- 日期范围：2025-01-01 至 2027-12-31（1095 天）
- 节气时区：北京时间（UTC+8）
- 支持 CORS，可直接从浏览器调用

## 部署

### 系统服务（systemd）

创建 `/etc/systemd/system/jievow-calendar.service`：

```ini
[Unit]
Description=Jievow Calendar API
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/jievow-calendar
ExecStart=/opt/jievow-calendar/bin/server
Environment=PORT=8900
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now jievow-calendar
sudo systemctl status jievow-calendar
```

### Docker

```dockerfile
FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server/

FROM alpine:3.21
COPY --from=build /server /app/server
COPY --from=build /src/data /app/data
WORKDIR /app
ENV PORT=8900
EXPOSE 8900
ENTRYPOINT ["/app/server"]
```

```bash
docker build -t jievow-calendar .
docker run -d --name calendar -p 8900:8900 jievow-calendar
```

### Nginx 反代

```nginx
location /api/ {
    proxy_pass http://127.0.0.1:8900;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
}
```

## 数据更新

```bash
# 重新生成 2025-2027 数据
go run ./cmd/datagen/

# 生成后重启服务
sudo systemctl restart jievow-calendar
```

## 开发

```bash
go test ./...        # 运行全部测试
go build ./...       # 编译检查
go run ./cmd/server/ # 开发启动
```

## 项目结构

```
cmd/server/       API 服务入口
cmd/datagen/      离线数据生成工具
calendar/         数据类型、存储层
api/              HTTP Handler、CORS
data/             预生成数据文件（已提交到 Git）
testdata/         基准校验数据
```
