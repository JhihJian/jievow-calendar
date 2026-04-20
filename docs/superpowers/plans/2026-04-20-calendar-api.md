# Calendar API Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use summ:subagent-driven-development (recommended) or summ:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a public Go API that returns lunar calendar and solar term data for a given Gregorian date, backed by a pre-generated JSON data file.

**Architecture:** Data-table-driven — an offline Go tool generates a JSON file covering 2025-01-01 to 2027-12-31 (~1096 records). The API server loads this file into memory at startup and serves queries via `GET /api/v1/date/{date}` with field selection. No database, no external runtime dependencies.

**Tech Stack:** Go 1.22+ (standard library HTTP routing with path parameters), `github.com/6tail/lunar-go` (data generation only, not imported at runtime).

---

## File Structure

```
jievow-calendar/
├── cmd/
│   ├── server/main.go          # API server entry point
│   └── datagen/main.go         # Offline data generation tool
├── calendar/
│   ├── types.go                # Data models, field groups, weekday helper
│   ├── store.go                # JSON loading, checksum, in-memory query
│   └── store_test.go           # Store unit tests
├── api/
│   ├── handler.go              # HTTP handler for GET /api/v1/date/{date}
│   ├── handler_test.go         # Handler unit tests
│   └── cors.go                 # CORS middleware
├── data/
│   ├── calendar.json           # Generated calendar data (committed)
│   └── calendar.json.sha256    # SHA256 checksum (committed)
├── testdata/
│   └── baselines.json          # Known-correct dates for accuracy tests
├── go.mod
└── go.sum
```

---

### Task 1: Project Scaffold & Data Types

**Files:**
- Create: `go.mod`
- Create: `calendar/types.go`

- [ ] **Step 1: Initialize Go module**

```bash
cd /data/dev/jievow-calendar
go mod init jievow-calendar
```

- [ ] **Step 2: Create `calendar/types.go`**

```go
package calendar

import "time"

const (
	FieldBasic     = "basic"
	FieldLunar     = "lunar"
	FieldSolarTerm = "solar_term"
	FieldSupplement = "supplement"
)

var ValidFields = map[string]bool{
	FieldBasic:      true,
	FieldLunar:      true,
	FieldSolarTerm:  true,
	FieldSupplement: true,
}

var DefaultFields = []string{FieldBasic, FieldLunar, FieldSolarTerm}

// DataFile is the top-level structure of calendar.json.
type DataFile struct {
	Version string           `json:"version"`
	Records []CalendarRecord `json:"records"`
}

// CalendarRecord represents one day's data in the data file.
type CalendarRecord struct {
	Date         string `json:"date"`
	LunarYear    int    `json:"lunar_year"`
	LunarMonth   int    `json:"lunar_month"`
	LunarDay     int    `json:"lunar_day"`
	IsLeapMonth  bool   `json:"is_leap_month"`
	YearGanzhi   string `json:"year_ganzhi"`
	MonthGanzhi  string `json:"month_ganzhi"`
	DayGanzhi    string `json:"day_ganzhi"`
	SolarTerm    string `json:"solar_term"`
}

var chineseWeekdays = [...]string{
	"星期日", "星期一", "星期二", "星期三",
	"星期四", "星期五", "星期六",
}

func WeekdayChinese(dateStr string) string {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return ""
	}
	return chineseWeekdays[t.Weekday()]
}

func ParseFields(raw string) []string {
	if raw == "" {
		return DefaultFields
	}
	return splitUnique(raw)
}

func splitUnique(s string) []string {
	seen := map[string]bool{}
	var result []string
	for _, f := range splitComma(s) {
		f = trimSpace(f)
		if f != "" && !seen[f] {
			seen[f] = true
			result = append(result, f)
		}
	}
	return result
}

func splitComma(s string) []string {
	var parts []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	return parts
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && s[start] == ' ' {
		start++
	}
	for end > start && s[end-1] == ' ' {
		end--
	}
	return s[start:end]
}
```

> Note: `splitComma` and `trimSpace` are hand-written to avoid importing `strings` and `strconv` for such simple logic. If the implementer prefers, `strings.Split` and `strings.TrimSpace` are fine alternatives.

- [ ] **Step 3: Verify it compiles**

Run: `go build ./calendar/`
Expected: no output (success)

- [ ] **Step 4: Commit**

