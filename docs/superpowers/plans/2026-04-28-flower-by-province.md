# 节气花卉查询 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use summ:subagent-driven-development (recommended) or summ:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 新增 `GET /api/v1/flowers` 端点，支持按省份和节气查询代表性花卉。

**Architecture:** 平行于现有 calendar 管线新增花卉管线：`data/flowers.json` → `FlowerStore`（内存索引）→ `HandleFlowers` HTTP handler。数据为手动维护的文化数据，非天文计算。

**Tech Stack:** Go 1.22+ stdlib only（`net/http`, `encoding/json`, `crypto/sha256`），无新依赖。

---

## File Structure

| File | Action | Responsibility |
|------|--------|----------------|
| `calendar/flowers.go` | Create | FlowerRecord 类型、ValidSolarTerms 常量、FlowerStore 加载/查询 |
| `calendar/flowers_test.go` | Create | Store 层单元测试 |
| `api/flowers.go` | Create | HandleFlowers HTTP handler、参数校验、响应构建 |
| `api/flowers_test.go` | Create | Handler 层单元测试 |
| `api/handler.go` | Modify | Handler 增加 flowers 字段、NewHandler 增加参数 |
| `api/handler_test.go` | Modify | 所有 NewHandler 调用增加 nil 参数 |
| `api/integration_test.go` | Modify | NewHandler 调用增加 nil 参数、新增花卉集成测试 |
| `cmd/server/main.go` | Modify | 加载 FlowerStore、注册路由、更新 NewHandler 调用 |
| `data/flowers.json` | Create | 816 条花卉数据 |
| `data/flowers.json.sha256` | Create | SHA256 校验文件 |

---

### Task 1: FlowerRecord 类型、常量、FlowerStore

**Files:**
- Create: `calendar/flowers.go`
- Create: `calendar/flowers_test.go`

- [ ] **Step 1: 写失败测试**

