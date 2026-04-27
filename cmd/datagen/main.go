package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"jievow-calendar/calendar"

	lunarcal "github.com/6tail/lunar-go/calendar"
)

func main() {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2027, 12, 31, 0, 0, 0, 0, time.UTC)

	var records []calendar.CalendarRecord
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		records = append(records, generateRecord(d))
	}

	computeActiveTerms(records)

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
	solar := lunarcal.NewSolar(year, int(month), day, 0, 0, 0)
	lunar := solar.GetLunar()

	lunarMonth := lunar.GetMonth()
	isLeap := lunarMonth < 0
	lunarDay := lunar.GetDay()
	yearGanzhi := lunar.GetYearInGanZhi()

	term := ""
	if jq := lunar.GetJieQi(); jq != "" {
		term = jq
	}

	return calendar.CalendarRecord{
		Date:        d.Format("2006-01-02"),
		LunarYear:   lunar.GetYear(),
		LunarMonth:  lunarMonth,
		LunarDay:    lunarDay,
		IsLeapMonth: isLeap,
		YearGanzhi:  yearGanzhi,
		MonthGanzhi: lunar.GetMonthInGanZhi(),
		DayGanzhi:   lunar.GetDayInGanZhi(),
		SolarTerm:   term,

		MonthDisplay: calendar.LunarMonthDisplay(lunarMonth, isLeap),
		DayDisplay:   calendar.LunarDayDisplay(lunarDay),
		Display:      calendar.LunarDisplay(lunarMonth, lunarDay, isLeap),
		YearDisplay:  calendar.LunarYearDisplay(yearGanzhi, lunarMonth, lunarDay, isLeap),
	}
}

func computeActiveTerms(records []calendar.CalendarRecord) {
	type termInfo struct {
		date time.Time
		name string
	}
	var terms []termInfo
	for _, r := range records {
		if r.SolarTerm != "" {
			t, _ := time.Parse("2006-01-02", r.Date)
			terms = append(terms, termInfo{t, r.SolarTerm})
		}
	}

	for i := range records {
		recDate, _ := time.Parse("2006-01-02", records[i].Date)

		idx := sort.Search(len(terms), func(j int) bool {
			return terms[j].date.After(recDate)
		})

		if idx > 0 {
			term := terms[idx-1]
			records[i].ActiveTerm = term.name
			records[i].IsTermDay = records[i].SolarTerm != ""
			records[i].TermStartDate = term.date.Format("2006-01-02")
			records[i].DayInTerm = int(recDate.Sub(term.date).Hours()/24) + 1
		}
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
