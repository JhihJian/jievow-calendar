package api

import (
	"encoding/json"
	"net/http"

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