```go
// calendar/flowers_test.go
package calendar

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestValidSolarTerms(t *testing.T) {
	if len(ValidSolarTerms) != 24 {
		t.Errorf("ValidSolarTerms len want 24 got %d", len(ValidSolarTerms))
	}
	lookup := map[string]bool{}
	for _, term := range ValidSolarTerms {
		lookup[term] = true
	}
	for _, term := range []string{"立春", "雨水", "惊蛰", "春分", "清明", "谷雨",
		"立夏", "小满", "芒种", "夏至", "小暑", "大暑",
		"立秋", "处暑", "白露", "秋分", "寒露", "霜降",
		"立冬", "小雪", "大雪", "冬至", "小寒", "大寒"} {
		if !lookup[term] {
			t.Errorf("missing solar term %q", term)
		}
	}
}

func TestFlowerStoreLoadAndQuery(t *testing.T) {
	dir := t.TempDir()
	records := []FlowerRecord{
		{Province: "北京", SolarTerm: "立春", Flowers: []string{"梅花", "山茶花"}},
		{Province: "浙江", SolarTerm: "立春", Flowers: []string{"梅花", "水仙"}},
		{Province: "北京", SolarTerm: "雨水", Flowers: []string{"迎春花", "玉兰花"}},
	}
	data, _ := json.MarshalIndent(records, "", "  ")
	dataFile := filepath.Join(dir, "flowers.json")
	dataWithNewline := append(data, '\n')
	os.WriteFile(dataFile, dataWithNewline, 0o644)

	hash := sha256.Sum256(dataWithNewline)
	checksumLine := fmt.Sprintf("%x  flowers.json\n", hash)
	os.WriteFile(filepath.Join(dir, "flowers.json.sha256"), []byte(checksumLine), 0o644)

	store, err := LoadFlowerStore(dir)
	if err != nil {
		t.Fatalf("LoadFlowerStore: %v", err)
	}

	flowers, ok := store.GetFlowers("北京", "立春")
	if !ok {
		t.Fatal("expected to find 北京 立春")
	}
	if len(flowers) != 2 || flowers[0] != "梅花" {
		t.Errorf("flowers want [梅花 山茶花] got %v", flowers)
	}

	flowers, ok = store.GetFlowers("浙江", "立春")
	if !ok {
		t.Fatal("expected to find 浙江 立春")
	}
	if flowers[0] != "梅花" {
		t.Errorf("first flower want 梅花 got %s", flowers[0])
	}

	_, ok = store.GetFlowers("上海", "立春")
	if ok {
		t.Error("expected not to find 上海 立春")
	}
}

func TestFlowerStoreListProvinces(t *testing.T) {
	dir := t.TempDir()
	records := []FlowerRecord{
		{Province: "浙江", SolarTerm: "立春", Flowers: []string{"梅花"}},
		{Province: "北京", SolarTerm: "立春", Flowers: []string{"梅花"}},
		{Province: "广东", SolarTerm: "立春", Flowers: []string{"桃花"}},
	}
	data, _ := json.MarshalIndent(records, "", "  ")
	dataFile := filepath.Join(dir, "flowers.json")
	dataWithNewline := append(data, '\n')
	os.WriteFile(dataFile, dataWithNewline, 0o644)

	hash := sha256.Sum256(dataWithNewline)
	checksumLine := fmt.Sprintf("%x  flowers.json\n", hash)
	os.WriteFile(filepath.Join(dir, "flowers.json.sha256"), []byte(checksumLine), 0o644)

	store, err := LoadFlowerStore(dir)
	if err != nil {
		t.Fatalf("LoadFlowerStore: %v", err)
	}

	provinces := store.ListProvinces()
	if len(provinces) != 3 {
		t.Fatalf("provinces want 3 got %d", len(provinces))
	}
	if provinces[0] != "北京" {
		t.Errorf("first province want 北京 got %s", provinces[0])
	}
}

func TestFlowerStoreBadChecksum(t *testing.T) {
	dir := t.TempDir()
	records := []FlowerRecord{{Province: "北京", SolarTerm: "立春", Flowers: []string{"梅花"}}}
	data, _ := json.Marshal(records)
	os.WriteFile(filepath.Join(dir, "flowers.json"), data, 0o644)
	os.WriteFile(filepath.Join(dir, "flowers.json.sha256"), []byte("0000  flowers.json\n"), 0o644)

	_, err := LoadFlowerStore(dir)
	if err == nil {
		t.Fatal("expected checksum error")
	}
}

func TestNewFlowerStore(t *testing.T) {
	records := []FlowerRecord{
		{Province: "北京", SolarTerm: "立春", Flowers: []string{"梅花"}},
	}
	store := NewFlowerStore(records)
	flowers, ok := store.GetFlowers("北京", "立春")
	if !ok {
		t.Fatal("expected to find 北京 立春")
	}
	if len(flowers) != 1 || flowers[0] != "梅花" {
		t.Errorf("flowers want [梅花] got %v", flowers)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./calendar/ -run "TestValidSolarTerms|TestFlowerStore|TestNewFlowerStore" -v`
Expected: FAIL（未定义的符号）

- [ ] **Step 3: 实现 calendar/flowers.go**