```bash
git add go.mod go.sum calendar/types.go
git commit -m "feat: project scaffold with data types and helpers"
```

---

### Task 2: Data Generation Tool

**Files:**
- Create: `cmd/datagen/main.go`
- Create: `data/calendar.json` (generated)
- Create: `data/calendar.json.sha256` (generated)

- [ ] **Step 1: Add lunar-go dependency**

```bash
cd /data/dev/jievow-calendar
go get github.com/6tail/lunar-go/calendar
```

- [ ] **Step 2: Create `cmd/datagen/main.go`**

```go
package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"jievow-calendar/calendar"

	lunarcal "github.com/6tail/lunar-go/calendar"
)

func main() {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2027, 12, 31, 0, 0, 0, 0, time.UTC)

	var records []calendar.CalendarRecord
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		rec := generateRecord(d)
		records = append(records, rec)
	}

	dataFile := calendar.DataFile{
		Version: "2025a",
		Records: records,
	}

	writeJSON("data/calendar.json", dataFile)
	writeChecksum("data/calendar.json", "data/calendar.json.sha256")
	fmt.Printf("Generated %d records\n", len(records))
}

func generateRecord(d time.Time) calendar.CalendarRecord {
	year, month, day := d.Date()
	solar := lunarcal.NewSolar(year, int(month), day)
	lunar := solar.ToLunar()

	term := ""
	jq := lunar.GetJieQi()
	if jq != "" {
		term = jq
	}

	return calendar.CalendarRecord{
		Date:        d.Format("2006-01-02"),
		LunarYear:   lunar.GetYear(),
		LunarMonth:  lunar.GetMonth(),
		LunarDay:    lunar.GetDay(),
		IsLeapMonth: lunar.IsLeapMonth(),
		YearGanzhi:  solar.GetYearInGanZhi(),
		MonthGanzhi: solar.GetMonthInGanZhi(),
		DayGanzhi:   solar.GetDayInGanZhi(),
		SolarTerm:   term,
	}
}

func writeJSON(path string, v any) {
	os.MkdirAll("data", 0o755)
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal error: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write error: %v\n", err)
		os.Exit(1)
	}
}

func writeChecksum(dataPath, checksumPath string) {
	data, err := os.ReadFile(dataPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read error: %v\n", err)
		os.Exit(1)
	}
	hash := sha256.Sum256(data)
	checksum := fmt.Sprintf("%x  calendar.json\n", hash)
	if err := os.WriteFile(checksumPath, []byte(checksum), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write checksum error: %v\n", err)
		os.Exit(1)
	}
}
```

> **Library API note:** The exact method names for `github.com/6tail/lunar-go/calendar` may differ across versions. The engineer should verify the API against the library's documentation at `pkg.go.dev/github.com/6tail/lunar-go`. Key methods used:
> - `lunarcal.NewSolar(y, m, d)` → create Solar date
> - `solar.ToLunar()` → get Lunar object
> - `lunar.GetYear()`, `.GetMonth()`, `.GetDay()`, `.IsLeapMonth()` → lunar date
> - `solar.GetYearInGanZhi()`, `.GetMonthInGanZhi()`, `.GetDayInGanZhi()` → GanZhi
> - `lunar.GetJieQi()` → solar term name (empty string if not a JieQi day)
>
> If the API differs, adjust method calls accordingly. The test in Task 3 will catch incorrect data.

- [ ] **Step 3: Run the generator**

```bash
cd /data/dev/jievow-calendar
go run ./cmd/datagen/
```

Expected output: `Generated 1096 records`

- [ ] **Step 4: Verify output**

```bash
# Check file exists and has reasonable size
ls -lh data/calendar.json
# Spot-check a known date (2026-04-20 should show 谷雨)
grep '2026-04-20' data/calendar.json
# Verify checksum file exists
cat data/calendar.json.sha256
```

Expected: `data/calendar.json` is ~200-300KB, the grep shows a record with `"solar_term": "谷雨"`.

- [ ] **Step 5: Commit**

```bash
git add cmd/datagen/main.go data/calendar.json data/calendar.json.sha256 go.mod go.sum
git commit -m "feat: data generation tool with 2025-2027 calendar data"
```

---

### Task 3: Baseline Accuracy Tests

**Files:**
- Create: `testdata/baselines.json`
- Create: `calendar/accuracy_test.go`

- [ ] **Step 1: Create `testdata/baselines.json`**

