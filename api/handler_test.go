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
