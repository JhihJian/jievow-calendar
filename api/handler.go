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