These are known-correct values verified against authoritative sources. The engineer **must** cross-check these against an online 万年历 (e.g., 香港天文台 or 紫金山天文台) before relying on them.

```json
[
  {
    "date": "2025-01-29",
    "lunar_year": 2025,
    "lunar_month": 1,
    "lunar_day": 1,
    "is_leap_month": false,
    "year_ganzhi": "乙巳",
    "solar_term": ""
  },
  {
    "date": "2025-02-03",
    "lunar_year": 2025,
    "lunar_month": 1,
    "lunar_day": 6,
    "is_leap_month": false,
    "solar_term": "立春"
  },
  {
    "date": "2025-05-31",
    "lunar_year": 2025,
    "lunar_month": 5,
    "lunar_day": 5,
    "is_leap_month": false,
    "solar_term": ""
  },
  {
    "date": "2025-10-06",
    "lunar_year": 2025,
    "lunar_month": 8,
    "lunar_day": 14,
    "is_leap_month": false,
    "solar_term": ""
  },
  {
    "date": "2026-01-20",
    "solar_term": "大寒"
  },
  {
    "date": "2026-02-17",
    "lunar_year": 2026,
    "lunar_month": 1,
    "lunar_day": 1,
    "is_leap_month": false,
    "year_ganzhi": "丙午",
    "solar_term": ""
  },
  {
    "date": "2026-04-20",
    "solar_term": "谷雨"
  },
  {
    "date": "2027-01-05",
    "solar_term": "小寒"
  }
]
```

> Only fields present in the baseline are checked. Missing fields are skipped in the comparison.

- [ ] **Step 2: Create `calendar/accuracy_test.go`**

```go
package calendar

import (
	"encoding/json"
	"os"
	"testing"
)

type baseline struct {
	Date        string `json:"date"`
	LunarYear   int    `json:"lunar_year,omitempty"`
	LunarMonth  int    `json:"lunar_month,omitempty"`
	LunarDay    int    `json:"lunar_day,omitempty"`
	IsLeapMonth *bool  `json:"is_leap_month,omitempty"`
	YearGanzhi  string `json:"year_ganzhi,omitempty"`
	SolarTerm   string `json:"solar_term,omitempty"`
}

func TestAccuracy(t *testing.T) {
	records := loadDataFile(t)
	baselines := loadBaselines(t)

	for _, b := range baselines {
		rec, ok := records[b.Date]
		if !ok {
			t.Errorf("date %s not found in data file", b.Date)
			continue
		}
		if b.LunarYear != 0 && rec.LunarYear != b.LunarYear {
			t.Errorf("%s: lunar_year want %d got %d", b.Date, b.LunarYear, rec.LunarYear)
		}
		if b.LunarMonth != 0 && rec.LunarMonth != b.LunarMonth {
			t.Errorf("%s: lunar_month want %d got %d", b.Date, b.LunarMonth, rec.LunarMonth)
		}
		if b.LunarDay != 0 && rec.LunarDay != b.LunarDay {
			t.Errorf("%s: lunar_day want %d got %d", b.Date, b.LunarDay, rec.LunarDay)
		}
		if b.IsLeapMonth != nil && rec.IsLeapMonth != *b.IsLeapMonth {
			t.Errorf("%s: is_leap_month want %v got %v", b.Date, *b.IsLeapMonth, rec.IsLeapMonth)
		}
		if b.YearGanzhi != "" && rec.YearGanzhi != b.YearGanzhi {
			t.Errorf("%s: year_ganzhi want %s got %s", b.Date, b.YearGanzhi, rec.YearGanzhi)
		}
		if b.SolarTerm != "" && rec.SolarTerm != b.SolarTerm {
			t.Errorf("%s: solar_term want %q got %q", b.Date, b.SolarTerm, rec.SolarTerm)
		}
	}
}

func loadDataFile(t *testing.T) map[string]CalendarRecord {
	t.Helper()
	data, err := os.ReadFile("../data/calendar.json")
	if err != nil {
		t.Fatalf("read data file: %v", err)
	}
	var df DataFile
	if err := json.Unmarshal(data, &df); err != nil {
		t.Fatalf("parse data file: %v", err)
	}
	m := make(map[string]CalendarRecord, len(df.Records))
	for _, r := range df.Records {
		m[r.Date] = r
	}
	return m
}

func loadBaselines(t *testing.T) []baseline {
	t.Helper()
	data, err := os.ReadFile("../testdata/baselines.json")
	if err != nil {
		t.Fatalf("read baselines: %v", err)
	}
	var bs []baseline
	if err := json.Unmarshal(data, &bs); err != nil {
		t.Fatalf("parse baselines: %v", err)
	}
	return bs
}
```