```go
// calendar/flowers.go
package calendar

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type FlowerRecord struct {
	Province   string   `json:"province"`
	SolarTerm  string   `json:"solar_term"`
	Flowers    []string `json:"flowers"`
}

var ValidSolarTerms = []string{
	"小寒", "大寒", "立春", "雨水", "惊蛰", "春分",
	"清明", "谷雨", "立夏", "小满", "芒种", "夏至",
	"小暑", "大暑", "立秋", "处暑", "白露", "秋分",
	"寒露", "霜降", "立冬", "小雪", "大雪", "冬至",
}

var validSolarTermSet map[string]bool

func init() {
	validSolarTermSet = make(map[string]bool, len(ValidSolarTerms))
	for _, t := range ValidSolarTerms {
		validSolarTermSet[t] = true
	}
}

func IsValidSolarTerm(term string) bool {
	return validSolarTermSet[term]
}

type FlowerStore struct {
	byProvinceTerm map[[2]string][]string
	provinces      []string
}

func LoadFlowerStore(dataDir string) (*FlowerStore, error) {
	dataPath := filepath.Join(dataDir, "flowers.json")
	checksumPath := dataPath + ".sha256"

	if err := verifyChecksum(dataPath, checksumPath); err != nil {
		return nil, fmt.Errorf("checksum verification failed: %w", err)
	}

	data, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, fmt.Errorf("read data file: %w", err)
	}

	var records []FlowerRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("parse data file: %w", err)
	}

	return NewFlowerStore(records), nil
}

func NewFlowerStore(records []FlowerRecord) *FlowerStore {
	byProvinceTerm := make(map[[2]string][]string, len(records))
	provinceSet := make(map[string]bool)
	for _, r := range records {
		key := [2]string{r.Province, r.SolarTerm}
		byProvinceTerm[key] = r.Flowers
		provinceSet[r.Province] = true
	}

	provinces := make([]string, 0, len(provinceSet))
	for p := range provinceSet {
		provinces = append(provinces, p)
	}
	sort.Strings(provinces)

	return &FlowerStore{byProvinceTerm: byProvinceTerm, provinces: provinces}
}

func (s *FlowerStore) GetFlowers(province, solarTerm string) ([]string, bool) {
	flowers, ok := s.byProvinceTerm[[2]string{province, solarTerm}]
	return flowers, ok
}

func (s *FlowerStore) ListProvinces() []string {
	return s.provinces
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./calendar/ -run "TestValidSolarTerms|TestFlowerStore|TestNewFlowerStore" -v`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add calendar/flowers.go calendar/flowers_test.go
git commit -m "feat: add FlowerRecord type, ValidSolarTerms, and FlowerStore"
```

---

### Task 2: 修改 Handler 结构体集成 FlowerStore

**Files:**
- Modify: `api/handler.go:12-18`
- Modify: `api/handler_test.go`（所有 `NewHandler(...)` 调用加 nil 参数）
- Modify: `api/integration_test.go`（所有 `NewHandler(...)` 调用加 nil 参数）

- [ ] **Step 1: 修改 Handler 结构体和 NewHandler**

将 `api/handler.go` 中：

```go
type Handler struct {
	store *calendar.Store
}

func NewHandler(store *calendar.Store) *Handler {
	return &Handler{store: store}
}
```

改为：

```go
type Handler struct {
	store   *calendar.Store
	flowers *calendar.FlowerStore
}

func NewHandler(store *calendar.Store, flowers *calendar.FlowerStore) *Handler {
	return &Handler{store: store, flowers: flowers}
}
```

- [ ] **Step 2: 更新 api/handler_test.go 中所有 NewHandler 调用**

所有 `NewHandler(store)` 改为 `NewHandler(store, nil)`，涉及的函数：
- `TestHandlerBasicQuery` — `NewHandler(newTestStore(t), nil)`
- `TestHandlerNonSolarTermDay` — `NewHandler(newTestStore(t), nil)`
- `TestHandlerFieldSelection` — `NewHandler(newTestStore(t), nil)`
- `TestHandlerInvalidDate` — `NewHandler(newTestStore(t), nil)`
- `TestHandlerDateOutOfRange` — `NewHandler(newTestStore(t), nil)`
- `TestHandlerInvalidFields` — `NewHandler(newTestStore(t), nil)`
- `TestHandleRange` — `NewHandler(store, nil)`
- `TestHandleRangeValidation` — `NewHandler(store, nil)`
- `TestHandleSolarTerms` — `NewHandler(store, nil)`
- `TestHandleSolarTermsValidation` — `NewHandler(store, nil)`

- [ ] **Step 3: 更新 api/integration_test.go 中所有 NewHandler 调用**

`integration_test.go` 中 `NewHandler(store)` 改为 `NewHandler(store, nil)`。

- [ ] **Step 4: 更新 cmd/server/main.go 中 NewHandler 调用**

暂传 nil，Task 4 中替换为实际 FlowerStore：

```go
h := api.NewHandler(store, nil)
```

- [ ] **Step 5: 运行全部测试确认无破坏**

Run: `go test ./... -v`
Expected: 全部 PASS

- [ ] **Step 6: 提交**

```bash
git add api/handler.go api/handler_test.go api/integration_test.go cmd/server/main.go
git commit -m "feat: add FlowerStore field to Handler, update NewHandler signature"
```

---

### Task 3: HandleFlowers API handler

**Files:**
- Create: `api/flowers.go`
- Create: `api/flowers_test.go`

- [ ] **Step 1: 写失败测试**

```go
// api/flowers_test.go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"jievow-calendar/calendar"
)

