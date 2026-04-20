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