- [ ] **Step 3: Run accuracy tests**

```bash
cd /data/dev/jievow-calendar
go test ./calendar/ -run TestAccuracy -v
```

Expected: PASS. If any baseline values are wrong, update `testdata/baselines.json` to match the authoritative source, not the generated data.

- [ ] **Step 4: Commit**

```bash
git add testdata/baselines.json calendar/accuracy_test.go
git commit -m "test: add baseline accuracy tests for calendar data"
```

---

### Task 4: Data Store (Load + Query)

**Files:**
- Create: `calendar/store.go`
- Create: `calendar/store_test.go`

- [ ] **Step 1: Write failing test for loading and querying**

Create `calendar/store_test.go`:

```go
package calendar

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestStoreLoadAndQuery(t *testing.T) {
	dir := t.TempDir()
	// Write a small test data file
	df := DataFile{
		Version: "test1",
		Records: []CalendarRecord{
			{
				Date: "2025-01-01", LunarYear: 2024, LunarMonth: 12,
				LunarDay: 2, YearGanzhi: "甲辰", MonthGanzhi: "丙子",
				DayGanzhi: "壬午",
			},
			{
				Date: "2025-04-04", LunarYear: 2025, LunarMonth: 3,
				LunarDay: 7, YearGanzhi: "乙巳", MonthGanzhi: "庚辰",
				DayGanzhi: "甲申", SolarTerm: "清明",
			},
		},
	}
	data, _ := json.MarshalIndent(df, "", "  ")
	dataFile := filepath.Join(dir, "calendar.json")
	dataWithNewline := append(data, '\n')
	os.WriteFile(dataFile, dataWithNewline, 0o644)

	// Write checksum
	hash := sha256.Sum256(dataWithNewline)
	checksumLine := fmt.Sprintf("%x  calendar.json\n", hash)
	os.WriteFile(filepath.Join(dir, "calendar.json.sha256"), []byte(checksumLine), 0o644)

	store, err := LoadStore(dir)
	if err != nil {
		t.Fatalf("LoadStore: %v", err)
	}

	// Query existing date
	rec, ok := store.Query("2025-01-01")
	if !ok {
		t.Fatal("expected to find 2025-01-01")
	}
	if rec.LunarMonth != 12 {
		t.Errorf("lunar_month want 12 got %d", rec.LunarMonth)
	}

	// Query solar term date
	rec, ok = store.Query("2025-04-04")
	if !ok {
		t.Fatal("expected to find 2025-04-04")
	}
	if rec.SolarTerm != "清明" {
		t.Errorf("solar_term want 清明 got %q", rec.SolarTerm)
	}

	// Query missing date
	_, ok = store.Query("2020-01-01")
	if ok {
		t.Error("expected not to find 2020-01-01")
	}

	if store.Version() != "test1" {
		t.Errorf("version want test1 got %s", store.Version())
	}
}

func TestStoreBadChecksum(t *testing.T) {
	dir := t.TempDir()
	df := DataFile{Version: "bad", Records: []CalendarRecord{}}
	data, _ := json.Marshal(df)
	os.WriteFile(filepath.Join(dir, "calendar.json"), data, 0o644)
	os.WriteFile(filepath.Join(dir, "calendar.json.sha256"), []byte("0000  calendar.json\n"), 0o644)

	_, err := LoadStore(dir)
	if err == nil {
		t.Fatal("expected checksum error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./calendar/ -run TestStore -v`
Expected: FAIL — `LoadStore` and `Store` type not defined.

- [ ] **Step 3: Implement `calendar/store.go`**