func newTestFlowerStore(t *testing.T) *calendar.FlowerStore {
	t.Helper()
	records := []calendar.FlowerRecord{
		{Province: "北京", SolarTerm: "立春", Flowers: []string{"梅花", "山茶花"}},
		{Province: "浙江", SolarTerm: "立春", Flowers: []string{"梅花", "水仙"}},
		{Province: "北京", SolarTerm: "雨水", Flowers: []string{"迎春花"}},
	}
	return calendar.NewFlowerStore(records)
}

func TestHandleFlowers(t *testing.T) {
	store := calendar.NewStore("testv", nil)
	flowerStore := newTestFlowerStore(t)
	h := NewHandler(store, flowerStore)

	req := httptest.NewRequest("GET", "/api/v1/flowers?solar_term=立春&province=北京", nil)
	w := httptest.NewRecorder()
	h.HandleFlowers(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status want 200 got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["solar_term"] != "立春" {
		t.Errorf("solar_term want 立春 got %v", resp["solar_term"])
	}
	if resp["province"] != "北京" {
		t.Errorf("province want 北京 got %v", resp["province"])
	}
	flowers := resp["flowers"].([]any)
	if len(flowers) != 2 {
		t.Fatalf("flowers want 2 got %d", len(flowers))
	}
	if flowers[0] != "梅花" {
		t.Errorf("first flower want 梅花 got %v", flowers[0])
	}
}

func TestHandleFlowersDefaultProvince(t *testing.T) {
	store := calendar.NewStore("testv", nil)
	flowerStore := newTestFlowerStore(t)
	h := NewHandler(store, flowerStore)

	req := httptest.NewRequest("GET", "/api/v1/flowers?solar_term=立春", nil)
	w := httptest.NewRecorder()
	h.HandleFlowers(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status want 200 got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["province"] != "北京" {
		t.Errorf("default province want 北京 got %v", resp["province"])
	}
}

func TestHandleFlowersMissingSolarTerm(t *testing.T) {
	store := calendar.NewStore("testv", nil)
	flowerStore := newTestFlowerStore(t)
	h := NewHandler(store, flowerStore)

	req := httptest.NewRequest("GET", "/api/v1/flowers?province=北京", nil)
	w := httptest.NewRecorder()
	h.HandleFlowers(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status want 400 got %d", w.Code)
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "invalid_params" {
		t.Errorf("error want invalid_params got %v", resp["error"])
	}
}

func TestHandleFlowersInvalidSolarTerm(t *testing.T) {
	store := calendar.NewStore("testv", nil)
	flowerStore := newTestFlowerStore(t)
	h := NewHandler(store, flowerStore)

	req := httptest.NewRequest("GET", "/api/v1/flowers?solar_term=不存在", nil)
	w := httptest.NewRecorder()
	h.HandleFlowers(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status want 400 got %d", w.Code)
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "invalid_solar_term" {
		t.Errorf("error want invalid_solar_term got %v", resp["error"])
	}
	// 应包含有效节气列表
	if resp["valid_terms"] == nil {
		t.Error("missing valid_terms")
	}
}

func TestHandleFlowersInvalidProvince(t *testing.T) {
	store := calendar.NewStore("testv", nil)
	flowerStore := newTestFlowerStore(t)
	h := NewHandler(store, flowerStore)

	req := httptest.NewRequest("GET", "/api/v1/flowers?solar_term=立春&province=火星", nil)
	w := httptest.NewRecorder()
	h.HandleFlowers(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status want 404 got %d", w.Code)
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "province_not_found" {
		t.Errorf("error want province_not_found got %v", resp["error"])
	}
	if resp["valid_provinces"] == nil {
		t.Error("missing valid_provinces")
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./api/ -run TestHandleFlowers -v`
Expected: FAIL（HandleFlowers 方法不存在）

- [ ] **Step 3: 实现 api/flowers.go**

```go
// api/flowers.go
package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"jievow-calendar/calendar"
)

const defaultProvince = "北京"

func (h *Handler) HandleFlowers(w http.ResponseWriter, r *http.Request) {
	solarTerm := r.URL.Query().Get("solar_term")
	if solarTerm == "" {
		writeError(w, http.StatusBadRequest, "invalid_params", "solar_term 参数必填")
		return
	}

	if !calendar.IsValidSolarTerm(solarTerm) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error":       "invalid_solar_term",
			"message":     "无效的节气名称",
			"valid_terms": calendar.ValidSolarTerms,
		})
		return
	}

	province := r.URL.Query().Get("province")
	if province == "" {
		province = defaultProvince
	}

	flowers, ok := h.flowers.GetFlowers(province, solarTerm)
	if !ok {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{
			"error":           "province_not_found",
			"message":         "不支持的省份",
			"valid_provinces": h.flowers.ListProvinces(),
		})
		return
	}

	resp := map[string]any{
		"solar_term": solarTerm,
		"province":   province,
		"flowers":    flowers,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(resp)
}

func joinStrings(ss []string, sep string) string {
	return strings.Join(ss, sep)
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./api/ -run TestHandleFlowers -v`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add api/flowers.go api/flowers_test.go
git commit -m "feat: add HandleFlowers API handler with validation"
```

---

### Task 4: Server 路由注册

**Files:**
- Modify: `cmd/server/main.go`

- [ ] **Step 1: 更新 main.go**

将 `cmd/server/main.go` 中的：

```go
	store, err := calendar.LoadStore("data")
	if err != nil {
		log.Fatalf("failed to load data: %v", err)
	}
	log.Printf("Loaded %d records (version %s)", store.Len(), store.Version())

	mux := http.NewServeMux()
	h := api.NewHandler(store, nil)
```

改为：

```go
	store, err := calendar.LoadStore("data")
	if err != nil {
		log.Fatalf("failed to load data: %v", err)
	}
	log.Printf("Loaded %d records (version %s)", store.Len(), store.Version())

	flowerStore, err := calendar.LoadFlowerStore("data")
	if err != nil {
		log.Fatalf("failed to load flower data: %v", err)
	}
	log.Printf("Loaded flower data for %d provinces", len(flowerStore.ListProvinces()))

	mux := http.NewServeMux()
	h := api.NewHandler(store, flowerStore)
```

并在路由注册部分添加：

```go
	mux.HandleFunc("GET /api/v1/flowers", h.HandleFlowers)
```

- [ ] **Step 2: 此时 flowers.json 还不存在，暂不运行。跳过提交，等 Task 5 数据就绪后一并测试。**

---

### Task 5: 创建 flowers.json 数据文件

**Files:**
- Create: `data/flowers.json`
- Create: `data/flowers.json.sha256`

- [ ] **Step 1: 编写 flowers.json**

816 条记录（34 省 × 24 节气），数据依据：
- 二十四番花信风（小寒→谷雨，传统固定花信）
- 各省现代物候观察（反映地域差异：南方花期早、北方迟、高海拔偏少）

数据组织原则：
- 每省每节气 2-5 种花
- 省份名与 spec 中 34 省份列表一致
- 节气名与 `calendar.ValidSolarTerms` 严格一致
- JSON 格式：扁平数组，每条 `{province, solar_term, flowers}`
- 文件末尾换行

具体数据内容须基于中国传统花卉物候知识整理，覆盖全部 816 条。

- [ ] **Step 2: 生成 SHA256 校验文件**

Run: `cd data && sha256sum flowers.json > flowers.json.sha256`
Expected: 生成 `data/flowers.json.sha256`，格式 `<hash>  flowers.json`

- [ ] **Step 3: 启动服务器验证**

Run: `go run ./cmd/server/`
Expected: 日志输出 `Loaded flower data for 34 provinces`

- [ ] **Step 4: 手动测试端点**

Run:
```bash
curl "http://localhost:8080/api/v1/flowers?solar_term=立春&province=浙江"
curl "http://localhost:8080/api/v1/flowers?solar_term=立春"
curl "http://localhost:8080/api/v1/flowers?solar_term=不存在"
curl "http://localhost:8080/api/v1/flowers?solar_term=立春&province=火星"
```
Expected: 分别返回 200 + 花卉数据、200 + 默认北京、400 + 有效节气、404 + 有效省份

- [ ] **Step 5: 提交**

```bash
git add cmd/server/main.go data/flowers.json data/flowers.json.sha256
git commit -m "feat: add flower data (34 provinces × 24 solar terms) and register route"
```

---

### Task 6: 集成测试

**Files:**
- Modify: `api/integration_test.go`

- [ ] **Step 1: 新增花卉集成测试**

在 `api/integration_test.go` 中添加：

```go
func setupTestFlowerServer(t *testing.T) *httptest.Server {
	t.Helper()
	store, err := calendar.LoadStore("../data")
	if err != nil {
		t.Fatalf("load store: %v", err)
	}
	flowerStore, err := calendar.LoadFlowerStore("../data")
	if err != nil {
		t.Fatalf("load flower store: %v", err)
	}
	mux := http.NewServeMux()
	h := NewHandler(store, flowerStore)
	mux.Handle("GET /api/v1/date/{date}", h)
	mux.HandleFunc("GET /api/v1/flowers", h.HandleFlowers)
	return httptest.NewServer(CORS(mux))
}

func TestIntegrationFlowers(t *testing.T) {
	srv := setupTestFlowerServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/flowers?solar_term=立春&province=浙江")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status want 200 got %d", resp.StatusCode)
	}

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)
	if body["solar_term"] != "立春" {
		t.Errorf("solar_term want 立春 got %v", body["solar_term"])
	}
	if body["province"] != "浙江" {
		t.Errorf("province want 浙江 got %v", body["province"])
	}
	flowers, ok := body["flowers"].([]any)
	if !ok || len(flowers) == 0 {
		t.Error("flowers should be a non-empty array")
	}
}

func TestIntegrationFlowersDefaultProvince(t *testing.T) {
	srv := setupTestFlowerServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/flowers?solar_term=雨水")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status want 200 got %d", resp.StatusCode)
	}

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)
	if body["province"] != "北京" {
		t.Errorf("default province want 北京 got %v", body["province"])
	}
}

func TestIntegrationFlowersErrors(t *testing.T) {
	srv := setupTestFlowerServer(t)
	defer srv.Close()

	tests := []struct {
		name       string
		url        string
		wantStatus int
		wantError  string
	}{
		{"missing solar_term", srv.URL + "/api/v1/flowers?province=北京", http.StatusBadRequest, "invalid_params"},
		{"invalid solar_term", srv.URL + "/api/v1/flowers?solar_term=不存在", http.StatusBadRequest, "invalid_solar_term"},
		{"invalid province", srv.URL + "/api/v1/flowers?solar_term=立春&province=火星", http.StatusNotFound, "province_not_found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(tt.url)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status want %d got %d", tt.wantStatus, resp.StatusCode)
			}

			var body map[string]any
			json.NewDecoder(resp.Body).Decode(&body)
			if body["error"] != tt.wantError {
				t.Errorf("error want %q got %q", tt.wantError, body["error"])
			}
		})
	}
}
```

- [ ] **Step 2: 运行全部测试**

Run: `go test ./... -v`
Expected: 全部 PASS

- [ ] **Step 3: 提交**

```bash
git add api/integration_test.go
git commit -m "test: add integration tests for flowers endpoint"
```
