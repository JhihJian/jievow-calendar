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
	store, err := calendar.LoadStore("../data")
	if err != nil {
		t.Fatalf("load store: %v", err)
	}
	return store
}

func TestHandlerBasicQuery(t *testing.T) {
	h := NewHandler(newTestStore(t))
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
	if lunar["display"] == nil {
		t.Error("missing lunar.display")
	}
	st, ok := resp["solar_term"].(map[string]any)
	if !ok || st["name"] != "谷雨" {
		t.Errorf("solar_term.name want 谷雨 got %v", resp["solar_term"])
	}
	if st["is_term_day"] != true {
		t.Errorf("solar_term.is_term_day want true got %v", st["is_term_day"])
	}
}

func TestHandlerNonSolarTermDay(t *testing.T) {
	h := NewHandler(newTestStore(t))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/date/2026-04-21", nil)
	r.SetPathValue("date", "2026-04-21")
	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status want 200 got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)

	st, ok := resp["solar_term"].(map[string]any)
	if !ok {
		t.Fatal("solar_term should be an object with active term info")
	}
	if st["name"] != "谷雨" {
		t.Errorf("solar_term.name want 谷雨 got %v", st["name"])
	}
	if st["is_term_day"] != false {
		t.Errorf("solar_term.is_term_day want false got %v", st["is_term_day"])
	}
	if st["day_in_term"] == nil {
		t.Error("missing solar_term.day_in_term")
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

func TestHandleRange(t *testing.T) {
	records := []calendar.CalendarRecord{
		{Date: "2026-04-01", LunarYear: 2026, LunarMonth: 3, LunarDay: 4, IsLeapMonth: false, YearGanzhi: "丙午", MonthGanzhi: "壬辰", DayGanzhi: "甲子", ActiveTerm: "春分", IsTermDay: false, TermStartDate: "2026-03-20", DayInTerm: 13, MonthDisplay: "三月", DayDisplay: "初四", Display: "三月初四", YearDisplay: "丙午年三月初四"},
		{Date: "2026-04-05", LunarYear: 2026, LunarMonth: 3, LunarDay: 8, IsLeapMonth: false, YearGanzhi: "丙午", MonthGanzhi: "壬辰", DayGanzhi: "戊寅", SolarTerm: "清明", ActiveTerm: "清明", IsTermDay: true, TermStartDate: "2026-04-05", DayInTerm: 1, MonthDisplay: "三月", DayDisplay: "初八", Display: "三月初八", YearDisplay: "丙午年三月初八"},
	}
	store := calendar.NewStore("testv", records)
	h := NewHandler(store)

	req := httptest.NewRequest("GET", "/api/v1/range?from=2026-04-01&to=2026-04-05&fields=basic,lunar,solar_term", nil)
	w := httptest.NewRecorder()
	h.HandleRange(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["from"] != "2026-04-01" {
		t.Errorf("from = %v, want 2026-04-01", resp["from"])
	}
	dates := resp["dates"].([]any)
	if len(dates) != 2 {
		t.Fatalf("len(dates) = %d, want 2", len(dates))
	}
}

func TestHandleRangeValidation(t *testing.T) {
	store := calendar.NewStore("testv", nil)
	h := NewHandler(store)

	tests := []struct {
		url  string
		code int
	}{
		{"/api/v1/range?from=2026-04-01", http.StatusBadRequest},
		{"/api/v1/range?from=2026-04-05&to=2026-04-01", http.StatusBadRequest},
		{"/api/v1/range?from=bad&to=2026-04-05", http.StatusBadRequest},
	}
	for _, tt := range tests {
		req := httptest.NewRequest("GET", tt.url, nil)
		w := httptest.NewRecorder()
		h.HandleRange(w, req)
		if w.Code != tt.code {
			t.Errorf("url=%s status=%d, want %d", tt.url, w.Code, tt.code)
		}
	}
}

func TestHandleSolarTerms(t *testing.T) {
	records := []calendar.CalendarRecord{
		{Date: "2026-01-05", SolarTerm: "小寒", MonthDisplay: "冬月"},
		{Date: "2026-01-20", SolarTerm: "大寒", MonthDisplay: "腊月"},
		{Date: "2026-04-05", SolarTerm: "清明", MonthDisplay: "三月"},
		{Date: "2026-06-15", SolarTerm: "", MonthDisplay: "五月"},
	}
	store := calendar.NewStore("testv", records)
	h := NewHandler(store)

	req := httptest.NewRequest("GET", "/api/v1/solar-terms?year=2026", nil)
	w := httptest.NewRecorder()
	h.HandleSolarTerms(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["year"] != float64(2026) {
		t.Errorf("year = %v, want 2026", resp["year"])
	}
	terms := resp["terms"].([]any)
	if len(terms) != 3 {
		t.Fatalf("len(terms) = %d, want 3", len(terms))
	}
	first := terms[0].(map[string]any)
	if first["name"] != "小寒" {
		t.Errorf("first term name = %v, want 小寒", first["name"])
	}
}

func TestHandleSolarTermsValidation(t *testing.T) {
	store := calendar.NewStore("testv", nil)
	h := NewHandler(store)

	tests := []struct {
		url  string
		code int
	}{
		{"/api/v1/solar-terms", http.StatusBadRequest},
		{"/api/v1/solar-terms?year=abc", http.StatusBadRequest},
		{"/api/v1/solar-terms?year=1999", http.StatusNotFound},
	}
	for _, tt := range tests {
		req := httptest.NewRequest("GET", tt.url, nil)
		w := httptest.NewRecorder()
		h.HandleSolarTerms(w, req)
		if w.Code != tt.code {
			t.Errorf("url=%s status=%d, want %d", tt.url, w.Code, tt.code)
		}
	}
}
