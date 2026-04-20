package calendar

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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