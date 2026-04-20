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

type DataFile struct {
	Version string           `json:"version"`
	Records []CalendarRecord `json:"records"`
}

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