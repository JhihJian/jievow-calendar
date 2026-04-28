// calendar/flowers.go
package calendar

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type FlowerRecord struct {
	Province   string   `json:"province"`
	SolarTerm  string   `json:"solar_term"`
	Flowers    []string `json:"flowers"`
}

var ValidSolarTerms = []string{
	"小寒", "大寒", "立春", "雨水", "惊蛰", "春分",
	"清明", "谷雨", "立夏", "小满", "芒种", "夏至",
	"小暑", "大暑", "立秋", "处暑", "白露", "秋分",
	"寒露", "霜降", "立冬", "小雪", "大雪", "冬至",
}

var validSolarTermSet map[string]bool

func init() {
	validSolarTermSet = make(map[string]bool, len(ValidSolarTerms))
	for _, t := range ValidSolarTerms {
		validSolarTermSet[t] = true
	}
}

func IsValidSolarTerm(term string) bool {
	return validSolarTermSet[term]
}

type FlowerStore struct {
	byProvinceTerm map[[2]string][]string
	provinces      []string
}

func LoadFlowerStore(dataDir string) (*FlowerStore, error) {
	dataPath := filepath.Join(dataDir, "flowers.json")
	checksumPath := dataPath + ".sha256"

	if err := verifyChecksum(dataPath, checksumPath); err != nil {
		return nil, fmt.Errorf("checksum verification failed: %w", err)
	}

	data, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, fmt.Errorf("read data file: %w", err)
	}

	var records []FlowerRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("parse data file: %w", err)
	}

	return NewFlowerStore(records), nil
}

func NewFlowerStore(records []FlowerRecord) *FlowerStore {
	byProvinceTerm := make(map[[2]string][]string, len(records))
	provinceSet := make(map[string]bool)
	for _, r := range records {
		key := [2]string{r.Province, r.SolarTerm}
		byProvinceTerm[key] = r.Flowers
		provinceSet[r.Province] = true
	}

	provinces := make([]string, 0, len(provinceSet))
	for p := range provinceSet {
		provinces = append(provinces, p)
	}
	sort.Strings(provinces)

	return &FlowerStore{byProvinceTerm: byProvinceTerm, provinces: provinces}
}

func (s *FlowerStore) GetFlowers(province, solarTerm string) ([]string, bool) {
	flowers, ok := s.byProvinceTerm[[2]string{province, solarTerm}]
	return flowers, ok
}

func (s *FlowerStore) ListProvinces() []string {
	return s.provinces
}
