package calendar

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestStoreLoadAndQuery(t *testing.T) {
	dir := t.TempDir()
	df := DataFile{
		Version: "test1",
		Records: []CalendarRecord{
			{
				Date: "2025-01-01", LunarYear: 2024, LunarMonth: 12,
				LunarDay: 2, YearGanzhi: "甲辰", MonthGanzhi: "丙子",
				DayGanzhi: "壬午",
			},
			{
				Date: "2025-04-04", LunarYear: 2025, LunarMonth: 3,
				LunarDay: 7, YearGanzhi: "乙巳", MonthGanzhi: "庚辰",
				DayGanzhi: "甲申", SolarTerm: "清明",
			},
		},
	}
	data, _ := json.MarshalIndent(df, "", "  ")
	dataFile := filepath.Join(dir, "calendar.json")
	dataWithNewline := append(data, '\n')
	os.WriteFile(dataFile, dataWithNewline, 0o644)

	hash := sha256.Sum256(dataWithNewline)
	checksumLine := fmt.Sprintf("%x  calendar.json\n", hash)
	os.WriteFile(filepath.Join(dir, "calendar.json.sha256"), []byte(checksumLine), 0o644)

	store, err := LoadStore(dir)
	if err != nil {
		t.Fatalf("LoadStore: %v", err)
	}

	rec, ok := store.Query("2025-01-01")
	if !ok {
		t.Fatal("expected to find 2025-01-01")
	}
	if rec.LunarMonth != 12 {
		t.Errorf("lunar_month want 12 got %d", rec.LunarMonth)
	}

	rec, ok = store.Query("2025-04-04")
	if !ok {
		t.Fatal("expected to find 2025-04-04")
	}
	if rec.SolarTerm != "清明" {
		t.Errorf("solar_term want 清明 got %q", rec.SolarTerm)
	}

	_, ok = store.Query("2020-01-01")
	if ok {
		t.Error("expected not to find 2020-01-01")
	}

	if store.Version() != "test1" {
		t.Errorf("version want test1 got %s", store.Version())
	}
}

func TestStoreBadChecksum(t *testing.T) {
	dir := t.TempDir()
	df := DataFile{Version: "bad", Records: []CalendarRecord{}}
	data, _ := json.Marshal(df)
	os.WriteFile(filepath.Join(dir, "calendar.json"), data, 0o644)
	os.WriteFile(filepath.Join(dir, "calendar.json.sha256"), []byte("0000  calendar.json\n"), 0o644)

	_, err := LoadStore(dir)
	if err == nil {
		t.Fatal("expected checksum error")
	}
}