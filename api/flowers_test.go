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
