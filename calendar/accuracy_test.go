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