package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"jievow-calendar/calendar"
)

type Handler struct {
	store   *calendar.Store
	flowers *calendar.FlowerStore
}

func NewHandler(store *calendar.Store, flowers *calendar.FlowerStore) *Handler {
	return &Handler{store: store, flowers: flowers}
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
			"month_display": rec.MonthDisplay,
			"day_display":   rec.DayDisplay,
			"display":       rec.Display,
			"year_display":  rec.YearDisplay,
		}
	}

	if fields[calendar.FieldSolarTerm] {
		if rec.ActiveTerm != "" {
			resp["solar_term"] = map[string]any{
				"name":        rec.ActiveTerm,
				"is_term_day": rec.IsTermDay,
				"start_date":  rec.TermStartDate,
				"day_in_term": rec.DayInTerm,
			}
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

func (h *Handler) HandleRange(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	if from == "" || to == "" {
		writeError(w, http.StatusBadRequest, "invalid_params", "from and to 参数必填")
		return
	}

	if _, err := time.Parse("2006-01-02", from); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_date", "日期格式应为 YYYY-MM-DD")
		return
	}
	if _, err := time.Parse("2006-01-02", to); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_date", "日期格式应为 YYYY-MM-DD")
		return
	}

	records, err := h.store.QueryRange(from, to)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_range", err.Error())
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

	fieldSet := make(map[string]bool, len(fields))
	for _, f := range fields {
		fieldSet[f] = true
	}

	dates := make([]map[string]any, 0, len(records))
	for _, rec := range records {
		entry := buildResponse(rec, fieldSet, "")
		delete(entry, "data_version")
		dates = append(dates, entry)
	}

	resp := map[string]any{
		"data_version": h.store.Version(),
		"from":         from,
		"to":           to,
		"dates":        dates,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) HandleSolarTerms(w http.ResponseWriter, r *http.Request) {
	yearStr := r.URL.Query().Get("year")
	if yearStr == "" {
		writeError(w, http.StatusBadRequest, "invalid_params", "year 参数必填")
		return
	}
	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 1 {
		writeError(w, http.StatusBadRequest, "invalid_year", "year 应为有效年份")
		return
	}

	records := h.store.SolarTermsByYear(year)
	if len(records) == 0 {
		writeError(w, http.StatusNotFound, "year_out_of_range",
			"支持的年份范围: 2025 至 2027")
		return
	}

	terms := make([]map[string]any, 0, len(records))
	for _, rec := range records {
		terms = append(terms, map[string]any{
			"name":          rec.SolarTerm,
			"date":          rec.Date,
			"month_display": rec.MonthDisplay,
		})
	}

	resp := map[string]any{
		"data_version": h.store.Version(),
		"year":         year,
		"terms":        terms,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(resp)
}