```go
package calendar

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Store struct {
	version string
	records map[string]CalendarRecord
}

func LoadStore(dataDir string) (*Store, error) {
	dataPath := filepath.Join(dataDir, "calendar.json")
	checksumPath := dataPath + ".sha256"

	if err := verifyChecksum(dataPath, checksumPath); err != nil {
		return nil, fmt.Errorf("checksum verification failed: %w", err)
	}

	data, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, fmt.Errorf("read data file: %w", err)
	}

	var df DataFile
	if err := json.Unmarshal(data, &df); err != nil {
		return nil, fmt.Errorf("parse data file: %w", err)
	}

	records := make(map[string]CalendarRecord, len(df.Records))
	for _, r := range df.Records {
		records[r.Date] = r
	}

	return &Store{version: df.Version, records: records}, nil
}

func (s *Store) Query(date string) (CalendarRecord, bool) {
	r, ok := s.records[date]
	return r, ok
}

func (s *Store) Version() string {
	return s.version
}

func verifyChecksum(dataPath, checksumPath string) error {
	data, err := os.ReadFile(dataPath)
	if err != nil {
		return err
	}
	expected := sha256.Sum256(data)
	actual := fmt.Sprintf("%x", expected)

	stored, err := os.ReadFile(checksumPath)
	if err != nil {
		return err
	}

	// stored format: "<hex>  calendar.json\n"
	if len(stored) < 64 {
		return fmt.Errorf("invalid checksum file")
	}
	if string(stored[:64]) != actual {
		return fmt.Errorf("checksum mismatch")
	}
	return nil
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./calendar/ -run TestStore -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add calendar/store.go calendar/store_test.go
git commit -m "feat: data store with checksum-verified loading and query"
```

---

### Task 5: API Handler

**Files:**
- Create: `api/handler.go`
- Create: `api/handler_test.go`

- [ ] **Step 1: Write failing test for handler**

Create `api/handler_test.go`:

```go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"jievow-calendar/calendar"
)

func newTestStore(t *testing.T) *calendar.Store {
	t.Helper()
	// Use the real data file for handler tests
	store, err := calendar.LoadStore("../data")
	if err != nil {
		t.Fatalf("load store: %v", err)
	}
	return store
}

func TestHandlerBasicQuery(t *testing.T) {
	h := NewHandler(newTestStore(t))

	// Test solar term day
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/date/2026-04-20?fields=basic,lunar,solar_term", nil)
	r.SetPathValue("date", "2026-04-20")
	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status want 200 got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["data_version"] == nil {
		t.Error("missing data_version")
	}
	if resp["date"] != "2026-04-20" {
		t.Errorf("date want 2026-04-20 got %v", resp["date"])
	}
	if resp["weekday"] != "星期一" {
		t.Errorf("weekday want 星期一 got %v", resp["weekday"])
	}
	lunar, ok := resp["lunar"].(map[string]any)
	if !ok {
		t.Fatal("lunar should be an object")
	}
	if lunar["year_ganzhi"] == nil {
		t.Error("missing lunar.year_ganzhi")
	}
	st, ok := resp["solar_term"].(map[string]any)
	if !ok || st["name"] != "谷雨" {
		t.Errorf("solar_term.name want 谷雨 got %v", resp["solar_term"])
	}
}

func TestHandlerNonSolarTermDay(t *testing.T) {
	h := NewHandler(newTestStore(t))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/date/2026-04-21", nil)
	r.SetPathValue("date", "2026-04-21")
	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status want 200 got %d", w.Code)
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["solar_term"] != nil {
		t.Errorf("solar_term should be null, got %v", resp["solar_term"])
	}
}

func TestHandlerFieldSelection(t *testing.T) {
	h := NewHandler(newTestStore(t))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/date/2026-04-20?fields=basic", nil)
	r.SetPathValue("date", "2026-04-20")
	h.ServeHTTP(w, r)

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["lunar"] != nil {
		t.Error("lunar should be omitted when not requested")
	}
	if resp["solar_term"] != nil {
		t.Error("solar_term should be omitted when not requested")
	}
	if resp["date"] == nil {
		t.Error("date should be present in basic field group")
	}
}

func TestHandlerInvalidDate(t *testing.T) {
	h := NewHandler(newTestStore(t))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/date/not-a-date", nil)
	r.SetPathValue("date", "not-a-date")
	h.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status want 400 got %d", w.Code)
	}
}

func TestHandlerDateOutOfRange(t *testing.T) {
	h := NewHandler(newTestStore(t))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/date/2020-01-01", nil)
	r.SetPathValue("date", "2020-01-01")
	h.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status want 404 got %d", w.Code)
	}
}

func TestHandlerInvalidFields(t *testing.T) {
	h := NewHandler(newTestStore(t))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/date/2026-04-20?fields=basic,invalid", nil)
	r.SetPathValue("date", "2026-04-20")
	h.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status want 400 got %d", w.Code)
	}
}
```

