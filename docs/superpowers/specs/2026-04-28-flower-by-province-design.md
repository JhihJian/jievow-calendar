# 节气花卉查询功能设计

## 概述

新增「每个节气在不同省份开什么花」的查询能力。数据为 34 省 × 24 节气的花卉名称，静态预生成，通过独立端点查询。

## 数据结构

**文件：`data/flowers.json`**

```json
[
  {
    "province": "浙江",
    "solar_term": "立春",
    "flowers": ["梅花", "山茶花", "水仙"]
  }
]
```

- 扁平数组，每条记录 = 一个省份 + 一个节气 + 花卉名称列表
- 总量：34 省 × 24 节气 = 816 条记录
- 省份名使用全称（浙江、广东等），节气名与现有 `solar_term` 字段一致（立春、雨水等）
- 附带 `data/flowers.json.sha256` 校验文件，与 calendar.json 保持一致
- 数据为手动维护的文化数据，非通过 datagen 生成

## Store 层

**文件：`calendar/flowers.go`**

**类型：**
- `FlowerRecord` — 对应 JSON 单条记录的 struct（Province, SolarTerm, Flowers []string）

**FlowerStore：**
- 内部索引 `byProvinceTerm map[[2]string][]string` — (省份, 节气) → 花卉列表，O(1) 查询
- 内部 `provinces []string` — 去重排序后的省份列表，供参数校验和列举
- `LoadFlowerStore(path string)` — 从 JSON 加载 + SHA256 校验 + 构建索引

**查询方法：**
- `GetFlowers(province, solarTerm string) ([]string, bool)` — 精确查询某省某节气
- `ListProvinces() []string` — 返回所有支持的省份

## API 层

**文件：`api/flowers.go`**

**端点：`GET /api/v1/flowers`**

| 参数 | 必填 | 说明 |
|------|------|------|
| `solar_term` | 是 | 节气名称，如「立春」 |
| `province` | 否 | 省份名称，默认「北京」 |

**响应格式：**

```json
{
  "solar_term": "立春",
  "province": "北京",
  "flowers": ["梅花", "山茶花", "水仙"]
}
```

**错误处理：**
- 缺少 `solar_term` → 400
- `solar_term` 不是有效节气名 → 400，附带有效节气列表
- 指定了不存在的 `province` → 404，附带有效省份列表

## 数据内容

**数据来源：**
1. 传统花信风（二十四番花信风）— 小寒至谷雨，每节气三候各一花，全国通用
2. 现代物候观察 — 不同省份实际盛开花卉，按省逐个整理

**整理原则：**
- 每省每节气 2-5 种代表性花卉
- 优先广为人知、有观赏价值的花
- 体现南方花期偏早、北方偏晚的规律
- 高海拔地区（西藏、青海等）部分节气允许空数组

**34 省份：** 北京、天津、河北、山西、内蒙古、辽宁、吉林、黑龙江、上海、江苏、浙江、安徽、福建、江西、山东、河南、湖北、湖南、广东、广西、海南、重庆、四川、贵州、云南、西藏、陕西、甘肃、青海、宁夏、新疆、香港、澳门、台湾

## 集成方式

- `cmd/server/main.go` 中加载 FlowerStore 并注入 Handler
- `api/handler.go` 的 Handler 结构体增加 `flowers *calendar.FlowerStore` 字段
- 在 `cmd/server/main.go` 的路由注册中添加 `/api/v1/flowers` 路由
- 不修改现有 CalendarStore 和现有端点的任何逻辑

## 实现步骤概要

1. 编写 `data/flowers.json`（816 条记录）及 SHA256 校验文件
2. 实现 `calendar/flowers.go`（FlowerRecord、FlowerStore、加载、查询）
3. 实现 `api/flowers.go`（HTTP handler、参数校验、响应构建）
4. 修改 `cmd/server/main.go`（加载 FlowerStore、注册路由）
5. 编写测试（Store 层单元测试、API 层集成测试）
