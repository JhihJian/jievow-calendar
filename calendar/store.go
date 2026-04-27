package calendar

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Store struct {
	version string
	records map[string]CalendarRecord
}

func LoadStore(dataDir string) (*Store, error) {
	dataPath := filepath.Join(dataDir, "calendar.json")
	checksumPath := dataPath + ".sha256"

	if err := verifyChecksum(dataPath, checksumPath); err != nil {
		return nil, fmt.Errorf("checksum verification failed: %w", err)
	}

	data, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, fmt.Errorf("read data file: %w", err)
	}

	var df DataFile
	if err := json.Unmarshal(data, &df); err != nil {
		return nil, fmt.Errorf("parse data file: %w", err)
	}

	records := make(map[string]CalendarRecord, len(df.Records))
	for _, r := range df.Records {
		records[r.Date] = r
	}

	return &Store{version: df.Version, records: records}, nil
}

func (s *Store) Query(date string) (CalendarRecord, bool) {
	r, ok := s.records[date]
	return r, ok
}

func (s *Store) Version() string {
	return s.version
}

func (s *Store) Len() int {
	return len(s.records)
}

func NewStore(version string, records []CalendarRecord) *Store {
	m := make(map[string]CalendarRecord, len(records))
	for _, r := range records {
		m[r.Date] = r
	}
	return &Store{version: version, records: m}
}

func (s *Store) QueryRange(from, to string) ([]CalendarRecord, error) {
	fromDate, err := time.Parse("2006-01-02", from)
	if err != nil {
		return nil, fmt.Errorf("invalid from date: %w", err)
	}
	toDate, err := time.Parse("2006-01-02", to)
	if err != nil {
		return nil, fmt.Errorf("invalid to date: %w", err)
	}
	if fromDate.After(toDate) {
		return nil, fmt.Errorf("from must be <= to")
	}
	days := int(toDate.Sub(fromDate).Hours()/24) + 1
	if days > 366 {
		return nil, fmt.Errorf("range exceeds 366 days")
	}
	var result []CalendarRecord
	for d := fromDate; !d.After(toDate); d = d.AddDate(0, 0, 1) {
		if rec, ok := s.records[d.Format("2006-01-02")]; ok {
			result = append(result, rec)
		}
	}
	return result, nil
}

func verifyChecksum(dataPath, checksumPath string) error {
	data, err := os.ReadFile(dataPath)
	if err != nil {
		return err
	}
	expected := sha256.Sum256(data)
	actual := fmt.Sprintf("%x", expected)

	stored, err := os.ReadFile(checksumPath)
	if err != nil {
		return err
	}

	if len(stored) < 64 {
		return fmt.Errorf("invalid checksum file")
	}
	if string(stored[:64]) != actual {
		return fmt.Errorf("checksum mismatch")
	}
	return nil
}