> Note: `r.SetPathValue` is available in Go 1.22+. If using an older version, adjust the test to extract the date from the path manually.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./api/ -v`
Expected: FAIL — `NewHandler` not defined.

- [ ] **Step 3: Implement `api/handler.go`**

```go
package api

import (
	"encoding/json"
	"net/http"
	"time"

	"jievow-calendar/calendar"
)

type Handler struct {
	store *calendar.Store
}

func NewHandler(store *calendar.Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dateStr := r.PathValue("date")
	if dateStr == "" {
		writeError(w, http.StatusBadRequest, "invalid_date", "日期格式应为 YYYY-MM-DD")
		return
	}

	if _, err := time.Parse("2006-01-02", dateStr); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_date", "日期格式应为 YYYY-MM-DD")
		return
	}

	fields := calendar.ParseFields(r.URL.Query().Get("fields"))
	for _, f := range fields {
		if !calendar.ValidFields[f] {
			writeError(w, http.StatusBadRequest, "invalid_fields",
				"合法值: basic, lunar, solar_term, supplement")
			return
		}
	}

	rec, ok := h.store.Query(dateStr)
	if !ok {
		writeError(w, http.StatusNotFound, "date_out_of_range",
			"支持的日期范围: 2025-01-01 至 2027-12-31")
		return
	}

	fieldSet := make(map[string]bool, len(fields))
	for _, f := range fields {
		fieldSet[f] = true
	}

	resp := buildResponse(rec, fieldSet, h.store.Version())
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(resp)
}

func buildResponse(rec calendar.CalendarRecord, fields map[string]bool, version string) map[string]any {
	resp := map[string]any{
		"data_version": version,
	}

	if fields[calendar.FieldBasic] {
		resp["date"] = rec.Date
		resp["weekday"] = calendar.WeekdayChinese(rec.Date)
	}

	if fields[calendar.FieldLunar] {
		resp["lunar"] = map[string]any{
			"year":          rec.LunarYear,
			"month":         rec.LunarMonth,
			"day":           rec.LunarDay,
			"is_leap_month": rec.IsLeapMonth,
			"year_ganzhi":   rec.YearGanzhi,
			"month_ganzhi":  rec.MonthGanzhi,
			"day_ganzhi":    rec.DayGanzhi,
		}
	}

	if fields[calendar.FieldSolarTerm] {
		if rec.SolarTerm != "" {
			resp["solar_term"] = map[string]any{"name": rec.SolarTerm}
		} else {
			resp["solar_term"] = nil
		}
	}

	if fields[calendar.FieldSupplement] {
		resp["supplement"] = nil
	}

	return resp
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   code,
		"message": message,
	})
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./api/ -v`
Expected: All PASS.

- [ ] **Step 5: Commit**

```bash
git add api/handler.go api/handler_test.go
git commit -m "feat: API handler with field selection and error handling"
```

---

### Task 6: CORS Middleware & Server Entry Point

**Files:**
- Create: `api/cors.go`
- Create: `cmd/server/main.go`

- [ ] **Step 1: Create `api/cors.go`**

```go
package api

import "net/http"

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
```

- [ ] **Step 2: Create `cmd/server/main.go`**

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"jievow-calendar/api"
	"jievow-calendar/calendar"
)

func main() {
	store, err := calendar.LoadStore("data")
	if err != nil {
		log.Fatalf("failed to load data: %v", err)
	}
	log.Printf("Loaded %d records (version %s)", len(store.Query("")), store.Version())

	mux := http.NewServeMux()
	mux.Handle("GET /api/v1/date/{date}", api.NewHandler(store))

	handler := api.CORS(mux)

	addr := ":8080"
	fmt.Printf("Calendar API listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
```

> Note: `len(store.Query(""))` won't work as-is since Query returns a record and bool. A `Len()` method may be needed. The engineer should add one to `store.go` or remove this log line. Simplest fix: add a `func (s *Store) Len() int { return len(s.records) }` method.

- [ ] **Step 3: Add `Len` method to `calendar/store.go`**

Add to the end of `calendar/store.go`:

```go
func (s *Store) Len() int {
	return len(s.records)
}
```

Update `cmd/server/main.go` log line to use `store.Len()`:

```go
log.Printf("Loaded %d records (version %s)", store.Len(), store.Version())
```

- [ ] **Step 4: Build and run**

