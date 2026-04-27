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
| `lunar` | 阴历年月日、闰月标记、年月日干支、中文展示字段 |
| `solar_term` | 当前所处节气区间（名称、是否节气日、开始日期、第几天） |
| `supplement` | 补充信息（预留，当前为 null） |

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
    "day_ganzhi": "甲子",
    "month_display": "三月",
    "day_display": "初四",
    "display": "三月初四",
    "year_display": "丙午年三月初四"
  },
  "solar_term": {
    "name": "谷雨",
    "is_term_day": true,
    "start_date": "2026-04-20",
    "day_in_term": 1
  }
}
```

```bash
# 非节气日 — 返回当前所处节气区间
curl http://localhost:8900/api/v1/date/2026-04-21
```

```json
{
  "solar_term": {
    "name": "谷雨",
    "is_term_day": false,
    "start_date": "2026-04-20",
    "day_in_term": 2
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

### 日期范围查询

```
GET /api/v1/range?from=2026-04-01&to=2026-04-05&fields=basic,lunar,solar_term
```

**查询参数：**

| 参数 | 说明 |
|------|------|
| `from` | 起始日期（含），格式 `YYYY-MM-DD`，必填 |
| `to` | 结束日期（含），格式 `YYYY-MM-DD`，必填 |
| `fields` | 字段组选择，同单日查询 |

**约束：** `from` ≤ `to`，范围不超过 366 天。超出数据覆盖范围的日期在响应中省略。

```bash
curl 'http://localhost:8900/api/v1/range?from=2026-04-19&to=2026-04-22&fields=basic,solar_term'
```

```json
{
  "data_version": "2025a",
  "from": "2026-04-19",
  "to": "2026-04-22",
  "dates": [
    { "date": "2026-04-19", "weekday": "星期日", "solar_term": { "name": "清明", "is_term_day": false, "start_date": "2026-04-05", "day_in_term": 15 } },
    { "date": "2026-04-20", "weekday": "星期一", "solar_term": { "name": "谷雨", "is_term_day": true, "start_date": "2026-04-20", "day_in_term": 1 } },
    { "date": "2026-04-21", "weekday": "星期二", "solar_term": { "name": "谷雨", "is_term_day": false, "start_date": "2026-04-20", "day_in_term": 2 } },
    { "date": "2026-04-22", "weekday": "星期三", "solar_term": { "name": "谷雨", "is_term_day": false, "start_date": "2026-04-20", "day_in_term": 3 } }
  ]
}
```

### 年度节气列表

```
GET /api/v1/solar-terms?year=2026
```

**查询参数：**

| 参数 | 说明 |
|------|------|
| `year` | 年份，必填 |

```bash
curl http://localhost:8900/api/v1/solar-terms?year=2026
```

```json
{
  "data_version": "2025a",
  "year": 2026,
  "terms": [
    { "name": "小寒", "date": "2026-01-05", "month_display": "冬月" },
    { "name": "大寒", "date": "2026-01-20", "month_display": "腊月" },
    { "name": "立春", "date": "2026-02-04", "month_display": "腊月" },
    ...
    { "name": "冬至", "date": "2026-12-22", "month_display": "冬月" }
  ]
}
```

### 错误响应

| 状态码 | 场景 |
|--------|------|
| 400 | 日期格式无效、参数缺失、范围无效、未知字段名 |
| 404 | 日期或年份超出支持范围 |

```json
{"error": "invalid_date", "message": "日期格式应为 YYYY-MM-DD"}
{"error": "invalid_range", "message": "from must be <= to"}
{"error": "invalid_year", "message": "year 应为有效年份"}
```

## 支持范围

- 日期范围：2025-01-01 至 2027-12-31（1095 天）
- 节气时区：北京时间（UTC+8）
- 支持 CORS，可直接从浏览器调用

## 部署

### PM2

```bash
CGO_ENABLED=0 go build -o bin/server ./cmd/server/
mkdir -p /opt/jievow-calendar
cp bin/server data /opt/jievow-calendar/ -r
pm2 start /opt/jievow-calendar/server --name jievow-calendar --cwd /opt/jievow-calendar -- --port 8900
pm2 save
```

更新版本时：

```bash
CGO_ENABLED=0 go build -o bin/server ./cmd/server/ && go run ./cmd/datagen/
cp bin/server /opt/jievow-calendar/ && cp -r data /opt/jievow-calendar/
pm2 restart jievow-calendar
```

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