```bash
cd /data/dev/jievow-calendar
go build -o bin/server ./cmd/server/
./bin/server
```

In another terminal:

```bash
curl http://localhost:8080/api/v1/date/2026-04-20
```

Expected: JSON response with `data_version`, `date`, `weekday`, `lunar`, `solar_term` fields.

```bash
curl http://localhost:8080/api/v1/date/not-a-date
```

Expected: `{"error":"invalid_date","message":"日期格式应为 YYYY-MM-DD"}`

- [ ] **Step 5: Commit**

```bash
git add api/cors.go cmd/server/main.go calendar/store.go
git commit -m "feat: server entry point with CORS middleware"
```

---

### Task 7: Integration Tests

**Files:**
- Create: `api/integration_test.go`

- [ ] **Step 1: Write integration test**

Create `api/integration_test.go`:

```go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"jievow-calendar/calendar"
)

func setupTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	store, err := calendar.LoadStore("../data")
	if err != nil {
		t.Fatalf("load store: %v", err)
	}
	mux := http.NewServeMux()
	mux.Handle("GET /api/v1/date/{date}", NewHandler(store))
	return httptest.NewServer(CORS(mux))
}

func TestIntegrationAllFields(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/date/2026-04-20?fields=basic,lunar,solar_term")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status want 200 got %d", resp.StatusCode)
	}

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)

	assertField(t, body, "data_version", "2025a")
	assertField(t, body, "date", "2026-04-20")
	assertField(t, body, "weekday", "星期一")

	lunar := body["lunar"].(map[string]any)
	if lunar["year_ganzhi"] == nil {
		t.Error("missing year_ganzhi")
	}

	st := body["solar_term"].(map[string]any)
	if st["name"] != "谷雨" {
		t.Errorf("solar_term want 谷雨 got %v", st["name"])
	}
}

func TestIntegrationNoSolarTerm(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/date/2026-04-21")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)

	if body["solar_term"] != nil {
		t.Errorf("solar_term should be null, got %v", body["solar_term"])
	}
}

func TestIntegrationOnlyBasic(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/date/2026-04-20?fields=basic")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)

	if _, ok := body["lunar"]; ok {
		t.Error("lunar should be absent")
	}
	if body["date"] == nil {
		t.Error("date should be present")
	}
}

func TestIntegrationCORS(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	req, _ := http.NewRequest("OPTIONS", srv.URL+"/api/v1/date/2026-04-20", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("OPTIONS status want 204 got %d", resp.StatusCode)
	}
	origin := resp.Header.Get("Access-Control-Allow-Origin")
	if origin != "*" {
		t.Errorf("CORS origin want * got %q", origin)
	}
}

func TestIntegrationErrorCases(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	tests := []struct {
		name       string
		url        string
		wantStatus int
		wantError  string
	}{
		{
			"invalid date format",
			srv.URL + "/api/v1/date/abc",
			http.StatusBadRequest,
			"invalid_date",
		},
		{
			"date out of range",
			srv.URL + "/api/v1/date/2020-01-01",
			http.StatusNotFound,
			"date_out_of_range",
		},
		{
			"invalid field",
			srv.URL + "/api/v1/date/2026-04-20?fields=nonexistent",
			http.StatusBadRequest,
			"invalid_fields",
		},
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

func assertField(t *testing.T, body map[string]any, key, expected string) {
	t.Helper()
	if body[key] != expected {
		t.Errorf("%s want %q got %v", key, expected, body[key])
	}
}
```

- [ ] **Step 2: Run integration tests**

```bash
cd /data/dev/jievow-calendar
go test ./api/ -v -run TestIntegration
```

Expected: All PASS.

- [ ] **Step 3: Run all tests**

```bash
go test ./... -v
```

Expected: All PASS across `calendar/` and `api/` packages.

- [ ] **Step 4: Commit**

```bash
git add api/integration_test.go
git commit -m "test: add integration tests for API endpoints"
```

---

## Verification Checklist

After all tasks are complete:

- [ ] `go test ./...` passes all tests
- [ ] `go build ./...` compiles without errors
- [ ] `curl http://localhost:8080/api/v1/date/2026-04-20` returns correct JSON
- [ ] `curl http://localhost:8080/api/v1/date/abc` returns 400 error
- [ ] `curl http://localhost:8080/api/v1/date/2020-01-01` returns 400 error (out of range)
- [ ] CORS headers present in response
